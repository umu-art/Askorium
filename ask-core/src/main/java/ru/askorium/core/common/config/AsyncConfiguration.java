package ru.askorium.core.common.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.task.AsyncTaskExecutor;
import org.springframework.scheduling.annotation.EnableAsync;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Configuration
@EnableAsync
public class AsyncConfiguration {

    @Bean
    public AsyncTaskExecutor asyncTaskExecutor() {
        var executor = new ThreadPoolTaskExecutor();
        executor.setThreadNamePrefix("v-async-");
        executor.setVirtualThreads(true);
        executor.setCorePoolSize(10);
        executor.setMaxPoolSize(10);
        executor.setWaitForTasksToCompleteOnShutdown(true);
        executor.setAwaitTerminationSeconds(30);
        executor.initialize();
        return executor;
    }

    @Bean
    public ExecutorService indexingExecutor() {
        return Executors.newSingleThreadExecutor(
                Thread.ofVirtual().name("indexing-", 0).factory()
        );
    }

}
