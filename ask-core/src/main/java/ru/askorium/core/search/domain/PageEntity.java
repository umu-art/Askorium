package ru.askorium.core.search.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.Getter;
import ru.askorium.core.common.BaseEntity;

import java.util.UUID;

@Getter
@Entity
@Table(name = "pages")
public class PageEntity extends BaseEntity {

    @Column(name = "source_id", nullable = false)
    private UUID sourceId;

    @Column(name = "url", nullable = false)
    private String url;

    @Column(name = "title")
    private String title;
}
