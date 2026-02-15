package ru.askorium.core.common;

import com.zaxxer.hikari.HikariDataSource;

public interface HikariDataSourceFactoryBean {
    HikariDataSource createByName(String name);
}
