package ru.askorium.core.search.workflow.activities.impl;

import com.google.common.collect.Streams;
import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import ru.askorium.core.index.IndexService;
import ru.askorium.core.index.IndexText;
import ru.askorium.core.index.IndexVector;
import ru.askorium.core.search.config.SearchProperties;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.jpa.PageBlockJpa;
import ru.askorium.core.search.jpa.PageDocumentJpa;
import ru.askorium.core.search.workflow.activities.RetrievalActivity;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class RetrievalActivityImpl extends AbstractQueryActivity implements RetrievalActivity {

    private final IndexService indexService;
    private final SearchProperties searchProperties;
    private final PageBlockJpa pageBlockJpa;
    private final PageDocumentJpa pageDocumentJpa;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void retrieve(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        if (CollectionUtils.isEmpty(query.getQueryVector())) {
            throw new IllegalStateException("Query vector is not generated");
        }

        if (StringUtils.isBlank(query.getNormalizedQuery())) {
            throw new IllegalStateException("Normalized query is not generated");
        }

        var params = searchProperties.forMode(query.getMode());

        var bm25Results = indexService.searchBM25(query.getNormalizedQuery(), params.getBm25Size());
        var knnResults = indexService.searchKnn(new IndexVector(null, query.getQueryVector(), 0f), params.getKnnSize());

        List<QuerySourceEntity> sources = mapToSources(bm25Results, knnResults, query.getSourceId());

        fillScores(sources, bm25Results, knnResults, params.getRrfK());

        sources.forEach(s -> s.setScoreFinal(s.getFusionScore()));

        if (searchProperties.isUseFullSource()) {
            sources = deduplicateByPage(sources);
        }

        sources.forEach(s -> s.setQuery(query));
        query.getSources().addAll(sources);
    }

    private List<QuerySourceEntity> mapToSources(List<IndexText> bm25Results, List<IndexVector> knnResults, UUID sourceId) {
        var keys = Streams.concat(
                        bm25Results.stream().map(IndexText::getKey),
                        knnResults.stream().map(IndexVector::getKey))
                .toList();

        var blockIds = keys.stream()
                .filter(StringUtils::isNotBlank)
                .filter(k -> k.startsWith("block:"))
                .map(k -> k.substring(6))
                .map(UUID::fromString)
                .distinct()
                .toList();

        var documentIds = keys.stream()
                .filter(StringUtils::isNotBlank)
                .filter(k -> k.startsWith("document:"))
                .map(k -> k.substring(9))
                .map(UUID::fromString)
                .distinct()
                .toList();

        var blocks = pageBlockJpa.findAllByIdInAndPage_SourceId(blockIds, sourceId);
        var documents = pageDocumentJpa.findAllByIdInAndPage_SourceId(documentIds, sourceId);

        return Streams.concat(
                blocks.stream().map(block -> {
                    var entity = new QuerySourceEntity();
                    entity.setIndexKey("block:" + block.getId());
                    entity.setText(block.getText());
                    entity.setUrl(block.getPage().getUrl());
                    entity.setTitle(block.getPage().getTitle());
                    return entity;
                }),
                documents.stream().map(doc -> {
                    var entity = new QuerySourceEntity();
                    entity.setIndexKey("document:" + doc.getId());
                    entity.setText(doc.getExtractedText());
                    entity.setUrl(doc.getUrl());
                    entity.setTitle(doc.getDescription());
                    return entity;
                })
        ).collect(Collectors.toCollection(ArrayList::new));
    }

    private List<QuerySourceEntity> deduplicateByPage(List<QuerySourceEntity> sources) {
        var blockSources = sources.stream()
                .filter(s -> s.getIndexKey().startsWith("block:"))
                .toList();

        var nonBlockSources = sources.stream()
                .filter(s -> !s.getIndexKey().startsWith("block:"))
                .toList();

        var blockIds = blockSources.stream()
                .map(s -> UUID.fromString(s.getIndexKey().substring(6)))
                .toList();

        var pageIdByKey = pageBlockJpa.findAllById(blockIds).stream()
                .collect(Collectors.toMap(
                        b -> "block:" + b.getId(),
                        b -> b.getPage().getId()
                ));

        var deduplicatedBlocks = blockSources.stream()
                .filter(s -> pageIdByKey.containsKey(s.getIndexKey()))
                .collect(Collectors.toMap(
                        s -> pageIdByKey.get(s.getIndexKey()),
                        s -> s,
                        (a, b) -> a.getFusionScore() >= b.getFusionScore() ? a : b
                ))
                .values();

        var result = new ArrayList<>(deduplicatedBlocks);
        result.addAll(nonBlockSources);
        return result;
    }

    private void fillScores(List<QuerySourceEntity> sources, List<IndexText> bm25Results, List<IndexVector> knnResults, int rrfK) {
        var bm25PositionMap = bm25Results.stream()
                .collect(Collectors.toMap(IndexText::getKey, bm25Results::indexOf));
        var bm25ScoreMap = bm25Results.stream()
                .collect(Collectors.toMap(IndexText::getKey, IndexText::getRank));

        var knnPositionMap = knnResults.stream()
                .collect(Collectors.toMap(IndexVector::getKey, knnResults::indexOf));
        var knnScoreMap = knnResults.stream()
                .collect(Collectors.toMap(IndexVector::getKey, IndexVector::getRank));

        sources.forEach(source -> {
            var rrfScore = 0.0f;

            var bm25Pos = bm25PositionMap.get(source.getIndexKey());
            if (bm25Pos != null) {
                rrfScore = Math.max(rrfScore, 1.0f / (rrfK + bm25Pos + 1));
                source.setScoreSparse(bm25ScoreMap.get(source.getIndexKey()));
            }

            var knnPos = knnPositionMap.get(source.getIndexKey());
            if (knnPos != null) {
                rrfScore = Math.max(rrfScore, 1.0f / (rrfK + knnPos + 1));
                source.setScoreDense(knnScoreMap.get(source.getIndexKey()));
            }

            source.setFusionScore(rrfScore);
        });
    }
}
