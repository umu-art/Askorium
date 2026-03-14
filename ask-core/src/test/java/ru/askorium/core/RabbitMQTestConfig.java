package ru.askorium.core;

import org.springframework.amqp.core.Queue;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.context.annotation.Bean;
import ru.askorium.core.source.config.ScrapperTasksProperties;

@TestConfiguration
public class RabbitMQTestConfig {

    @Autowired
    ScrapperTasksProperties scrapperTasksProperties;

    @Bean
    public Queue crawlerOutputQueue() {
        return new Queue(scrapperTasksProperties.getScrapperResponseQueueName(), true);
    }
}
