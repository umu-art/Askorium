package ru.askorium.core.source.sync;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.askorium.core.index.IndexService;
import ru.askorium.core.source.jpa.PageJpa;

import java.util.ArrayList;

@Slf4j
@Service
@RequiredArgsConstructor
public class IndexSyncService {

    private final IndexService indexService;
    private final PageJpa pageJpa;

    @Transactional(transactionManager = "sourcesTransactionManager", readOnly = true)
    public void cleanupStaleIndexEntries() {
        var validKeys = new ArrayList<String>();
        validKeys.addAll(pageJpa.findAllBlockIndexIds());
        validKeys.addAll(pageJpa.findAllDocumentIndexIds());
        log.info("Cleaning up stale index entries, {} valid keys in DB", validKeys.size());
        indexService.deleteStaleKeys(validKeys);
    }

}
