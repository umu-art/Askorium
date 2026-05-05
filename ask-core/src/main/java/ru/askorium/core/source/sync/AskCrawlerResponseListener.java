package ru.askorium.core.source.sync;

import io.temporal.api.enums.v1.WorkflowIdReusePolicy;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;
import ru.askorium.api.model.CrawlEvent;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.source.workflow.IndexingWorkflow;
import ru.askorium.core.source.domain.SyncTaskEntity;
import ru.askorium.core.source.domain.SyncTaskStatus;
import ru.askorium.core.source.jpa.SyncTaskJpa;

import java.util.Collections;
import java.util.Objects;

import static java.util.Objects.nonNull;

@Slf4j
@Service
@RequiredArgsConstructor
public class AskCrawlerResponseListener {

    private final SyncTaskJpa syncTaskJpa;
    private final PageProcessor pageProcessor;
    private final SyncDispatcher syncDispatcher;
    private final WorkflowClient workflowClient;

    @RabbitListener(queues = "${askorium.scrapper.scrapper-response-queue-name}")
    public void handle(CrawlEvent event) {
        if (CollectionUtils.isEmpty(event.getPages())) {
            event.setPages(Collections.emptyList());
        }

        var taskId = event.getTaskId();
        var task = syncTaskJpa.findById(taskId)
                .orElse(null);

        if (task == null || task.getStatus() != SyncTaskStatus.RUNNING) {
            log.warn("Task {} not found or not running, skipping response", taskId);
            return;
        }

        log.debug("Processing scrape event for task {}", taskId);
        log.debug("type {}, pages {}, error {}", event.getType(), event.getPages().size(), event.getError());

        var stats = event.getStats();
        if (nonNull(stats)) {
            task.setPagesScraped(Objects.requireNonNullElse(stats.getPagesScraped(), 0));
            task.setPagesFailed(Objects.requireNonNullElse(stats.getPagesFailed(), 0));
        }

        syncTaskJpa.save(task);

        switch (event.getType()) {
            case TASK_COMPLETED -> {
                processPageBatch(task, event);
                syncDispatcher.markCompleted(taskId);
            }
            case TASK_FAILED -> syncDispatcher.markFailed(taskId, event.getError().getMessage());
            case PAGE_BATCH -> processPageBatch(task, event);
            default -> log.warn("Unknown event type: {}", event.getType());
        }
    }

    private void processPageBatch(SyncTaskEntity task, CrawlEvent event) {
        log.debug("processPageBatch {}", event.getPages().size());

        for (int i = 0; i < event.getPages().size(); i++) {
            var page = event.getPages().get(i);
            log.trace("processPage {}: {} of {}", page.getUrl(), i + 1, event.getPages().size());
            processPage(page, task);
        }
    }

    private void processPage(ScrappedPage page, SyncTaskEntity task) {
        try {
            var savedPage = pageProcessor.processPage(page, task.getSourceId(), task.isForceSync());

            if (nonNull(savedPage)) {
                log.trace("Triggering IndexingWorkflow for page {}", savedPage.getUrl());

                var stub = workflowClient.newWorkflowStub(
                        IndexingWorkflow.class,
                        WorkflowOptions.newBuilder()
                                .setTaskQueue("askorium-search")
                                .setWorkflowId("index-page-" + savedPage.getId())
                                .setWorkflowIdReusePolicy(WorkflowIdReusePolicy.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY)
                                .build()
                );

                WorkflowClient.start(stub::index, savedPage.getId());
            }
        } catch (Exception e) {
            log.error("Failed to process page {} for task {}: {}", page.getUrl(), task.getId(), e.getMessage(), e);
        }
    }
}
