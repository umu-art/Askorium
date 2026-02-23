package ru.askorium.core.user.jpa;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.askorium.core.user.domain.UserEntity;

import java.util.UUID;

@Repository
public interface UserJpa extends JpaRepository<UserEntity, UUID> {
}
