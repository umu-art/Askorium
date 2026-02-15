package ru.askorium.core.search.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import ru.askorium.core.search.domain.SearchResultItemEntity;

import java.util.UUID;

public interface SearchResultItemJpa extends JpaRepository<SearchResultItemEntity, UUID> {
}
