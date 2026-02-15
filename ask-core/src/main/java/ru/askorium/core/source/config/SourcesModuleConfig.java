package ru.askorium.core.source.config;

import jakarta.persistence.EntityManagerFactory;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.config.BeanDefinition;
import org.springframework.beans.factory.support.BeanDefinitionReaderUtils;
import org.springframework.beans.factory.support.BeanDefinitionRegistry;
import org.springframework.beans.factory.support.BeanNameGenerator;
import org.springframework.boot.orm.jpa.EntityManagerFactoryBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;
import org.springframework.lang.NonNull;
import org.springframework.orm.jpa.JpaTransactionManager;
import org.springframework.orm.jpa.LocalContainerEntityManagerFactoryBean;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.annotation.EnableTransactionManagement;
import ru.askorium.core.common.HikariDataSourceFactoryBean;

@Configuration
@EnableTransactionManagement
@EnableJpaRepositories(
        basePackages = "ru.askorium.core.source.jpa",
        entityManagerFactoryRef = "sourcesEntityManagerFactory",
        transactionManagerRef = "sourcesTransactionManager",
        repositoryImplementationPostfix = "sourcesModuleJpa",
        nameGenerator = SourcesModuleConfig.CustomNameGenerator.class
)
@ComponentScan(basePackages = "ru.askorium.core.source")
@RequiredArgsConstructor
public class SourcesModuleConfig {

    @Bean("sourcesEntityManagerFactory")
    public LocalContainerEntityManagerFactoryBean entityManagerFactory(
            HikariDataSourceFactoryBean factory,
            EntityManagerFactoryBuilder builder) {

        return builder
                .dataSource(factory.createByName("sources"))
                .packages("ru.askorium.core.sources.domain")
                .persistenceUnit("sources")
                .build();
    }

    @Bean("sourcesTransactionManager")
    public PlatformTransactionManager transactionManager(
            @Qualifier("sourcesEntityManagerFactory") EntityManagerFactory entityManagerFactory) {
        return new JpaTransactionManager(entityManagerFactory);
    }

    public static class CustomNameGenerator implements BeanNameGenerator {

        @NonNull
        @Override
        public String generateBeanName(@NonNull BeanDefinition definition, @NonNull BeanDefinitionRegistry registry) {
            return BeanDefinitionReaderUtils.generateBeanName(definition, registry) + "-sourcesModule";
        }
    }
}
