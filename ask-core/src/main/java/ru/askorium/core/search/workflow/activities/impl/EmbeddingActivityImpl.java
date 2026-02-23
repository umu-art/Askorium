package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.ask_encoder_api.AskEncoderService;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.workflow.activities.EmbeddingActivity;

import java.util.List;
import java.util.UUID;

@Component
@ActivityImpl
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
        var vector = askEncoderService.generateEmbeddings(List.of(query.getNormalizedQuery())).getFirst();
        query.setQueryVector(vector);
    }
}
