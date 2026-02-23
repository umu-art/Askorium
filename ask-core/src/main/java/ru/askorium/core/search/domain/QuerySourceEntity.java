package ru.askorium.core.search.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import ru.askorium.core.common.BaseEntity;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Entity
@Table(name = "search_query_sources")
public class QuerySourceEntity extends BaseEntity {

    @ManyToOne
    @JoinColumn(name = "query_id", referencedColumnName = "id", nullable = false)
    private QueryEntity query;

    @Column(name = "index_key", nullable = false)
    private String indexKey;

    @Column(name = "url", nullable = false)
    private String url;

    @Column(name = "title", nullable = false)
    private String title;

    @Column(name = "date")
    private OffsetDateTime date;

    @Column(name = "text", nullable = false)
    private String text;

    @Column(name = "score_sparse")
    private Float scoreSparse;

    @Column(name = "score_dense")
    private Float scoreDense;

    @Column(name = "fusion_score")
    private Float fusionScore;

    @Column(name = "score_final", nullable = false)
    private Float scoreFinal;

}
