package applib

import (
	"sync"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
)

// PageStatus описывает жизненный цикл одной страницы в рамках задачи.
type PageStatus int

const (
	PageStatusQueued    PageStatus = iota // в frontier, ещё не отправлена
	PageStatusRendering                   // отправлена в renderer, ждём ответа
	PageStatusParsed                      // получен результат, ждёт батча
	PageStatusBatched                     // включена в CrawlEvent, отправлена
	PageStatusFailed                      // ошибка на любом этапе
)

// PageProgress хранит полный lifecycle одной страницы внутри задачи краулинга.
type PageProgress struct {
	URL   string
	Depth int32
	IsDoc bool // true для документных ссылок

	Status PageStatus

	EnqueuedAt time.Time
	SentAt     time.Time // момент отправки в renderer
	ParsedAt   time.Time
	BatchedAt  time.Time

	Result *crawler_model.ScrappedPage // заполняется после Parsed
	Error  string
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
	visited      VisitedStore
	frontier     Frontier
	progress     map[string]*PageProgress // key: URL
	pagesScraped int32
	pagesFailed  int32
}

func NewJobState(taskID, domain string, maxDepth, maxPages, concurrency int32, metadata map[string]interface{}, frontier Frontier, visited VisitedStore) *JobState {
	return &JobState{
		TaskID:      taskID,
		Domain:      domain,
		MaxDepth:    maxDepth,
		MaxPages:    maxPages,
		Concurrency: concurrency,
		Metadata:    metadata,
		StartedAt:   time.Now(),
		visited:     visited,
		frontier:    frontier,
		progress:    make(map[string]*PageProgress),
	}
}

// Enqueue добавляет URL во frontier если он ещё не был посещён.
func (j *JobState) Enqueue(c URLCandidate) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.visited.Contains(c.URL) {
		return
	}
	j.visited.Add(c.URL)
	j.progress[c.URL] = &PageProgress{
		URL:        c.URL,
		Depth:      c.Depth,
		IsDoc:      c.ContentTypeHint != "",
		Status:     PageStatusQueued,
		EnqueuedAt: time.Now(),
	}
	j.frontier.Push(c)
}

// Dequeue извлекает следующий кандидат из frontier. Возвращает false если очередь пуста.
func (j *JobState) Dequeue() (URLCandidate, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.frontier.Pop()
}

func (j *JobState) FrontierLen() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.frontier.Len()
}

// LimitReached возвращает true если достигнут лимит HTML-страниц.
func (j *JobState) LimitReached() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.pagesScraped >= j.MaxPages
}

func (j *JobState) MarkRendering(url string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if p, ok := j.progress[url]; ok {
		p.Status = PageStatusRendering
		p.SentAt = time.Now()
	}
}

func (j *JobState) MarkParsed(url string, result crawler_model.ScrappedPage) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if p, ok := j.progress[url]; ok {
		p.Status = PageStatusParsed
		p.ParsedAt = time.Now()
		p.Result = &result
		j.pagesScraped++
	}
}

func (j *JobState) MarkFailed(url string, reason string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if p, ok := j.progress[url]; ok {
		p.Status = PageStatusFailed
		p.Error = reason
		j.pagesFailed++
	}
}

// InFlightCount возвращает число страниц в статусе Rendering.
func (j *JobState) InFlightCount() int32 {
	j.mu.Lock()
	defer j.mu.Unlock()
	var count int32
	for _, p := range j.progress {
		if p.Status == PageStatusRendering {
			count++
		}
	}
	return count
}

// IsDone возвращает true когда frontier пуст и нет страниц в состоянии Rendering или Parsed.
func (j *JobState) IsDone() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.frontier.Len() > 0 {
		return false
	}
	for _, p := range j.progress {
		if p.Status == PageStatusRendering || p.Status == PageStatusParsed {
			return false
		}
	}
	return true
}

// DrainParsed возвращает до limit страниц в статусе Parsed и переводит их в Batched.
func (j *JobState) DrainParsed(limit int) []*PageProgress {
	j.mu.Lock()
	defer j.mu.Unlock()
	var result []*PageProgress
	for _, p := range j.progress {
		if p.Status == PageStatusParsed {
			p.Status = PageStatusBatched
			p.BatchedAt = time.Now()
			result = append(result, p)
			if len(result) >= limit {
				break
			}
		}
	}
	return result
}

// ExpireInFlight удаляет из Rendering записи, превысившие timeout,
// переводит их в Failed и возвращает список просроченных URL.
func (j *JobState) ExpireInFlight(timeout time.Duration) []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	var expired []string
	now := time.Now()
	for url, p := range j.progress {
		if p.Status == PageStatusRendering && now.Sub(p.SentAt) > timeout {
			p.Status = PageStatusFailed
			p.Error = "in-flight TTL expired"
			j.pagesFailed++
			expired = append(expired, url)
		}
	}
	return expired
}

// Stats возвращает текущую статистику: scraped, failed, frontier size.
func (j *JobState) Stats() (scraped, failed, frontier int32) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.pagesScraped, j.pagesFailed, int32(j.frontier.Len())
}
