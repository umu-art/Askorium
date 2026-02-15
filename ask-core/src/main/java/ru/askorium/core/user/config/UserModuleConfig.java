package ru.askorium.core.user.config;

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
        basePackages = "ru.askorium.core.user.jpa",
        entityManagerFactoryRef = "userEntityManagerFactory",
        transactionManagerRef = "userTransactionManager",
        repositoryImplementationPostfix = "userModuleJpa",
        nameGenerator = UserModuleConfig.CustomNameGenerator.class
)
@ComponentScan(basePackages = "ru.askorium.core.user")
@RequiredArgsConstructor
public class UserModuleConfig {

    @Bean("userEntityManagerFactory")
    public LocalContainerEntityManagerFactoryBean entityManagerFactory(
            HikariDataSourceFactoryBean factory,
            EntityManagerFactoryBuilder builder) {

        return builder
                .dataSource(factory.createByName("user"))
                .packages("ru.askorium.core.user.domain")
                .persistenceUnit("user")
                .build();
    }

    @Bean("userTransactionManager")
    public PlatformTransactionManager transactionManager(
            @Qualifier("userEntityManagerFactory") EntityManagerFactory entityManagerFactory) {
        return new JpaTransactionManager(entityManagerFactory);
    }

    public static class CustomNameGenerator implements BeanNameGenerator {

        @NonNull
        @Override
        public String generateBeanName(@NonNull BeanDefinition definition, @NonNull BeanDefinitionRegistry registry) {
            return BeanDefinitionReaderUtils.generateBeanName(definition, registry) + "-userModule";
        }
    }
}