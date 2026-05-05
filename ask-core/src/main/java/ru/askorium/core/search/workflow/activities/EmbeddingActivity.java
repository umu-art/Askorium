package ru.askorium.core.search.workflow.activities;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

import java.util.List;
import java.util.UUID;

@ActivityInterface
public interface EmbeddingActivity {

    @ActivityMethod
    String getNormalizedQuery(UUID queryId);

    @ActivityMethod
    void saveEmbedding(UUID queryId, List<Float> vector);

}
