package ru.askorium.core.index.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "askorium.index")
public class IndexProperties {

    private String textIndexName;

    private String vectorIndexName;

    private int vectorDimension;

    private int hnswM;

    private int hnswEfConstruction;

}
