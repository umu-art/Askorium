package ru.askorium.core.common.config;

import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.config.RetryInterceptorBuilder;
import org.springframework.amqp.rabbit.config.SimpleRabbitListenerContainerFactory;
import org.springframework.context.annotation.Configuration;

@Slf4j
@Configuration
@RequiredArgsConstructor
public class RabbitRetryBackOffConfiguration {

    private final SimpleRabbitListenerContainerFactory rabbitListenerContainerFactory;

    @PostConstruct
    public void configureBackOff() {
        rabbitListenerContainerFactory.setAdviceChain(RetryInterceptorBuilder
                .stateless()
                .maxAttempts(3)
                .backOffOptions(500, 2.0, 10000)
                .build());

        rabbitListenerContainerFactory.setDefaultRequeueRejected(false);
        log.info("Configured RabbitMQ listener container factory to self-retry and not requeue rejected messages.");
    }

}