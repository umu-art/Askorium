package ru.askorium.core.common.datasource;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import ru.askorium.core.common.HikariDataSourceFactoryBean;

@Component
@RequiredArgsConstructor
public class HikariDataSourceFactoryBeanImpl implements HikariDataSourceFactoryBean {

    private final MasterHikariConfig masterHikariConfig;

    public HikariDataSource createByName(String name) {
        HikariConfig config = new HikariConfig();
        masterHikariConfig.copyStateTo(config);
        config.setPoolName("hikari-pool-" + name);

        return new HikariDataSource(config);
    }
}
