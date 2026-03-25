package ru.askorium.core.source.sync;

import lombok.AccessLevel;
import lombok.NoArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import ru.askorium.api.model.ContentBlock;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.source.config.ScrapperTasksProperties;

import java.util.ArrayList;
import java.util.List;

@Slf4j
@NoArgsConstructor(access = AccessLevel.PRIVATE)
public class BlockLenUtils {
    public static void mergeShortBlocks(ScrappedPage page, ScrapperTasksProperties scrapperTasksProperties) {
        var blocks = page.getBlocks();
        if (blocks.size() <= 1) return;

        var minLength = scrapperTasksProperties.getMinBlockLength();
        var maxLength = scrapperTasksProperties.getMaxBlockLength();

        var allInRange = blocks.stream()
                .map(ContentBlock::getText)
                .filter(StringUtils::isNotBlank)
                .allMatch(t -> t.length() >= minLength && t.length() <= maxLength);

        if (allInRange) {
            return;
        }

        log.warn("Page '{}' has blocks outside [{}, {}] range, normalizing",
                page.getUrl(), minLength, maxLength);

        // Split oversized blocks
        var splitBlocks = new ArrayList<ContentBlock>();
        for (var block : blocks) {
            var text = block.getText();
            if (text.length() <= maxLength) {
                splitBlocks.add(block);
                continue;
            }

            int start = 0;
            while (start < text.length()) {
                int end = Math.min(start + maxLength, text.length());
                var chunk = new ContentBlock(block.getType(), text.substring(start, end));
                chunk.setHtmlId(start == 0 ? block.getHtmlId() : null);
                chunk.setHeadingLevel(block.getHeadingLevel());
                splitBlocks.add(chunk);
                start = end;
            }
        }

        // Merge short blocks
        var result = new ArrayList<>(List.of(splitBlocks.getFirst()));
        for (int i = 1; i < splitBlocks.size(); i++) {
            var block = splitBlocks.get(i);
            var text = block.getText();
            if (text != null && text.length() < minLength) {
                var prev = result.getLast();
                prev.setText(prev.getText() + " " + text);
            } else {
                result.add(block);
            }
        }

        page.setBlocks(result);
    }
}
