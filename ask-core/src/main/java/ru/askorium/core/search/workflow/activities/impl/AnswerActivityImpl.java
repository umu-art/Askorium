package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.apache.commons.lang3.StringUtils;
import org.springframework.util.CollectionUtils;
import ru.askorium.api.model.SearchMode;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.workflow.activities.AnswerActivity;

import java.util.Comparator;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class AnswerActivityImpl extends AbstractQueryActivity implements AnswerActivity {

    private final ChatClient chatClient;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void generateAnswer(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        if (!SearchMode.DEEP.equals(query.getMode())) {
            return;
        }

        if (StringUtils.isBlank(query.getNormalizedQuery())) {
            throw new IllegalStateException("Normalized query is not generated");
        }

        if (CollectionUtils.isEmpty(query.getSources())) {
            throw new IllegalStateException("Sources are not retrieved");
        }

        var context = query.getSources().stream()
                .filter(s -> s.getScoreFinal() != null)
                .sorted(Comparator.comparing(QuerySourceEntity::getScoreFinal).reversed())
                .map(QuerySourceEntity::getText)
                .collect(Collectors.joining("\n\n"));

        var answer = chatClient.prompt()
                .user(u -> u.text("Question: " + query.getNormalizedQuery() + "\n\nContext:\n" + context))
                .call()
                .content();

        query.setAnswer(answer);
    }
}
