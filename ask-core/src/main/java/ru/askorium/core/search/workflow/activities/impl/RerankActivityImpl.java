package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.workflow.activities.RerankActivity;

import java.util.UUID;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class RerankActivityImpl extends AbstractQueryActivity implements RerankActivity {

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void rerank(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        // reranking via Temporal encoder activity is implemented in SearchWorkflowImpl
    }
}
