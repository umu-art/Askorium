package ru.askorium.core.search.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import lombok.Getter;
import ru.askorium.core.common.BaseEntity;

@Getter
@Entity
@Table(name = "page_blocks")
public class PageBlockEntity extends BaseEntity {

    @ManyToOne
    @JoinColumn(name = "page_id", nullable = false)
    private PageEntity page;

    @Column(name = "text", nullable = false)
    private String text;
}
