package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.io.Resource;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.apache.commons.lang3.StringUtils;
import org.springframework.util.CollectionUtils;
import ru.askorium.api.model.SearchMode;
import ru.askorium.core.search.domain.PageBlockEntity;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.jpa.PageBlockJpa;
import ru.askorium.core.search.jpa.PageDocumentJpa;
import ru.askorium.core.search.workflow.activities.AnswerActivity;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashSet;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class AnswerActivityImpl extends AbstractQueryActivity implements AnswerActivity {

    private final ChatClient chatClient;
    private final PageBlockJpa pageBlockJpa;
    private final PageDocumentJpa pageDocumentJpa;

    @Value("classpath:prompts/answer.st")
    private Resource promptResource;

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

        var seen = new LinkedHashSet<String>();
        var contextParts = new ArrayList<String>();
        query.getSources().stream()
                .filter(s -> s.getScoreFinal() != null)
                .sorted(Comparator.comparing(QuerySourceEntity::getScoreFinal).reversed())
                .forEach(source -> {
                    var key = source.getIndexKey();
                    if (key.startsWith("block:")) {
                        var blockId = UUID.fromString(key.substring(6));
                        pageBlockJpa.findById(blockId).ifPresent(block -> {
                            var pageId = block.getPage().getId().toString();
                            if (seen.add(pageId)) {
                                var pageText = pageBlockJpa.findAllByPage_Id(block.getPage().getId()).stream()
                                        .map(PageBlockEntity::getText)
                                        .collect(Collectors.joining("\n"));
                                contextParts.add(pageText);
                            }
                        });
                    } else if (key.startsWith("document:")) {
                        if (seen.add(key)) {
                            var documentId = UUID.fromString(key.substring(9));
                            pageDocumentJpa.findById(documentId)
                                    .ifPresent(doc -> contextParts.add(doc.getExtractedText()));
                        }
                    }
                });

        var context = String.join("\n\n", contextParts);

        var answer = chatClient.prompt()
                .user(u -> u.text(promptResource)
                        .param("question", query.getNormalizedQuery())
                        .param("context", context))
                .call()
                .content();

        query.setAnswer(answer);
    }
}
