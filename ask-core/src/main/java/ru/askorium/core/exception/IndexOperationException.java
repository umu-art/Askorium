package ru.askorium.core.exception;

import org.springframework.http.HttpStatus;

public class IndexOperationException extends AskCoreException {

    public IndexOperationException(String message) {
        super(message, HttpStatus.INTERNAL_SERVER_ERROR);
    }

    public IndexOperationException(String message, Throwable cause) {
        super(message, cause, HttpStatus.INTERNAL_SERVER_ERROR);
    }

}

