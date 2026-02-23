package ru.askorium.core.feedback.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.core.feedback.domain.FeedbackEntity;

import java.util.UUID;

@Repository
public interface FeedbackJpa extends JpaRepository<FeedbackEntity, UUID> {
}
