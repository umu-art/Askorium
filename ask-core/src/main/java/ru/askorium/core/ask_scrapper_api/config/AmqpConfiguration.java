package ru.askorium.core.ask_scrapper_api.config;

import lombok.RequiredArgsConstructor;
import org.springframework.amqp.core.Queue;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
@RequiredArgsConstructor
public class AmqpConfiguration {

    private final AmqpProperties amqpProperties;

    @Bean
    public Queue scrapperRequestQueue() {
        return new Queue(amqpProperties.getScrapperRequestQueueName(), true);
    }

    @Bean
    public Queue scrapperResponseQueue() {
        return new Queue(amqpProperties.getScrapperResponseQueueName(), true);
    }

}
