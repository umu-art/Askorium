package ru.askorium.core.source.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.core.source.domain.SyncTaskUrlEntity;
import ru.askorium.core.source.domain.SyncTaskUrlStatus;

import java.util.List;
import java.util.UUID;

@Repository
public interface SyncTaskUrlJpa extends JpaRepository<SyncTaskUrlEntity, UUID> {

    boolean existsByTaskIdAndUrl(UUID taskId, String url);

    long countByTaskIdAndStatus(UUID taskId, SyncTaskUrlStatus status);

    List<SyncTaskUrlEntity> findByTaskIdAndStatus(UUID taskId, SyncTaskUrlStatus status);
}
