package ru.askorium.core.ask_scrapper_api.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "askorium.amqp")
public class AmqpProperties {

    private String scrapperRequestQueueName;

    private String scrapperResponseQueueName;

}
