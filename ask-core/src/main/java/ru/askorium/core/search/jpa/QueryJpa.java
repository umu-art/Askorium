package ru.askorium.core.search.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.api.model.SearchStatus;
import ru.askorium.core.search.domain.QueryEntity;

import java.time.OffsetDateTime;
import java.util.UUID;

@Repository
public interface QueryJpa extends JpaRepository<QueryEntity, UUID> {
    boolean existsByUserIdAndStatus(UUID userId, SearchStatus searchStatus);
    boolean existsByUserIdAndStatusAndCreatedAfter(UUID userId, SearchStatus status, OffsetDateTime after);
}
