package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.io.Resource;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import ru.askorium.core.search.config.SearchProperties;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.service.ContextBuilderService;
import ru.askorium.core.search.workflow.activities.AnswerActivity;

import java.util.Comparator;
import java.util.UUID;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class AnswerActivityImpl extends AbstractQueryActivity implements AnswerActivity {

    private final ChatClient chatClient;
    private final ContextBuilderService contextBuilderService;
    private final SearchProperties searchProperties;

    @Value("classpath:prompts/answer.st")
    private Resource promptResource;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void generateAnswer(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        if (StringUtils.isBlank(query.getNormalizedQuery())) {
            throw new IllegalStateException("Normalized query is not generated");
        }

        if (CollectionUtils.isEmpty(query.getSources())) {
            throw new IllegalStateException("Sources are not retrieved");
        }

        var params = searchProperties.forMode(query.getMode());

        var answerBlocks = query.getSources().stream()
                .sorted(Comparator.comparing(QuerySourceEntity::getScoreFinal).reversed())
                .limit(params.getFinalTopN())
                .toList();

        var context = String.join("\n\n", contextBuilderService.buildContext(answerBlocks));

        var answer = chatClient.prompt()
                .user(u -> u.text(promptResource)
                        .param("question", query.getNormalizedQuery())
                        .param("context", context))
                .call()
                .content();

        query.setAnswer(answer);
    }
}
