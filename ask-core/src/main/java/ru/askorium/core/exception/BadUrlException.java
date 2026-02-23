package ru.askorium.core.exception;

import org.springframework.http.HttpStatus;

public class BadUrlException extends AskCoreException {
    public BadUrlException(String url, Throwable e) {
        super("Bad url: " + url, e, HttpStatus.BAD_REQUEST);
    }
}
