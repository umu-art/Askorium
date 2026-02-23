package ru.askorium.core.search;

import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import lombok.RequiredArgsConstructor;
import lombok.SneakyThrows;
import org.redisson.api.RedissonClient;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import ru.askorium.api.model.SearchCreateRequest;
import ru.askorium.api.model.SearchCreateResponse;
import ru.askorium.api.model.SearchGetResponse;
import ru.askorium.api.model.SearchStatus;
import ru.askorium.api.server.SearchApi;
import ru.askorium.core.exception.EntityNotFoundException;
import ru.askorium.core.exception.ForbiddenException;
import ru.askorium.core.exception.LockedException;
import ru.askorium.core.search.jpa.QueryJpa;
import ru.askorium.core.search.mapper.SearchMapper;
import ru.askorium.core.search.workflow.SearchWorkflow;

import java.util.Objects;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

import static ru.askorium.core.common.UserUtils.getUserId;

@Controller
@RequiredArgsConstructor
public class SearchController implements SearchApi {

    private final QueryJpa queryJpa;
    private final SearchMapper searchMapper;
    private final WorkflowClient workflowClient;
    private final RedissonClient redissonClient;

    @SneakyThrows
    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public ResponseEntity<SearchCreateResponse> createSearchQuery(SearchCreateRequest request) {

        var lock = redissonClient.getLock("search_query_lock:" + getUserId());
        if (!lock.tryLock(10, 1, TimeUnit.MINUTES)) {
            throw new LockedException();
        }

        try {
            var hasActiveQuery = queryJpa.existsByUserIdAndStatus(getUserId(), SearchStatus.RUNNING);
            if (hasActiveQuery) {
                throw new LockedException();
            }

            var entity = searchMapper.toEntity(request);
            entity.setUserId(getUserId());
            entity.setStatus(SearchStatus.RUNNING);

            queryJpa.save(entity);

            var queryId = entity.getId();
            TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
                @Override
                public void afterCommit() {
                    var stub = workflowClient.newWorkflowStub(SearchWorkflow.class, WorkflowOptions.newBuilder()
                            .setTaskQueue("askorium-search")
                            .setWorkflowId(queryId.toString())
                            .build());
                    WorkflowClient.start(stub::search, queryId);
                }
            });

            return ResponseEntity.accepted()
                    .body(new SearchCreateResponse().queryId(entity.getId()));
        } finally {
            lock.unlock();
        }
    }

    @Override
    public ResponseEntity<SearchGetResponse> getSearchQueryResult(UUID queryId) {
        var query = queryJpa.findById(queryId)
                .orElseThrow(() -> new EntityNotFoundException("Query", queryId));

        if (!Objects.equals(query.getUserId(), getUserId())) {
            throw new ForbiddenException();
        }

        return ResponseEntity.ok(searchMapper.toGetResponse(query));
    }
}
