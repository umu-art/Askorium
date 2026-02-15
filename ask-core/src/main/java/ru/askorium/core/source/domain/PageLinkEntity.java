package ru.askorium.core.source.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.FetchType;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.ToString;
import ru.askorium.api.client.model.LinkType;
import ru.askorium.core.common.BaseEntity;

import java.util.Objects;
import java.util.UUID;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true, exclude = "page")
@ToString(exclude = "page")
@Entity
@Table(name = "page_links")
public class PageLinkEntity extends BaseEntity {

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "page_id", nullable = false)
    private PageEntity page;

    @Column(name = "block_id")
    private UUID blockId;

    @Column(name = "href", nullable = false)
    private String href;

    @Column(name = "type", nullable = false)
    private LinkType type;

    @Column(name = "anchor_text")
    private String anchorText;

    @Column(name = "snippet")
    private String snippet;

    @Column(name = "position", nullable = false)
    private int position;

}
