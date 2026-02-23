package ru.askorium.core.source.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.MappingTarget;
import ru.askorium.api.model.SourceAutoSyncPolicy;
import ru.askorium.api.model.SourceDto;
import ru.askorium.core.common.ToBaseEntity;
import ru.askorium.core.source.domain.SourceEntity;
import ru.askorium.core.source.domain.SourceSyncPolicyEntity;

@Mapper(componentModel = "spring")
public interface SourceMapper {

    @ToBaseEntity
    void updateEntityFromDto(@MappingTarget SourceEntity entity, SourceDto dto);

    SourceDto toDto(SourceEntity entity);

    SourceAutoSyncPolicy toSyncPolicyDto(SourceSyncPolicyEntity entity);

}
