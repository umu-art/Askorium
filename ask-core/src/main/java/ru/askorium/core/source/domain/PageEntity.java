package ru.askorium.core.source.domain;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import ru.askorium.core.common.BaseEntity;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true, exclude = {"blocks", "links", "documents"})
@Entity
@Table(name = "pages")
public class PageEntity extends BaseEntity {

    @Column(name = "source_id", nullable = false)
    private UUID sourceId;

    @Column(name = "url", nullable = false)
    private String url;

    @Column(name = "title")
    private String title;

    @Column(name = "preview_url")
    private String previewUrl;

    @Column(name = "icon_url")
    private String iconUrl;

    @Column(name = "description")
    private String description;

    @Column(name = "language")
    private String language;

    @Column(name = "content_hash")
    private String contentHash;

    @OneToMany(mappedBy = "page", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<PageBlockEntity> blocks = new ArrayList<>();

    @OneToMany(mappedBy = "page", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<PageLinkEntity> links = new ArrayList<>();

    @OneToMany(mappedBy = "page", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<PageDocumentEntity> documents = new ArrayList<>();

}
