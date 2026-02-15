package ru.askorium.core.ask_encoder_api;

import ru.askorium.api.client.model.RerankBlock;
import ru.askorium.api.client.model.RerankResult;

import java.util.List;

public interface AskEncoderService {

    List<List<Float>> generateEmbeddings(List<String> texts);

    List<RerankResult> rerank(String query, List<RerankBlock> items);

}
