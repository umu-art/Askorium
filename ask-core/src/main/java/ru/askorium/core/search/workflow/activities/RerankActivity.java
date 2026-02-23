package ru.askorium.core.search.workflow.activities;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

import java.util.UUID;

@ActivityInterface
public interface RerankActivity {

    @ActivityMethod
    void rerank(UUID queryId);

}
