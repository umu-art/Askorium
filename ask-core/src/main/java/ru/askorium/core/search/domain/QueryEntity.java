package ru.askorium.core.search.domain;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.OneToMany;
import jakarta.persistence.OrderBy;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.Fetch;
import org.hibernate.annotations.FetchMode;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;
import ru.askorium.api.model.SearchMode;
import ru.askorium.api.model.SearchStatus;
import ru.askorium.core.common.BaseEntity;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Entity
@Table(name = "search_queries")
public class QueryEntity extends BaseEntity {

    @Column(name = "user_id", nullable = false)
    private UUID userId;

    @Column(name = "source_id", nullable = false)
    private UUID sourceId;

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false)
    private SearchStatus status;

    @Column(name = "query", nullable = false)
    private String query;

    @Column(name = "normalized_query")
    private String normalizedQuery;

    @JdbcTypeCode(SqlTypes.JSON)
    @Column(name = "query_vector", columnDefinition = "jsonb")
    private List<Float> queryVector;

    @Enumerated(EnumType.STRING)
    @Column(name = "mode", nullable = false)
    private SearchMode mode;

    @Column(name = "answer")
    private String answer;

    @Column(name = "error")
    private String error;

    @Column(name = "finished_at")
    private OffsetDateTime finishedAt;

    @Fetch(FetchMode.JOIN)
    @OrderBy("scoreFinal desc")
    @OneToMany(mappedBy = "query", fetch = FetchType.EAGER, cascade = CascadeType.ALL, orphanRemoval = true)
    private List<QuerySourceEntity> sources;

}
