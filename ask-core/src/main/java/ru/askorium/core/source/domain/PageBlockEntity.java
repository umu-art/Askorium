package ru.askorium.core.source.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.ToString;
import ru.askorium.api.model.ContentBlockType;
import ru.askorium.core.common.BaseEntity;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true, exclude = "page")
@ToString(exclude = "page")
@Entity
@Table(name = "page_blocks")
public class PageBlockEntity extends BaseEntity {

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "page_id", nullable = false)
    private PageEntity page;

    @Column(name = "html_id")
    private String htmlId;

    @Enumerated(EnumType.STRING)
    @Column(name = "type", nullable = false)
    private ContentBlockType type;

    @Column(name = "heading_level")
    private Integer headingLevel;

    @Column(name = "text", nullable = false)
    private String text;

    public String getIndexId() {
        return "block:" + getId();
    }
}
