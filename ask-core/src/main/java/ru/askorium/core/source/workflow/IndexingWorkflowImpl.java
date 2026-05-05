package ru.askorium.core.source.workflow;

import io.temporal.activity.ActivityOptions;
import io.temporal.common.RetryOptions;
import io.temporal.spring.boot.WorkflowImpl;
import io.temporal.workflow.Workflow;
import ru.askorium.core.encoder.EncoderActivity;
import ru.askorium.core.index.IndexText;
import ru.askorium.core.source.workflow.activities.IndexingActivity;

import java.time.Duration;
import java.util.List;
import java.util.UUID;

@WorkflowImpl(taskQueues = "askorium-search")
public class IndexingWorkflowImpl implements IndexingWorkflow {

    private static final ActivityOptions LOCAL_OPTIONS = ActivityOptions.newBuilder()
            .setStartToCloseTimeout(Duration.ofMinutes(2))
            .setRetryOptions(RetryOptions.newBuilder()
                    .setInitialInterval(Duration.ofSeconds(5))
                    .setMaximumAttempts(3)
                    .build())
            .build();

    private static final ActivityOptions ENCODER_OPTIONS = ActivityOptions.newBuilder()
            .setTaskQueue("askorium-encoder")
            .setStartToCloseTimeout(Duration.ofMinutes(10))
            .setRetryOptions(RetryOptions.newBuilder()
                    .setInitialInterval(Duration.ofSeconds(5))
                    .setMaximumAttempts(3)
                    .build())
            .build();

    private final IndexingActivity indexingActivity =
            Workflow.newActivityStub(IndexingActivity.class, LOCAL_OPTIONS);

    private final EncoderActivity encoderActivity =
            Workflow.newActivityStub(EncoderActivity.class, ENCODER_OPTIONS);

    @Override
    public void index(UUID pageId) {
        List<IndexText> texts = indexingActivity.loadPageTexts(pageId);
        if (texts.isEmpty()) {
            return;
        }

        List<String> rawTexts = texts.stream().map(IndexText::getText).toList();
        List<List<Float>> vectors = encoderActivity.generateEmbeddings(rawTexts);

        indexingActivity.saveToOpenSearch(texts, vectors);
    }

}
