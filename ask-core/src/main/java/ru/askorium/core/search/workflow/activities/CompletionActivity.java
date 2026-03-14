package ru.askorium.core.search.workflow.activities;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

import java.util.UUID;

@ActivityInterface
public interface CompletionActivity {

    @ActivityMethod
    void markDone(UUID queryId);

    @ActivityMethod
    void markFailed(UUID queryId, String errorMessage);

}
