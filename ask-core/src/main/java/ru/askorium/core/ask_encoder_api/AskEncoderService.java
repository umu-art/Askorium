package ru.askorium.core.ask_encoder_api;

import ru.askorium.api.model.RerankBlock;
import ru.askorium.api.model.RerankResult;

import java.util.List;

public interface AskEncoderService {

    List<List<Float>> generateEmbeddings(List<String> texts);

    List<RerankResult> rerank(String query, List<RerankBlock> items);

}
