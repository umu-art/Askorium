package ru.askorium.core.source.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.MappingTarget;
import ru.askorium.api.client.model.ContentBlock;
import ru.askorium.api.client.model.Document;
import ru.askorium.api.client.model.Link;
import ru.askorium.api.client.model.ScrappedPage;
import ru.askorium.core.common.ToBaseEntity;
import ru.askorium.core.source.domain.PageBlockEntity;
import ru.askorium.core.source.domain.PageDocumentEntity;
import ru.askorium.core.source.domain.PageEntity;
import ru.askorium.core.source.domain.PageLinkEntity;

@Mapper(componentModel = "spring")
public interface PageMapper {

    @ToBaseEntity
    @Mapping(target = "sourceId", ignore = true)
    @Mapping(target = "contentHash", ignore = true)
    @Mapping(target = "blocks", ignore = true)
    @Mapping(target = "links", ignore = true)
    @Mapping(target = "documents", ignore = true)
    void updatePage(@MappingTarget PageEntity entity, ScrappedPage scrapped);

    @ToBaseEntity
    @Mapping(target = "page", ignore = true)
    PageBlockEntity toBlockEntity(ContentBlock block);

    @ToBaseEntity
    @Mapping(target = "page", ignore = true)
    @Mapping(target = "blockId", ignore = true)
    PageLinkEntity toLinkEntity(Link link);

    @ToBaseEntity
    @Mapping(target = "page", ignore = true)
    PageDocumentEntity toDocumentEntity(Document doc);

}
