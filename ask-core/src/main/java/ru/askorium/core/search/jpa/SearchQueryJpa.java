package ru.askorium.core.search.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import ru.askorium.core.search.domain.SearchQueryEntity;

import java.util.UUID;

public interface SearchQueryJpa extends JpaRepository<SearchQueryEntity, UUID> {
}
