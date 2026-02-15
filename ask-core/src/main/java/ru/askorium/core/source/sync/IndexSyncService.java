package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import ru.askorium.core.ask_encoder_api.AskEncoderService;
import ru.askorium.core.indexes.IndexService;
import ru.askorium.core.indexes.IndexText;
import ru.askorium.core.indexes.IndexVector;
import ru.askorium.core.source.domain.PageEntity;

import java.util.ArrayList;
import java.util.stream.Stream;

@Slf4j
@Service
@RequiredArgsConstructor
public class IndexSyncService {

    private final IndexService indexService;
    private final AskEncoderService askEncoderService;

    public void syncIndexes(ArrayList<PageEntity> updatedPages) {
        log.debug("syncIndexes for {} updated pages", updatedPages.size());

        var textsFromPages = updatedPages.stream()
                .flatMap(page -> {
                    var blocksTexts = page.getBlocks()
                            .stream()
                            .map(block ->
                                    new IndexText(
                                            block.getIndexId(),
                                            block.getText(),
                                            0
                                    ));

                    var documentsText = page.getDocuments()
                            .stream()
                            .map(document ->
                                    new IndexText(
                                            document.getIndexId(),
                                            document.getExtractedText(),
                                            0
                                    ));

                    return Stream.concat(blocksTexts, documentsText);
                })
                .toList();

        var textsToEmbedding = textsFromPages.stream()
                .map(IndexText::getText)
                .toList();

        var embeddings = askEncoderService.generateEmbeddings(textsToEmbedding);

        var indexTextsWithEmbeddings = new ArrayList<IndexVector>();
        for (int i = 0; i < textsFromPages.size(); i++) {
            indexTextsWithEmbeddings.add(
                    new IndexVector(
                            textsFromPages.get(i).getKey(),
                            embeddings.get(i),
                            0
                    )
            );
        }

        indexService.saveTexts(textsFromPages);
        indexService.saveVectors(indexTextsWithEmbeddings);

        log.debug("Synced {} texts and {} vectors for {} updated pages",
                textsFromPages.size(),
                indexTextsWithEmbeddings.size(),
                updatedPages.size()
        );
    }

}
