package ru.askorium.core.feedback.mapper;

import org.mapstruct.Mapper;
import ru.askorium.api.server.model.FeedbackDto;
import ru.askorium.core.common.ToBaseEntity;
import ru.askorium.core.feedback.domain.FeedbackEntity;

@Mapper(componentModel = "spring")
public interface FeedbackMapper {

    @ToBaseEntity
    FeedbackEntity toEntity(FeedbackDto feedbackDto);

}
