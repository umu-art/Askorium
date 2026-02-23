package ru.askorium.core.source.config;

import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import ru.askorium.api.client.ApiClient;
import ru.askorium.api.client.SourceApi;
import ru.askorium.core.common.RestTemplateFactory;

@Configuration
@RequiredArgsConstructor
public class SelfSourceApiConfiguration {

    private final RestTemplateFactory restTemplateFactory;

    @Value("${askorium.api.self-url}")
    private String selfUrl;

    @Bean
    public SourceApi selfSourceApi() {
        var apiClient = new ApiClient(restTemplateFactory.createRestTemplate());
        apiClient.setBasePath(selfUrl);
        return new SourceApi(apiClient);
    }

}
