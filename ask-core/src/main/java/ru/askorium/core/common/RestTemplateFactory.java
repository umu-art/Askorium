package ru.askorium.core.common;

import lombok.extern.slf4j.Slf4j;
import org.jetbrains.annotations.NotNull;
import org.springframework.http.HttpMethod;
import org.springframework.http.client.BufferingClientHttpRequestFactory;
import org.springframework.http.client.ClientHttpRequestInterceptor;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Service;
import org.springframework.util.StreamUtils;
import org.springframework.web.client.RequestCallback;
import org.springframework.web.client.ResponseExtractor;
import org.springframework.web.client.RestClientException;
import org.springframework.web.client.RestTemplate;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.URI;
import java.nio.charset.StandardCharsets;

@Slf4j
@Service
public class RestTemplateFactory {

    private static final int CONNECT_TIMEOUT_MS = 2_000;
    private static final int READ_TIMEOUT_MS = 60_000;

    public RestTemplate createRestTemplate() {
        var restTemplate = new RepeatableRestTemplate();

        var baseFactory = new SimpleClientHttpRequestFactory();
        baseFactory.setConnectTimeout(CONNECT_TIMEOUT_MS);
        baseFactory.setReadTimeout(READ_TIMEOUT_MS);

        restTemplate.setRequestFactory(new BufferingClientHttpRequestFactory(baseFactory));
        restTemplate.getInterceptors().add(loggingInterceptor());
        return restTemplate;
    }

    private ClientHttpRequestInterceptor loggingInterceptor() {
        return (request, body, execution) -> {
            byte[] requestBody = body.clone();

            ClientHttpResponse response;
            try {
                response = execution.execute(request, body);
            } catch (IOException ex) {
                log.error("Request failed. Method: {}, URI: {}, RequestBody: {}",
                        request.getMethod(), request.getURI(), new String(requestBody, StandardCharsets.UTF_8));
                throw ex;
            }

            ByteArrayOutputStream buffer = new ByteArrayOutputStream();
            StreamUtils.copy(response.getBody(), buffer);

            if (response.getStatusCode().is2xxSuccessful()) {
                log.info("Request successful. Status: {}, URI: {}, ResponseBody: {}",
                        response.getStatusCode(), request.getURI(), buffer.toString(StandardCharsets.UTF_8));

                return response;
            } else {
                log.warn("Request failed. Status: {}, URI: {}, RequestBody: {}, ResponseBody: {}",
                        response.getStatusCode(), request.getURI(),
                        new String(requestBody, StandardCharsets.UTF_8), buffer.toString(StandardCharsets.UTF_8));
            }

            return response;
        };
    }

    private static class RepeatableRestTemplate extends RestTemplate {
        @Override
        protected <T> T doExecute(@NotNull URI url, String uriTemplate, HttpMethod method, RequestCallback requestCallback, ResponseExtractor<T> responseExtractor) throws RestClientException {
            for (int i = 0; i < 3; i++) {
                try {
                    return super.doExecute(url, uriTemplate, method, requestCallback, responseExtractor);
                } catch (Exception e) {
                    log.error("Failed http request attempt {} / 3", i + 1, e);
                }
            }
            throw new RestClientException("Http request failed");
        }
    }
}
