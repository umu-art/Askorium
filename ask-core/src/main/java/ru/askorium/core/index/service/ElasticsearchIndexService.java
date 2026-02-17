package ru.askorium.core.index.service;

import co.elastic.clients.elasticsearch.ElasticsearchClient;
import co.elastic.clients.elasticsearch._types.mapping.DenseVectorIndexOptionsType;
import co.elastic.clients.elasticsearch._types.mapping.DenseVectorSimilarity;
import co.elastic.clients.elasticsearch.core.BulkRequest;
import co.elastic.clients.elasticsearch.core.search.Hit;
import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;
import ru.askorium.core.exception.IndexOperationException;
import ru.askorium.core.index.IndexService;
import ru.askorium.core.index.IndexText;
import ru.askorium.core.index.IndexVector;
import ru.askorium.core.index.config.IndexProperties;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.Objects;

@Slf4j
@Service
@RequiredArgsConstructor
public class ElasticsearchIndexService implements IndexService {

    private final ElasticsearchClient esClient;
    private final IndexProperties properties;

    @PostConstruct
    void ensureIndices() throws IOException {
        createTextIndexIfNotExists();
        createVectorIndexIfNotExists();
    }

    @Override
    public void saveTexts(List<IndexText> texts) {
        if (CollectionUtils.isEmpty(texts)) {
            return;
        }

        var documents = texts.stream()
                .map(text -> Map.of(
                        "key", text.getKey(),
                        "text", text.getText()
                ))
                .toList();

        try {
            bulkSave(documents, properties.getTextIndexName());
        } catch (IOException e) {
            throw new IndexOperationException("Failed to bulk index texts", e);
        }
    }

    @Override
    public void saveVectors(List<IndexVector> vectors) {
        if (CollectionUtils.isEmpty(vectors)) {
            return;
        }

        var documents = vectors.stream()
                .map(vector -> Map.of(
                        "key", vector.getKey(),
                        "values", vector.getValues()
                ))
                .toList();

        try {
            bulkSave(documents, properties.getVectorIndexName());
        } catch (IOException e) {
            throw new IndexOperationException("Failed to bulk index vectors", e);
        }
    }

    private <T> void bulkSave(List<Map<String, T>> documents, String indexName) throws IOException {
        var bulkRequest = new BulkRequest.Builder();
        for (var doc : documents) {
            bulkRequest.operations(op -> op
                    .index(idx -> idx
                            .index(indexName)
                            .id((String) doc.get("key"))
                            .document(doc)
                    )
            );
        }

        var response = esClient.bulk(bulkRequest.build());
        if (response.errors()) {
            response.items().stream()
                    .filter(item -> item.error() != null)
                    .forEach(item -> log.error("Failed to index document {}: {}",
                            item.id(), item.error().reason()));

            throw new IndexOperationException("Bulk indexing completed with errors");
        }
    }

    @Override
    public List<IndexText> searchBM25(String query, int size) {
        try {
            var response = esClient.search(s -> s
                            .index(properties.getTextIndexName())
                            .query(q -> q
                                    .match(m -> m
                                            .field("text")
                                            .query(query)
                                    )
                            )
                            .size(size),
                    IndexText.class
            );

            return response.hits().hits().stream()
                    .map(this::toIndexText)
                    .toList();

        } catch (IOException e) {
            throw new IndexOperationException("Failed to search BM25", e);
        }
    }

    @Override
    public List<IndexVector> searchKnn(IndexVector vector, int size) {
        int numCandidates = Math.max(size * 10, 100);

        try {
            var response = esClient.search(s -> s
                            .index(properties.getVectorIndexName())
                            .knn(k -> k
                                    .field("values")
                                    .queryVector(vector.getValues())
                                    .k(size)
                                    .numCandidates(numCandidates)
                            ),
                    IndexVector.class
            );

            return response.hits().hits().stream()
                    .map(this::toIndexVector)
                    .toList();

        } catch (IOException e) {
            throw new IndexOperationException("Failed to search kNN", e);
        }
    }

    private void createTextIndexIfNotExists() throws IOException {
        var indexName = properties.getTextIndexName();

        var exists = esClient.indices()
                .exists(e -> e.index(indexName))
                .value();

        if (exists) {
            return;
        }

        esClient.indices()
                .create(c -> c
                        .index(indexName)
                        .mappings(m -> m
                                .properties("key", p -> p.keyword(k -> k))
                                .properties("text", p -> p.text(t -> t.analyzer("standard")))
                        )
                );

        log.info("Created text index: {}", indexName);
    }

    private void createVectorIndexIfNotExists() throws IOException {
        var indexName = properties.getVectorIndexName();

        var exists = esClient.indices()
                .exists(e -> e.index(indexName))
                .value();

        if (exists) {
            return;
        }

        esClient.indices()
                .create(c -> c
                        .index(indexName)
                        .mappings(m -> m
                                .properties("key", p -> p.keyword(k -> k))
                                .properties("values", p -> p.denseVector(dv -> dv
                                        .dims(properties.getVectorDimension())
                                        .similarity(DenseVectorSimilarity.Cosine)
                                        .index(true)
                                        .indexOptions(io -> io
                                                .type(DenseVectorIndexOptionsType.Hnsw)
                                                .m(properties.getHnswM())
                                                .efConstruction(properties.getHnswEfConstruction())
                                        )
                                ))
                        )
                );

        log.info("Created vector index: {}", indexName);
    }

    private IndexText toIndexText(Hit<IndexText> hit) {
        var doc = hit.source();
        assert doc != null;
        doc.setRank(Objects.requireNonNullElse(hit.score(), 0d).floatValue());
        return doc;
    }

    private IndexVector toIndexVector(Hit<IndexVector> hit) {
        var doc = hit.source();
        assert doc != null;
        doc.setRank(Objects.requireNonNullElse(hit.score(), 0d).floatValue());
        return doc;
    }

}
