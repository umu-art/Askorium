package ru.askorium.core.common;

import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpMethod;
import org.springframework.http.client.BufferingClientHttpRequestFactory;
import org.springframework.http.client.ClientHttpRequestInterceptor;
import org.springframework.http.client.ClientHttpResponse;
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

    public RestTemplate createRestTemplate() {
        var restTemplate = new RepeatableRestTemplate();
        restTemplate.setRequestFactory(new BufferingClientHttpRequestFactory(restTemplate.getRequestFactory()));
        restTemplate.getInterceptors().add(loggingInterceptor());
        return restTemplate;
    }

    private static class RepeatableRestTemplate extends RestTemplate {
        @Override
        protected <T> T doExecute(URI url, String uriTemplate, HttpMethod method, RequestCallback requestCallback, ResponseExtractor<T> responseExtractor) throws RestClientException {
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
}
