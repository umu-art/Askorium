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
}
