package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Service;
import ru.askorium.api.model.ScrapeResponse;
import ru.askorium.core.source.domain.SyncTaskStatus;
import ru.askorium.core.source.domain.SyncTaskUrlStatus;
import ru.askorium.core.source.jpa.SyncTaskJpa;
import ru.askorium.core.source.jpa.SyncTaskUrlJpa;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;

@Slf4j
@Service
@RequiredArgsConstructor
public class AskParserSender {

    private final SyncTaskJpa syncTaskJpa;
    private final SyncTaskUrlJpa syncTaskUrlJpa;
    private final PageProcessor pageProcessor;
    private final IndexSyncService indexSyncService;
    private final SyncDispatcher syncDispatcher;
    private final RenderRequestSender renderRequestSender;
    private final ExecutorService indexingExecutor;

    @RabbitListener(queues = "${askorium.scrapper.scrapper-response-queue-name}")
    public void handleScrappedPage(ScrapeResponse response) {
        var urlEntity = syncTaskUrlJpa.findById(response.getTaskId()).orElse(null);
        if (urlEntity == null) {
            log.warn("Unknown URL entity for response taskId {}, skipping", response.getTaskId());
            return;
        }

        var taskId = urlEntity.getTaskId();
        var task = syncTaskJpa.findById(taskId).orElse(null);
        if (task == null || task.getStatus() != SyncTaskStatus.RUNNING) {
            log.warn("Task {} not found or not running, skipping response for URL {}", taskId, urlEntity.getUrl());
            return;
        }

        log.debug("Processing scrape response for URL {} (task {})", urlEntity.getUrl(), taskId);

        boolean taskComplete;

        if (Boolean.TRUE.equals(response.getSuccess()) && response.getPage() != null) {
            var result = processResponse(urlEntity.getId(), taskId, task.getSourceId(), task.isForceSync(), response);

            final var savedPage = result.savedPage();
            if (savedPage != null) {
                CompletableFuture.runAsync(
                        () -> indexSyncService.syncIndexes(new ArrayList<>(List.of(savedPage))),
                        indexingExecutor
                );
            }

            for (var url : result.newUrls()) {
                renderRequestSender.send(taskId, url);
            }

            taskComplete = syncTaskUrlJpa.countByTaskIdAndStatus(taskId, SyncTaskUrlStatus.PENDING) == 0;
        } else {
            var errorMsg = response.getError() != null ? response.getError().getMessage() : "unknown error";
            log.warn("Scrape failed for URL {}: {}", urlEntity.getUrl(), errorMsg);
            taskComplete = markFailed(urlEntity.getId(), taskId);
        }

        if (taskComplete) {
            syncDispatcher.completeTask(taskId);
        }
    }

    private PageProcessor.ScrapeResult processResponse(UUID urlEntityId, UUID taskId, UUID sourceId, boolean forceSync, ScrapeResponse response) {
        var urlEntity = syncTaskUrlJpa.findById(urlEntityId).orElseThrow();
        urlEntity.setStatus(SyncTaskUrlStatus.SCRAPED);
        syncTaskUrlJpa.save(urlEntity);
        syncTaskJpa.incrementPagesScraped(taskId);

        return pageProcessor.processPage(response.getPage(), sourceId, forceSync);
    }

    private boolean markFailed(UUID urlEntityId, UUID taskId) {
        var urlEntity = syncTaskUrlJpa.findById(urlEntityId).orElseThrow();
        urlEntity.setStatus(SyncTaskUrlStatus.FAILED);
        syncTaskUrlJpa.save(urlEntity);
        syncTaskJpa.incrementPagesFailed(taskId);

        return syncTaskUrlJpa.countByTaskIdAndStatus(taskId, SyncTaskUrlStatus.PENDING) == 0;
    }
}
