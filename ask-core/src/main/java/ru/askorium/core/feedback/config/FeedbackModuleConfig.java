package ru.askorium.core.feedback.config;

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
        basePackages = "ru.askorium.core.feedback.jpa",
        entityManagerFactoryRef = "feedbackEntityManagerFactory",
        transactionManagerRef = "feedbackTransactionManager",
        repositoryImplementationPostfix = "feedbackModuleJpa",
        nameGenerator = FeedbackModuleConfig.CustomNameGenerator.class
)
@ComponentScan(basePackages = "ru.askorium.core.feedback")
@RequiredArgsConstructor
public class FeedbackModuleConfig {

    @Bean("feedbackEntityManagerFactory")
    public LocalContainerEntityManagerFactoryBean entityManagerFactory(
            HikariDataSourceFactoryBean factory,
            EntityManagerFactoryBuilder builder) {

        return builder
                .dataSource(factory.createByName("feedback"))
                .packages("ru.askorium.core.feedback.domain")
                .persistenceUnit("feedback")
                .build();
    }

    @Bean("feedbackTransactionManager")
    public PlatformTransactionManager transactionManager(
            @Qualifier("feedbackEntityManagerFactory") EntityManagerFactory entityManagerFactory) {
        return new JpaTransactionManager(entityManagerFactory);
    }

    public static class CustomNameGenerator implements BeanNameGenerator {

        @NonNull
        @Override
        public String generateBeanName(@NonNull BeanDefinition definition, @NonNull BeanDefinitionRegistry registry) {
            return BeanDefinitionReaderUtils.generateBeanName(definition, registry) + "-feedbackModule";
        }
    }
}
