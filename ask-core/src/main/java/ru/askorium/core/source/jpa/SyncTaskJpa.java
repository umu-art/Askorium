package ru.askorium.core.source.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.source.domain.SyncTaskEntity;

import java.util.UUID;

@Repository
public interface SyncTaskJpa extends JpaRepository<SyncTaskEntity, UUID> {

    @Transactional
    @Modifying(clearAutomatically = true)
    @Query("UPDATE SyncTaskEntity t SET t.pagesDiscovered = t.pagesDiscovered + 1 WHERE t.id = :taskId")
    void incrementPagesDiscovered(@Param("taskId") UUID taskId);

    @Transactional
    @Modifying(clearAutomatically = true)
    @Query("UPDATE SyncTaskEntity t SET t.pagesScraped = t.pagesScraped + 1 WHERE t.id = :taskId")
    void incrementPagesScraped(@Param("taskId") UUID taskId);

    @Transactional
    @Modifying(clearAutomatically = true)
    @Query("UPDATE SyncTaskEntity t SET t.pagesFailed = t.pagesFailed + 1 WHERE t.id = :taskId")
    void incrementPagesFailed(@Param("taskId") UUID taskId);
}
