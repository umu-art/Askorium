package ru.askorium.core.source;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.SourceDto;
import ru.askorium.api.model.SourceSyncRequest;
import ru.askorium.api.server.SourceApi;
import ru.askorium.core.source.domain.SourceEntity;
import ru.askorium.core.source.jpa.SourceJpa;
import ru.askorium.core.common.UrlUtils;
import ru.askorium.core.source.mapper.SourceMapper;
import ru.askorium.core.source.sync.AutoSyncManager;
import ru.askorium.core.source.sync.IndexSyncService;
import ru.askorium.core.source.sync.SyncDispatcher;

import java.util.List;
import java.util.UUID;

import static java.util.Objects.nonNull;

@Controller
@RequiredArgsConstructor
public class SourceController implements SourceApi {

    private final SyncDispatcher syncDispatcher;
    private final AutoSyncManager autoSyncManager;
    private final IndexSyncService indexSyncService;
    private final SourceJpa sourceJpa;
    private final SourceMapper sourceMapper;

    @Override
    public ResponseEntity<List<SourceDto>> listSources() {
        var sources = sourceJpa.findAll().stream()
                .map(sourceMapper::toDto)
                .toList();

        return ResponseEntity.ok(sources);
    }

    @Override
    @Transactional(transactionManager = "sourcesTransactionManager")
    public ResponseEntity<SourceDto> upsertSource(SourceDto source) {
        SourceEntity entity;

        if (nonNull(source.getId())) {
            entity = sourceJpa.findById(source.getId())
                    .orElseThrow(() -> new RuntimeException("Source not found with id: " + source.getId()));
        } else {
            entity = sourceJpa.findBySourceUrl(source.getSourceUrl())
                    .orElse(new SourceEntity());
        }

        sourceMapper.updateEntityFromDto(entity, source);
        entity.setSourceUrl(UrlUtils.normalizeUrl(entity.getSourceUrl()));
        entity.getSyncPolicy().setSource(entity);
        entity = sourceJpa.save(entity);

        return ResponseEntity.ok(sourceMapper.toDto(entity));
    }

    @Override
    @Transactional(transactionManager = "sourcesTransactionManager")
    public ResponseEntity<Void> deleteSource(UUID sourceId) {
        sourceJpa.deleteById(sourceId);
        return ResponseEntity.ok().build();
    }

    @Override
    public ResponseEntity<Void> autoSyncSource() {
        autoSyncManager.autoSync();
        return ResponseEntity.ok().build();
    }

    @Override
    public ResponseEntity<Void> cleanupIndexes() {
        indexSyncService.cleanupStaleIndexEntries();
        return ResponseEntity.ok().build();
    }

    @Override
    public ResponseEntity<Void> syncSource(SourceSyncRequest sourceSyncRequest) {
        syncDispatcher.sync(sourceSyncRequest);
        return ResponseEntity.ok().build();
    }
}
