# Askorium

[![ArgoCD](https://argocd.kazenin.ru/api/badge?project=askorium&revision=true)](https://argocd.kazenin.ru/api/badge?project=askorium&revision=true)
[![Assembling this whole mess](https://github.com/umu-art/Askorium/actions/workflows/entrypoint.yml/badge.svg)](https://github.com/umu-art/Askorium/actions/workflows/entrypoint.yml)

Данная курсовая работа направлена на разработку системы семантического поиска
«Аскориум», предназначенной для работы в рамках одного сайта. Система автоматически
собирает актуальный контент, обрабатывает запрос на естественном языке и выдает ответ
на основе источников.

# О проекте

Инструмент собирает контент из источников (сайт, набор документов, документация, услуги и т.п.)
и реализует осмысленный семантический поиск по нему: принимает запрос на естественном языке и возвращает
релевантный ответ с указанием ссылок.

Основная проблема — поиск информации в обширном контенте. Решение состоит из трёх частей:
быстрый и качественный скраппинг, хранение и обновление текстово-векторных данных, и семантический поиск по ним.

# Технологии

| Сервис          | Язык / стек                                          |
|-----------------|------------------------------------------------------|
| ask-core        | Java 21, Spring Boot 3.5, Spring Modulith            |
| ask-ui          | React 18, TypeScript, Vite, Tailwind CSS             |
| ask-encoder     | Python 3.11, FastAPI, sentence-transformers, PyTorch |
| ask-parser      | Go 1.25                                              |
| ask-crawler     | Go 1.25                                              |
| ask-renderer    | Python 3.11, FastAPI, Playwright (Chromium)          |
| ask-trafilatura | Python, trafilatura                                  |

**Инфраструктура:**

- PostgreSQL 17 — основная БД
- OpenSearch 3 — полнотекстовый (BM25) и векторный (kNN) поиск
- Redis 7 — distributed locking
- RabbitMQ 4 — очередь сообщений между сервисами
- Temporal 1.29 — оркестрация workflow

**API:** OpenAPI 3.1 (spec-first) + OpenAPI Generator — генерация клиентов/серверов на Java, Python, Go, TypeScript

**CI/CD:** GitHub Actions, ArgoCD (GitOps), Docker, Kubernetes

# Как поднимать из образов

## 1. Инфраструктура

```bash
docker-compose up -d
```

| Сервис      | Порт(ы)      | Примечание                 |
|-------------|--------------|----------------------------|
| PostgreSQL  | 5432         | Основная БД                |
| OpenSearch  | 9200         | Поиск                      |
| Redis       | 6379         |                            |
| RabbitMQ    | 5672 / 15672 | Management UI: guest/guest |
| Temporal    | 7233         |                            |
| Temporal UI | 8233         |                            |

## 2. Сборка и запуск сервисов

Все Dockerfile лежат в `iac/images/`. Сборка выполняется из корня репозитория:

```bash
docker build -f iac/images/ask-core/Dockerfile        -t ask-core        .
docker build -f iac/images/ask-ui/Dockerfile           -t ask-ui          .
docker build -f iac/images/ask-encoder/Dockerfile      -t ask-encoder     .
docker build -f iac/images/ask-renderer/Dockerfile     -t ask-renderer    .
docker build -f iac/images/ask-parser/Dockerfile       -t ask-parser      .
docker build -f iac/images/ask-crawler/Dockerfile      -t ask-crawler     .
docker build -f iac/images/ask-trafilatura/Dockerfile  -t ask-trafilatura .
```

Переменные окружения задаются через `.env` (см. `.env.example`).