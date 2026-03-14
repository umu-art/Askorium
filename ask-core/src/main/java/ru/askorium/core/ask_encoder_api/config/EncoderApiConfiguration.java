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

    private static final int CONNECT_TIMEOUT_MS = 20_000;
    private static final int READ_TIMEOUT_MS = 600_000;

    @Value("${askorium.api.encoder-url}")
    private String encoderUrl;

    @Bean
    public EncoderApi encoderApi() {
        var apiClient = new ApiClient(restTemplateFactory.createRestTemplate(CONNECT_TIMEOUT_MS, READ_TIMEOUT_MS));
        apiClient.setBasePath(encoderUrl);
        return new EncoderApi(apiClient);
    }

}
