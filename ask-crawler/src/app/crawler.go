package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
	scrapper_model "github.com/omo-ri/askorium/go-parser-api"
	render_model "github.com/omo-ri/askorium/go-renderer-api"

	applib "ask-crawler/src/app/lib"
	infra_amqp "ask-crawler/src/infra/amqp"
)

const (
	defaultMaxDepth int32 = 5
	defaultMaxPages int32 = 1000
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
	if req.Options != nil {
		if req.Options.MaxDepth != nil {
			maxDepth = *req.Options.MaxDepth
		}
		if req.Options.MaxPages != nil {
			maxPages = *req.Options.MaxPages
		}
	}

	job := applib.NewJobState(req.TaskId, req.Domain, maxDepth, maxPages, req.Metadata)

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

	return s.processNext(ctx, job)
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

	if resp.Success && resp.Page != nil {
		page := applib.MapPage(*resp.Page)

		// Сначала enqueue ссылки — чтобы frontier_size в stats был актуальным
		s.enqueueLinks(job, resp.Page.Links)
		job.IncrScraped()
		job.AddPage(page)
	} else {
		job.IncrFailed()
		s.logger.Warn("crawler: scrape failed", "task_id", resp.TaskId, "error", resp.Error)
	}

	if job.LimitReached() {
		return s.completeJob(ctx, job, crawler_model.COMPLETIONREASON_MAX_PAGES_REACHED)
	}

	entry, ok := job.Dequeue()
	if !ok {
		return s.completeJob(ctx, job, crawler_model.COMPLETIONREASON_FRONTIER_EMPTY)
	}
	return s.sendToRenderer(ctx, job, entry)
}

func (s *CrawlerService) processNext(ctx context.Context, job *applib.JobState) error {
	entry, ok := job.Dequeue()
	if !ok {
		return s.completeJob(ctx, job, crawler_model.COMPLETIONREASON_FRONTIER_EMPTY)
	}
	return s.sendToRenderer(ctx, job, entry)
}

func (s *CrawlerService) sendToRenderer(ctx context.Context, job *applib.JobState, entry applib.URLEntry) error {
	if job.MaxDepth > 0 && entry.Depth > job.MaxDepth {
		s.logger.Debug("crawler: skipping URL, depth limit",
			"task_id", job.TaskID,
			"url", entry.URL,
			"depth", entry.Depth,
		)
		next, ok := job.Dequeue()
		if !ok {
			return s.completeJob(ctx, job, crawler_model.COMPLETIONREASON_MAX_DEPTH_REACHED)
		}
		return s.sendToRenderer(ctx, job, next)
	}

	renderInput := render_model.RenderInput{
		TaskId:   job.TaskID,
		Url:      entry.URL,
		Metadata: map[string]interface{}{"depth": entry.Depth},
	}

	body, err := json.Marshal(renderInput)
	if err != nil {
		return fmt.Errorf("crawler: marshal RenderInput: %w", err)
	}
	if err := s.renderPublisher.Publish(ctx, body); err != nil {
		return fmt.Errorf("crawler: publish RenderInput: %w", err)
	}

	s.logger.Info("crawler: sent to render.requests",
		"task_id", job.TaskID,
		"url", entry.URL,
		"depth", entry.Depth,
	)
	return nil
}

func (s *CrawlerService) enqueueLinks(job *applib.JobState, links []scrapper_model.Link) {
	for _, l := range links {
		if l.Type != scrapper_model.LINKTYPE_INTERNAL {
			continue
		}
		job.Enqueue(l.Href, 1)
	}
}

// completeJob публикует task.completed со всеми накопленными страницами.
func (s *CrawlerService) completeJob(ctx context.Context, job *applib.JobState, reason crawler_model.CompletionReason) error {
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

func (s *CrawlerService) publishEvent(ctx context.Context, event crawler_model.CrawlEvent, routingKey string) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("crawler: marshal event %q: %w", routingKey, err)
	}
	if err := s.eventPublisher.PublishWithKey(ctx, routingKey, body); err != nil {
		return fmt.Errorf("crawler: publish event %q: %w", routingKey, err)
	}
	s.logger.Info("crawler: published event", "routing_key", routingKey)
	return nil
}
