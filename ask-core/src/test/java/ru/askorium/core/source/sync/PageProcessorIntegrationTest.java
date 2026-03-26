package ru.askorium.core.source.sync;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.api.model.ContentBlock;
import ru.askorium.api.model.ContentBlockType;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.AbstractIntegrationTest;
import ru.askorium.core.common.BaseEntity;
import ru.askorium.core.index.IndexService;
import ru.askorium.core.source.domain.PageBlockEntity;
import ru.askorium.core.source.domain.SourceEntity;
import ru.askorium.core.source.jpa.PageJpa;
import ru.askorium.core.source.jpa.SourceJpa;
import ru.askorium.core.text_processing.TextProcessingService;

import java.util.List;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.lenient;

class PageProcessorIntegrationTest extends AbstractIntegrationTest {

    @MockitoBean
    TextProcessingService textProcessingService;
    @MockitoBean
    IndexService indexService;

    @Autowired
    PageProcessor pageProcessor;
    @Autowired
    PageJpa pageJpa;
    @Autowired
    SourceJpa sourceJpa;

    private UUID sourceId;

    @BeforeEach
    void setUp() {
        lenient().when(textProcessingService.normalizeText(any()))
                .thenAnswer(inv -> inv.getArgument(0));

        var source = new SourceEntity();
        source.setSourceUrl("https://example.com");
        sourceId = sourceJpa.save(source).getId();
    }

    @AfterEach
    void tearDown() {
        sourceJpa.deleteAll();
    }

    @Test
    @Transactional(transactionManager = "sourcesTransactionManager")
    void newPage_savedToDatabase() {
        var result = pageProcessor.processPage(
                buildPage("https://example.com/new",
                        block("h1", ContentBlockType.HEADING, "Hello", 1),
                        block("p1", ContentBlockType.PARAGRAPH, "World", null)),
                sourceId, false);

        assertThat(result).isNotNull();
        assertThat(result.getId()).isNotNull();

        var fromDb = pageJpa.findByUrl(result.getUrl());
        assertThat(fromDb).isPresent();
        assertThat(fromDb.get().getBlocks()).hasSize(1);
        assertThat(fromDb.get().getBlocks())
                .extracting(PageBlockEntity::getText)
                .containsExactlyInAnyOrder("Hello World");
    }

    @Test
    void rescanSameBlocks_preservesExistingBlockIds() {
        var first = pageProcessor.processPage(
                buildPage("https://example.com/same",
                        block("h1", ContentBlockType.HEADING, "Title", 1),
                        block("p1", ContentBlockType.PARAGRAPH, "Body", null)),
                sourceId, true);
        assertThat(first).isNotNull();
        var originalIds = first.getBlocks().stream().map(BaseEntity::getId).toList();

        var second = pageProcessor.processPage(
                buildPage("https://example.com/same",
                        block("h1", ContentBlockType.HEADING, "Title", 1),
                        block("p1", ContentBlockType.PARAGRAPH, "Body", null)),
                sourceId, true);

        assertThat(second).isNotNull();
        assertThat(second.getBlocks())
                .extracting(BaseEntity::getId)
                .containsExactlyInAnyOrderElementsOf(originalIds);
    }

    @Test
    @Transactional(transactionManager = "sourcesTransactionManager")
    void rescanChangedBlocks_updatesBlocksInDb() {
        var url = "https://example.com/changed";
        pageProcessor.processPage(
                buildPage(url, block("h1", ContentBlockType.HEADING, "Old Title", 1)),
                sourceId, true);

        var result = pageProcessor.processPage(
                buildPage(url, block("h1", ContentBlockType.HEADING, "New Title", 1)),
                sourceId, true);

        assertThat(result).isNotNull();
        assertThat(result.getBlocks()).hasSize(1);
        assertThat(result.getBlocks().getFirst().getText()).isEqualTo("New Title");

        var fromDb = pageJpa.findByUrl(url);
        assertThat(fromDb).isPresent();
        assertThat(fromDb.get().getBlocks()).hasSize(1);
        assertThat(fromDb.get().getBlocks().getFirst().getText()).isEqualTo("New Title");
    }

    @Test
    void unchangedPage_notForced_returnsNull() {
        var url = "https://example.com/unchanged";
        var first = pageProcessor.processPage(
                buildPage(url, block("h1", ContentBlockType.HEADING, "Static", 1)),
                sourceId, false);
        assertThat(first).isNotNull();

        var second = pageProcessor.processPage(
                buildPage(url, block("h1", ContentBlockType.HEADING, "Static", 1)),
                sourceId, false);

        assertThat(second).isNull();
    }

    @Test
    @Transactional(transactionManager = "sourcesTransactionManager")
    void changedPage_removeBlocksInDb(){
        var url = "https://example.com/remove";
        pageProcessor.processPage(
                buildPage(url,
                        block("h1", ContentBlockType.HEADING, "Title", 1),
                        block("p1", ContentBlockType.PARAGRAPH, "Body", null)),
                sourceId, true);

        var result = pageProcessor.processPage(
                buildPage(url, block("h1", ContentBlockType.HEADING, "Title", 1)),
                sourceId, true);

        assertThat(result).isNotNull();
        assertThat(result.getBlocks()).hasSize(1);
        assertThat(result.getBlocks().getFirst().getText()).isEqualTo("Title");

        var fromDb = pageJpa.findByUrl(url);
        assertThat(fromDb).isPresent();
        assertThat(fromDb.get().getBlocks()).hasSize(1);
        assertThat(fromDb.get().getBlocks().getFirst().getText()).isEqualTo("Title");
    }

    private ScrappedPage buildPage(String url, ContentBlock... blocks) {
        var page = new ScrappedPage();
        page.setUrl(url);
        page.setTitle("Test Page");
        page.setBlocks(List.of(blocks));
        return page;
    }

    private ContentBlock block(String htmlId, ContentBlockType type, String text, Integer headingLevel) {
        var b = new ContentBlock();
        b.setHtmlId(htmlId);
        b.setType(type);
        b.setText(text);
        b.setHeadingLevel(headingLevel);
        return b;
    }
}
