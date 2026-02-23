package ru.askorium.core.user.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "askorium.user")
public class UserProperties {

    private RateLimit rateLimit = new RateLimit();

    @Data
    public static class RateLimit {
        private int requestsPerMinute = 60;
    }
}
