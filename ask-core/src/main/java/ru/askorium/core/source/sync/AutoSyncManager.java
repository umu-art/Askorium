package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.core.task.AsyncTaskExecutor;
import org.springframework.stereotype.Service;
import ru.askorium.api.client.SourceApi;
import ru.askorium.api.model.SourceSyncRequest;
import ru.askorium.core.exception.AutoSyncFailedException;
import ru.askorium.core.source.domain.SourceSyncPolicyEntity;
import ru.askorium.core.source.jpa.SourceJpa;

import java.time.OffsetDateTime;
import java.util.concurrent.atomic.AtomicBoolean;

@Slf4j
@Service
@RequiredArgsConstructor
public class AutoSyncManager {

    private final SourceApi selfSourceApi;
    private final SourceJpa sourceJpa;
    private final AsyncTaskExecutor asyncTaskExecutor;

    public void autoSync() {
        log.info("Starting auto-sync process");

        var wasFailed = new AtomicBoolean(false);

        var sourcesToSync = sourceJpa.findAll()
                .stream()
                .filter(s -> isNeedAutoSync(s.getSyncPolicy()))
                .toList();

        var futures = sourcesToSync.stream()
                .map(source -> asyncTaskExecutor.submit(() -> {
                    try {
                        var syncRequest = new SourceSyncRequest();
                        syncRequest.setSourceId(source.getId());
                        selfSourceApi.syncSource(syncRequest);
                    } catch (Exception e) {
                        log.error("Failed to auto-sync source with id {}", source.getId(), e);
                        wasFailed.set(true);
                    }
                }))
                .toList();

        for (var future : futures) {
            try {
                future.get();
            } catch (Exception e) {
                log.error("Error while waiting for auto-sync task to complete", e);
            }
        }

        if (wasFailed.get()) {
            throw new AutoSyncFailedException();
        }
    }

    private boolean isNeedAutoSync(SourceSyncPolicyEntity syncPolicy) {
        if (syncPolicy == null) {
            return false;
        }

        if (!syncPolicy.isEnabled()) {
            return false;
        }

        var lastSync = syncPolicy.getLastSyncedAt();
        if (lastSync == null) {
            return true;
        }

        var nextSyncTime = lastSync.plusMinutes(syncPolicy.getIntervalMinutes());

        return nextSyncTime.isBefore(OffsetDateTime.now());
    }

}
