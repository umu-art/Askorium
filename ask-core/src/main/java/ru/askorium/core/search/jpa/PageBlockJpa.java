package ru.askorium.core.search.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.core.search.domain.PageBlockEntity;

import java.util.UUID;

@Repository
public interface PageBlockJpa extends JpaRepository<PageBlockEntity, UUID> {
}
