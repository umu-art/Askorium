package applib

import (
	"sync"
	"time"
)

type URLEntry struct {
	URL   string
	Depth int32
}

type JobState struct {
	TaskID   string
	Domain   string
	MaxDepth int32
	MaxPages int32
	Metadata map[string]interface{}

	StartedAt time.Time

	mu           sync.Mutex
	visited      map[string]bool
	frontier     []URLEntry
	pagesScraped int32
	pagesFailed  int32
}

func NewJobState(taskID, domain string, maxDepth, maxPages int32, metadata map[string]interface{}) *JobState {
	return &JobState{
		TaskID:    taskID,
		Domain:    domain,
		MaxDepth:  maxDepth,
		MaxPages:  maxPages,
		Metadata:  metadata,
		StartedAt: time.Now(),
		visited:   make(map[string]bool),
	}
}

func (j *JobState) Enqueue(url string, depth int32) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.visited[url] {
		return
	}
	j.visited[url] = true
	j.frontier = append(j.frontier, URLEntry{URL: url, Depth: depth})
}

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

func (j *JobState) Stats() (scraped, failed, frontier int32) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.pagesScraped, j.pagesFailed, int32(len(j.frontier))
}
