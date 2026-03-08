package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	amqpimpl "ask-parser/internal/amqp/impl"
	"ask-parser/internal/handler"

	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

// --- mocks ---

type mockParser struct {
	page *scrappermodel.ScrappedPage
	err  error
}

func (m *mockParser) Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error) {
	return m.page, m.err
}

type mockPublisher struct {
	published []byte
	err       error
}

func (m *mockPublisher) Publish(ctx context.Context, body []byte) error {
	m.published = body
	return m.err
}

func (m *mockPublisher) PublishWithKey(ctx context.Context, routingKey string, body []byte) error {
	m.published = body
	return m.err
}

var noopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// --- helpers ---

func makeRenderOutput(taskID string, success bool, html string, meta map[string]interface{}) []byte {
	msg := map[string]interface{}{
		"task_id": taskID,
		"url":     "https://example.com",
		"success": success,
	}
	if html != "" {
		msg["html"] = html
	}
	if meta != nil {
		msg["metadata"] = meta
	}
	b, _ := json.Marshal(msg)
	return b
}

func makeRenderError(taskID string, code string, message string) []byte {
	msg := map[string]interface{}{
		"task_id": taskID,
		"url":     "https://example.com",
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	b, _ := json.Marshal(msg)
	return b
}

func unmarshalResponse(t *testing.T, data []byte) scrappermodel.ScrapeResponse {
	t.Helper()
	var resp scrappermodel.ScrapeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return resp
}

// --- tests ---

func TestHandler_InvalidJSON(t *testing.T) {
	h := handler.NewHandler(&mockParser{}, &mockPublisher{}, noopLogger)
	action := h(context.Background(), []byte("not json"))

	if action != amqpimpl.NackDiscard {
		t.Errorf("expected NackDiscard, got %v", action)
	}
}

func TestHandler_RenderFailed_ForwardsError(t *testing.T) {
	pub := &mockPublisher{}
	h := handler.NewHandler(&mockParser{}, pub, noopLogger)

	msg := makeRenderError("task-1", "TIMEOUT", "page timed out")
	action := h(context.Background(), msg)

	if action != amqpimpl.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}

	resp := unmarshalResponse(t, pub.published)
	if resp.GetSuccess() {
		t.Error("expected success=false")
	}
	if resp.GetTaskId() != "task-1" {
		t.Errorf("expected task_id=task-1, got %s", resp.GetTaskId())
	}
	if !resp.HasError() {
		t.Fatal("expected error to be set")
	}
	scrapeErr, _ := resp.GetErrorOk()
	if string(scrapeErr.GetCode()) != "TIMEOUT" {
		t.Errorf("expected error code TIMEOUT, got %s", scrapeErr.GetCode())
	}
}

func TestHandler_RenderFailed_NoError(t *testing.T) {
	pub := &mockPublisher{}
	h := handler.NewHandler(&mockParser{}, pub, noopLogger)

	msg := makeRenderOutput("task-2", false, "", nil)
	action := h(context.Background(), msg)

	if action != amqpimpl.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}

	resp := unmarshalResponse(t, pub.published)
	if resp.GetSuccess() {
		t.Error("expected success=false")
	}
	if resp.HasError() {
		t.Error("expected no error field when render output has no error")
	}
}

func TestHandler_ParseSuccess(t *testing.T) {
	page := scrappermodel.NewScrappedPage(
		"https://example.com",
		"Test Page",
		[]scrappermodel.ContentBlock{
			*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_HEADING, "Hello"),
		},
	)
	pub := &mockPublisher{}
	h := handler.NewHandler(&mockParser{page: page}, pub, noopLogger)

	msg := makeRenderOutput("task-3", true, "<html><body><h1>Hello</h1></body></html>", nil)
	action := h(context.Background(), msg)

	if action != amqpimpl.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}

	resp := unmarshalResponse(t, pub.published)
	if !resp.GetSuccess() {
		t.Error("expected success=true")
	}
	page, ok := resp.GetPageOk()
	if !ok {
		t.Fatal("expected page to be set")
	}
	if page.GetTitle() != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%s'", page.GetTitle())
	}
}

func TestHandler_ParseError(t *testing.T) {
	pub := &mockPublisher{}
	p := &mockParser{err: errors.New("broken html")}
	h := handler.NewHandler(p, pub, noopLogger)

	msg := makeRenderOutput("task-4", true, "<html>bad</html>", nil)
	action := h(context.Background(), msg)

	if action != amqpimpl.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}

	resp := unmarshalResponse(t, pub.published)
	if resp.GetSuccess() {
		t.Error("expected success=false")
	}
	parseErr, _ := resp.GetErrorOk()
	if string(parseErr.GetCode()) != "PARSE_ERROR" {
		t.Errorf("expected PARSE_ERROR, got %s", parseErr.GetCode())
	}
}

func TestHandler_PublishFailed(t *testing.T) {
	pub := &mockPublisher{err: errors.New("connection lost")}
	page := scrappermodel.NewScrappedPage("https://example.com", "T", nil)
	h := handler.NewHandler(&mockParser{page: page}, pub, noopLogger)

	msg := makeRenderOutput("task-5", true, "<html></html>", nil)
	action := h(context.Background(), msg)

	if action != amqpimpl.NackRequeue {
		t.Errorf("expected NackRequeue, got %v", action)
	}
}

func TestHandler_MetadataPassthrough(t *testing.T) {
	page := scrappermodel.NewScrappedPage("https://example.com", "T", nil)
	pub := &mockPublisher{}
	h := handler.NewHandler(&mockParser{page: page}, pub, noopLogger)

	meta := map[string]interface{}{"crawl_id": "abc", "depth": float64(2)}
	msg := makeRenderOutput("task-6", true, "<html></html>", meta)
	action := h(context.Background(), msg)

	if action != amqpimpl.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}

	resp := unmarshalResponse(t, pub.published)
	if !resp.HasMetadata() {
		t.Fatal("expected metadata to be set")
	}
	got := resp.GetMetadata()
	if got["crawl_id"] != "abc" {
		t.Errorf("expected crawl_id=abc, got %v", got["crawl_id"])
	}
	if got["depth"] != float64(2) {
		t.Errorf("expected depth=2, got %v", got["depth"])
	}
}

func TestHandler_NoMetadata(t *testing.T) {
	page := scrappermodel.NewScrappedPage("https://example.com", "T", nil)
	pub := &mockPublisher{}
	h := handler.NewHandler(&mockParser{page: page}, pub, noopLogger)

	msg := makeRenderOutput("task-7", true, "<html></html>", nil)
	h(context.Background(), msg)

	resp := unmarshalResponse(t, pub.published)
	if resp.HasMetadata() {
		t.Error("expected no metadata when input has none")
	}
}
