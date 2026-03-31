# ask-crawler benchmark

Измеряет throughput (pages/min) краулера при разных конфигурациях:
число рендереров, тип frontier (memory / Redis), глубина обхода, тип сайта.

---

## Структура

```
.crawler-benchmark/
├── docker-compose.yml   — стек бенчмарка (RabbitMQ, Redis, crawler, parser, renderer×N)
├── run.py               — запуск прогонов, сбор метрик, запись summary.csv
├── targets.json         — сайты-цели (12 штук, 3 группы) и профили обхода
├── build.sh             — пересобрать образы crawler/parser из Go-исходников
├── .venv/               — Python-окружение (только pika)
└── results/
    ├── summary.csv                    — сводная таблица всех прогонов
    └── <YYYYMMDD_HHMMSS_label>/
        ├── config.json                — параметры прогона
        ├── batch_NNNN.json            — данные страниц по батчам
        ├── task.completed.json        — финальное событие от краулера
        └── timing.json                — метрики времени
```

---

## Быстрый старт

### 1. Окружение

```bash
cd .crawler-benchmark
python3 -m venv .venv
.venv/bin/pip install pika
```

### 2. Docker-образы

`askorium/ask-renderer:latest` — уже собран (содержит Playwright/Chromium, ~4 GB),
не трогать. Crawler и parser пересобрать при изменении Go-кода:

```bash
bash build.sh          # crawler + parser
bash build.sh crawler  # только crawler
bash build.sh parser   # только parser
```

### 3. Запуск стека

```bash
docker compose up -d --wait
```

### 4. Запуск прогонов

**Один таргет:**
```bash
.venv/bin/python run.py docs.python.org
.venv/bin/python run.py docs.python.org --profile deep
.venv/bin/python run.py docs.python.org --label "r2-redis"
```

**Все таргеты последовательно:**
```bash
.venv/bin/python run.py --all --profile shallow
.venv/bin/python run.py --all --profile deep --label "r3-memory"
```

**Фильтрация по группе:**
```bash
# по тегу (AND-фильтр)
.venv/bin/python run.py --all --tags docs
.venv/bin/python run.py --all --tags ru,university
.venv/bin/python run.py --all --tags en,university

# исключить отдельные сайты
.venv/bin/python run.py --all --skip hse.ru msu.ru
```

### 5. Остановка стека

```bash
docker compose down
```

---

## Матрица экспериментов

Конфиг задаётся в `docker-compose.yml` **перед каждой группой**.

| Параметр | Где менять | Значения |
|---|---|---|
| Число рендереров | `deploy.replicas` (renderer) | 1 / 2 / 3 / 5 |
| Frontier | `REDIS_URL` (ask-crawler) | `""` = memory, `"redis://redis:6379"` = Redis |

После смены конфига — перезапустить стек:
```bash
docker compose down && docker compose up -d --wait
```

### Группа A — масштабирование рендереров

```bash
# replicas: 1 → 2 → 3 → 5, REDIS_URL: ""
.venv/bin/python run.py --all --tags docs --profile shallow --label "r1-memory"
# (сменить replicas, перезапустить стек)
.venv/bin/python run.py --all --tags docs --profile shallow --label "r2-memory"
# ...
```

### Группа B — FIFO vs Priority frontier

Priority включается через `options.priority: true` в сообщении — задаётся
в `targets.json` или через отдельный `--priority` флаг (TODO: добавить).
Пока: два прогона с `priority: false` (FIFO) и `priority: true` в коде.

### Группа C — In-memory vs Redis

```bash
# REDIS_URL: ""  →  REDIS_URL: "redis://redis:6379"
.venv/bin/python run.py --all --tags docs --profile shallow --label "memory"
# (сменить REDIS_URL, перезапустить стек)
.venv/bin/python run.py --all --tags docs --profile shallow --label "redis"
```

---

## Таргеты

12 сайтов в трёх группах:

| Группа | Таргеты | Особенность |
|---|---|---|
| `ru,university` | hse.ru, msu.ru, spbu.ru, urfu.ru, nsu.ru | Российские вузы; VPN нужен при работе с хоста |
| `en,university` | mit.edu, stanford.edu, ox.ac.uk, eth.ch | Крупные международные университеты |
| `docs` | docs.python.org, docs.docker.com, docs.ansible.com | Документация; доступны без VPN |

> **VPN:** российские сайты недоступны из Docker-контейнера при включённом VPN на хосте.
> Для группы `ru` нужно отключить VPN или использовать `--tags docs`.

Профили обхода:

| Профиль | max_pages | max_depth |
|---|---|---|
| `shallow` | 100 | 2 |
| `deep` | 500 | 6 |
| `full` | 2000 | 10 |

---

## Анализ результатов

Все прогоны накапливаются в `results/summary.csv`:

```
run,target,profile,frontier,replicas,elapsed_sec,throughput_ppm,
pages_scraped,pages_failed,completion_reason,n_batches
```

```python
import pandas as pd

df = pd.read_csv("results/summary.csv")

# Группа A: throughput vs replicas
df.groupby("replicas")["throughput_ppm"].mean().plot(marker="o")

# Группа C: memory vs Redis
df.groupby("frontier")[["elapsed_sec", "throughput_ppm"]].mean()

# По типу сайта
df.groupby("target")["throughput_ppm"].mean().sort_values()
```

---

## Известные ограничения

- **VPN** не пробрасывается в Docker — российские сайты работают только без VPN.
- **docs.ansible.com** медленно рендерится (~6 pages/min на shallow). При необходимости увеличить `TIMEOUT` в `run.py` (по умолчанию 600 с).
- **docs.docker.com** — Next.js SPA, высокая нагрузка на Playwright.
- После `docker compose down` очереди RabbitMQ сбрасываются — незавершённые задачи теряются.
