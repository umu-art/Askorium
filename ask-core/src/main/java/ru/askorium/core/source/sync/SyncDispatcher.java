package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.redisson.api.RedissonClient;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import ru.askorium.api.model.CrawlTaskRequest;
import ru.askorium.api.model.SourceSyncRequest;
import ru.askorium.core.exception.EntityNotFoundException;
import ru.askorium.core.source.config.ScrapperTasksProperties;
import ru.askorium.core.source.domain.SyncTaskEntity;
import ru.askorium.core.source.domain.SyncTaskStatus;
import ru.askorium.core.source.jpa.SourceJpa;
import ru.askorium.core.source.jpa.SyncTaskJpa;

import java.time.OffsetDateTime;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class SyncDispatcher {

    private final SourceJpa sourceJpa;
    private final SyncTaskJpa syncTaskJpa;
    private final RedissonClient redissonClient;
    private final RabbitTemplate rabbitTemplate;
    private final ScrapperTasksProperties scrapperTasksProperties;

    public void sync(SourceSyncRequest request) {
        var sourceId = request.getSourceId();
        var source = sourceJpa.findById(sourceId)
                .orElseThrow(() -> new IllegalArgumentException("Source not found: " + sourceId));

        var task = new SyncTaskEntity();
        task.setSourceId(sourceId);
        task.setStatus(SyncTaskStatus.RUNNING);
        task.setForceSync(Boolean.TRUE.equals(request.getForce()));
        task.setStartedAt(OffsetDateTime.now());
        task = syncTaskJpa.saveAndFlush(task);

        log.info("Starting sync task {} for source {}", task.getId(), sourceId);

        sendToCrawler(task.getId(), source.getSourceUrl());
    }

    private void sendToCrawler(UUID id, String sourceUrl) {
        var request = new CrawlTaskRequest()
                .taskId(id)
                .domain(sourceUrl);

        rabbitTemplate.convertAndSend(scrapperTasksProperties.getScrapperRequestQueueName(), request);
        log.debug("Sent crawl task {} to queue for source {}", id, sourceUrl);
    }

    @Transactional(transactionManager = "sourcesTransactionManager")
    public void markFailed(UUID taskId, String message) {
        var task = syncTaskJpa.findById(taskId)
                .orElseThrow(() -> new EntityNotFoundException("Task", taskId));

        if (task.getStatus() != SyncTaskStatus.RUNNING) {
            log.warn("Attempted to mark task {} as failed, but it is in status {}", taskId, task.getStatus());
            return;
        }

        task.setStatus(SyncTaskStatus.FAILED);
        task.setErrorMessage(message);
        task.setFinishedAt(OffsetDateTime.now());
        syncTaskJpa.save(task);
    }

    @Transactional(transactionManager = "sourcesTransactionManager")
    public void markCompleted(UUID taskId) {
        var task = syncTaskJpa.findById(taskId)
                .orElseThrow(() -> new EntityNotFoundException("Task", taskId));

        if (task.getStatus() != SyncTaskStatus.RUNNING) {
            log.warn("Attempted to mark task {} as completed, but it is in status {}", taskId, task.getStatus());
            return;
        }

        task.setStatus(SyncTaskStatus.COMPLETED);
        task.setFinishedAt(OffsetDateTime.now());
        syncTaskJpa.save(task);
    }
}
