# Askorium Api

Директория `api/` — единственный источник истины для контрактов между сервисами Askorium.
Здесь хранятся OpenAPI 3.1 спецификации и конфигурации кодогенерации. Никакой бизнес-логики нет.

## Зачем

Проект следует принципу **Api First**: спецификации пишутся вручную, а серверные интерфейсы
и клиентский код генерируются из них автоматически. Это гарантирует, что контракты между
Java-оркестратором (ask-core), Python-сервисами и Go-сервисами всегда согласованы.

## Структура

```
api/
├── build/                         # Сгенерированный код (git-ignored)│
├── config/                        # Конфигурации кодогенерации
│   └── lint.yaml                  # Правила линтинга Redocly
├── src/                           # OpenAPI 3.1 спецификации (источник истины)│
└── build.py                       # Скрипт кодогенерации
```

## Как пользоваться

### Требования

| Инструмент | Версия | Зачем |
|---|---|---|
| Python | 3.x | запуск `build.py` |
| Java | 17+ | запуск OpenAPI Generator CLI JAR |
| Redocly CLI | latest | бандлинг спецификаций (`npm i -g @redocly/cli`) |
| wget | — | скачивание JAR (первый запуск) |

+ набор инструментов для сборки и публикации пакетов (Maven для Java, build для Python, npm для TypeScript)

### Документация

* https://redocly.com/docs/cli
* https://openapi-generator.tech/

### Запуск

```bash
# Полная сборка всех языков
python api/build.py

# Только нужные языки
python api/build.py --langs java
python api/build.py --langs python go
python api/build.py --langs java python go

# Чистая пересборка (удаляет output-директории перед генерацией)
python api/build.py --force
python api/build.py --langs java --force
```

### Что происходит при запуске

1. Скачивается `openapi-generator-cli-7.14.0.jar` в `build/` (однократно).
2. Для каждого подходящего конфига в `config/` (параллельно):
   - Исходные спеки из `src/` объединяются в `combined.yaml` через `$ref`-ссылки.
   - Redocly разворачивает все `$ref` в единый `processed.yaml`.
   - OpenAPI Generator генерирует код по `processed.yaml` и конфигу.
3. Каждый `build/*-api/` устанавливается локально (параллельно):
   - `java` — `mvn clean install` → артефакт в локальный Maven-репозиторий.
   - `python` — `python -m build --wheel` → `pip install`.
   - `ts` — `npm install && npm run build && npm link`.

## Конфигурационные файлы

Файлы в `config/` — расширение конфигурации OpenAPI Generator с двумя дополнительными полями.

### Поля

| Поле | Тип | Описание |
|---|---|---|
| `sources` | `null` / строка / список | Какие спеки из `src/` включать. `null` или `'*'` — все, строка — glob-паттерн, список — несколько паттернов. |
| `target` | строка | Поддиректория в `build/`, куда пишется результат. Должна начинаться с имени языка: `java-*`, `python-*`, `go-*`, `ts-*`. |
| `generatorName` | строка | Генератор OpenAPI Generator (`spring`, `java`, `go`, `python-fastapi` и др.). |
| `additionalProperties` | объект | Параметры генератора (пакеты, артефакты, флаги). |
| `globalProperties` | объект | Управление набором генерируемых файлов (`supportingFiles`, `models`, `apis` и т.д.). |

### Пример: добавить новый язык

1. Написать спецификацию в `src/my-service-api.yaml`.
2. Создать `config/ts-my-service-config.yaml`:

```yaml
sources:
  - my-service-api.yaml
target: ts-my-service-api
generatorName: typescript-axios

additionalProperties:
  npmName: my-service-api
  supportsES6: true
```

3. Запустить `python api/build.py --langs ts`.

## Соглашения

- **Имя конфига**: `{lang}-{name}-config.yaml`. Префикс `{lang}` используется `build.py` для фильтрации по `--langs`.
- **Имя target**: `{lang}-{name}-api`. Тот же префикс — `build.py` определяет по нему способ установки.
- **Спеки не трогаем после генерации**: всё в `build/` — производный артефакт. Правки вносятся только в `src/`.
- **После изменения спеки** — обязательно перегенерировать: `python api/build.py`.
- **Линтинг**: `redocly lint --config api/config/lint.yaml api/src/*.yaml`.

## Известные баги

- Большая часть вспомогательных файлов генераторов отключена в конфиге, чтобы не мусорить
- Избегаем oneOf, anyOf, allOf, т.к. они плохо поддерживаются генераторами
- Очень осторожно с default значениями в схемах, т.к. они могут криво обрабатываться