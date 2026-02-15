package ru.askorium.core.common.datasource;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import ru.askorium.core.common.HikariDataSourceFactoryBean;

import javax.sql.DataSource;

@Configuration
class DataSourceConfig {

    @Primary
    @Bean("primaryDataSource")
    DataSource primaryDataSource(HikariDataSourceFactoryBean factory) {
        return factory.createByName("primary");
    }

}
