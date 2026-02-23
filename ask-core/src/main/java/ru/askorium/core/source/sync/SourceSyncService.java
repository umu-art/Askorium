package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import org.redisson.api.RedissonClient;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.Document;
import ru.askorium.api.model.Link;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.api.model.SourceSyncRequest;
import ru.askorium.core.ask_scrapper_api.AskScrapperService;
import ru.askorium.core.common.ObjectCompareUtils;
import ru.askorium.core.common.UrlUtils;
import ru.askorium.core.exception.BadUrlException;
import ru.askorium.core.source.domain.PageEntity;
import ru.askorium.core.source.jpa.PageJpa;
import ru.askorium.core.source.jpa.SourceJpa;
import ru.askorium.core.source.mapper.PageMapper;
import ru.askorium.core.text_processing.TextProcessingService;

import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.function.BiPredicate;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class SourceSyncService {

    private final PageJpa pageJpa;
    private final SourceJpa sourceJpa;
    private final RedissonClient redissonClient;
    private final AskScrapperService askScrapperService;
    private final IndexSyncService indexSyncService;
    private final TextProcessingService textProcessingService;
    private final PageMapper pageMapper;

    @SneakyThrows
    @Transactional(transactionManager = "sourcesTransactionManager")
    public void sync(SourceSyncRequest request) {
        var sourceId = request.getSourceId();

        var lock = redissonClient.getLock("source-sync:" + sourceId);
        if (!lock.tryLock(5, 30, TimeUnit.MINUTES)) {
            log.warn("Could not acquire lock for source {}, another sync in progress?", sourceId);
            return;
        }

        try {
            doSync(sourceId, Boolean.TRUE.equals(request.getForce()));
        } finally {
            lock.unlock();
        }
    }

    private void doSync(UUID sourceId, boolean force) {
        var source = sourceJpa.findById(sourceId)
                .orElseThrow(() -> new IllegalArgumentException("Source not found: " + sourceId));

        log.debug("Starting sync for source {} (force={})", source.getSourceUrl(), force);

        var scrappedPages = askScrapperService.scrapSource(source.getSourceUrl());

        normalizeTexts(scrappedPages);
        normalizeUrls(scrappedPages);

        log.debug("Scrapped {} pages for source {}", scrappedPages.size(), source.getSourceUrl());

        var existingPages = pageJpa.findBySourceId(sourceId);

        var updatedPages = new ArrayList<PageEntity>();

        var existingByUrl = existingPages.stream()
                .collect(Collectors.toMap(PageEntity::getUrl, Function.identity()));

        var scrappedUrls = scrappedPages.stream()
                .map(ScrappedPage::getUrl)
                .collect(Collectors.toSet());

        for (var scrappedPage : scrappedPages) {
            var url = scrappedPage.getUrl();
            var contentHash = String.valueOf(scrappedPage.hashCode());

            var existing = existingByUrl.get(url);
            if (existing != null && contentHash.equals(existing.getContentHash()) && !force) {
                log.debug("Page unchanged, skipping: {}", url);
                continue;
            }

            var page = Objects.requireNonNullElse(existing, new PageEntity());
            syncPage(page, scrappedPage, sourceId, contentHash);
            updatedPages.add(pageJpa.save(page));
        }

        var stalePages = existingPages.stream()
                .filter(page -> !scrappedUrls.contains(page.getUrl()))
                .toList();

        if (!stalePages.isEmpty()) {
            log.info("Deleting {} stale pages for source {}", stalePages.size(), sourceId);
            pageJpa.deleteAll(stalePages);
        }

        source.getSyncPolicy().setLastSyncedAt(OffsetDateTime.now());
        sourceJpa.save(source);

        indexSyncService.syncIndexes(updatedPages);

        log.info("Sync completed for source {}. Scraped {} pages, removed {} stale.",
                sourceId, scrappedPages.size(), stalePages.size());
    }

    private void normalizeUrls(List<ScrappedPage> pages) {
        pages.removeIf(page -> {
            try {
                page.setUrl(UrlUtils.normalizeUrl(page.getUrl()));
            } catch (BadUrlException e) {
                log.warn("Bad page URL, skipping: {}", page.getUrl());
                return true;
            }

            if (page.getLinks() != null) {
                page.getLinks().removeIf(link -> {
                    try {
                        link.setHref(UrlUtils.normalizeUrl(link.getHref()));
                        return false;
                    } catch (BadUrlException e) {
                        log.warn("Bad link URL, skipping: {}", link.getHref());
                        return true;
                    }
                });
            }

            if (page.getDocuments() != null) {
                page.getDocuments().removeIf(doc -> {
                    try {
                        doc.setUrl(UrlUtils.normalizeUrl(doc.getUrl()));
                        return false;
                    } catch (BadUrlException e) {
                        log.warn("Bad doc URL, skipping: {}", doc.getUrl());
                        return true;
                    }
                });
            }

            return false;
        });
    }

    private void normalizeTexts(List<ScrappedPage> scrappedPages) {
        scrappedPages.forEach(page -> {
            var blocks = page.getBlocks();
            blocks.forEach(block -> block.setText(textProcessingService.normalizeText(block.getText())));

            var links = Objects.requireNonNullElse(page.getLinks(), new ArrayList<Link>());
            links.forEach(link -> link.setAnchorText(textProcessingService.normalizeText(link.getAnchorText())));

            var documents = Objects.requireNonNullElse(page.getDocuments(), new ArrayList<Document>());
            documents.forEach(doc -> doc.setExtractedText(textProcessingService.normalizeText(doc.getExtractedText())));
        });
    }

    private void syncPage(PageEntity page, ScrappedPage scrappedPage, UUID sourceId, String contentHash) {
        pageMapper.updatePage(page, scrappedPage);

        page.setSourceId(sourceId);
        page.setContentHash(contentHash);

        var blockCandidates = scrappedPage.getBlocks().stream()
                .map(pageMapper::toBlockEntity)
                .toList();

        var linkCandidates = Objects.requireNonNullElse(scrappedPage.getLinks(), new ArrayList<Link>())
                .stream()
                .map(pageMapper::toLinkEntity)
                .toList();

        var docCandidates = Objects.requireNonNullElse(scrappedPage.getDocuments(), new ArrayList<Document>())
                .stream()
                .map(pageMapper::toDocumentEntity)
                .toList();

        syncCollection(page.getBlocks(), blockCandidates,
                contentEquals(), b -> b.setPage(page));

        syncCollection(page.getLinks(), linkCandidates,
                contentEquals(), l -> l.setPage(page));

        syncCollection(page.getDocuments(), docCandidates,
                contentEquals(), d -> d.setPage(page));
    }

    private <T> void syncCollection(
            List<T> existing,
            List<T> candidates,
            BiPredicate<T, T> contentEquals,
            Consumer<T> setPage
    ) {
        var remainingCandidates = new ArrayList<>(candidates);
        var toRemove = new ArrayList<T>();

        for (var entity : existing) {
            var matched = false;
            var it = remainingCandidates.iterator();
            while (it.hasNext()) {
                if (contentEquals.test(entity, it.next())) {
                    it.remove();
                    matched = true;
                    break;
                }
            }
            if (!matched) {
                toRemove.add(entity);
            }
        }

        existing.removeAll(toRemove);

        for (T candidate : remainingCandidates) {
            setPage.accept(candidate);
            existing.add(candidate);
        }
    }

    private <T> BiPredicate<T, T> contentEquals() {
        return (a, b) -> ObjectCompareUtils.equalsObjectsExcludeFields(a, b, "id", "created", "updated");
    }

}
