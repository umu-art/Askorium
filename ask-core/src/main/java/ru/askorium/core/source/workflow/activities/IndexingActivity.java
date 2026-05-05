package ru.askorium.core.source.workflow.activities;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;
import ru.askorium.core.index.IndexText;

import java.util.List;
import java.util.UUID;

@ActivityInterface
public interface IndexingActivity {

    @ActivityMethod
    List<IndexText> loadPageTexts(UUID pageId);

    @ActivityMethod
    void saveToOpenSearch(List<IndexText> texts, List<List<Float>> vectors);

}
