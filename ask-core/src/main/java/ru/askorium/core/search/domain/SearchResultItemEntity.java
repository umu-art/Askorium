package ru.askorium.core.search.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import ru.askorium.core.common.BaseEntity;

import java.util.UUID;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Entity
@Table(name = "search_result_items")
public class SearchResultItemEntity extends BaseEntity {

    @Column(name = "query_id", nullable = false)
    private UUID queryId;

    @Column(name = "rank", nullable = false)
    private int rank;

    @Column(name = "block_id", nullable = false)
    private UUID blockId;

    @Column(name = "page_id", nullable = false)
    private UUID pageId;

    @Column(name = "score_sparse")
    private Float scoreSparse;

    @Column(name = "score_dense")
    private Float scoreDense;

    @Column(name = "score_final", nullable = false)
    private float scoreFinal;

    @Column(name = "rerank_score")
    private Float rerankScore;

}
