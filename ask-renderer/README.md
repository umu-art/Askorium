# Ask-Renderer

Сервис рендеринга страниц. Получает URL из RabbitMQ, рендерит через headless Chromium (Playwright), отдаёт HTML обратно в RabbitMQ.

## Место в пайплайне

```
              render.input              render.output
Ask-Crawler ──────────────► Ask-Renderer ──────────────► Ask-Parser
```

- Потребляет `RenderInput` из очереди `render.input`
- Публикует `RenderOutput` в очередь `render.output`
- Модели сообщений генерируются из `api/src/renderer-api.yaml` (OpenAPI Generator, **не** ручные)

## Быстрый старт

```bash
# 1. RabbitMQ
docker run -d --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:management

# 2. Генерация моделей (из корня проекта)
python api/build.py --langs python

# 3. Зависимости
cd ask-renderer
pip install -r requirements.txt
pip install ../api/build/python-renderer-api/
playwright install chromium

# 4. Запуск
python -m app.main
```

## Docker

```bash
# Сначала сгенерировать модели
python api/build.py --langs python

# Собирать из корня проекта (нужен доступ к api/build/)
docker build -f ask-renderer/Dockerfile -t ask-renderer .

docker run --rm \
  -e AMQP_URL=amqp://guest:guest@host.docker.internal:5672/ \
  ask-renderer
```

## Конфигурация (env)

| Переменная | Дефолт | Описание |
|---|---|---|
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ |
| `INPUT_QUEUE` | `render.input` | Входная очередь |
| `OUTPUT_QUEUE` | `render.output` | Выходная очередь |
| `MAX_CONCURRENT_PAGES` | `5` | Макс. одновременных вкладок Chromium (= prefetch) |
| `DEFAULT_TIMEOUT_MS` | `15000` | Таймаут загрузки страницы |
| `MAX_RETRIES` | `2` | Повторы для retriable-ошибок (итого 3 попытки) |
| `MESSAGE_TTL_MS` | `60000` | TTL сообщений во входной очереди |
| `HEALTH_PORT` | `8080` | Порт health check + metrics |

## Формат сообщений

### RenderInput (входящее)

```json
{
  "task_id": "uuid",
  "url": "https://example.com",
  "timeout_ms": 15000,
  "metadata": {}
}
```

`timeout_ms` и `metadata` — опциональны. `metadata` прозрачно передаётся в ответ без изменений.

### RenderOutput (исходящее)

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
  "error": { "code": "TIMEOUT", "message": "page load timed out: ..." },
  "metadata": {}
}
```

## Обработка ошибок

**Retriable** (повторяются до `MAX_RETRIES` раз): `TIMEOUT`, `NETWORK_ERROR`

**Non-retriable** (сразу error response): `PAGE_NOT_FOUND`, `ACCESS_DENIED`, `INVALID_URL`, `PARSE_ERROR`

Ошибки рендеринга **не** уходят в DLQ — всегда отправляется `RenderOutput{success: false}` вниз по пайплайну. В DLQ попадают только невалидные сообщения (битый JSON, отсутствие обязательных полей) и сообщения с истёкшим TTL.

## DLQ

| Компонент | Имя |
|---|---|
| Dead Letter Exchange | `render.dlx` (FANOUT) |
| Dead Letter Queue | `render.dlq` |

## Эндпоинты

**`GET /`** — health check. Проверяет Chromium и RabbitMQ-соединение.

```json
{ "status": "ok", "checks": { "browser": "ok", "rabbitmq": "ok" } }
```

**`GET /metrics`** — Prometheus-метрики:

| Метрика | Тип | Описание |
|---|---|---|
| `renderer_messages_total` | Counter | Обработанные сообщения (`status=success\|error\|dlq`) |
| `renderer_errors_total` | Counter | Ошибки по коду (`code=TIMEOUT\|...`) |
| `renderer_retries_total` | Counter | Общее число повторных попыток |
| `renderer_duration_seconds` | Histogram | Время рендеринга |
| `renderer_active_tasks` | Gauge | Задачи в обработке прямо сейчас |

## Ключевое поведение

- **Параллелизм** контролируется через RabbitMQ `prefetch_count`, а не семафорами. Один Chromium-процесс, одна вкладка на сообщение.
- **Graceful shutdown**: по SIGTERM/SIGINT прекращает приём новых сообщений, дожидается завершения активных задач, закрывает соединения. Неподтверждённые сообщения автоматически возвращаются в очередь.
- **Reconnect**: используется `aio_pika.connect_robust()` — автоматическое переподключение к RabbitMQ.
- **Персистентность**: очереди `durable=True`, сообщения `PERSISTENT`.
- **Manual ack**: сообщение подтверждается только после успешной публикации результата. При ошибке публикации — `nack(requeue=true)`.

## Структура

```
ask-renderer/
├── app/
│   ├── main.py               # Точка входа, сигналы, lifecycle
│   ├── config.py              # Конфигурация (env)
│   ├── browser_manager.py     # Chromium: запуск, рендеринг, классификация ошибок
│   ├── message_handler.py     # Обработка сообщений: десериализация → рендеринг → публикация
│   ├── rabbitmq_consumer.py   # RabbitMQ: подключение, очереди, prefetch, consume
│   ├── health.py              # HTTP: health check + /metrics
│   └── metrics.py             # Prometheus-метрики
├── requirements.txt
└── Dockerfile
```
