package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.apache.commons.lang3.StringUtils;
import org.springframework.util.CollectionUtils;
import ru.askorium.api.model.RerankBlock;
import ru.askorium.api.model.RerankResult;
import ru.askorium.core.ask_encoder_api.AskEncoderService;
import ru.askorium.core.search.config.SearchProperties;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.workflow.activities.RerankActivity;

import java.util.Comparator;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
@ActivityImpl(taskQueues = "askorium-search")
@RequiredArgsConstructor
public class RerankActivityImpl extends AbstractQueryActivity implements RerankActivity {

    private final AskEncoderService askEncoderService;
    private final SearchProperties searchProperties;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void rerank(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        if (StringUtils.isBlank(query.getNormalizedQuery())) {
            throw new IllegalStateException("Normalized query is not generated");
        }

        if (CollectionUtils.isEmpty(query.getSources())) {
            throw new IllegalStateException("Sources are not retrieved or empty");
        }

        var params = searchProperties.forMode(query.getMode());

        var topSources = query.getSources().stream()
                .sorted(Comparator.comparing(QuerySourceEntity::getFusionScore).reversed())
                .limit(params.getRerankTopN())
                .toList();

        var rerankBlocks = topSources.stream()
                .map(s -> new RerankBlock().id(s.getId()).text(s.getText()))
                .toList();

        var scoreMap = askEncoderService.rerank(query.getNormalizedQuery(), rerankBlocks).stream()
                .collect(Collectors.toMap(RerankResult::getId, RerankResult::getScore));

        topSources.forEach(s -> s.setScoreFinal(scoreMap.get(s.getId())));
    }
}
