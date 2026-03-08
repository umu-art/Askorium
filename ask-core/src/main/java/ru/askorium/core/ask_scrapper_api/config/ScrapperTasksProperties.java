package ru.askorium.core.ask_scrapper_api.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.time.Duration;

@Data
@Component
@ConfigurationProperties(prefix = "askorium.scrapper")
public class ScrapperTasksProperties {

    private String scrapperRequestQueueName;

    private String scrapperResponseQueueName;

    private String redisKeyPrefix;

    private Duration waitTimeout;

}
