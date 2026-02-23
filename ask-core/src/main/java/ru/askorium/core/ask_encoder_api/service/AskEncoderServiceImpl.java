package ru.askorium.core.ask_encoder_api.service;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import ru.askorium.api.client.EncoderApi;
import ru.askorium.api.model.EmbeddingRequest;
import ru.askorium.api.model.RerankBlock;
import ru.askorium.api.model.RerankRequest;
import ru.askorium.api.model.RerankResult;
import ru.askorium.core.ask_encoder_api.AskEncoderService;

import java.util.List;

@Service
@RequiredArgsConstructor
public class AskEncoderServiceImpl implements AskEncoderService {

    private final EncoderApi encoderApi;

    @Override
    public List<List<Float>> generateEmbeddings(List<String> texts) {
        var request = new EmbeddingRequest()
                .texts(texts);

        return encoderApi.generateEmbeddings(request)
                .getEmbeddings();
    }

    @Override
    public List<RerankResult> rerank(String query, List<RerankBlock> items) {
        var request = new RerankRequest()
                .query(query)
                .blocks(items);

        return encoderApi.rerank(request)
                .getResults();
    }
}
