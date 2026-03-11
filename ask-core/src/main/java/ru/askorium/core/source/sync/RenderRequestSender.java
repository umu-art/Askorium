package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.RenderInput;
import ru.askorium.core.source.config.ScrapperTasksProperties;
import ru.askorium.core.source.domain.SyncTaskUrlEntity;
import ru.askorium.core.source.domain.SyncTaskUrlStatus;
import ru.askorium.core.source.jpa.SyncTaskJpa;
import ru.askorium.core.source.jpa.SyncTaskUrlJpa;

import java.net.URI;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class RenderRequestSender {

    private final SyncTaskJpa syncTaskJpa;
    private final SyncTaskUrlJpa syncTaskUrlJpa;
    private final RabbitTemplate rabbitTemplate;
    private final ScrapperTasksProperties scrapperTasksProperties;

    @Transactional(transactionManager = "sourcesTransactionManager")
    public void send(UUID taskId, String url) {
        if (syncTaskUrlJpa.existsByTaskIdAndUrl(taskId, url)) return;

        var urlEntity = new SyncTaskUrlEntity();
        urlEntity.setTaskId(taskId);
        urlEntity.setUrl(url);
        urlEntity.setStatus(SyncTaskUrlStatus.PENDING);
        urlEntity = syncTaskUrlJpa.saveAndFlush(urlEntity);
        syncTaskJpa.incrementPagesDiscovered(taskId);

        var request = new RenderInput()
                .url(URI.create(url))
                .taskId(urlEntity.getId());
        rabbitTemplate.convertAndSend(scrapperTasksProperties.getScrapperRequestQueueName(), request);
        log.debug("Sent render request for URL {} (urlEntityId={})", url, urlEntity.getId());
    }
}
