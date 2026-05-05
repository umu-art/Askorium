package ru.askorium.core.search.workflow.activities;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@ActivityInterface
public interface RerankActivity {

    @ActivityMethod
    RerankData getRerankData(UUID queryId);

    @ActivityMethod
    void saveRerankScores(UUID queryId, List<Map<String, Object>> results);

}
