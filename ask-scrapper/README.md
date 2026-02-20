# ask-scrapper

Модуль сбора и парсинга веб-страниц для системы семантического поиска Askorium. Получает задачи через RabbitMQ, вызывает [ask-renderer](../ask-renderer) для получения HTML, извлекает структурированные данные и возвращает результат.

## Архитектура

```
ask-core (планировщик задач)
    │
    │  ScrapeRequest (RabbitMQ: scrape.request)
    ▼
ask-scrapper (данный модуль)
    │
    │  POST /render (HTTP)
    ▼
ask-renderer (headless Chromium)
    │
    │  HTML
    ▼
ask-scrapper: парсинг → ScrappedPage
    │
    │  ScrapeResponse (RabbitMQ: scrape.response)
    ▼
ask-core (потребитель результатов)
```

## Что извлекается

| Данные | Описание |
|--------|----------|
| **Метаданные** | title, description, og:image, favicon, язык, дата изменения |
| **Контент-блоки** | Заголовки (h1–h6), абзацы, элементы списков |
| **Ссылки** | Внутренние/внешние, anchor text, контекст |
| **Документы** | PDF, изображения, вложения с описанием |

## Конфигурация

Через переменные окружения:

| Переменная | Обязательная | По умолчанию | Описание |
|------------|:---:|---|---|
| `RENDERER_URL` | да | — | Адрес сервиса рендеринга |
| `AMQP_URL` | да | — | Строка подключения к RabbitMQ |
| `AMQP_REQUEST_QUEUE` | нет | `scrape.request` | Очередь входящих задач |
| `AMQP_RESPONSE_QUEUE` | нет | `scrape.response` | Очередь результатов |
| `PREFETCH_COUNT` | нет | `5` | Макс. параллельных задач |
| `LOG_LEVEL` | нет | `info` | Уровень логирования |

## Запуск

пока не знаю

## Структура проекта

```
ask-scrapper/
├── cmd/scrapper/main.go              # Точка входа
├── internal/
│   ├── config/config.go              # Чтение переменных окружения
│   ├── renderer/renderer.go          # Обёртка над сгенерированным клиентом renderer
│   ├── parser/
│   │   ├── parser.go                 # Общий вход: HTML → ScrappedPage
│   │   ├── metadata.go               # Извлечение метаданных страницы
│   │   ├── blocks.go                 # Извлечение контент-блоков
│   │   ├── links.go                  # Извлечение и классификация ссылок
│   │   └── documents.go              # Извлечение документов и изображений
│   ├── pipeline/pipeline.go          # Оркестрация: renderer → parser → ответ
│   └── handler/consumer.go           # Потребитель RabbitMQ
├── go.mod
└── go.sum
```

## Обработка ошибок

Все ошибки упаковываются в `ScrapeResponse` с `success: false`:

| Код ошибки | Описание | Повтор |
|------------|----------|:---:|
| `TIMEOUT` | Таймаут загрузки страницы | да |
| `NETWORK_ERROR` | Сетевая ошибка | да |
| `RATE_LIMITED` | Превышен лимит запросов | да |
| `PAGE_NOT_FOUND` | Страница не найдена (404) | нет |
| `ACCESS_DENIED` | Доступ запрещён (401/403) | нет |
| `INVALID_URL` | Невалидный URL | нет |
| `PARSE_ERROR` | Ошибка парсинга HTML | нет |

## Технологии

| Компонент | Технология |
|-----------|------------|
| Язык | Go |
| Логирование | slog (стандартная библиотека) |
| HTML-парсинг | goquery |
| Очереди | RabbitMQ (amqp091-go) |
| API-модели | OpenAPI Generator |