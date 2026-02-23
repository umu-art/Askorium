package ru.askorium.core.exception;

import org.springframework.http.HttpStatus;

public class LockedException extends AskCoreException {
    public LockedException() {
        super("Wait for the lock to be released", HttpStatus.LOCKED);
    }
}
