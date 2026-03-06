package ru.askorium.core.exception;

import org.springframework.http.HttpStatus;

public class AutoSyncFailedException extends AskCoreException {

    public AutoSyncFailedException() {
        super("Failed to preform auto-sync", HttpStatus.INTERNAL_SERVER_ERROR);
    }

}
