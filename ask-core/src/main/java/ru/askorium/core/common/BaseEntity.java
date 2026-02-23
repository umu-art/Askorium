package ru.askorium.core.common;

import jakarta.persistence.Column;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.Id;
import jakarta.persistence.MappedSuperclass;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;
import java.util.UUID;

import static java.util.Objects.isNull;

@Data
@NoArgsConstructor
@MappedSuperclass
public class BaseEntity {
    @Id
    @GeneratedValue
    private UUID id;

    @Column(name = "created", nullable = false, updatable = false)
    private OffsetDateTime created;

    @Column(name = "updated")
    private OffsetDateTime updated;

    @PrePersist
    public void prePersist() {
        if (isNull(this.created)) {
            this.created = OffsetDateTime.now();
        }
    }

    @PreUpdate
    public void preUpdate() {
        this.updated = OffsetDateTime.now();
    }

}
