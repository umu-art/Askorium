package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.SearchStatus;
import ru.askorium.core.search.jpa.QueryJpa;
import ru.askorium.core.search.workflow.activities.CompletionActivity;

import java.time.OffsetDateTime;
import java.util.UUID;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class CompletionActivityImpl implements CompletionActivity {

    private final QueryJpa queryJpa;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void markDone(UUID queryId) {
        var query = queryJpa.findById(queryId)
                .orElseThrow(() -> new RuntimeException("Query not found: " + queryId));
        query.setStatus(SearchStatus.DONE);
        query.setFinishedAt(OffsetDateTime.now());
        queryJpa.save(query);
    }

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void markFailed(UUID queryId, String errorMessage) {
        var query = queryJpa.findById(queryId)
                .orElseThrow(() -> new RuntimeException("Query not found: " + queryId));
        query.setStatus(SearchStatus.FAILED);
        query.setError(errorMessage);
        query.setFinishedAt(OffsetDateTime.now());
        queryJpa.save(query);
    }
}
