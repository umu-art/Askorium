package ru.askorium.core.search.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.Getter;
import ru.askorium.core.common.BaseEntity;

@Getter
@Entity
@Table(name = "page_documents")
public class PageDocumentEntity extends BaseEntity {

    @Column(name = "url", nullable = false)
    private String url;

    @Column(name = "extracted_text")
    private String extractedText;

    @Column(name = "description")
    private String description;

}
