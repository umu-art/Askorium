package ru.askorium.core.exception;

import org.springframework.http.HttpStatus;

import java.util.UUID;

public class EntityNotFoundException extends AskCoreException {
    public EntityNotFoundException(String entity, UUID key) {
        super(String.format("%s with id %s not found", entity, key), HttpStatus.NOT_FOUND);
    }
}
