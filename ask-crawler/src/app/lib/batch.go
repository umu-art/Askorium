package applib

import (
	"sync"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
)

// BatchConfig описывает настройки сборки батчей страниц.
type BatchConfig struct {
	MaxSize        int           // макс. страниц в одном батче (default: 20)
	MaxAge         time.Duration // макс. ожидание с момента первой страницы (default: 15s)
	MaxDocsPerPage int           // лимит документов в метаданных страницы (default: 10)
}

func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxSize:        20,
		MaxAge:         15 * time.Second,
		MaxDocsPerPage: 10,
	}
}

// BatchBuilder накапливает страницы и отправляет батчи по триггеру.
//
// Триггеры flush:
//   - len(pending) >= MaxSize  — после каждого Add
//   - now - firstAt >= MaxAge  — через Tick()
//   - принудительный Flush()   — при завершении задачи
type BatchBuilder struct {
	cfg     BatchConfig
	mu      sync.Mutex
	pending []*crawler_model.ScrappedPage
	firstAt time.Time
	seq     int // порядковый номер батча, начиная с 1
	flush   func(pages []crawler_model.ScrappedPage, seq int)
}

// NewBatchBuilder создаёт BatchBuilder. flush вызывается при каждом сбросе батча.
func NewBatchBuilder(cfg BatchConfig, flush func(pages []crawler_model.ScrappedPage, seq int)) *BatchBuilder {
	return &BatchBuilder{
		cfg:   cfg,
		flush: flush,
	}
}

// Add добавляет страницу. Возвращает true если был выполнен flush.
func (b *BatchBuilder) Add(page *crawler_model.ScrappedPage) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.trimDocs(page)

	if len(b.pending) == 0 {
		b.firstAt = time.Now()
	}
	b.pending = append(b.pending, page)

	if len(b.pending) >= b.cfg.MaxSize {
		b.doFlush()
		return true
	}
	return false
}

// Tick проверяет возраст батча и сбрасывает если MaxAge превышен.
// Вызывать периодически (например, раз в секунду через time.Ticker).
// Возвращает true если был выполнен flush.
func (b *BatchBuilder) Tick() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.pending) == 0 {
		return false
	}
	if time.Since(b.firstAt) >= b.cfg.MaxAge {
		b.doFlush()
		return true
	}
	return false
}

// Flush принудительно сбрасывает накопленные страницы.
// Вызывается при завершении задачи.
func (b *BatchBuilder) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.pending) > 0 {
		b.doFlush()
	}
}

// Seq возвращает номер следующего батча (сколько батчей уже было отправлено + 1).
func (b *BatchBuilder) Seq() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.seq + 1
}

// doFlush выполняет сброс; вызывается под мьютексом.
func (b *BatchBuilder) doFlush() {
	b.seq++
	pages := make([]crawler_model.ScrappedPage, len(b.pending))
	for i, p := range b.pending {
		pages[i] = *p
	}
	b.pending = b.pending[:0]
	b.flush(pages, b.seq)
}

// trimDocs ограничивает число документов в метаданных страницы до MaxDocsPerPage.
// Стратегия: оставляем первые N (парсер уже сортирует по контексту).
// Сами документы как страницы продолжают обрабатываться через frontier.
func (b *BatchBuilder) trimDocs(page *crawler_model.ScrappedPage) {
	if b.cfg.MaxDocsPerPage <= 0 || len(page.Documents) <= b.cfg.MaxDocsPerPage {
		return
	}
	page.Documents = page.Documents[:b.cfg.MaxDocsPerPage]
}
