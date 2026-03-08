package ru.askorium.core.ask_scrapper_api.config;

import lombok.RequiredArgsConstructor;
import org.springframework.amqp.core.Queue;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
@RequiredArgsConstructor
public class AmqpConfiguration {

    private final ScrapperTasksProperties scrapperTasksProperties;

    @Bean
    public Queue scrapperRequestQueue() {
        return new Queue(scrapperTasksProperties.getScrapperRequestQueueName(), true);
    }

    @Bean
    public Queue scrapperResponseQueue() {
        return new Queue(scrapperTasksProperties.getScrapperResponseQueueName(), true);
    }

}
