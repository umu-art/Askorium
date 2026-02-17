package ru.askorium.core.exception;

import lombok.Getter;
import org.springframework.http.HttpStatus;

@Getter
public class AskCoreException extends RuntimeException {

    private final int code;

    public AskCoreException(String message, int code) {
        super(message);
        this.code = code;
    }

    public AskCoreException(String message, HttpStatus code) {
        super(message);
        this.code = code.value();
    }

    public AskCoreException(String message, Throwable cause, HttpStatus code) {
        super(message, cause);
        this.code = code.value();
    }
}