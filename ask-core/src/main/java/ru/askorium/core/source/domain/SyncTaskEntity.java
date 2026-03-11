package ru.askorium.core.source.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
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
@Table(name = "sync_tasks")
public class SyncTaskEntity extends BaseEntity {

    @Column(name = "source_id", nullable = false)
    private UUID sourceId;

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false)
    private SyncTaskStatus status;

    @Column(name = "force_sync", nullable = false)
    private boolean forceSync;

    @Column(name = "pages_discovered", nullable = false)
    private int pagesDiscovered;

    @Column(name = "pages_scraped", nullable = false)
    private int pagesScraped;

    @Column(name = "pages_failed", nullable = false)
    private int pagesFailed;

    @Column(name = "error_message")
    private String errorMessage;

    @Column(name = "started_at")
    private OffsetDateTime startedAt;

    @Column(name = "finished_at")
    private OffsetDateTime finishedAt;
}
