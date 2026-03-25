package ru.askorium.core.source.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "askorium.scrapper")
public class ScrapperTasksProperties {

    private String scrapperRequestQueueName;

    private String scrapperResponseQueueName;

    private int minBlockLength;

    private int maxBlockLength;

}
