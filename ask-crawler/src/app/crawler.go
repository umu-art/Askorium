package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
	scrapper_model "github.com/omo-ri/askorium/go-parser-api"
	render_model "github.com/omo-ri/askorium/go-renderer-api"

	applib "ask-crawler/src/app/lib"
	infra_amqp "ask-crawler/src/infra/amqp"
)

const (
	defaultMaxDepth    int32 = 5
	defaultMaxPages    int32 = 100
	defaultConcurrency int32 = 3

	// inFlightTTL — максимальное время ожидания результата от parser'а.
	// Если за это время ответ не пришёл (например, сообщение ушло в DLQ),
	// URL считается failed и pipeline продолжается.
	inFlightTTL = 5 * time.Minute
)

type CrawlerService struct {
	renderPublisher infra_amqp.MessagePublisher
	eventPublisher  infra_amqp.MessagePublisher
	logger          *slog.Logger

	jobs   map[string]*applib.JobState
	jobsMu sync.Mutex
}

func NewCrawlerService(
	renderPublisher infra_amqp.MessagePublisher,
	eventPublisher infra_amqp.MessagePublisher,
	logger *slog.Logger,
) *CrawlerService {
	return &CrawlerService{
		renderPublisher: renderPublisher,
		eventPublisher:  eventPublisher,
		logger:          logger,
		jobs:            make(map[string]*applib.JobState),
	}
}

func (s *CrawlerService) HandleTask(ctx context.Context, msg []byte) error {
	var req crawler_model.CrawlTaskRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		return fmt.Errorf("crawler: unmarshal CrawlTaskRequest: %w", err)
	}

	s.logger.Info("crawler: received task",
		"task_id", req.TaskId,
		"domain", req.Domain,
		"seed_urls", req.SeedUrls,
	)

	maxDepth := defaultMaxDepth
	maxPages := defaultMaxPages
	concurrency := defaultConcurrency
	if req.Options != nil {
		if req.Options.MaxDepth != nil {
			maxDepth = *req.Options.MaxDepth
		}
		if req.Options.MaxPages != nil {
			maxPages = *req.Options.MaxPages
		}
		if req.Options.Concurrency != nil {
			concurrency = *req.Options.Concurrency
		}
	}

	job := applib.NewJobState(req.TaskId, req.Domain, maxDepth, maxPages, concurrency, req.Metadata)

	if len(req.SeedUrls) > 0 {
		for _, u := range req.SeedUrls {
			job.Enqueue(u, 0)
		}
	} else {
		job.Enqueue(req.Domain, 0)
	}

	s.jobsMu.Lock()
	s.jobs[req.TaskId] = job
	s.jobsMu.Unlock()

	return s.fillPipeline(ctx, job)
}

func (s *CrawlerService) HandleScrapeResult(ctx context.Context, msg []byte) error {
	var resp scrapper_model.ScrapeResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		return fmt.Errorf("crawler: unmarshal ScrapeResponse: %w", err)
	}

	s.logger.Info("crawler: received scrape result",
		"task_id", resp.TaskId,
		"success", resp.Success,
	)

	s.jobsMu.Lock()
	job, ok := s.jobs[resp.TaskId]
	s.jobsMu.Unlock()

	if !ok {
		s.logger.Warn("crawler: unknown task_id, ignoring", "task_id", resp.TaskId)
		return nil
	}

	srcURL := extractSrcURL(resp.Metadata)
	currentDepth := extractDepth(resp.Metadata)

	if resp.Success && resp.Page != nil {
		sourcePageURL := extractSourcePageURL(resp.Metadata)
		page := applib.MapPage(*resp.Page, sourcePageURL)
		s.enqueueLinks(job, resp.Page.Links, currentDepth)
		s.enqueueDocuments(job, resp.Page.Documents, resp.Page.Url)
		job.IncrScraped()
		job.AddPage(page)
	} else {
		job.IncrFailed()
		s.logger.Warn("crawler: scrape failed", "task_id", resp.TaskId, "error", resp.Error)
	}

	job.DecrInFlight(srcURL)

	return s.fillPipeline(ctx, job)
}

func (s *CrawlerService) fillPipeline(ctx context.Context, job *applib.JobState) error {
	for {
		if job.InFlightCount() >= job.Concurrency {
			break
		}
		entry, ok := job.Dequeue()
		if !ok {
			break
		}
		isDoc := entry.ContentTypeHint != ""
		// HTML pages respect the page limit; documents are always processed.
		if !isDoc && job.LimitReached() {
			continue
		}
		if !isDoc && job.MaxDepth > 0 && entry.Depth > job.MaxDepth {
			s.logger.Debug("crawler: skipping URL, depth limit",
				"task_id", job.TaskID,
				"url", entry.URL,
				"depth", entry.Depth,
			)
			continue
		}
		if err := s.sendToRenderer(ctx, job, entry); err != nil {
			return err
		}
	}

	done := job.IsDone()
	limitedAndDrained := job.LimitReached() && job.InFlightCount() == 0

	if done || limitedAndDrained {
		reason := crawler_model.COMPLETIONREASON_FRONTIER_EMPTY
		if job.LimitReached() {
			reason = crawler_model.COMPLETIONREASON_MAX_PAGES_REACHED
		}
		return s.finishJob(ctx, job, reason)
	}

	return nil
}

func (s *CrawlerService) sendToRenderer(ctx context.Context, job *applib.JobState, entry applib.URLEntry) error {
	metadata := map[string]interface{}{
		"depth":   entry.Depth,
		"src_url": entry.URL,
	}
	if entry.ContentTypeHint != "" {
		metadata["content_type_hint"] = entry.ContentTypeHint
	}
	if entry.SourcePageURL != "" {
		metadata["source_page_url"] = entry.SourcePageURL
	}

	renderInput := render_model.RenderInput{
		TaskId:   job.TaskID,
		Url:      entry.URL,
		Metadata: metadata,
	}

	body, err := json.Marshal(renderInput)
	if err != nil {
		return fmt.Errorf("crawler: marshal RenderInput: %w", err)
	}
	if err := s.renderPublisher.Publish(ctx, body); err != nil {
		return fmt.Errorf("crawler: publish RenderInput: %w", err)
	}

	job.IncrInFlight(entry.URL)

	s.logger.Debug("crawler: sent to renderer",
		"task_id", job.TaskID,
		"url", entry.URL,
		"depth", entry.Depth,
	)
	return nil
}

func (s *CrawlerService) enqueueLinks(job *applib.JobState, links []scrapper_model.Link, currentDepth int32) {
	for _, l := range links {
		if l.Type != scrapper_model.LINKTYPE_INTERNAL {
			continue
		}
		job.Enqueue(l.Href, currentDepth+1)
	}
}

func (s *CrawlerService) finishJob(ctx context.Context, job *applib.JobState, reason crawler_model.CompletionReason) error {
	scraped, failed, _ := job.Stats()
	frontier := int32(0)
	pages := job.Pages()

	event := crawler_model.CrawlEvent{
		TaskId:           job.TaskID,
		Type:             crawler_model.CRAWLEVENTTYPE_TASK_COMPLETED,
		Timestamp:        time.Now().UTC(),
		Pages:            pages,
		CompletionReason: &reason,
		Stats: &crawler_model.CrawlProgressStats{
			PagesScraped: &scraped,
			PagesFailed:  &failed,
			FrontierSize: &frontier,
		},
	}

	if err := s.publishEvent(ctx, event, fmt.Sprintf("%s.task.completed", job.TaskID)); err != nil {
		return err
	}

	s.removeJob(job.TaskID)
	s.logger.Info("crawler: task completed",
		"task_id", job.TaskID,
		"reason", reason,
		"pages_scraped", scraped,
		"pages_failed", failed,
	)
	return nil
}

// failJob публикует task.failed с описанием ошибки.
func (s *CrawlerService) failJob(ctx context.Context, job *applib.JobState, code crawler_model.CrawlerErrorCode, message string) error {
	scraped, failed, _ := job.Stats()
	frontier := int32(0)

	crawlerErr := crawler_model.CrawlerError{
		Code:    code,
		Message: message,
	}

	event := crawler_model.CrawlEvent{
		TaskId:    job.TaskID,
		Type:      crawler_model.CRAWLEVENTTYPE_TASK_FAILED,
		Timestamp: time.Now().UTC(),
		Error:     &crawlerErr,
		Stats: &crawler_model.CrawlProgressStats{
			PagesScraped: &scraped,
			PagesFailed:  &failed,
			FrontierSize: &frontier,
		},
	}

	if err := s.publishEvent(ctx, event, fmt.Sprintf("%s.task.failed", job.TaskID)); err != nil {
		return err
	}

	s.removeJob(job.TaskID)
	s.logger.Error("crawler: task failed",
		"task_id", job.TaskID,
		"code", code,
		"message", message,
	)
	return nil
}

func (s *CrawlerService) removeJob(taskID string) {
	s.jobsMu.Lock()
	delete(s.jobs, taskID)
	s.jobsMu.Unlock()
}

func (s *CrawlerService) publishEvent(ctx context.Context, event crawler_model.CrawlEvent, eventType string) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("crawler: marshal event %q: %w", eventType, err)
	}
	if err := s.eventPublisher.Publish(ctx, body); err != nil {
		return fmt.Errorf("crawler: publish event %q: %w", eventType, err)
	}
	s.logger.Info("crawler: published event", "event_type", eventType)
	return nil
}

// WatchTTL запускает фоновый goroutine, который периодически проверяет
// in-flight страницы всех активных задач и снимает зависшие (TTL истёк).
// Вызывать до broker.Start.
func (s *CrawlerService) WatchTTL(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.expireInFlight(ctx)
		}
	}
}

func (s *CrawlerService) expireInFlight(ctx context.Context) {
	s.jobsMu.Lock()
	jobs := make([]*applib.JobState, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	s.jobsMu.Unlock()

	for _, job := range jobs {
		expired := job.ExpireInFlight(inFlightTTL)
		if len(expired) == 0 {
			continue
		}
		for _, url := range expired {
			s.logger.Warn("crawler: in-flight TTL expired, counting as failed",
				"task_id", job.TaskID,
				"url", url,
			)
		}
		if err := s.fillPipeline(ctx, job); err != nil {
			s.logger.Error("crawler: fillPipeline after TTL expiry failed",
				"task_id", job.TaskID,
				"error", err,
			)
		}
	}
}

func (s *CrawlerService) enqueueDocuments(job *applib.JobState, docs []scrapper_model.Document, sourcePageURL string) {
	for _, d := range docs {
		hint := extensionHint(d.MimeType)
		if hint == "" {
			continue
		}
		job.EnqueueDocument(d.Url, hint, sourcePageURL)
	}
}

func extensionHint(mime string) string {
	switch {
	case mime == "application/pdf":
		return "pdf"
	case strings.Contains(mime, "wordprocessingml"):
		return "docx"
	case strings.Contains(mime, "spreadsheetml"):
		return "xlsx"
	case strings.Contains(mime, "vnd.ms-excel"):
		return "xls"
	default:
		return ""
	}
}

// extractSrcURL читает оригинальный URL из metadata сообщения.
func extractSrcURL(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	v, ok := metadata["src_url"]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func extractSourcePageURL(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	v, ok := metadata["source_page_url"]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// extractDepth reads the "depth" field from RabbitMQ message metadata.
// JSON numbers unmarshal as float64 in Go, so we handle that type explicitly.
func extractDepth(metadata map[string]interface{}) int32 {
	if metadata == nil {
		return 0
	}
	v, ok := metadata["depth"]
	if !ok {
		return 0
	}
	switch d := v.(type) {
	case float64:
		return int32(d)
	case int32:
		return d
	case int:
		return int32(d)
	}
	return 0
}
