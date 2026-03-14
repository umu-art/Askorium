package ru.askorium.core.source.sync;

import jakarta.persistence.EntityManager;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.test.util.ReflectionTestUtils;
import ru.askorium.api.model.ContentBlock;
import ru.askorium.api.model.ContentBlockType;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.source.domain.PageBlockEntity;
import ru.askorium.core.source.domain.PageEntity;
import ru.askorium.core.source.jpa.PageJpa;
import ru.askorium.core.source.mapper.PageMapper;
import ru.askorium.core.text_processing.TextProcessingService;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class PageProcessorTest {

    @Mock PageJpa pageJpa;
    @Mock TextProcessingService textProcessingService;
    @Mock PageMapper pageMapper;
    @Mock EntityManager entityManager;

    @InjectMocks PageProcessor pageProcessor;

    private static final UUID SOURCE_ID = UUID.randomUUID();

    @BeforeEach
    void setUp() {
        ReflectionTestUtils.setField(pageProcessor, "entityManager", entityManager);
        lenient().when(textProcessingService.normalizeText(any()))
                .thenAnswer(inv -> inv.getArgument(0));
        lenient().when(pageMapper.toBlockEntity(any())).thenAnswer(inv -> {
            ContentBlock cb = inv.getArgument(0);
            return createBlockEntity(cb.getHtmlId(), cb.getType(), cb.getText(), cb.getHeadingLevel());
        });
    }

    @Test
    void rescanSameBlocks_preservesExistingAndDoesNotFlush() {
        var existingPage = createExistingPage();
        var block1Id = UUID.randomUUID();
        var block2Id = UUID.randomUUID();
        addBlock(existingPage, block1Id, "h1", ContentBlockType.HEADING, "Title", 1);
        addBlock(existingPage, block2Id, "p1", ContentBlockType.PARAGRAPH, "Text", null);

        when(pageJpa.findByUrl("https://example.com/page")).thenReturn(Optional.of(existingPage));
        when(pageJpa.save(any())).thenAnswer(inv -> inv.getArgument(0));

        var scrappedPage = createScrappedPage(
                createContentBlock("h1", ContentBlockType.HEADING, "Title", 1),
                createContentBlock("p1", ContentBlockType.PARAGRAPH, "Text", null)
        );

        var result = pageProcessor.processPage(scrappedPage, SOURCE_ID, true);

        assertThat(result.getBlocks()).hasSize(2);
        assertThat(result.getBlocks())
                .extracting(PageBlockEntity::getId)
                .containsExactlyInAnyOrder(block1Id, block2Id);
        verify(entityManager, never()).flush();
    }

    @Test
    void rescanChangedBlocks_removesOldAddsNewAndFlushes() {
        var existingPage = createExistingPage();
        addBlock(existingPage, UUID.randomUUID(), "h1", ContentBlockType.HEADING, "Old Title", 1);

        when(pageJpa.findByUrl("https://example.com/page")).thenReturn(Optional.of(existingPage));
        when(pageJpa.save(any())).thenAnswer(inv -> inv.getArgument(0));

        var scrappedPage = createScrappedPage(
                createContentBlock("h1", ContentBlockType.HEADING, "New Title", 1)
        );

        var result = pageProcessor.processPage(scrappedPage, SOURCE_ID, true);

        assertThat(result.getBlocks()).hasSize(1);
        assertThat(result.getBlocks().get(0).getText()).isEqualTo("New Title");
        verify(entityManager).flush();
    }

    @Test
    void rescanPartiallyChangedBlocks_preservesMatchedReplacesRest() {
        var existingPage = createExistingPage();
        var keptId = UUID.randomUUID();
        addBlock(existingPage, keptId, "h1", ContentBlockType.HEADING, "Kept", 1);
        addBlock(existingPage, UUID.randomUUID(), "p1", ContentBlockType.PARAGRAPH, "Removed", null);

        when(pageJpa.findByUrl("https://example.com/page")).thenReturn(Optional.of(existingPage));
        when(pageJpa.save(any())).thenAnswer(inv -> inv.getArgument(0));

        var scrappedPage = createScrappedPage(
                createContentBlock("h1", ContentBlockType.HEADING, "Kept", 1),
                createContentBlock("p2", ContentBlockType.PARAGRAPH, "Added", null)
        );

        var result = pageProcessor.processPage(scrappedPage, SOURCE_ID, true);

        assertThat(result.getBlocks()).hasSize(2);
        assertThat(result.getBlocks())
                .extracting(PageBlockEntity::getText)
                .containsExactlyInAnyOrder("Kept", "Added");
        assertThat(result.getBlocks())
                .filteredOn(b -> "Kept".equals(b.getText()))
                .first()
                .extracting(PageBlockEntity::getId)
                .isEqualTo(keptId);
        verify(entityManager).flush();
    }

    @Test
    void rescanWithEmptyHtmlId_preservesBlockAndDoesNotFlush() {
        var existingPage = createExistingPage();
        var blockId = UUID.randomUUID();
        addBlock(existingPage, blockId, "", ContentBlockType.PARAGRAPH, "Some text", null);

        when(pageJpa.findByUrl("https://example.com/page")).thenReturn(Optional.of(existingPage));
        when(pageJpa.save(any())).thenAnswer(inv -> inv.getArgument(0));

        var scrappedPage = createScrappedPage(
                createContentBlock("", ContentBlockType.PARAGRAPH, "Some text", null)
        );

        var result = pageProcessor.processPage(scrappedPage, SOURCE_ID, true);

        assertThat(result.getBlocks()).hasSize(1);
        assertThat(result.getBlocks().get(0).getId()).isEqualTo(blockId);
        verify(entityManager, never()).flush();
    }

    private PageEntity createExistingPage() {
        var page = new PageEntity();
        page.setId(UUID.randomUUID());
        page.setUrl("https://example.com/page");
        page.setSourceId(SOURCE_ID);
        page.setBlocks(new ArrayList<>());
        page.setLinks(new ArrayList<>());
        page.setDocuments(new ArrayList<>());
        return page;
    }

    private void addBlock(PageEntity page, UUID id, String htmlId, ContentBlockType type, String text, Integer headingLevel) {
        var block = createBlockEntity(htmlId, type, text, headingLevel);
        block.setId(id);
        block.setPage(page);
        page.getBlocks().add(block);
    }

    private PageBlockEntity createBlockEntity(String htmlId, ContentBlockType type, String text, Integer headingLevel) {
        var entity = new PageBlockEntity();
        entity.setHtmlId(htmlId);
        entity.setType(type);
        entity.setText(text);
        entity.setHeadingLevel(headingLevel);
        return entity;
    }

    private ScrappedPage createScrappedPage(ContentBlock... blocks) {
        var page = new ScrappedPage();
        page.setUrl("https://example.com/page");
        page.setBlocks(List.of(blocks));
        return page;
    }

    private ContentBlock createContentBlock(String htmlId, ContentBlockType type, String text, Integer headingLevel) {
        var block = new ContentBlock();
        block.setHtmlId(htmlId);
        block.setType(type);
        block.setText(text);
        block.setHeadingLevel(headingLevel);
        return block;
    }
}
