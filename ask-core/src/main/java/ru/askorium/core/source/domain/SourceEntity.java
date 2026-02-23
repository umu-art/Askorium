package ru.askorium.core.source.domain;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.OneToOne;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import ru.askorium.core.common.BaseEntity;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@Entity
@Table(name = "sources")
public class SourceEntity extends BaseEntity {

    @Column(name = "source_url", nullable = false)
    private String sourceUrl;

    @OneToOne(mappedBy = "source", cascade = CascadeType.ALL, orphanRemoval = true)
    private SourceSyncPolicyEntity syncPolicy;

}
