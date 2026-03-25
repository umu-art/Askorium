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
	defaultMaxDepth    int32 = 100
	defaultMaxPages    int32 = 100000
	defaultConcurrency int32 = 3

	// inFlightTTL — максимальное время ожидания результата от parser'а.
	// Если за это время ответ не пришёл (например, сообщение ушло в DLQ),
	// URL считается failed и pipeline продолжается.
	inFlightTTL = 5 * time.Minute
)

// activeCrawl объединяет состояние одной активной задачи: JobState, BatchBuilder
// и временный аккумулятор страниц (Phase 1 — до появления page.batch события).
type activeCrawl struct {
	job     *applib.JobState
	builder *applib.BatchBuilder
	mu      sync.Mutex
	pages   []crawler_model.ScrappedPage // Phase 1: накапливается здесь, отправляется в finishJob
}

func (ac *activeCrawl) addBatch(pages []crawler_model.ScrappedPage) {
	ac.mu.Lock()
	ac.pages = append(ac.pages, pages...)
	ac.mu.Unlock()
}

func (ac *activeCrawl) drainPages() []crawler_model.ScrappedPage {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	out := ac.pages
	ac.pages = nil
	return out
}

// CrawlerService управляет активными задачами краулинга.
type CrawlerService struct {
	renderPublisher infra_amqp.MessagePublisher
	eventPublisher  infra_amqp.MessagePublisher
	logger          *slog.Logger

	filter    applib.URLFilter
	scorer    applib.LinkScorer
	batchCfg  applib.BatchConfig
	batchMode bool // true → стриминг page.batch; false → Phase 1 (всё в task.completed)

	jobsMu sync.Mutex
	jobs   map[string]*activeCrawl
}

func NewCrawlerService(
	renderPublisher infra_amqp.MessagePublisher,
	eventPublisher infra_amqp.MessagePublisher,
	logger *slog.Logger,
	batchMode bool,
) *CrawlerService {
	return &CrawlerService{
		renderPublisher: renderPublisher,
		eventPublisher:  eventPublisher,
		logger:          logger,
		filter:          applib.NewDefaultChainFilter(applib.DefaultFilterConfig()),
		scorer:          applib.DefaultCompositeScorer(),
		batchCfg:        applib.DefaultBatchConfig(),
		batchMode:       batchMode,
		jobs:            make(map[string]*activeCrawl),
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

	s.logger.Info("crawler: task options",
		"max_depth", maxDepth,
		"max_pages", maxPages,
		"concurrency", concurrency,
	)

	var frontier applib.Frontier
	if s.batchMode {
		frontier = applib.NewPriorityFrontier()
	} else {
		frontier = applib.NewFIFOFrontier()
	}
	job := applib.NewJobState(req.TaskId, req.Domain, maxDepth, maxPages, concurrency, req.Metadata, frontier)

	seeds := req.SeedUrls
	if len(seeds) == 0 {
		seeds = []string{req.Domain}
	}
	for _, u := range seeds {
		job.Enqueue(applib.URLCandidate{URL: u, Depth: 0})
	}

	ac := &activeCrawl{job: job}
	ac.builder = applib.NewBatchBuilder(s.batchCfg, func(pages []crawler_model.ScrappedPage, seq int) {
		if s.batchMode {
			// Phase 2: публикуем батч немедленно; используем Background чтобы не
			// зависеть от контекста HandleTask/WatchBatch.
			if err := s.publishPageBatch(context.Background(), job.TaskID, pages, seq); err != nil {
				s.logger.Error("crawler: failed to publish page.batch",
					"task_id", job.TaskID,
					"batch_seq", seq,
					"error", err,
				)
			}
		} else {
			// Phase 1: аккумулируем, отправим в task.completed.
			ac.addBatch(pages)
		}
	})

	s.jobsMu.Lock()
	s.jobs[req.TaskId] = ac
	s.jobsMu.Unlock()

	return s.fillPipeline(ctx, ac)
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
	ac, ok := s.jobs[resp.TaskId]
	s.jobsMu.Unlock()

	if !ok {
		s.logger.Warn("crawler: unknown task_id, ignoring", "task_id", resp.TaskId)
		return nil
	}

	job := ac.job
	srcURL := extractSrcURL(resp.Metadata)
	currentDepth := extractDepth(resp.Metadata)

	if resp.Success && resp.Page != nil {
		sourcePageURL := extractSourcePageURL(resp.Metadata)
		page := applib.MapPage(*resp.Page, sourcePageURL)
		job.MarkParsed(srcURL, page)
		ac.builder.Add(&page)
		s.enqueueLinks(job, resp.Page.Links, currentDepth)
		s.enqueueDocuments(job, resp.Page.Documents, resp.Page.Url)
	} else {
		errMsg := ""
		if resp.Error != nil {
			errMsg = resp.Error.Message
		}
		job.MarkFailed(srcURL, errMsg)
		s.logger.Warn("crawler: scrape failed", "task_id", resp.TaskId, "error", errMsg)
	}

	return s.fillPipeline(ctx, ac)
}

func (s *CrawlerService) fillPipeline(ctx context.Context, ac *activeCrawl) error {
	job := ac.job
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
		return s.finishJob(ctx, ac, reason)
	}

	return nil
}

func (s *CrawlerService) sendToRenderer(ctx context.Context, job *applib.JobState, entry applib.URLCandidate) error {
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

	job.MarkRendering(entry.URL)

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
		anchorText := ""
		if l.AnchorText != nil {
			anchorText = *l.AnchorText
		}
		c := applib.URLCandidate{
			URL:        l.Href,
			Depth:      currentDepth + 1,
			AnchorText: anchorText,
		}
		if !s.filter.Allow(c) {
			continue
		}
		c.Score = s.scorer.Score(c)
		job.Enqueue(c)
	}
}

func (s *CrawlerService) enqueueDocuments(job *applib.JobState, docs []scrapper_model.Document, sourcePageURL string) {
	for _, d := range docs {
		hint := extensionHint(d.MimeType)
		if hint == "" {
			continue
		}
		job.Enqueue(applib.URLCandidate{
			URL:             d.Url,
			ContentTypeHint: hint,
			SourcePageURL:   sourcePageURL,
		})
	}
}

func (s *CrawlerService) finishJob(ctx context.Context, ac *activeCrawl, reason crawler_model.CompletionReason) error {
	job := ac.job
	scraped, failed, _ := job.Stats()
	frontier := int32(0)

	// Flush финального батча.
	// Phase 1: flush складывает в ac.pages → включаем в task.completed.
	// Phase 2: flush публикует page.batch → task.completed идёт без pages.
	ac.builder.Flush()

	var pages []crawler_model.ScrappedPage
	if !s.batchMode {
		pages = ac.drainPages()
	}

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

// publishPageBatch публикует промежуточный батч страниц (Phase 2).
func (s *CrawlerService) publishPageBatch(ctx context.Context, taskID string, pages []crawler_model.ScrappedPage, seq int) error {
	batchSeq := int32(seq)
	event := crawler_model.CrawlEvent{
		TaskId:    taskID,
		Type:      crawler_model.CRAWLEVENTTYPE_PAGE_BATCH,
		Timestamp: time.Now().UTC(),
		BatchSeq:  &batchSeq,
		Pages:     pages,
	}
	return s.publishEvent(ctx, event, fmt.Sprintf("%s.page.batch.%d", taskID, seq))
}

func (s *CrawlerService) failJob(ctx context.Context, ac *activeCrawl, code crawler_model.CrawlerErrorCode, message string) error {
	job := ac.job
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

// WatchBatch запускает фоновый goroutine, который раз в секунду вызывает Tick
// на BatchBuilder всех активных задач — для age-based flush (Phase 2).
func (s *CrawlerService) WatchBatch(ctx context.Context) {
	if !s.batchMode {
		return // в Phase 1 age-based flush не нужен
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.jobsMu.Lock()
			crawls := make([]*activeCrawl, 0, len(s.jobs))
			for _, ac := range s.jobs {
				crawls = append(crawls, ac)
			}
			s.jobsMu.Unlock()
			for _, ac := range crawls {
				ac.builder.Tick()
			}
		}
	}
}

// WatchTTL запускает фоновый goroutine, который периодически проверяет
// in-flight страницы всех активных задач и снимает зависшие (TTL истёк).
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
	crawls := make([]*activeCrawl, 0, len(s.jobs))
	for _, ac := range s.jobs {
		crawls = append(crawls, ac)
	}
	s.jobsMu.Unlock()

	for _, ac := range crawls {
		expired := ac.job.ExpireInFlight(inFlightTTL)
		if len(expired) == 0 {
			continue
		}
		for _, url := range expired {
			s.logger.Warn("crawler: in-flight TTL expired, counting as failed",
				"task_id", ac.job.TaskID,
				"url", url,
			)
		}
		if err := s.fillPipeline(ctx, ac); err != nil {
			s.logger.Error("crawler: fillPipeline after TTL expiry failed",
				"task_id", ac.job.TaskID,
				"error", err,
			)
		}
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
