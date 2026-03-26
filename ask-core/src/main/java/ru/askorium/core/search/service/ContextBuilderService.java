package ru.askorium.core.search.service;

import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Component;
import ru.askorium.core.search.config.SearchProperties;
import ru.askorium.core.search.domain.PageBlockEntity;
import ru.askorium.core.search.domain.PageDocumentEntity;
import ru.askorium.core.search.domain.QuerySourceEntity;
import ru.askorium.core.search.jpa.PageBlockJpa;
import ru.askorium.core.search.jpa.PageDocumentJpa;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
@RequiredArgsConstructor
public class ContextBuilderService {

    private final PageBlockJpa pageBlockJpa;
    private final PageDocumentJpa pageDocumentJpa;
    private final SearchProperties searchProperties;

    public List<String> buildContext(List<QuerySourceEntity> sources) {
        var seen = new LinkedHashSet<String>();
        var parts = new ArrayList<String>();
        int sourceNum = 1;

        var sorted = sources.stream()
                .filter(s -> s.getScoreFinal() != null)
                .sorted(Comparator.comparing(QuerySourceEntity::getScoreFinal).reversed())
                .toList();

        for (var source : sorted) {
            var key = source.getIndexKey();

            if (key.startsWith("block:")) {
                var blockId = UUID.fromString(key.substring(6));
                var blockOpt = pageBlockJpa.findById(blockId);
                if (blockOpt.isEmpty()) continue;

                var block = blockOpt.get();
                var pageId = block.getPage().getId().toString();
                if (!seen.add(pageId)) continue;

                var description = pageDocumentJpa.findByUrl(source.getUrl())
                        .map(PageDocumentEntity::getDescription)
                        .orElse(null);

                var content = searchProperties.isUseFullSource()
                        ? pageBlockJpa.findAllByPage_Id(block.getPage().getId()).stream()
                                .map(PageBlockEntity::getText)
                                .collect(Collectors.joining("\n"))
                        : source.getText();

                parts.add(format(sourceNum++, source.getUrl(), source.getTitle(), description, content));

            } else if (key.startsWith("document:")) {
                if (!seen.add(key)) continue;

                var documentId = UUID.fromString(key.substring(9));
                var docOpt = pageDocumentJpa.findById(documentId);
                if (docOpt.isEmpty()) continue;

                var doc = docOpt.get();
                var content = searchProperties.isUseFullSource() ? doc.getExtractedText() : source.getText();

                parts.add(format(sourceNum++, source.getUrl(), source.getTitle(), doc.getDescription(), content));
            }
        }

        return parts;
    }

    private String format(int num, String url, String title, String description, String content) {
        var sb = new StringBuilder();
        sb.append("[").append(num).append("] ").append(url).append("\n");
        sb.append("title: ").append(title).append("\n");
        if (StringUtils.isNotBlank(description)) {
            sb.append("description: ").append(description).append("\n");
        }
        sb.append("content: ").append(content);
        return sb.toString();
    }
}
