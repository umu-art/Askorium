package ru.askorium.core.source.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.core.source.domain.SourceEntity;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface SourceJpa extends JpaRepository<SourceEntity, UUID> {
    Optional<SourceEntity> findBySourceUrl(String sourceUrl);
}
