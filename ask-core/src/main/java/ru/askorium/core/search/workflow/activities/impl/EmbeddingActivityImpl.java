package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.workflow.activities.EmbeddingActivity;

import java.util.List;
import java.util.UUID;

@Slf4j
@Component
@ActivityImpl(taskQueues = "askorium-search")
public class EmbeddingActivityImpl extends AbstractQueryActivity implements EmbeddingActivity {

    @Override
    @Transactional(transactionManager = "searchTransactionManager", readOnly = true)
    public String getNormalizedQuery(UUID queryId) {
        var query = queryJpa.findById(queryId)
                .orElseThrow(() -> new RuntimeException("Query not found: " + queryId));
        if (StringUtils.isBlank(query.getNormalizedQuery())) {
            throw new IllegalStateException("Normalized query is not generated for query: " + queryId);
        }
        return query.getNormalizedQuery();
    }

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void saveEmbedding(UUID queryId, List<Float> vector) {
        var query = queryJpa.findById(queryId)
                .orElseThrow(() -> new RuntimeException("Query not found: " + queryId));
        query.setQueryVector(vector);
        queryJpa.save(query);
        log.debug("[{}] Embedding saved, dimension={}", queryId, vector.size());
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        // not used — methods implemented directly above
    }

}
