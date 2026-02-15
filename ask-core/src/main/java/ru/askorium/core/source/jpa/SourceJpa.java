package ru.askorium.core.source.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import ru.askorium.core.source.domain.SourceEntity;

import java.util.UUID;

public interface SourceJpa extends JpaRepository<SourceEntity, UUID> {
}
