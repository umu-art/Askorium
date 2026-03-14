package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.apache.commons.lang3.StringUtils;
import ru.askorium.core.ask_encoder_api.AskEncoderService;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.workflow.activities.EmbeddingActivity;

import java.util.List;
import java.util.UUID;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class EmbeddingActivityImpl extends AbstractQueryActivity implements EmbeddingActivity {

    private final AskEncoderService askEncoderService;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void generateEmbedding(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        if (StringUtils.isBlank(query.getNormalizedQuery())) {
            throw new IllegalStateException("Normalized query is not generated");
        }

        var vector = askEncoderService.generateEmbeddings(List.of(query.getNormalizedQuery())).getFirst();
        query.setQueryVector(vector);
    }
}
