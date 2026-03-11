# ask-core Architecture

ask-core is the Spring Boot orchestrator of the Askorium system. It is a modular monolith built with Spring Modulith, responsible for source management, content indexing, search orchestration, and answer generation.

## Module Overview

Eight Spring Modulith modules under `ru.askorium.core`:

| Module | Package | Status | Responsibility |
|---|---|---|---|
| **user** | `user` | Implemented | Cookie-based identity, Spring Security config, global error handling |
| **source** | `source` | Implemented | Source CRUD, sync scheduling, scrapper integration, indexing orchestration |
| **feedback** | `feedback` | Implemented | User feedback collection on search results |
| **index** | `index` | Implemented | Elasticsearch abstraction: BM25 text index + kNN vector index |
| **search** | `search` | Stub | Query orchestration (domain entities + JPA + controller stub, no business logic yet) |
| **textProcessing** | `text_processing` | Stub | Text normalization interface (returns text as-is, TODO) |
| **askEncoderApi** | `ask_encoder_api` | Interface only | HTTP client for ask-encoder service (embeddings, reranking) |
| **askScrapperApi** | `ask_scrapper_api` | Interface only | Client for ask-scrapper service (web scraping) |

Module boundaries are enforced by Spring Modulith + ArchUnit tests (`ModularityTests`). Modules communicate via Spring application events, not direct bean references across boundaries.

## Module: user

The `user` module owns identity and request context propagation.

### Cookie-based identity
- `AuthenticationFilter` (`OncePerRequestFilter`) runs on every request
- Reads `ask_uid` cookie (UUID). If valid and found in Postgres — updates `lastSeenAt` and `lastSeenIp`
- If cookie absent or user not found — creates a new `UserEntity` in Postgres, records `firstVisitUserAgent` and `firstVisitHeaders`, sets `Set-Cookie: ask_uid=<uuid>; HttpOnly; Secure; Path=/; Max-Age=315360000` (~10 years)
- Stores `userId` in `RequestAttributes.SCOPE_REQUEST` — retrieved via `UserUtils.getUserId()` anywhere in request scope

### Security
- `SecurityConfiguration` disables CSRF, sessions, form login, anonymous, logout, rememberMe
- Adds `AuthenticationFilter` before `UsernamePasswordAuthenticationFilter`

### Error handling
- `ExceptionControllerAdvice` — global `@ControllerAdvice` handling `AskCoreException` (custom status code), binding/conversion errors (400), `ErrorResponse` (framework status), and fallback (500)

## Module: source

The `source` module manages content sources and the full scraping → indexing pipeline.

### Source CRUD
- `SourceController` implements generated `SourceApi` interface
- Operations: list all, upsert (create or update by id), delete, trigger auto-sync, trigger manual sync
- `SourceEntity` has `sourceUrl` + `SourceSyncPolicyEntity` (enabled, intervalMinutes, lastSyncedAt)

### Sync pipeline (`SourceSyncService`)
1. Acquire distributed lock via Redisson (`source-sync:<sourceId>`, 5s wait, 30min lease)
2. Call `AskParserSender.scrapSource(url)` → `List<ScrappedPage>`
3. Normalize all texts via `TextProcessingService.normalizeText()` (blocks, link anchors, document extracted text)
4. Diff against existing pages by URL:
   - Skip unchanged pages (by `contentHash`) unless force=true
   - Update changed pages: sync blocks, links, documents collections via content-equals diffing (`ObjectCompareUtils`)
   - Delete stale pages no longer returned by scrapper
5. Update `syncPolicy.lastSyncedAt`
6. Trigger `IndexSyncService.syncIndexes()` for updated pages

### Auto-sync (`AutoSyncManager`)
- Filters sources with enabled sync policy and elapsed interval
- Runs sync for each eligible source in parallel via `AsyncTaskExecutor`
- Calls self via generated `SourceApi` client (HTTP loopback)
- Throws `AutoSyncFailedException` if any sync fails

### Index sync (`IndexSyncService`)
1. Extracts texts from `PageBlockEntity` (prefix `page:<id>`) and `PageDocumentEntity` (prefix `document:<id>`)
2. Generates embeddings via `AskEncoderService.generateEmbeddings()`
3. Saves to both Elasticsearch indices: `IndexService.saveTexts()` + `IndexService.saveVectors()`

### Domain entities
- `PageEntity` — url, title, previewUrl, iconUrl, description, language, contentHash. Has `@OneToMany` to blocks, links, documents (cascade ALL, orphanRemoval)
- `PageBlockEntity` — htmlId, type (`ContentBlockType`), headingLevel, text
- `PageLinkEntity` — href, type (`LinkType`), anchorText, snippet, position, blockId
- `PageDocumentEntity` — url, mimeType, sizeBytes, extractedText, description, descriptionSource (`DocumentDescriptionSourceType`)

## Module: feedback

- `FeedbackController` implements generated `FeedbackApi` — single endpoint `submitFeedback`
- Maps DTO via `FeedbackMapper` (MapStruct), sets `userId` from request context, saves to Postgres
- `FeedbackEntity` — queryId, userId, rating (int), text

## Module: index

Elasticsearch abstraction providing text (BM25) and vector (kNN) search.

### `ElasticsearchIndexService`
- On startup (`@PostConstruct`): creates text and vector indices if they don't exist
- **Text index**: `key` (keyword) + `text` (text, standard analyzer)
- **Vector index**: `key` (keyword) + `values` (dense_vector, cosine similarity, HNSW with configurable M and ef_construction)
- `saveTexts()` / `saveVectors()` — bulk indexing with error reporting
- `searchBM25(query, size)` — match query on `text` field
- `searchKnn(vector, size)` — kNN search with `numCandidates = max(size * 10, 100)`

### Configuration (`IndexProperties`)
Bound to `askorium.index.*`:
- `textIndexName` = `askorium-texts`
- `vectorIndexName` = `askorium-vectors`
- `vectorDimension` = 1024
- `hnswM` = 16
- `hnswEfConstruction` = 100

## Module: search (stub)

Domain layer is defined but no business logic is implemented yet.

- `SearchController` implements `SearchApi` — both `createSearchQuery` and `getSearchQueryResult` return null
- `SearchQueryEntity` — userId, sourceId, status, query, mode, answer, error, finishedAt
- `SearchResultItemEntity` — queryId, rank, blockId, pageId, scoreSparse, scoreDense, scoreFinal, rerankScore
- `SearchQueryJpa`, `SearchResultItemJpa` — repositories defined
- `SearchModuleConfig` — isolated datasource configured
- `search/serivce/` directory exists but is empty

## Storage Split

Strict separation — Elasticsearch holds **only** search indices, everything else lives in PostgreSQL.

### PostgreSQL (primary domain store)
- Users, Sources (+ sync policies), Pages, Blocks, Links, Documents
- Search queries, result items, feedback
- Managed via Flyway migrations (`db/migration/`)

### Elasticsearch (search indices only)
- **Text index** (`askorium-texts`): `key` + `text` for BM25
- **Vector index** (`askorium-vectors`): `key` + `values` (1024-dim, HNSW, cosine) for kNN

### Redis
- Distributed locking for source sync (via Redisson)

## Key Entities

### Users

| Entity | Table | Fields | Description |
|---|---|---|---|
| **UserEntity** | `users` | lastSeenAt, lastSeenIp, firstVisitUserAgent, firstVisitHeaders | User identified by `ask_uid` cookie |

### Sources & Content (PostgreSQL)

| Entity | Table | Fields | Description |
|---|---|---|---|
| **SourceEntity** | `sources` | sourceUrl, syncPolicy (1:1) | Website to index and search within |
| **SourceSyncPolicyEntity** | `source_sync_policies` | source (1:1), enabled, intervalMinutes (default 720), lastSyncedAt | Auto-sync configuration |
| **PageEntity** | `pages` | sourceId, url, title, previewUrl, iconUrl, description, language, contentHash, blocks (1:N), links (1:N), documents (1:N) | Scraped web page |
| **PageBlockEntity** | `page_blocks` | page, htmlId, type (ContentBlockType), headingLevel, text | Content block within page |
| **PageLinkEntity** | `page_links` | page, blockId, href, type (LinkType), anchorText, snippet, position | Link within page |
| **PageDocumentEntity** | `page_documents` | page, url, mimeType, sizeBytes, extractedText, description, descriptionSource | Attached document |

### Search Lifecycle (PostgreSQL)

| Entity | Table | Fields | Description |
|---|---|---|---|
| **SearchQueryEntity** | `search_queries` | userId, sourceId, status, query, mode, answer, error, finishedAt | Search request (entities defined, logic not implemented) |
| **SearchResultItemEntity** | `search_result_items` | queryId, rank, blockId, pageId, scoreSparse, scoreDense, scoreFinal, rerankScore | Search result (entities defined, logic not implemented) |
| **FeedbackEntity** | `feedbacks` | queryId, userId, rating (int), text | User feedback on search quality |

All entities extend `BaseEntity` (UUID id, created, updated timestamps).

## Infrastructure Integration

### PostgreSQL
- Primary domain store for all entities
- Flyway migrations in `src/main/resources/db/migration/`
- Each module has isolated HikariCP datasource, EntityManager, and TransactionManager

### Elasticsearch
- Text index for BM25 search
- Vector index for kNN search (HNSW, cosine, 1024-dim)
- Auto-created on startup via `ElasticsearchIndexService.@PostConstruct`

### Redis (Redisson)
- Distributed locking for source sync (`source-sync:<sourceId>`)

### RabbitMQ
- Configured: `RabbitConverterConfiguration` (JSON message converter), `RabbitRetryBackOffConfiguration` (retry policy)
- Not yet wired to scrapper service (AskParserSender is interface only)

## Cross-Cutting Concerns

- **Async**: `AsyncConfiguration` provides `AsyncTaskExecutor` (used by `AutoSyncManager`)
- **Retry**: `RetryableConfiguration` enables Spring Retry
- **Module events**: Spring Modulith application events for inter-module communication
- **Observability**: Spring Boot Actuator + Prometheus metrics export (dependencies configured)
- **Validation**: `spring-boot-starter-validation` on API request DTOs
