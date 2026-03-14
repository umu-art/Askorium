package ru.askorium.core.source.config;


import lombok.extern.slf4j.Slf4j;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.Scheduled;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

@Slf4j
@Configuration
public class EncoderSchedulerConfiguration {

    private final LinkedBlockingQueue<Runnable> taskQueue = new LinkedBlockingQueue<>();

    @Bean
    public ExecutorService indexingExecutor() {
        return new ThreadPoolExecutor(1, 1,
                0L, TimeUnit.MILLISECONDS,
                taskQueue,
                Thread.ofVirtual().name("indexing-", 0).factory());
    }

    @Scheduled(fixedDelay = 60_000)
    public void printTaskCount() {
        log.debug("Tasks in queue: {}", taskQueue.size());
    }
}
