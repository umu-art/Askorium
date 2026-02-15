package ru.askorium.core.feedback;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.server.FeedbackApi;
import ru.askorium.api.server.model.FeedbackDto;
import ru.askorium.core.feedback.jpa.FeedbackJpa;
import ru.askorium.core.feedback.mapper.FeedbackMapper;

import static ru.askorium.core.common.UserUtils.getUserId;

@Controller
@RequiredArgsConstructor
public class FeedbackController implements FeedbackApi {

    private final FeedbackJpa feedbackJpa;
    private final FeedbackMapper feedbackMapper;

    @Override
    @Transactional
    public ResponseEntity<Void> submitFeedback(FeedbackDto feedback) {
        var entity = feedbackMapper.toEntity(feedback);
        entity.setUserId(getUserId());
        feedbackJpa.save(entity);

        return ResponseEntity.ok().build();
    }
}
