package applib

import (
	"sync"
	"testing"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
)

func newPage(url string) *crawler_model.ScrappedPage {
	return &crawler_model.ScrappedPage{Url: url}
}

func TestBatchBuilder_SizeFlush(t *testing.T) {
	cfg := BatchConfig{MaxSize: 3, MaxAge: time.Hour, MaxDocsPerPage: 10}
	var mu sync.Mutex
	var flushed [][]crawler_model.ScrappedPage

	b := NewBatchBuilder(cfg, func(pages []crawler_model.ScrappedPage, seq int) {
		mu.Lock()
		flushed = append(flushed, pages)
		mu.Unlock()
	})

	b.Add(newPage("a"))
	b.Add(newPage("b"))
	if len(flushed) != 0 {
		t.Fatal("expected no flush after 2 pages")
	}
	flushed3 := b.Add(newPage("c")) // triggers size flush
	if !flushed3 {
		t.Error("Add should return true when flush triggered")
	}
	mu.Lock()
	n := len(flushed)
	mu.Unlock()
	if n != 1 {
		t.Fatalf("expected 1 flush, got %d", n)
	}
	if len(flushed[0]) != 3 {
		t.Errorf("expected batch of 3, got %d", len(flushed[0]))
	}
}

func TestBatchBuilder_AgeFlush(t *testing.T) {
	cfg := BatchConfig{MaxSize: 100, MaxAge: 10 * time.Millisecond, MaxDocsPerPage: 10}
	var mu sync.Mutex
	var flushed [][]crawler_model.ScrappedPage

	b := NewBatchBuilder(cfg, func(pages []crawler_model.ScrappedPage, seq int) {
		mu.Lock()
		flushed = append(flushed, pages)
		mu.Unlock()
	})

	b.Add(newPage("x"))

	// Tick до истечения age — не должно флашить
	ticked := b.Tick()
	if ticked {
		t.Error("Tick should not flush before MaxAge")
	}

	time.Sleep(20 * time.Millisecond)

	ticked = b.Tick()
	if !ticked {
		t.Error("Tick should flush after MaxAge")
	}
	mu.Lock()
	n := len(flushed)
	mu.Unlock()
	if n != 1 {
		t.Fatalf("expected 1 flush after age, got %d", n)
	}
}

func TestBatchBuilder_ForcedFlush(t *testing.T) {
	cfg := BatchConfig{MaxSize: 100, MaxAge: time.Hour, MaxDocsPerPage: 10}
	var flushed [][]crawler_model.ScrappedPage

	b := NewBatchBuilder(cfg, func(pages []crawler_model.ScrappedPage, seq int) {
		flushed = append(flushed, pages)
	})

	b.Add(newPage("p1"))
	b.Add(newPage("p2"))
	b.Flush()

	if len(flushed) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(flushed))
	}
	if len(flushed[0]) != 2 {
		t.Errorf("expected batch of 2, got %d", len(flushed[0]))
	}
}

func TestBatchBuilder_FlushEmpty(t *testing.T) {
	cfg := DefaultBatchConfig()
	called := false
	b := NewBatchBuilder(cfg, func(_ []crawler_model.ScrappedPage, _ int) {
		called = true
	})
	b.Flush() // should not call flush callback on empty
	if called {
		t.Error("Flush on empty builder should not invoke callback")
	}
}

func TestBatchBuilder_SeqIncrement(t *testing.T) {
	cfg := BatchConfig{MaxSize: 1, MaxAge: time.Hour, MaxDocsPerPage: 10}
	var seqs []int
	b := NewBatchBuilder(cfg, func(_ []crawler_model.ScrappedPage, seq int) {
		seqs = append(seqs, seq)
	})

	b.Add(newPage("1"))
	b.Add(newPage("2"))
	b.Add(newPage("3"))

	if len(seqs) != 3 {
		t.Fatalf("expected 3 flushes, got %d", len(seqs))
	}
	for i, seq := range seqs {
		if seq != i+1 {
			t.Errorf("seq[%d] = %d, want %d", i, seq, i+1)
		}
	}
}

func TestBatchBuilder_TrimDocs(t *testing.T) {
	cfg := BatchConfig{MaxSize: 100, MaxAge: time.Hour, MaxDocsPerPage: 2}
	var flushed []crawler_model.ScrappedPage

	b := NewBatchBuilder(cfg, func(pages []crawler_model.ScrappedPage, _ int) {
		flushed = append(flushed, pages...)
	})

	page := &crawler_model.ScrappedPage{
		Url: "https://example.com/",
		Documents: []crawler_model.Document{
			{Url: "a.pdf"}, {Url: "b.pdf"}, {Url: "c.pdf"},
		},
	}
	b.Add(page)
	b.Flush()

	if len(flushed) != 1 {
		t.Fatal("expected 1 page")
	}
	if len(flushed[0].Documents) != 2 {
		t.Errorf("expected 2 docs (trimmed), got %d", len(flushed[0].Documents))
	}
}
