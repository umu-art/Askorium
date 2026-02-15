package ru.askorium.core.common.datasource;

import com.zaxxer.hikari.HikariConfig;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "spring.datasource.hikari")
class MasterHikariConfig extends HikariConfig {
}
    