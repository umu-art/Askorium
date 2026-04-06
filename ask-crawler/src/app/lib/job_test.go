package applib

import (
	"testing"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
)

func newTestJob() *JobState {
	factory := NewMemFrontierFactory()
	frontier := factory.NewFrontier("test", false)
	visited := factory.NewVisitedStore("test")
	return NewJobState("task1", "example.com", 3, 10, 5, nil, frontier, visited)
}

func TestJobState_Enqueue_Basic(t *testing.T) {
	j := newTestJob()
	j.Enqueue(URLCandidate{URL: "https://example.com/a"})
	if j.FrontierLen() != 1 {
		t.Errorf("expected FrontierLen=1, got %d", j.FrontierLen())
	}
}

func TestJobState_Enqueue_Dedup(t *testing.T) {
	j := newTestJob()
	c := URLCandidate{URL: "https://example.com/a"}
	j.Enqueue(c)
	j.Enqueue(c)
	if j.FrontierLen() != 1 {
		t.Errorf("expected FrontierLen=1 after duplicate enqueue, got %d", j.FrontierLen())
	}
}

func TestJobState_Enqueue_TwoDistinct(t *testing.T) {
	j := newTestJob()
	j.Enqueue(URLCandidate{URL: "https://example.com/a"})
	j.Enqueue(URLCandidate{URL: "https://example.com/b"})
	if j.FrontierLen() != 2 {
		t.Errorf("expected FrontierLen=2, got %d", j.FrontierLen())
	}
}

func TestJobState_Dequeue(t *testing.T) {
	j := newTestJob()
	j.Enqueue(URLCandidate{URL: "https://example.com/a"})
	got, ok := j.Dequeue()
	if !ok {
		t.Fatal("Dequeue returned false, expected true")
	}
	if got.URL != "https://example.com/a" {
		t.Errorf("got %q, want 'https://example.com/a'", got.URL)
	}
}

func TestJobState_Dequeue_Empty(t *testing.T) {
	j := newTestJob()
	_, ok := j.Dequeue()
	if ok {
		t.Error("Dequeue on empty frontier should return false")
	}
}

func TestJobState_MarkRendering(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	if j.InFlightCount() != 1 {
		t.Errorf("expected InFlightCount=1, got %d", j.InFlightCount())
	}
}

func TestJobState_MarkRendering_Unknown(t *testing.T) {
	j := newTestJob()
	// should be a no-op — no panic
	j.MarkRendering("https://unknown.com/")
	if j.InFlightCount() != 0 {
		t.Errorf("expected InFlightCount=0 for unknown URL, got %d", j.InFlightCount())
	}
}

func TestJobState_MarkParsed(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	j.MarkParsed(url, crawler_model.ScrappedPage{})
	if j.InFlightCount() != 0 {
		t.Errorf("expected InFlightCount=0 after MarkParsed, got %d", j.InFlightCount())
	}
	scraped, _, _ := j.Stats()
	if scraped != 1 {
		t.Errorf("expected pagesScraped=1, got %d", scraped)
	}
}

func TestJobState_MarkFailed(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	j.MarkFailed(url, "timeout")
	if j.InFlightCount() != 0 {
		t.Errorf("expected InFlightCount=0 after MarkFailed, got %d", j.InFlightCount())
	}
	_, failed, _ := j.Stats()
	if failed != 1 {
		t.Errorf("expected pagesFailed=1, got %d", failed)
	}
}

func TestJobState_LimitReached_Initially(t *testing.T) {
	j := newTestJob()
	if j.LimitReached() {
		t.Error("limit should not be reached on fresh job")
	}
}

func TestJobState_LimitReached_AfterMaxPages(t *testing.T) {
	factory := NewMemFrontierFactory()
	j := NewJobState("task", "example.com", 3, 2, 5, nil,
		factory.NewFrontier("t", false), factory.NewVisitedStore("t"))

	for _, u := range []string{"https://example.com/1", "https://example.com/2"} {
		j.Enqueue(URLCandidate{URL: u})
		j.Dequeue()
		j.MarkParsed(u, crawler_model.ScrappedPage{})
	}

	if !j.LimitReached() {
		t.Error("limit should be reached after maxPages scraped")
	}
}

func TestJobState_IsDone_Empty(t *testing.T) {
	j := newTestJob()
	if !j.IsDone() {
		t.Error("fresh empty job should be done")
	}
}

func TestJobState_IsDone_WithFrontierItem(t *testing.T) {
	j := newTestJob()
	j.Enqueue(URLCandidate{URL: "https://example.com/a"})
	if j.IsDone() {
		t.Error("job with items in frontier should not be done")
	}
}

func TestJobState_IsDone_WithInFlight(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	if j.IsDone() {
		t.Error("job should not be done while page is rendering")
	}
}

func TestJobState_IsDone_WithParsedNotBatched(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	j.MarkParsed(url, crawler_model.ScrappedPage{})
	if j.IsDone() {
		t.Error("job should not be done while page is in Parsed (not yet batched) state")
	}
}

func TestJobState_IsDone_AfterDrain(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	j.MarkParsed(url, crawler_model.ScrappedPage{})
	j.DrainParsed(10)
	if !j.IsDone() {
		t.Error("job should be done when all pages are batched and frontier is empty")
	}
}

func TestJobState_IsDone_AllFailed(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)
	j.MarkFailed(url, "error")
	if !j.IsDone() {
		t.Error("job should be done when all pages have failed and frontier is empty")
	}
}

func TestJobState_DrainParsed_LimitRespected(t *testing.T) {
	j := newTestJob()
	for _, u := range []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/c",
	} {
		j.Enqueue(URLCandidate{URL: u})
		j.Dequeue()
		j.MarkRendering(u)
		j.MarkParsed(u, crawler_model.ScrappedPage{})
	}

	first := j.DrainParsed(2)
	if len(first) != 2 {
		t.Errorf("DrainParsed(2) should return 2, got %d", len(first))
	}
	second := j.DrainParsed(10)
	if len(second) != 1 {
		t.Errorf("expected 1 remaining after second drain, got %d", len(second))
	}
}

func TestJobState_DrainParsed_Empty(t *testing.T) {
	j := newTestJob()
	drained := j.DrainParsed(10)
	if len(drained) != 0 {
		t.Errorf("expected 0 drained from empty job, got %d", len(drained))
	}
}

func TestJobState_ExpireInFlight_Expired(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)

	expired := j.ExpireInFlight(0)
	if len(expired) != 1 || expired[0] != url {
		t.Errorf("expected [%q] expired, got %v", url, expired)
	}
	_, failed, _ := j.Stats()
	if failed != 1 {
		t.Errorf("expected pagesFailed=1 after expiry, got %d", failed)
	}
}

func TestJobState_ExpireInFlight_NotExpired(t *testing.T) {
	j := newTestJob()
	url := "https://example.com/a"
	j.Enqueue(URLCandidate{URL: url})
	j.Dequeue()
	j.MarkRendering(url)

	expired := j.ExpireInFlight(10 * time.Minute)
	if len(expired) != 0 {
		t.Errorf("expected no expired URLs, got %v", expired)
	}
}

func TestJobState_ExpireInFlight_OnlyRendering(t *testing.T) {
	j := newTestJob()

	urlRendering := "https://example.com/rendering"
	urlParsed := "https://example.com/parsed"

	j.Enqueue(URLCandidate{URL: urlRendering})
	j.Enqueue(URLCandidate{URL: urlParsed})
	j.Dequeue()
	j.Dequeue()
	j.MarkRendering(urlRendering)
	j.MarkRendering(urlParsed)
	j.MarkParsed(urlParsed, crawler_model.ScrappedPage{})

	expired := j.ExpireInFlight(0)
	if len(expired) != 1 || expired[0] != urlRendering {
		t.Errorf("only rendering URL should expire, got %v", expired)
	}
}

func TestJobState_Stats(t *testing.T) {
	j := newTestJob()
	scraped, failed, frontier := j.Stats()
	if scraped != 0 || failed != 0 || frontier != 0 {
		t.Errorf("initial stats should be (0,0,0), got (%d,%d,%d)", scraped, failed, frontier)
	}

	j.Enqueue(URLCandidate{URL: "https://example.com/a"})
	j.Enqueue(URLCandidate{URL: "https://example.com/b"})
	j.Dequeue()
	j.Dequeue()
	j.MarkParsed("https://example.com/a", crawler_model.ScrappedPage{})
	j.MarkFailed("https://example.com/b", "err")

	scraped, failed, frontier = j.Stats()
	if scraped != 1 {
		t.Errorf("scraped=%d, want 1", scraped)
	}
	if failed != 1 {
		t.Errorf("failed=%d, want 1", failed)
	}
	if frontier != 0 {
		t.Errorf("frontier=%d, want 0", frontier)
	}
}
