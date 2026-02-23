package ru.askorium.core.search.workflow.activities.impl;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.jpa.QueryJpa;

import java.util.UUID;

@Slf4j
@RequiredArgsConstructor
public abstract class AbstractQueryActivity {

    @Autowired
    protected QueryJpa queryJpa;

    protected void executeWithQuery(UUID queryId) {
        var query = queryJpa.findById(queryId)
                .orElseThrow(() -> new RuntimeException("Query not found: " + queryId));
        log.debug("[{}] {}: start", queryId, getClass().getSimpleName());
        doWithQuery(query);
        queryJpa.save(query);
        log.debug("[{}] {}: done", queryId, getClass().getSimpleName());
    }

    protected abstract void doWithQuery(QueryEntity query);
}
