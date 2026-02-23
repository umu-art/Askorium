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
import ru.askorium.api.model.DocumentDescriptionSourceType;
import ru.askorium.core.common.BaseEntity;

@Data
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true, exclude = "page")
@ToString(exclude = "page")
@Entity
@Table(name = "page_documents")
public class PageDocumentEntity extends BaseEntity {

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "page_id", nullable = false)
    private PageEntity page;

    @Column(name = "url", nullable = false)
    private String url;

    @Column(name = "mime_type", nullable = false)
    private String mimeType;

    @Column(name = "size_bytes")
    private Integer sizeBytes;

    @Column(name = "extracted_text")
    private String extractedText;

    @Column(name = "description")
    private String description;

    @Enumerated(EnumType.STRING)
    @Column(name = "description_source")
    private DocumentDescriptionSourceType descriptionSource;

    public String getIndexId() {
        return "document:" + getId();
    }
}
