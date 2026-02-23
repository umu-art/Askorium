package ru.askorium.core.exception;

import org.springframework.http.HttpStatus;

public class ForbiddenException extends AskCoreException {
    public ForbiddenException() {
        super("Forbidden", HttpStatus.FORBIDDEN);
    }
}
