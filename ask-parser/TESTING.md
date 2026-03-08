# Ask-Parser — Тестирование

## Запуск

```bash
cd ask-parser/src

go test ./...                  # все тесты
go test ./... -v               # с подробным выводом
go test ./... -cover           # с покрытием
```

## Покрытие тестами

### handler (`internal/handler`)

Мокаются: `parser.Parser`, `amqp.MessagePublisher`.

| Тест | Сценарий | Ожидаемый результат |
|---|---|---|
| `InvalidJSON` | Невалидный JSON на входе | `NackDiscard` → DLQ |
| `RenderFailed_ForwardsError` | `success:false` с error | Транслирует error code (TIMEOUT) в ScrapeResponse |
| `RenderFailed_NoError` | `success:false` без error | `success:false`, поле error не установлено |
| `ParseSuccess` | Успешный парсинг HTML | `success:true`, page заполнен |
| `ParseError` | Parser возвращает ошибку | `success:false`, `error.code=PARSE_ERROR` |
| `PublishFailed` | Publish в RabbitMQ не удался | `NackRequeue` → повтор |
| `MetadataPassthrough` | Входное сообщение содержит metadata | metadata копируется в ответ без изменений |
| `NoMetadata` | Входное сообщение без metadata | Поле metadata отсутствует в ответе |

### sanitizer (`internal/parser`)

| Тест | Сценарий |
|---|---|
| `RemovesNav` | `<nav>` удаляется, `<p>` остаётся |
| `RemovesFooterHeaderAside` | `<footer>`, `<header>`, `<aside>` удаляются |
| `RemovesHiddenElements` | `[hidden]`, `[aria-hidden]`, `display:none` удаляются |
| `RemovesScriptStyleNoscript` | `<script>`, `<style>`, `<noscript>` удаляются |
| `PreservesMainContent` | `<h1>`, `<p>`, `<li>` внутри `<main>` сохраняются |

### extractor (`internal/parser`)

| Тест | Сценарий |
|---|---|
| `SelectContentRoot_Main` | При наличии `<main>` выбирается он |
| `SelectContentRoot_Article` | При отсутствии `<main>` выбирается `<article>` |
| `SelectContentRoot_Body` | Fallback на `<body>` |
| `ClassifyElement` | h1–h6 → heading + level, p → paragraph, li → listItem |
| `CalcPosition_RuneBased` | Позиция считается в рунах, не в байтах (кириллица) |
| `MakeSnippet_RuneSafe` | Сниппет не ломает UTF-8 символы |
| `ExtractMetadata` | title, description, previewUrl, iconUrl, language из HTML head |
| `ExtractMetadata_Fallbacks` | og:title, og:description как fallback |
| `ExtractMetadata_Untitled` | Пустой title → "Untitled" |
| `ContentExtractor_FullPage` | Полная страница: блоки, ссылки (internal/external), документы |
| `ContentExtractor_FiltersJunkHrefs` | `#`, `javascript:`, `mailto:`, `tel:` отфильтрованы |
| `ContentExtractor_EmptyTextSkipped` | Пустые элементы не попадают в блоки |

### normalizer (`internal/parser`)

| Тест | Сценарий |
|---|---|
| `FilterShortBlocks` | Блоки с текстом < 2 символов удаляются |
| `CollapseWhitespace` | Множественные пробелы/переносы → один пробел |
| `NormalizeURL_RemovesUTM` | `utm_*` параметры и `#fragment` удаляются |
| `DeduplicateLinks` | Дублирующиеся href сохраняются только один раз |

## Что можно mock

| Граница | Интерфейс | Область видимости | Кто мокает |
|---|---|---|---|
| handler → parser | `parser.Parser` | Экспортируемый | `handler_test.go` |
| handler → amqp | `amqp.MessagePublisher` | Экспортируемый | `handler_test.go` |
| parser → sanitizer | `sanitizer` | Пакетный (unexported) | `parser/*_test.go` |
| parser → extractor | `extractor` | Пакетный (unexported) | `parser/*_test.go` |
| parser → normalizer | `normalizer` | Пакетный (unexported) | `parser/*_test.go` |

## Не покрыто тестами

| Компонент | Причина |
|---|---|
| `amqp/connection.go` | Требует реального RabbitMQ (интеграционный тест) |
| `amqp/broker.go` | Оркестрация consumer/publisher — интеграционный тест |
| `amqp/impl/consumer.go` | Зависит от AMQP channel — интеграционный тест |
| `amqp/impl/publisher.go` | Зависит от AMQP channel — интеграционный тест |
| `config/config.go` | Чтение env-переменных — тривиальная логика |