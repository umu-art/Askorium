package ru.askorium.core.index.service;

import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch.OpenSearchClient;
import org.opensearch.client.opensearch._types.FieldValue;
import org.opensearch.client.opensearch.core.BulkRequest;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;
import ru.askorium.core.exception.IndexOperationException;
import ru.askorium.core.index.IndexService;
import ru.askorium.core.index.IndexText;
import ru.askorium.core.index.IndexVector;
import ru.askorium.core.index.config.IndexProperties;

import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;

@Slf4j
@Service
@RequiredArgsConstructor
public class OpenSearchIndexService implements IndexService {

    private final OpenSearchClient client;
    private final IndexProperties properties;

    @PostConstruct
    void ensureIndices() throws IOException {
        createTextIndexIfNotExists();
        createVectorIndexIfNotExists();
        log.info("OpenSearch ready");
    }

    @Override
    public void saveTexts(List<IndexText> texts) {
        if (CollectionUtils.isEmpty(texts)) return;

        List<Map<String, Object>> docs = texts.stream()
                .map(t -> {
                    Map<String, Object> doc = new HashMap<>();
                    doc.put("key", t.getKey());
                    doc.put("text", t.getText());
                    return doc;
                })
                .toList();

        bulkSave(docs, properties.getTextIndexName());
    }

    @Override
    public void saveVectors(List<IndexVector> vectors) {
        if (CollectionUtils.isEmpty(vectors)) return;

        List<Map<String, Object>> docs = vectors.stream()
                .map(v -> {
                    Map<String, Object> doc = new HashMap<>();
                    doc.put("key", v.getKey());
                    doc.put("values", v.getValues());
                    return doc;
                })
                .toList();

        bulkSave(docs, properties.getVectorIndexName());
    }

    private void bulkSave(List<Map<String, Object>> docs, String indexName) {
        var bulk = new BulkRequest.Builder();
        for (var doc : docs) {
            bulk.operations(op -> op
                    .index(idx -> idx
                            .index(indexName)
                            .id((String) doc.get("key"))
                            .document(doc)));
        }

        try {
            var response = client.bulk(bulk.build());
            if (response.errors()) {
                response.items().stream()
                        .filter(item -> item.error() != null)
                        .forEach(item -> log.error("Bulk index error [{}]: {}", item.id(), item.error().reason()));
                throw new IndexOperationException("Bulk indexing completed with errors");
            }
        } catch (IOException e) {
            throw new IndexOperationException("Failed to bulk index into " + indexName, e);
        }
    }

    @Override
    public List<IndexText> searchBM25(String query, int size) {
        try {
            var response = client.search(s -> s
                            .index(properties.getTextIndexName())
                            .size(size)
                            .query(q -> q
                                    .match(m -> m
                                            .field("text")
                                            .query(FieldValue.of(query))
                                    )
                            ),
                    IndexText.class);

            return response.hits().hits().stream()
                    .map(hit -> {
                        var doc = hit.source();
                        Objects.requireNonNull(doc).setRank(scoreOf(hit.score()));
                        return doc;
                    })
                    .toList();

        } catch (IOException e) {
            throw new IndexOperationException("BM25 search failed", e);
        }
    }

    @Override
    public List<IndexVector> searchKnn(IndexVector vector, int size) {
        try {
            var response = client.search(s -> s
                            .index(properties.getVectorIndexName())
                            .size(size)
                            .query(q -> q
                                    .knn(k -> k
                                            .field("values")
                                            .vector(vector.getValues())
                                            .k(size)
                                    )
                            ),
                    IndexVector.class);

            return response.hits().hits().stream()
                    .map(hit -> {
                        var doc = hit.source();
                        Objects.requireNonNull(doc).setRank(scoreOf(hit.score()));
                        return doc;
                    })
                    .toList();

        } catch (IOException e) {
            throw new IndexOperationException("kNN search failed", e);
        }
    }

    private void createTextIndexIfNotExists() throws IOException {
        var name = properties.getTextIndexName();
        if (indexExists(name)) return;

        client.indices().create(c -> c
                .index(name)
                .mappings(m -> m
                        .properties("key", p -> p.keyword(k -> k))
                        .properties("text", p -> p.text(t -> t.analyzer("standard")))));

        log.info("Created text index: {}", name);
    }

    private void createVectorIndexIfNotExists() throws IOException {
        var name = properties.getVectorIndexName();
        if (indexExists(name)) return;

        client.indices().create(c -> c
                .index(name)
                .settings(s -> s.knn(true))
                .mappings(m -> m
                        .properties("key", p -> p.keyword(k -> k))
                        .properties("values", p -> p.knnVector(kv -> kv
                                .dimension(properties.getVectorDimension())
                                .method(method -> method
                                        .name("hnsw")
                                        .engine("lucene")
                                        .spaceType("cosinesimil")
                                        .parameters(Map.of(
                                                "m", JsonData.of(properties.getHnswM()),
                                                "ef_construction", JsonData.of(properties.getHnswEfConstruction())
                                        ))
                                )
                        ))));

        log.info("Created vector index: {}", name);
    }

    private boolean indexExists(String name) throws IOException {
        return client.indices().exists(e -> e.index(name)).value();
    }

    @Override
    public void deleteStaleKeys(List<String> validKeys) {
        var fieldValues = validKeys.stream().map(FieldValue::of).toList();
        deleteStaleFromIndex(properties.getTextIndexName(), fieldValues);
        deleteStaleFromIndex(properties.getVectorIndexName(), fieldValues);
    }

    private void deleteStaleFromIndex(String indexName, List<FieldValue> validKeyValues) {
        try {
            var response = client.deleteByQuery(d -> d
                    .index(indexName)
                    .query(q -> q
                            .bool(b -> b
                                    .mustNot(mn -> mn
                                            .terms(t -> t
                                                    .field("key")
                                                    .terms(tv -> tv.value(validKeyValues))
                                            )
                                    )
                            )
                    )
            );
            log.info("Deleted {} stale entries from index '{}'", response.deleted(), indexName);
        } catch (IOException e) {
            throw new IndexOperationException("Failed to delete stale keys from " + indexName, e);
        }
    }

    private float scoreOf(Double score) {
        return Objects.requireNonNullElse(score, 0d).floatValue();
    }
}