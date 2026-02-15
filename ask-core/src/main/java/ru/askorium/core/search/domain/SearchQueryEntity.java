package ru.askorium.core.search.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import ru.askorium.core.common.BaseEntity;

import java.time.OffsetDateTime;
import java.util.UUID;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Entity
@Table(name = "search_queries")
public class SearchQueryEntity extends BaseEntity {

    @Column(name = "user_id", nullable = false)
    private UUID userId;

    @Column(name = "source_id", nullable = false)
    private UUID sourceId;

    @Column(name = "status", nullable = false)
    private String status;

    @Column(name = "query", nullable = false)
    private String query;

    @Column(name = "mode", nullable = false)
    private String mode;

    @Column(name = "answer")
    private String answer;

    @Column(name = "error")
    private String error;

    @Column(name = "finished_at")
    private OffsetDateTime finishedAt;

}
