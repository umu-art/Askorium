package ru.askorium.core.user.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import ru.askorium.core.user.domain.UserEntity;

import java.util.UUID;

public interface UserJpa extends JpaRepository<UserEntity, UUID> {
}
