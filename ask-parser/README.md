# Ask-Parser

Сервис парсинга HTML-страниц. Получает отрендеренный HTML из RabbitMQ, извлекает структурированные данные (текстовые блоки, ссылки, документы, метаданные), отдаёт результат обратно в RabbitMQ.

## Место в пайплайне

```
              render.output             parser.output
Ask-Renderer ──────────────► Ask-Parser ──────────────► Ask-Crawler
```

- Потребляет `RenderOutput` из очереди `render.output`
- Публикует `ScrapeResponse` в очередь `parser.output`
- Модели сообщений генерируются из `api/src/renderer-api.yaml` и `api/src/scrapper-api.yaml` (OpenAPI Generator, **не** ручные)

## Быстрый старт

```bash
# 1. RabbitMQ
docker run -d --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:management

# 2. Генерация моделей (из корня проекта)
python api/build.py --langs go

# 3. Сборка и запуск
cd ask-parser/src
AMQP_URL=amqp://guest:guest@localhost:5672/ go run ./cmd/main.go
```

## Docker

Из корня проекта:

```bash
# Сборка образа
docker build -f iac/images/ask-parser/Dockerfile -t ask-parser .

# Запуск
docker run -e AMQP_URL=amqp://guest:guest@host.docker.internal:5672/ ask-parser
```

## Конфигурация (env)

| Переменная | Дефолт | Описание |
|---|---|---|
| `AMQP_URL` | *(обязательна)* | Адрес RabbitMQ |
| `INPUT_QUEUE` | `render.output` | Входная очередь |
| `OUTPUT_QUEUE` | `parser.output` | Выходная очередь |
| `PREFETCH_COUNT` | `10` | Prefetch (количество одновременно обрабатываемых сообщений) |
| `LOG_LEVEL` | `info` | Уровень логирования (`debug`, `info`, `warn`, `error`) |

## Формат сообщений

### RenderOutput (входящее)

Успех:

```json
{
  "task_id": "uuid",
  "url": "https://example.com",
  "success": true,
  "html": "<html>...</html>",
  "elapsed_ms": 1234,
  "metadata": {}
}
```

Ошибка:

```json
{
  "task_id": "uuid",
  "url": "https://example.com",
  "success": false,
  "error": { "code": "TIMEOUT", "message": "..." },
  "metadata": {}
}
```

### ScrapeResponse (исходящее)

Успех:

```json
{
  "task_id": "uuid",
  "success": true,
  "page": {
    "url": "https://example.com",
    "title": "Example",
    "description": "...",
    "language": "en",
    "blocks": [
      { "htmlId": "block-0", "type": "heading", "headingLevel": 1, "text": "Hello" },
      { "htmlId": "block-1", "type": "paragraph", "text": "Some text..." }
    ],
    "links": [
      { "href": "https://example.com/about", "type": "internal", "anchorText": "About", "contextBlockId": "block-1", "position": 5, "snippet": "..." }
    ],
    "documents": [
      { "url": "https://example.com/file.pdf", "mimeType": "application/pdf", "description": "Download" }
    ]
  },
  "metadata": {}
}
```

Ошибка:

```json
{
  "task_id": "uuid",
  "success": false,
  "error": { "code": "PARSE_ERROR", "message": "failed to parse HTML: ..." },
  "metadata": {}
}
```

## Обработка сообщений

| Сценарий | Действие |
|---|---|
| `RenderOutput` с `success: false` | Ошибка Renderer прозрачно транслируется в `ScrapeResponse{success: false}` |
| `RenderOutput` с `success: true` | HTML парсится → `ScrapeResponse{success: true, page: ...}` |
| Ошибка парсинга | `ScrapeResponse{success: false, error: PARSE_ERROR}` |
| Невалидный JSON | `nack(requeue=false)` → DLQ |
| Ошибка публикации ответа | `nack(requeue=true)` → повторная обработка |

Поле `metadata` прозрачно копируется из входящего сообщения в исходящее без изменений.

## Пайплайн парсинга

Парсер обрабатывает HTML в три последовательных этапа:

```
Raw HTML
  │
  ▼
Sanitizer   — удаляет шумовые узлы из DOM
  │
  ▼
Extractor   — извлекает метаданные, блоки, ссылки, документы
  │
  ▼
Normalizer  — нормализует URL, убирает дубликаты, чистит текст
  │
  ▼
ScrappedPage
```

### Sanitizer

Удаляет из DOM узлы, не несущие полезного контента:

| Селектор | Причина удаления |
|---|---|
| `nav` | Навигация — повторяется на каждой странице |
| `aside` | Боковые панели, реклама |
| `footer` | Подвал — копирайт, повторные ссылки |
| `header` | Шапка — логотип, навигация |
| `noscript`, `script`, `style` | Служебный код |
| `svg`, `iframe` | Встроенные объекты без текстового контента |
| `[hidden]`, `[aria-hidden="true"]` | Скрытые элементы |
| `[style*="display:none"]` | Элементы, скрытые через inline-стили |

### Extractor

#### Метаданные страницы

| Поле | Источник |
|---|---|
| `title` | `<title>` → `og:title` → `"Untitled"` |
| `description` | `meta[name=description]` → `og:description` |
| `previewUrl` | `og:image` |
| `iconUrl` | `link[rel=icon]` → `link[rel="shortcut icon"]` |
| `language` | `<html lang>` |
| `lastModified` | `article:modified_time` → `http-equiv=last-modified` |

#### Контентные блоки

Парсер выбирает корневой элемент контента: `<main>` → `<article>` → `<body>` (fallback).

Внутри корня обрабатываются элементы `h1`–`h6`, `<p>`, `<li>`:
- **heading** — заголовки с указанием уровня
- **paragraph** — абзацы
- **listItem** — элементы списка

Каждый блок получает `htmlId` (из атрибута `id` элемента, либо генерируется `block-0`, `block-1`, ...).

#### Ссылки

Из каждого блока извлекаются `<a>` — с позицией в тексте (rune-based), сниппетом контекста, классификацией (internal/external по хосту).

Фильтруются: пустые href, `#` (якоря), `javascript:`, `mailto:`, `tel:`.

#### Документы

- `<a>` со ссылкой на файл (`.pdf`, `.doc`, `.docx`, `.xls`, `.xlsx`, `.ppt`, `.pptx`, `.zip`, `.rar`) → `Document`
- `<img>` → `Document` (с alt-текстом как описание)

### Normalizer

| Действие | Описание |
|---|---|
| Фильтрация коротких блоков | Блоки с текстом < 2 символов удаляются (убирает шум вроде кнопок "A") |
| Нормализация пробелов | Множественные пробелы/переносы → одиночный пробел |
| Нормализация URL | Удаление `utm_*` параметров и фрагментов (`#...`) |
| Дедупликация ссылок | Одинаковые URL сохраняются только при первом вхождении |

## Планируемые расширения

### Расширение ContentBlockType

Текущий парсер извлекает только `h1-h6`, `<p>`, `<li>`. Для полноценного семантического поиска Ask-Core необходимо расширить набор извлекаемых элементов:

| Элемент | Тип (планируемый) | Обоснование |
|---|---|---|
| `<table>` (`<td>`, `<th>`) | `tableCell` | Ключевая информация часто в таблицах (расписания, тарифы, параметры API) |
| `<pre>`, `<code>` | `code` | Основной контент технических документаций и туториалов |
| `<blockquote>` | `quote` | Цитаты — высокоценный контент для поиска (интервью, статьи) |
| `<dt>`, `<dd>` | `definition` | FAQ-страницы используют definition lists — пара вопрос/ответ |
| `<figcaption>` | `paragraph` | Описания изображений — ценный текст для контекстного поиска |

Требует обновления `ContentBlockType` enum в `scrapper-api.yaml` и согласования с Ask-Core.

### Извлечение содержимого документов

Поле `extractedText` в модели `Document` предусмотрено для текста, извлечённого из файлов (PDF, DOCX, OCR изображений).

Рекомендуемая архитектура — отдельный микросервис **Ask-DocExtractor**:

```
Ask-Parser ──► parser.output ──► Ask-Crawler
     │
     └──► doc.input ──► Ask-DocExtractor ──► doc.output ──► Ask-Crawler
```

Причина вынесения: загрузка файлов + извлечение текста — тяжёлый IO, не должно блокировать парсинг HTML.

## DLQ

| Компонент | Имя |
|---|---|
| Input DLX | `render.output.dlx` (FANOUT) |
| Input DLQ | `render.output.dlq` |
| Output DLX | `parser.output.dlx` (FANOUT) |
| Output DLQ | `parser.output.dlq` |

## Ключевое поведение

- **Manual ack**: сообщение подтверждается только после успешной публикации результата.
- **Reconnect**: автоматическое переподключение к RabbitMQ с экспоненциальным backoff (3s → 30s).
- **Graceful shutdown**: по `SIGTERM`/`SIGINT` дожидается завершения текущих задач. Неподтверждённые сообщения автоматически возвращаются в очередь.
- **Персистентность**: очереди `durable=true`, сообщения `PERSISTENT`.

## Структура

```
ask-parser/
└── src/
    ├── cmd/
    │   └── main.go                     # Точка входа
    ├── internal/
    │   ├── config/
    │   │   └── config.go               # Конфигурация (env)
    │   ├── amqp/
    │   │   ├── connection.go           # RabbitMQ-соединение с auto-reconnect
    │   │   ├── broker.go               # Оркестрация consumer/publisher
    │   │   ├── interfaces.go           # MessagePublisher, MessageConsumer
    │   │   └── impl/
    │   │       ├── consumer.go         # Consumer: topology, prefetch, consume loop
    │   │       └── publisher.go        # Publisher: exchange/queue declaration, publish
    │   ├── parser/
    │   │   ├── interfaces.go           # Parser (exported) + sanitizer/extractor/normalizer (internal)
    │   │   ├── parser.go              # htmlParser: Sanitize → Extract → Normalize
    │   │   ├── sanitize.go            # DOM-очистка шумовых узлов
    │   │   ├── extract.go             # Извлечение метаданных, блоков, ссылок, документов
    │   │   ├── normalize.go           # URL-нормализация, дедупликация, очистка текста
    │   │   └── documents.go           # Определение файлов, MIME-типы
    │   └── handler/
    │       └── handler.go              # Бизнес-логика обработки сообщений
    ├── go.mod
    └── go.sum
```
