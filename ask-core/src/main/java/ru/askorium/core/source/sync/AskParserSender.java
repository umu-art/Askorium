package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;
import ru.askorium.api.model.RenderInput;
import ru.askorium.api.model.ScrapeResponse;
import ru.askorium.core.source.config.ScrapperTasksProperties;
import ru.askorium.core.source.domain.SyncTaskStatus;
import ru.askorium.core.source.jpa.SyncTaskJpa;
import ru.askorium.core.source.jpa.SyncTaskUrlJpa;

import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class AskParserSender {

    private final ScrapperTasksProperties scrapperTasksProperties;
    private final RabbitTemplate rabbitTemplate;
    private final SyncTaskJpa syncTaskJpa;
    private final SyncTaskUrlJpa syncTaskUrlJpa;
    private final PageProcessor pageProcessor;
    private final SyncDispatcher syncDispatcher;
    private final IndexSyncService indexSyncService;

    @RabbitListener(queues = "#{scrapperResponseQueue.name}")
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
            var result = pageProcessor.processScrapedUrl(
                    urlEntity.getId(), taskId, task.getSourceId(), task.isForceSync(), response.getPage());

            if (result.savedPage() != null) {
                indexSyncService.syncIndexes(new ArrayList<>(List.of(result.savedPage())));
            }

            for (var newUrl : result.newUrls()) {
                sendRenderRequest(newUrl.getId(), newUrl.getUrl());
            }

            taskComplete = result.taskComplete();
        } else {
            var errorMsg = response.getError() != null ? response.getError().getMessage() : "unknown error";
            log.warn("Scrape failed for URL {}: {}", urlEntity.getUrl(), errorMsg);
            taskComplete = pageProcessor.markUrlFailed(urlEntity.getId(), taskId);
        }

        if (taskComplete) {
            syncDispatcher.completeTask(taskId);
        }
    }

    void sendRenderRequest(UUID urlEntityId, String url) {
        var request = new RenderInput()
                .url(URI.create(url))
                .taskId(urlEntityId);
        rabbitTemplate.convertAndSend(scrapperTasksProperties.getScrapperRequestQueueName(), request);
    }
}