package ru.askorium.core.exceptions;

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
}