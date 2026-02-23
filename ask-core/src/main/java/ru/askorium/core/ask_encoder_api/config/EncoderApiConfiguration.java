package ru.askorium.core.ask_encoder_api.config;

import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import ru.askorium.api.client.ApiClient;
import ru.askorium.api.client.EncoderApi;
import ru.askorium.core.common.RestTemplateFactory;

@Configuration
@RequiredArgsConstructor
public class EncoderApiConfiguration {

    private final RestTemplateFactory restTemplateFactory;

    @Value("${askorium.api.encoder-url}")
    private String encoderUrl;

    @Bean
    public EncoderApi encoderApi() {
        var apiClient = new ApiClient(restTemplateFactory.createRestTemplate());
        apiClient.setBasePath(encoderUrl);
        return new EncoderApi(apiClient);
    }

}
