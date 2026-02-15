package ru.askorium.core.source;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.server.SourceApi;
import ru.askorium.api.server.model.SourceDto;
import ru.askorium.api.server.model.SourceSyncRequest;
import ru.askorium.core.source.domain.SourceEntity;
import ru.askorium.core.source.jpa.SourceJpa;
import ru.askorium.core.source.mapper.SourceMapper;
import ru.askorium.core.source.sync.AutoSyncManager;
import ru.askorium.core.source.sync.SourceSyncService;

import java.util.List;
import java.util.UUID;

import static java.util.Objects.nonNull;

@Controller
@RequiredArgsConstructor
public class SourceController implements SourceApi {

    private final SourceSyncService sourceSyncService;
    private final AutoSyncManager autoSyncManager;
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
    @Transactional
    public ResponseEntity<SourceDto> upsertSource(SourceDto source) {
        SourceEntity entity;

        if (nonNull(source.getId())) {
            entity = sourceJpa.findById(source.getId())
                    .orElseThrow(() -> new RuntimeException("Source not found with id: " + source.getId()));
        } else {
            entity = new SourceEntity();
        }

        sourceMapper.updateEntityFromDto(entity, source);
        entity = sourceJpa.save(entity);

        return ResponseEntity.ok(sourceMapper.toDto(entity));
    }

    @Override
    @Transactional
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
    public ResponseEntity<Void> syncSource(SourceSyncRequest sourceSyncRequest) {
        sourceSyncService.sync(sourceSyncRequest);
        return ResponseEntity.ok().build();
    }
}
