package ru.askorium.core.exception;

import org.apache.http.HttpStatus;

public class AutoSyncFailedException extends AskCoreException {

    public AutoSyncFailedException() {
        super("Failed to preform auto-sync", HttpStatus.SC_INTERNAL_SERVER_ERROR);
    }

}
