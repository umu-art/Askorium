# E2E тест: ask-renderer + ask-parser

Ручной end-to-end тест пайплайна рендеринга и парсинга.

## Суть теста

Проверяет связку двух сервисов через RabbitMQ:

```
send.py → askorium.render.input → ask-renderer → askorium.render.output → ask-parser → askorium.parser.output → listen.py
```

1. `send.py` публикует задачу на рендеринг (URL страницы) в очередь `render.input`
2. `ask-renderer` рендерит страницу через Playwright и кладёт HTML в `render.output`
3. `ask-parser` парсит HTML и кладёт структурированный `ScrapeResponse` в `parser.output`
4. `listen.py` забирает результат и сохраняет в `result_<task_id>.json`

Пример результата: `wiki-res.json` — парсинг страницы Wikipedia "Web scraping".

## Требования

- Docker (инфраструктура: RabbitMQ)
- Python 3.10+ с `pika` (`pip install pika`)
- Запущенные `ask-renderer` и `ask-parser`

## Запуск

### 1. Инфраструктура

```bash
cd ..
docker compose up -d
```

### 2. ask-renderer

```bash
cd ask-renderer
pip install -r requirements.txt
pip install ../api/build/python-renderer-api/
playwright install chromium
AMQP_URL=amqp://guest:guest@localhost:5672/ python -m app.main
```

### 3. ask-parser

```bash
cd ask-parser/src
AMQP_URL=amqp://guest:guest@localhost:5672/ go run main.go
```

### 4. Слушаем результат (отдельный терминал)

```bash
cd .e2e-test-scrapper
python listen.py
```

### 5. Отправляем задачу

```bash
python send.py
# Выведет: Sent: <task_id>
```

`listen.py` сохранит файл `result_<task_id>.json` с полным `ScrapeResponse`.

## Изменить URL

В `send.py` поменять поле `url`:

```python
"url": "https://example.com/some-article",
```
