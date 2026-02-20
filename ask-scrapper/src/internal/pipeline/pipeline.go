package pipeline

import (
	"ask-scrapper/internal/parser"
	"ask-scrapper/internal/renderer"
	"context"
	"errors"
	"log/slog"

	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

const (
	defaultTimeoutMs  = 15000
	defaultMaxRetries = 0
)

type Pipeline interface {
	Process(ctx context.Context, req *scrappermodel.ScrapeRequest) *scrappermodel.ScrapeResponse
}

type pipelineImpl struct {
	renderer renderer.Renderer
	parser   parser.Parser
}

func NewPipeline(r renderer.Renderer, p parser.Parser) Pipeline {
	return &pipelineImpl{renderer: r, parser: p}
}

func (p *pipelineImpl) Process(ctx context.Context, req *scrappermodel.ScrapeRequest) *scrappermodel.ScrapeResponse {
	timeoutMs, maxRetries := p.extractOptions(req)

	slog.Info("rendering page", "task_id", req.GetTaskId(), "url", req.GetUrl())
	html, err := p.renderWithRetry(ctx, req.GetUrl(), timeoutMs, maxRetries, req.GetTaskId())
	if err != nil {
		slog.Error("render failed", "task_id", req.GetTaskId(), "error", err)
		return p.failResponse(req, err)
	}

	page, err := p.parser.Parse(html, req.GetUrl())
	if err != nil {
		slog.Error("parse failed", "task_id", req.GetTaskId(), "error", err)
		return p.errorResponse(req, scrappermodel.ERRORCODE_PARSE_ERROR, err.Error())
	}

	slog.Info("scrape succeeded", "task_id", req.GetTaskId(), "url", req.GetUrl())
	resp := scrappermodel.NewScrapeResponse(req.GetTaskId(), true)
	resp.SetPage(*page)
	resp.SetMetadata(req.GetMetadata())
	return resp
}

func (p *pipelineImpl) extractOptions(req *scrappermodel.ScrapeRequest) (timeoutMs int, maxRetries int) {
	timeoutMs = defaultTimeoutMs
	maxRetries = defaultMaxRetries

	if req.HasOptions() {
		opts := req.GetOptions()
		if opts.HasTimeoutMs() {
			timeoutMs = int(opts.GetTimeoutMs())
		}
		if opts.HasMaxRetries() {
			maxRetries = int(opts.GetMaxRetries())
		}
	}
	return
}

func (p *pipelineImpl) renderWithRetry(ctx context.Context, url string, timeoutMs int, maxRetries int, taskID string) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		html, _, err := p.renderer.Render(ctx, url, timeoutMs)
		if err == nil {
			return html, nil
		}
		lastErr = err

		var re *renderer.RenderError
		if errors.As(err, &re) && !isRetryable(re.Code) {
			return "", err
		}

		if attempt < maxRetries {
			slog.Warn("render failed, retrying",
				"task_id", taskID,
				"attempt", attempt+1,
				"error", err,
			)
		}
	}
	return "", lastErr
}

func (p *pipelineImpl) failResponse(req *scrappermodel.ScrapeRequest, err error) *scrappermodel.ScrapeResponse {
	code := scrappermodel.ERRORCODE_NETWORK_ERROR
	message := err.Error()

	var re *renderer.RenderError
	if errors.As(err, &re) {
		code = scrappermodel.ErrorCode(re.Code)
		message = re.Message
	}

	return p.errorResponse(req, code, message)
}

func (p *pipelineImpl) errorResponse(req *scrappermodel.ScrapeRequest, code scrappermodel.ErrorCode, message string) *scrappermodel.ScrapeResponse {
	resp := scrappermodel.NewScrapeResponse(req.GetTaskId(), false)
	resp.SetError(*scrappermodel.NewScrapeError(code, message))
	resp.SetMetadata(req.GetMetadata())
	return resp
}

func isRetryable(code string) bool {
	return code == string(scrappermodel.ERRORCODE_TIMEOUT) ||
		code == string(scrappermodel.ERRORCODE_NETWORK_ERROR)
}
