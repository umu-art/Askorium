package ru.askorium.core.search.workflow;

import io.temporal.activity.ActivityOptions;
import io.temporal.common.RetryOptions;
import io.temporal.spring.boot.WorkflowImpl;
import io.temporal.workflow.Workflow;
import org.springframework.stereotype.Component;
import ru.askorium.core.search.workflow.activities.AnswerActivity;
import ru.askorium.core.search.workflow.activities.EmbeddingActivity;
import ru.askorium.core.search.workflow.activities.NormalizationActivity;
import ru.askorium.core.search.workflow.activities.RerankActivity;
import ru.askorium.core.search.workflow.activities.RetrievalActivity;

import java.time.Duration;
import java.util.UUID;

@WorkflowImpl(taskQueues = "askorium-search")
public class SearchWorkflowImpl implements SearchWorkflow {

    private static final ActivityOptions DEFAULT_OPTIONS = ActivityOptions.newBuilder()
            .setStartToCloseTimeout(Duration.ofMinutes(2))
            .setRetryOptions(RetryOptions.newBuilder()
                    .setInitialInterval(Duration.ofSeconds(5))
                    .setMaximumAttempts(3)
                    .build())
            .build();

    private final NormalizationActivity normalizationActivity =
            Workflow.newActivityStub(NormalizationActivity.class, DEFAULT_OPTIONS);

    private final EmbeddingActivity embeddingActivity =
            Workflow.newActivityStub(EmbeddingActivity.class, DEFAULT_OPTIONS);

    private final RetrievalActivity retrievalActivity =
            Workflow.newActivityStub(RetrievalActivity.class, DEFAULT_OPTIONS);

    private final RerankActivity rerankActivity =
            Workflow.newActivityStub(RerankActivity.class, DEFAULT_OPTIONS);

    private final AnswerActivity answerActivity =
            Workflow.newActivityStub(AnswerActivity.class, DEFAULT_OPTIONS);

    @Override
    public void search(UUID queryId) {
        normalizationActivity.normalize(queryId);
        embeddingActivity.generateEmbedding(queryId);
        retrievalActivity.retrieve(queryId);
        rerankActivity.rerank(queryId);
        answerActivity.generateAnswer(queryId);
    }
}
