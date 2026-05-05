package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.search.config.SearchProperties;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.jpa.QueryJpa;
import ru.askorium.core.search.workflow.activities.RerankActivity;
import ru.askorium.core.search.workflow.activities.RerankData;

import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class RerankActivityImpl implements RerankActivity {

    private final QueryJpa queryJpa;
    private final SearchProperties searchProperties;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public RerankData getRerankData(UUID queryId) {
        var query = load(queryId);
        var params = searchProperties.forMode(query.getMode());

        var blocks = query.getSources().stream()
                .sorted(Comparator.comparing(QuerySourceEntity::getFusionScore).reversed())
                .limit(params.getRerankTopN())
                .map(s -> Map.<String, Object>of("id", s.getIndexKey(), "text", s.getText()))
                .toList();

        return new RerankData(query.getNormalizedQuery(), blocks);
    }

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void saveRerankScores(UUID queryId, List<Map<String, Object>> results) {
        var query = load(queryId);

        var scoreByKey = results.stream()
                .collect(Collectors.toMap(
                        r -> (String) r.get("id"),
                        r -> ((Number) r.get("score")).floatValue()
                ));

        query.getSources().forEach(source -> {
            var score = scoreByKey.get(source.getIndexKey());
            source.setScoreFinal(score != null ? score : source.getFusionScore());
        });

        queryJpa.save(query);
    }

    private QueryEntity load(UUID queryId) {
        return queryJpa.findById(queryId)
                .orElseThrow(() -> new RuntimeException("Query not found: " + queryId));
    }
}
