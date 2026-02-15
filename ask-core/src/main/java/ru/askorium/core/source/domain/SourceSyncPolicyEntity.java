package ru.askorium.core.source.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.OneToOne;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.ToString;
import ru.askorium.core.common.BaseEntity;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true, exclude = "source")
@ToString(exclude = "source")
@Entity
@Table(name = "source_sync_policies")
public class SourceSyncPolicyEntity extends BaseEntity {

    @OneToOne
    @JoinColumn(name = "source_id", nullable = false)
    private SourceEntity source;

    @Column(name = "enabled", nullable = false)
    private boolean enabled = true;

    @Column(name = "interval_minutes", nullable = false)
    private int intervalMinutes = 720;

    @Column(name = "last_synced_at")
    private OffsetDateTime lastSyncedAt;

}
