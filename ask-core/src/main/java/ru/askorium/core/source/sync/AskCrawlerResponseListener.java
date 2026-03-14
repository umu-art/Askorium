package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Service;
import ru.askorium.api.model.CrawlEvent;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.source.domain.SyncTaskEntity;
import ru.askorium.core.source.domain.SyncTaskStatus;
import ru.askorium.core.source.jpa.SyncTaskJpa;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;

import static java.util.Objects.nonNull;

@Slf4j
@Service
@RequiredArgsConstructor
public class AskCrawlerResponseListener {

    private final SyncTaskJpa syncTaskJpa;
    private final PageProcessor pageProcessor;
    private final IndexSyncService indexSyncService;
    private final SyncDispatcher syncDispatcher;
    private final ExecutorService indexingExecutor;

    @RabbitListener(queues = "${askorium.scrapper.scrapper-response-queue-name}")
    public void handle(CrawlEvent event) {
        var taskId = event.getTaskId();
        var task = syncTaskJpa.findById(taskId)
                .orElse(null);

        if (task == null || task.getStatus() != SyncTaskStatus.RUNNING) {
            log.warn("Task {} not found or not running, skipping response", taskId);
            return;
        }

        log.debug("Processing scrape response for task {}", taskId);

        if (nonNull(event.getError())) {
            log.warn("Scrape failed: {}", event.getError());
            syncDispatcher.markFailed(taskId, event.getError().getMessage());
            return;
        }

        for (var page : event.getPages()) {
            processPage(page, task, taskId);
        }

        var stats = event.getStats();
        if (nonNull(stats)) {
            task.setPagesScraped(Objects.requireNonNullElse(stats.getPagesScraped(), 0));
            task.setPagesFailed(Objects.requireNonNullElse(stats.getPagesFailed(), 0));
        }

        syncDispatcher.markCompleted(taskId);
    }

    private void processPage(ScrappedPage page, SyncTaskEntity task, UUID taskId) {
        try {
            var savedPage = pageProcessor.processPage(page, task.getSourceId(), task.isForceSync());

            if (nonNull(savedPage)) {
                CompletableFuture.runAsync(
                        () -> indexSyncService.syncIndexes(new ArrayList<>(List.of(savedPage))),
                        indexingExecutor
                );
            }
        } catch (Exception e) {
            log.error("Failed to process page {} for task {}: {}", page.getUrl(), taskId, e.getMessage(), e);
        }
    }
}
