# Сравнение конфигураций парсера

Как протестировать разные варианты пайплайна (сегментеры, сериализаторы таблиц) на одних и тех же страницах.

## Идея

Запускаем парсер с одной конфигурацией, прогоняем набор URL, сохраняем результаты.
Меняем конфигурацию, повторяем. Сравниваем JSON-выходы.

## Что менять

В `ask-parser/internal/parser/pipeline.go`, функция `NewParser()`:

### Базовый сегментер (текущий default)

```go
segment.NewDOMBlockSegmenter(),
```

### Enhanced + CompactTableSerializer (default)

```go
segment.NewEnhancedDOMBlockSegmenter(nil),
```

### Enhanced + LinearizedTableSerializer

```go
segment.NewEnhancedDOMBlockSegmenter(&segment.LinearizedTableSerializer{}),
```

## Процесс

1. Выбрать набор URL с разной структурой:
   - статья (Wikipedia, Хабр) — проверка текста, заголовков, списков
   - страница с таблицами — проверка сериализации
   - FAQ/документация — проверка dt/dd, figcaption
   - страница с навигационными списками — проверка группировки li

2. Запустить с первой конфигурацией:
   ```bash
   # В pipeline.go: segment.NewDOMBlockSegmenter()
   cd ask-parser/src && go run main.go
   ```
   Прогнать URL через `send.py` + `listen.py`, сохранить результаты:
   ```bash
   mkdir -p results/base
   mv result_*.json results/base/
   ```

3. Поменять сегментер в `pipeline.go`, перезапустить парсер, повторить:
   ```bash
   mkdir -p results/enhanced-compact
   mv result_*.json results/enhanced-compact/
   ```

4. Сравнить:
   ```bash
   # Количество блоков
   jq '.contentBlocks | length' results/base/*.json
   jq '.contentBlocks | length' results/enhanced-compact/*.json

   # Текст блоков (для визуального сравнения)
   jq -r '.contentBlocks[].text' results/base/result_xxx.json > base.txt
   jq -r '.contentBlocks[].text' results/enhanced-compact/result_xxx.json > enhanced.txt
   diff base.txt enhanced.txt
   ```

## На что смотреть

- **Количество блоков**: enhanced должен давать меньше за счёт группировки li
- **Таблицы**: base выдаёт отдельные td, enhanced — один блок на таблицу
- **Потеря контента**: убедиться, что dt/dd/figcaption не теряются
- **Linearized vs Compact**: linearized даёт более длинные блоки с повторением заголовков — проверить, не завышает ли это score у селектора
