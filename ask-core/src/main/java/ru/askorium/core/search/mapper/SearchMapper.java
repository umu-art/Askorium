package ru.askorium.core.search.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import ru.askorium.api.model.SearchCreateRequest;
import ru.askorium.api.model.SearchGetResponse;
import ru.askorium.core.common.ToBaseEntity;
import ru.askorium.core.search.domain.QueryEntity;

@Mapper(componentModel = "spring")
public interface SearchMapper {

    @Mapping(target = "queryId", source = "id")
    SearchGetResponse toGetResponse(QueryEntity entity);

    @ToBaseEntity
    QueryEntity toEntity(SearchCreateRequest request);
}
