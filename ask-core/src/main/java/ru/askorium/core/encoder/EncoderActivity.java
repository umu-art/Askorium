package ru.askorium.core.encoder;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

import java.util.List;
import java.util.Map;

@ActivityInterface
public interface EncoderActivity {

    @ActivityMethod
    List<List<Float>> generateEmbeddings(List<String> texts);

    @ActivityMethod
    List<Map<String, Object>> rerank(String query, List<Map<String, Object>> blocks);

}
