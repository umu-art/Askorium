package applib

import (
	"sync"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
)

// URLEntry — элемент очереди обхода.
type URLEntry struct {
	URL   string
	Depth int32
}

// JobState хранит состояние одной задачи краулинга в памяти.
type JobState struct {
	TaskID      string
	Domain      string
	MaxDepth    int32
	MaxPages    int32
	Concurrency int32
	Metadata    map[string]interface{}

	StartedAt time.Time

	mu           sync.Mutex
	visited      map[string]bool
	frontier     []URLEntry
	pages        []crawler_model.ScrappedPage
	pagesScraped int32
	pagesFailed  int32
	inFlightAt   map[string]time.Time // key: оригинальный URL, отправленный в renderer
}

func NewJobState(taskID, domain string, maxDepth, maxPages, concurrency int32, metadata map[string]interface{}) *JobState {
	return &JobState{
		TaskID:      taskID,
		Domain:      domain,
		MaxDepth:    maxDepth,
		MaxPages:    maxPages,
		Concurrency: concurrency,
		Metadata:    metadata,
		StartedAt:   time.Now(),
		visited:     make(map[string]bool),
		inFlightAt:  make(map[string]time.Time),
	}
}

// Enqueue добавляет URL в очередь если он ещё не был посещён.
func (j *JobState) Enqueue(url string, depth int32) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.visited[url] {
		return
	}
	j.visited[url] = true
	j.frontier = append(j.frontier, URLEntry{URL: url, Depth: depth})
}

// Dequeue извлекает следующий URL из очереди. Возвращает false если очередь пуста.
func (j *JobState) Dequeue() (URLEntry, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if len(j.frontier) == 0 {
		return URLEntry{}, false
	}
	entry := j.frontier[0]
	j.frontier = j.frontier[1:]
	return entry, true
}

// LimitReached возвращает true если достигнут лимит страниц.
func (j *JobState) LimitReached() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.pagesScraped >= j.MaxPages
}

func (j *JobState) IncrScraped() {
	j.mu.Lock()
	j.pagesScraped++
	j.mu.Unlock()
}

func (j *JobState) IncrFailed() {
	j.mu.Lock()
	j.pagesFailed++
	j.mu.Unlock()
}

// AddPage добавляет успешно обработанную страницу в накопитель.
func (j *JobState) AddPage(page crawler_model.ScrappedPage) {
	j.mu.Lock()
	j.pages = append(j.pages, page)
	j.mu.Unlock()
}

// Pages возвращает копию накопленных страниц.
func (j *JobState) Pages() []crawler_model.ScrappedPage {
	j.mu.Lock()
	defer j.mu.Unlock()
	result := make([]crawler_model.ScrappedPage, len(j.pages))
	copy(result, j.pages)
	return result
}

func (j *JobState) IncrInFlight(url string) {
	j.mu.Lock()
	j.inFlightAt[url] = time.Now()
	j.mu.Unlock()
}

func (j *JobState) DecrInFlight(url string) {
	j.mu.Lock()
	delete(j.inFlightAt, url)
	j.mu.Unlock()
}

func (j *JobState) InFlightCount() int32 {
	j.mu.Lock()
	defer j.mu.Unlock()
	return int32(len(j.inFlightAt))
}

// IsDone возвращает true когда frontier пуст и нет страниц в обработке.
func (j *JobState) IsDone() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return len(j.frontier) == 0 && len(j.inFlightAt) == 0
}

// ExpireInFlight удаляет из in-flight записи, превысившие timeout,
// инкрементирует pagesFailed и возвращает список просроченных URL.
func (j *JobState) ExpireInFlight(timeout time.Duration) []string {
	j.mu.Lock()
	defer j.mu.Unlock()

	var expired []string
	now := time.Now()
	for url, sentAt := range j.inFlightAt {
		if now.Sub(sentAt) > timeout {
			expired = append(expired, url)
			delete(j.inFlightAt, url)
			j.pagesFailed++
		}
	}
	return expired
}

// Stats возвращает текущую статистику: scraped, failed, frontier size.
func (j *JobState) Stats() (scraped, failed, frontier int32) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.pagesScraped, j.pagesFailed, int32(len(j.frontier))
}
