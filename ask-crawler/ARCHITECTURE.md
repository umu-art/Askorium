# Crawler Service

Центральный оркестратор системы веб-краулинга. Принимает задачи на обход домена, управляет очередью URL, координирует работу Scraper и публикует результаты.

## Архитектура системы

```
External System
      │  CrawlTaskRequest
      ▼
  [crawler.tasks]
      │
   Crawler ──[render.requests]──► Renderer ──[scraper.requests]──► Scraper
      ▲                                                                 │
      └─────────────────────[scraper.results]──────────────────────────┘
      │
  [crawler.events]
      │  CrawlEvent
      ▼
External System
```

## Структура проекта

```
crawler/
├── main.go
└── src/
    ├── domain/
    │   ├── frontier/        # interface Frontier, публичные типы
    │   │   ├── lib/         # внутренние хелперы (ключи Redis и т.п.)
    │   │   ├── impl/        # RedisFrontier
    │   │   └── test/
    │   ├── filter/          # interface URLFilter, публичные типы
    │   │   ├── lib/         # внутренние хелперы
    │   │   ├── impl/        # ChainFilter, DomainFilter, DepthFilter, PatternFilter
    │   │   └── test/
    │   ├── robots/          # interface RobotsChecker, публичные типы
    │   │   ├── lib/         # внутренние структуры парсера
    │   │   ├── impl/
    │   │   └── test/
    │   └── job/             # interface JobStore, типы Job, JobStatus
    │       ├── lib/         # внутренние хелперы (не нужны снаружи)
    │       ├── impl/        # RedisJobStore
    │       └── test/
    ├── app/
    │   ├── crawler.go       # CrawlerService — приём задач, запуск воркеров
    │   └── worker.go        # Worker — цикл: frontier → render.requests → scraper.results
    ├── infra/
    │   ├── redis/           # Клиент, хелперы
    │   ├── postgres/        # sqlc-генерированный код, репозитории
    │   └── amqp/
    │       ├── consumer/    # Получение CrawlTaskRequest от внешней системы
    │       └── publisher/   # Публикация событий во внешнюю систему
    └── config/              # Конфиг из env, структуры настроек
```

## Слои и зависимости

| Слой | Зависит от | Не знает о |
|------|-----------|------------|
| `domain` | — | infra, app |
| `app` | `domain/lib` (интерфейсы) | impl-деталях |
| `infra` | `domain/lib` | `app` |
| `main.go` | всех | — |

Зависимости прокидываются через конструкторы. `main.go` — единственное место, где собираются impl.

## Основной поток

1. `consumer` получает `CrawlTaskRequest`, передаёт в `CrawlerService`
2. `CrawlerService` инициализирует `Job`, загружает robots.txt, формирует seed URLs, запускает воркеры
3. Каждый `Worker` берёт URL из `Frontier` → публикует в `render.requests`
4. Renderer рендерит страницу → публикует в `scraper.requests`
5. Scraper парсит HTML → публикует `ScrapeResponse` в `scraper.results`
6. `Worker` получает результат, обновляет `Job`, фильтрует новые ссылки, добавляет в `Frontier`
7. `publisher` отправляет `PageScrapedEvent` внешней системе
8. Когда frontier пуст или достигнут лимит — `TaskCompletedEvent`

## MVP scope

- [ ] Frontier + visited set (Redis)
- [ ] Фильтрация URL (домен, глубина, паттерны)
- [ ] robots.txt
- [ ] Concurrency control
- [ ] Retry зависших in-flight URL (TTL в Redis)
- [ ] Структурированные логи
- [ ] Метрики Prometheus
- [ ] Graceful shutdown
- [ ] Rate limiting *(пока фиксированная задержка)*
- [ ] sitemap
- [ ] Восстановление после падения
- [ ] OpenTelemetry трейсинг
