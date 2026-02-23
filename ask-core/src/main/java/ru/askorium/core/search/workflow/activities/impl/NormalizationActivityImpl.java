package ru.askorium.core.search.workflow.activities.impl;

import io.temporal.spring.boot.ActivityImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.apache.commons.lang3.StringUtils;
import ru.askorium.core.search.domain.QueryEntity;
import ru.askorium.core.search.workflow.activities.NormalizationActivity;
import ru.askorium.core.text_processing.TextProcessingService;

import java.util.UUID;

@Component
@ActivityImpl
@RequiredArgsConstructor
public class NormalizationActivityImpl extends AbstractQueryActivity implements NormalizationActivity {

    private final TextProcessingService textProcessingService;

    @Override
    @Transactional(transactionManager = "searchTransactionManager")
    public void normalize(UUID queryId) {
        executeWithQuery(queryId);
    }

    @Override
    protected void doWithQuery(QueryEntity query) {
        if (StringUtils.isBlank(query.getQuery())) {
            throw new IllegalStateException("Query is blank");
        }

        var text = query.getQuery();
        var normalizedText = textProcessingService.normalizeText(text);
        query.setNormalizedQuery(normalizedText);
    }
}
