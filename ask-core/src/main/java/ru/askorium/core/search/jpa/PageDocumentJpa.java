package ru.askorium.core.search.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.core.search.domain.PageDocumentEntity;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PageDocumentJpa extends JpaRepository<PageDocumentEntity, UUID> {
    Optional<PageDocumentEntity> findByUrl(String url);
    List<PageDocumentEntity> findAllByIdInAndPage_SourceId(List<UUID> ids, UUID sourceId);
}
