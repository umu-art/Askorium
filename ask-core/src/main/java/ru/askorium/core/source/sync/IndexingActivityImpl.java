package ru.askorium.core.source.sync;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.index.IndexService;
import ru.askorium.core.index.IndexText;
import ru.askorium.core.index.IndexVector;
import ru.askorium.core.source.workflow.activities.IndexingActivity;
import ru.askorium.core.source.jpa.PageJpa;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.stream.Stream;

@Slf4j
@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class IndexingActivityImpl implements IndexingActivity {

    private final PageJpa pageJpa;
    private final IndexService indexService;

    @Override
    @Transactional(transactionManager = "sourcesTransactionManager", readOnly = true)
    public List<IndexText> loadPageTexts(UUID pageId) {
        var page = pageJpa.findById(pageId)
                .orElseThrow(() -> new RuntimeException("Page not found: " + pageId));

        var blockTexts = page.getBlocks().stream()
                .map(block -> new IndexText(block.getIndexId(), block.getText(), 0f));

        var documentTexts = page.getDocuments().stream()
                .filter(doc -> doc.getExtractedText() != null && !doc.getExtractedText().isBlank())
                .map(doc -> new IndexText(doc.getIndexId(), doc.getExtractedText(), 0f));

        var result = Stream.concat(blockTexts, documentTexts).toList();
        log.debug("[{}] Loaded {} texts for indexing", pageId, result.size());
        return result;
    }

    @Override
    public void saveToOpenSearch(List<IndexText> texts, List<List<Float>> vectors) {
        var indexVectors = new ArrayList<IndexVector>(texts.size());
        for (int i = 0; i < texts.size(); i++) {
            indexVectors.add(new IndexVector(texts.get(i).getKey(), vectors.get(i), 0f));
        }

        indexService.saveTexts(texts);
        indexService.saveVectors(indexVectors);

        log.debug("Indexed {} texts and {} vectors", texts.size(), indexVectors.size());
    }

}
