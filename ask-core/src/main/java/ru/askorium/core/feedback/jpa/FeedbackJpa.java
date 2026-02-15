package ru.askorium.core.feedback.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import ru.askorium.core.feedback.domain.FeedbackEntity;

import java.util.UUID;

public interface FeedbackJpa extends JpaRepository<FeedbackEntity, UUID> {
}
