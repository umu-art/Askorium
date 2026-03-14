package ru.askorium.core.source.sync;

import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.Document;
import ru.askorium.api.model.Link;
import ru.askorium.api.model.LinkType;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.common.ObjectCompareUtils;
import ru.askorium.core.common.UrlUtils;
import ru.askorium.core.exception.BadUrlException;
import ru.askorium.core.source.domain.PageEntity;
import ru.askorium.core.source.jpa.PageJpa;
import ru.askorium.core.source.mapper.PageMapper;
import ru.askorium.core.text_processing.TextProcessingService;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.UUID;
import java.util.function.BiPredicate;
import java.util.function.Consumer;

import static java.util.Objects.isNull;

@Slf4j
@Service
@RequiredArgsConstructor
public class PageProcessor {

    private final PageJpa pageJpa;
    private final TextProcessingService textProcessingService;
    private final PageMapper pageMapper;

    @PersistenceContext(unitName = "sources")
    private EntityManager entityManager;

    @Transactional(transactionManager = "sourcesTransactionManager")
    public PageEntity processPage(ScrappedPage page, UUID sourceId, boolean forceSync) {
        normalizePageTexts(page);

        if (!normalizePageUrl(page)) {
            return null;
        }

        return savePageDelta(page, sourceId, forceSync);
    }

    private void normalizePageTexts(ScrappedPage page) {
        if (isNull(page.getBlocks())) {
            page.setBlocks(new ArrayList<>());
        }

        page.getBlocks().forEach(block ->
                block.setText(textProcessingService.normalizeText(block.getText())));

        var links = Objects.requireNonNullElse(page.getLinks(), new ArrayList<Link>());
        links.forEach(link ->
                link.setAnchorText(textProcessingService.normalizeText(link.getAnchorText())));

        var documents = Objects.requireNonNullElse(page.getDocuments(), new ArrayList<Document>());
        documents.forEach(doc ->
                doc.setExtractedText(textProcessingService.normalizeText(doc.getExtractedText())));
    }

    private boolean normalizePageUrl(ScrappedPage page) {
        try {
            page.setUrl(UrlUtils.normalizeUrl(page.getUrl()));
        } catch (BadUrlException e) {
            log.warn("Bad page URL, skipping: {}", page.getUrl());
            return false;
        }

        if (page.getLinks() != null) {
            page.getLinks().removeIf(link -> {
                try {
                    link.setHref(UrlUtils.normalizeUrl(link.getHref()));
                    return false;
                } catch (BadUrlException e) {
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
                    return true;
                }
            });
        }

        return true;
    }

    private PageEntity savePageDelta(ScrappedPage scrappedPage, UUID sourceId, boolean force) {
        var url = scrappedPage.getUrl();
        var contentHash = String.valueOf(scrappedPage.hashCode());

        var existing = pageJpa.findByUrl(url).orElse(null);
        if (existing != null && contentHash.equals(existing.getContentHash()) && !force) {
            log.debug("Page unchanged, skipping: {}", url);
            return null;
        }

        var page = Objects.requireNonNullElse(existing, new PageEntity());
        syncPage(page, scrappedPage, sourceId, contentHash);
        return pageJpa.save(page);
    }

    private void syncPage(PageEntity page, ScrappedPage scrappedPage, UUID sourceId, String contentHash) {
        pageMapper.updatePage(page, scrappedPage);
        page.setSourceId(sourceId);
        page.setContentHash(contentHash);

        syncCollection(page.getBlocks(),
                scrappedPage.getBlocks().stream().map(pageMapper::toBlockEntity).toList(),
                contentEquals(), b -> b.setPage(page));

        syncCollection(page.getLinks(),
                Objects.requireNonNullElse(scrappedPage.getLinks(), new ArrayList<Link>())
                        .stream().map(pageMapper::toLinkEntity).toList(),
                contentEquals(), l -> l.setPage(page));

        syncCollection(page.getDocuments(),
                Objects.requireNonNullElse(scrappedPage.getDocuments(), new ArrayList<Document>())
                        .stream().map(pageMapper::toDocumentEntity).toList(),
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

        if (!toRemove.isEmpty() && !remainingCandidates.isEmpty()) {
            entityManager.flush();
        }

        for (T candidate : remainingCandidates) {
            setPage.accept(candidate);
            existing.add(candidate);
        }
    }

    private <T> BiPredicate<T, T> contentEquals() {
        return (a, b) -> ObjectCompareUtils.equalsObjectsExcludeFields(a, b, "id", "created", "updated", "page");
    }

    public record ScrapeResult(PageEntity savedPage, List<String> newUrls) {}
}
