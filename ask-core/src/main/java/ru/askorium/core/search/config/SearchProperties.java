package ru.askorium.core.search.config;

import jakarta.annotation.PostConstruct;
import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;
import ru.askorium.api.model.SearchMode;

import java.util.Arrays;
import java.util.EnumMap;
import java.util.Map;
import java.util.stream.Collectors;

@Data
@Component
@ConfigurationProperties(prefix = "askorium.search")
public class SearchProperties {

    private boolean useFullSource = true;
    private Map<SearchMode, ModeParams> modes = new EnumMap<>(SearchMode.class);

    @PostConstruct
    public void validate() {
        var missing = Arrays.stream(SearchMode.values())
                .filter(mode -> !modes.containsKey(mode))
                .map(SearchMode::name)
                .collect(Collectors.joining(", "));

        if (!missing.isEmpty()) {
            throw new IllegalStateException(
                    "Missing search configuration for modes: " + missing +
                            ". Configure askorium.search.modes.<MODE> for each SearchMode.");
        }
    }

    public ModeParams forMode(SearchMode mode) {
        return modes.get(mode);
    }

    @Data
    public static class ModeParams {
        private int bm25Size = 50;
        private int knnSize = 50;
        private int rrfK = 60;
        private int rerankTopN = 10;
    }
}
