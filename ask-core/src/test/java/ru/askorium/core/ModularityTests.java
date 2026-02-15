package ru.askorium.core;

import com.tngtech.archunit.base.DescribedPredicate;
import com.tngtech.archunit.core.domain.JavaClass;
import org.junit.jupiter.api.Test;
import org.springframework.modulith.core.ApplicationModules;
import org.springframework.modulith.docs.Documenter;

import java.util.List;

class ModularityTests {

    private final ApplicationModules modules = ApplicationModules.of(AskCoreApplication.class);

    @Test
    void verifiesModularStructure() {
        modules.forEach(System.out::println);
        modules.verify();
    }

    @Test
    void createModuleDocumentation() {

        var businessModules = ApplicationModules.of(AskCoreApplication.class);
        new Documenter(businessModules).writeDocumentation();

    }
}