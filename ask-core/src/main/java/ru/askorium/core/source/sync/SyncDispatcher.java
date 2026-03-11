package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.RenderInput;
import ru.askorium.api.model.SourceSyncRequest;
import ru.askorium.core.source.config.ScrapperTasksProperties;
import ru.askorium.core.source.domain.SyncTaskEntity;
import ru.askorium.core.source.domain.SyncTaskStatus;
import ru.askorium.core.source.domain.SyncTaskUrlEntity;
import ru.askorium.core.source.domain.SyncTaskUrlStatus;
import ru.askorium.core.source.jpa.PageJpa;
import ru.askorium.core.source.jpa.SourceJpa;
import ru.askorium.core.source.jpa.SyncTaskJpa;
import ru.askorium.core.source.jpa.SyncTaskUrlJpa;

import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;

import java.net.URI;
import java.time.OffsetDateTime;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class SyncDispatcher {

    private final SourceJpa sourceJpa;
    private final PageJpa pageJpa;
    private final SyncTaskJpa syncTaskJpa;
    private final SyncTaskUrlJpa syncTaskUrlJpa;
    private final RabbitTemplate rabbitTemplate;
    private final ScrapperTasksProperties scrapperTasksProperties;

    private final ConcurrentHashMap<UUID, CompletableFuture<Void>> pendingTasks = new ConcurrentHashMap<>();

    public void sync(SourceSyncRequest request) {
        var sourceId = request.getSourceId();
        var source = sourceJpa.findById(sourceId)
                .orElseThrow(() -> new IllegalArgumentException("Source not found: " + sourceId));

        var task = new SyncTaskEntity();
        task.setSourceId(sourceId);
        task.setStatus(SyncTaskStatus.RUNNING);
        task.setForceSync(Boolean.TRUE.equals(request.getForce()));
        task.setPagesDiscovered(1);
        task.setStartedAt(OffsetDateTime.now());

        try {
            task = syncTaskJpa.saveAndFlush(task);
        } catch (DataIntegrityViolationException e) {
            log.warn("Another sync is already running for source {}", sourceId);
            return;
        }

        var urlEntity = new SyncTaskUrlEntity();
        urlEntity.setTaskId(task.getId());
        urlEntity.setUrl(source.getSourceUrl());
        urlEntity.setStatus(SyncTaskUrlStatus.PENDING);
        urlEntity = syncTaskUrlJpa.save(urlEntity);

        var future = new CompletableFuture<Void>();
        pendingTasks.put(task.getId(), future);

        log.info("Starting sync task {} for source {}", task.getId(), sourceId);
        sendRenderRequest(urlEntity.getId(), source.getSourceUrl());

        try {
            future.get(scrapperTasksProperties.getWaitTimeout().toMillis(), TimeUnit.MILLISECONDS);
        } catch (TimeoutException e) {
            log.error("Sync task {} timed out for source {}", task.getId(), sourceId);
            failTask(task.getId(), "Sync timed out");
        } catch (ExecutionException e) {
            log.error("Sync task {} failed for source {}", task.getId(), sourceId, e.getCause());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            failTask(task.getId(), "Sync interrupted");
        } finally {
            pendingTasks.remove(task.getId());
        }
    }

    void sendRenderRequest(UUID urlEntityId, String url) {
        var request = new RenderInput()
                .url(URI.create(url))
                .taskId(urlEntityId);
        rabbitTemplate.convertAndSend(scrapperTasksProperties.getScrapperRequestQueueName(), request);
    }

    @Transactional(transactionManager = "sourcesTransactionManager")
    public void completeTask(UUID taskId) {
        var task = syncTaskJpa.findById(taskId).orElse(null);
        if (task == null || task.getStatus() != SyncTaskStatus.RUNNING) return;

        var sourceId = task.getSourceId();
        var scrapedUrls = syncTaskUrlJpa.findByTaskIdAndStatus(taskId, SyncTaskUrlStatus.SCRAPED)
                .stream()
                .map(SyncTaskUrlEntity::getUrl)
                .collect(Collectors.toSet());

        var stalePages = pageJpa.findBySourceId(sourceId).stream()
                .filter(p -> !scrapedUrls.contains(p.getUrl()))
                .toList();

        if (!stalePages.isEmpty()) {
            log.info("Deleting {} stale pages for source {}", stalePages.size(), sourceId);
            pageJpa.deleteAll(stalePages);
        }

        var source = sourceJpa.findById(sourceId).orElseThrow();
        source.getSyncPolicy().setLastSyncedAt(OffsetDateTime.now());
        sourceJpa.save(source);

        task.setStatus(SyncTaskStatus.COMPLETED);
        task.setFinishedAt(OffsetDateTime.now());
        syncTaskJpa.save(task);

        log.info("Sync task {} completed. Scraped {}, deleted {} stale pages.",
                taskId, scrapedUrls.size(), stalePages.size());

        TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
            @Override
            public void afterCommit() {
                notifyTaskDone(taskId, null);
            }
        });
    }

    void failTask(UUID taskId, String errorMessage) {
        var task = syncTaskJpa.findById(taskId).orElse(null);
        if (task != null && task.getStatus() == SyncTaskStatus.RUNNING) {
            task.setStatus(SyncTaskStatus.FAILED);
            task.setErrorMessage(errorMessage);
            task.setFinishedAt(OffsetDateTime.now());
            syncTaskJpa.save(task);
        }
        notifyTaskDone(taskId, new RuntimeException(errorMessage));
    }

    private void notifyTaskDone(UUID taskId, Exception error) {
        var future = pendingTasks.remove(taskId);
        if (future == null) return;
        if (error != null) {
            future.completeExceptionally(error);
        } else {
            future.complete(null);
        }
    }
}