package ru.askorium.core.source.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;
import ru.askorium.core.source.domain.PageEntity;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PageJpa extends JpaRepository<PageEntity, UUID> {

    List<PageEntity> findBySourceId(UUID sourceId);

    Optional<PageEntity> findByUrl(String url);

    @Query("SELECT CONCAT('block:', b.id) FROM PageEntity p JOIN p.blocks b")
    List<String> findAllBlockIndexIds();

    @Query("SELECT CONCAT('document:', d.id) FROM PageEntity p JOIN p.documents d")
    List<String> findAllDocumentIndexIds();

}
