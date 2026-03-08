package ru.askorium.core.ask_scrapper_api.service;

import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.redisson.api.RedissonClient;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;
import ru.askorium.api.model.CrawlTaskRequest;
import ru.askorium.api.model.ScrapeResponse;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.ask_scrapper_api.AskScrapperService;
import ru.askorium.core.ask_scrapper_api.config.ScrapperTasksProperties;

import java.util.List;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class AskScrapperServiceImpl implements AskScrapperService {

    private static final UUID currentClientId = UUID.randomUUID();

    private final ScrapperTasksProperties scrapperTasksProperties;
    private final RabbitTemplate rabbitTemplate;
    private final RedissonClient redissonClient;

    @PostConstruct
    public void init() {
        log.info("AskScrapperService initialized with clientId: {}", currentClientId);
    }

    @Override
    public List<ScrappedPage> scrapSource(String sourceUrl) {
        var prefix = scrapperTasksProperties.getRedisKeyPrefix();

        var request = new CrawlTaskRequest()
                .taskId(UUID.randomUUID())
                .domain(sourceUrl)
                .putMetadataItem("currentClientId", currentClientId);

        var semaphore = redissonClient.getSemaphore(prefix + ":" + request.getTaskId() + ":semaphore");

        try {
            semaphore.trySetPermits(0);

            rabbitTemplate.convertAndSend(scrapperTasksProperties.getScrapperRequestQueueName(), request);

            boolean acquired;
            try {
                acquired = semaphore.tryAcquire(scrapperTasksProperties.getWaitTimeout());
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                acquired = false;
            }

            if (!acquired) {
                log.warn("Timeout while waiting for scrape response for taskId: {}", request.getTaskId());
                return List.of();
            }

            var responseKey = prefix + ":" + request.getTaskId() + ":response";
            var response = redissonClient.<ScrapeResponse>getBucket(responseKey).get();

            redissonClient.getBucket(responseKey).delete();

            if (Boolean.FALSE.equals(response.getSuccess())) {
                log.warn("Scrape task failed for taskId: {}. Error: {}", request.getTaskId(), response.getError());
                return List.of();
            }

            return List.of(response.getPage());
        } finally {
            semaphore.delete();
        }
    }

    @RabbitListener(queues = "#{scrapperResponseQueue.name}")
    public void handleScrappedPage(ScrapeResponse scrapeResponse) {
        var taskId = scrapeResponse.getTaskId();
        var prefix = scrapperTasksProperties.getRedisKeyPrefix();

        redissonClient.<ScrapeResponse>getBucket(prefix + ":" + taskId + ":response")
                .set(scrapeResponse, scrapperTasksProperties.getWaitTimeout());

        redissonClient.getSemaphore(prefix + ":" + taskId + ":semaphore").release();
    }
}
