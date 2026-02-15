package ru.askorium.core.exceptions;

import org.apache.http.HttpStatus;

public class AutoSyncFailedException extends AskCoreException {

    public AutoSyncFailedException() {
        super("Failed to preform auto-sync", HttpStatus.SC_INTERNAL_SERVER_ERROR);
    }

}
