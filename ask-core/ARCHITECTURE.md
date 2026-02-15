# ask-core Architecture

ask-core is the Spring Boot orchestrator of the Askorium system. It is a modular monolith built with Spring Modulith, responsible for source management, content indexing, search orchestration, and answer generation.

## Module Overview

Six Spring Modulith modules under `ru.askorium.core`:

| Module | Responsibility |
|---|---|
| **user** | User lifecycle (cookie-based identity, CRUD), rate limiting (via Redis), request context propagation |
| **sources** | Source CRUD, sync scheduling, scrapper integration, indexing orchestration (chunking, doc2query, send to ES) |
| **search** | Query orchestration: sparse + dense retrieval, fusion, reranking, LLM answer generation |
| **textProcessing** | Text normalization: standardize words, expand abbreviations, remove stop-words, expand synonyms. Internal API for `sources` and `search` |
| **askSearchApi** | HTTP client for ask-search service (embeddings, reranking) |
| **askScrapperApi** | RabbitMQ producer/consumer for ask-scrapper service |

Module boundaries are enforced by Spring Modulith + ArchUnit tests (`ModularityTests`). Modules communicate via Spring application events, not direct bean references across boundaries.

## Module: user

The `user` module owns identity, rate limiting, and request context propagation.

### Cookie-based identity
- Issues and validates a `user_id` cookie (UUID)
- If the cookie is absent or invalid: creates a new User in Postgres, sets `Set-Cookie: user_id=<uuid>; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=...`
- All domain entities and events carry `userId`

### Rate limiting
- Enforces per-user request rate limits via Redis (sliding window)
- Falls back to IP-based limiting before the cookie is issued
- Configurable limits per endpoint / operation type

### Implementation
- HTTP layer: `OncePerRequestFilter` / `HandlerInterceptor` resolves `userId` into a `RequestContext` (`ThreadLocal` or `@RequestScope` bean), checks rate limits
- Modulith events: all events include `userId` field, propagated through pipelines

## Module: textProcessing

The `textProcessing` module provides text normalization as an internal API consumed by `sources` (indexing pipeline) and `search` (query preparation).

### Normalization steps
- Standardize words (lowercase, lemmatization)
- Expand abbreviations and domain-specific terms
- Remove stop-words
- Expand synonyms

Does **not** include chunking or doc2query — those remain in the `sources` module.

## Storage Split

Strict separation — Elasticsearch holds **only** search indices, everything else lives in PostgreSQL.

### PostgreSQL (primary domain store)
- Users, Sources, Pages, Blocks metadata, Documents, Links
- SearchQuery lifecycle (RUNNING/DONE/FAILED), results, audit
- Feedback
- Sync jobs / scheduling state

### Elasticsearch (search indices only)
- **Sparse index** (BM25): `chunk_text` + `doc_queries` + join fields (`block_id`, `source_id`, `page_id`)
- **Dense index** (kNN): `embedding_vector` + join fields (`block_id`, `source_id`, `page_id`), optionally `chunk_text` for debugging/highlighting

ES does **not** store: sources, pages, blocks as domain entities, links, documents, snippet metadata, users, queries, feedback, sync state.

### Redis
- Query result caching by key `(userId, sourceId, normalizedQuery, mode)`
- Rate limiting by `userId` (IP as fallback before cookie is issued)

## Data Flow: Scrapping Pipeline

```
addSource / syncSource (HTTP, user_id cookie)
        │
        ▼
   sources module
        │  persists Source{userId,...} in Postgres
        │  publishes ScrapperRequest {userId, sourceId, url}
        ▼
  askScrapperApi ──► RabbitMQ task queue ──► ask-scrapper
                                                  │
                                                  ▼
                                            ScrapperResult
                                             └─ ScrapedPage
                                                 ├─ blocks: ContentBlock[]
                                                 ├─ links: Link[]
                                                 └─ documents: Document[]
                                                  │
  askScrapperApi ◄── RabbitMQ result queue ◄──────┘
        │  emits event {userId, sourceId, scrapedPage}
        ▼
   sources module
        │  stores ALL content in Postgres (pages, blocks, links, documents)
        │  triggers indexing
        ▼
   Indexing Pipeline
```

The scrapper returns a `ScrapedPage` containing structured content blocks (headings, paragraphs, list items), internal/external links with positional context, and documents with extracted text (OCR/PDF).

## Data Flow: Indexing Pipeline

After receiving a `ScrapedPage`:

1. **Persist structured content to Postgres**
   - Upsert `IndexedPage`
   - Insert/update `IndexedBlock` (raw block text, pre-chunking)
   - Store links (`PageLink`) and documents (`PageDocument`) with extracted text
2. **Text normalization** — via `textProcessing` module (standardize words, expand abbreviations, remove stop-words, expand synonyms)
3. **Overlap chunking** — normalized chunks saved to `IndexedBlock` (Postgres) as source of truth
4. **doc2query enrichment** — generated queries saved to `IndexedBlock.docQueries` (Postgres)
5. **Embedding generation** — askSearchApi `POST /embeddings` (batch up to 256)
6. **Index to Elasticsearch (indices only)**
   - Sparse index: `block_id`, `source_id`, `page_id`, `chunk_text`, `doc_queries`
   - Dense index: `block_id`, `source_id`, `page_id`, `embedding_vector` (+ optionally `chunk_text`/`doc_queries`)

## Data Flow: Search Pipeline

`POST /ask/query` creates an async search query bound to `user_id`:

```
SearchCreateRequest {query, sourceId, mode} + user_id cookie
        │
        ▼
0. Auth/Identity
   - resolve userId from cookie (create if missing)
   - validate sourceId belongs to this userId (Postgres)
        │
        ▼
1. Persist SearchQuery{userId, sourceId, query, mode, status=RUNNING} in Postgres
        │
        ▼
2. Query Preparation
   ├─ Text normalization (via textProcessing module)
   └─ Expand with generated broad queries
        │
        ▼
3. Retriever (parallel)
   ├─ N from BM25 sparse search (ES → block_id list + scores)
   └─ N from dense search:
      embed query via askSearchApi POST /embeddings → ES kNN → block_id list + scores
        │
        ▼
4. Fusion
   Convex Score Combination: candidates → M
        │
        ▼
5. Reranker
   Cross-encoder (BGE-reranker-v2-m3) via askSearchApi POST /rerank
   ├─ Early Exit: skip candidates below relevance threshold mid-computation
   └─ Batch mode: multiple candidates per query in one call (up to 128)
        │
        ▼
6. Load metadata FROM Postgres by block_id/page_id
   → build top-K SourceSnippets {url, title, date, snippet}
        │
        ▼
7. RAG: LLM (Google GenAI via Spring AI) generates answer from top-K snippets
        │
        ▼
8. Persist results in Postgres:
   - SearchQuery.status=DONE, answer
   - SearchResultItem[] with ranked block/page refs
        │
        ▼
9. GET /ask/query/{queryId} polls for status and result
```

Target latency for `fast` mode: < 0.5 sec (assuming warm ES + cache + fast rerank/LLM paths).

## Key Entities

### Users

| Entity | Fields | Description |
|---|---|---|
| **User** | userId (UUID), createdAt, lastSeenAt, userAgent?, ipHash? | User identified by `user_id` cookie |

### Sources & Content (PostgreSQL)

| Entity | Fields | Description |
|---|---|---|
| **Source** | sourceId, userId, sourceUrl, syncPolicy(...), createdAt | Website to index and search within |
| **IndexedPage** | pageId, sourceId, userId, url, title, language, lastModified, hash, createdAt/updatedAt | Page metadata |
| **IndexedBlock** | blockId, pageId, sourceId, userId, chunkText, docQueries, chunkNo, offsetStart/End, createdAt | Canonical chunk storage and metadata |
| **PageLink** | linkId, pageId, sourceId, userId, fromBlockId?, url, type (internal/external), position | Links from page |
| **PageDocument** | docId, pageId, sourceId, userId, url, mime, extractedText, metaJson | Documents (PDF/OCR) and extracted text |

### Search Lifecycle (PostgreSQL)

| Entity | Fields | Description |
|---|---|---|
| **SearchQuery** | queryId, userId, sourceId, status, query, mode, answer, createdAt, finishedAt, error? | Async search request |
| **SearchResultItem** | queryId, rank, blockId, pageId, scoreSparse?, scoreDense?, scoreFinal, rerankScore? | Search results as block/page references |
| **Feedback** | feedbackId, queryId, userId, rating (1-5), text, createdAt | User feedback on search quality |

## Infrastructure Integration

### PostgreSQL
- Primary domain store: users, sources, pages, blocks, docs, links
- Search queries, results, feedback
- Sync scheduling state, retries, audit

### Elasticsearch
- Sparse index for BM25
- Dense index for kNN vectors
- Stores minimal identifiers (`block_id`, `page_id`, `source_id`) for reverse lookup of metadata in Postgres

### RabbitMQ
- Task/result queues for scrapper, DLQ with retry counter
- Messages carry `userId` + `sourceId` for validation and routing

### Redis
- Cache by key `(userId, sourceId, normalizedQuery, mode)`
- Rate limiting by `userId` via `user` module (IP fallback before cookie is issued)

### Google GenAI (via Spring AI)
- RAG answer generation from top-K source snippets

## Cross-Cutting Concerns

- **Resilience**: Spring Retry for ask-search HTTP calls; RabbitMQ DLQ for scrapper failures
- **Module events**: Spring Modulith application events for inter-module communication (e.g., source synced triggers reindex)
- **Observability**: Spring Boot Actuator + Prometheus metrics export
- **Validation**: `spring-boot-starter-validation` on API request DTOs
- **Deployment**: Kubernetes with 2 replicas per module, load balancing via K8s services
