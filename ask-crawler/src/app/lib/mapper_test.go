package applib

import (
	"testing"
	"time"

	scrapper_model "github.com/omo-ri/askorium/go-parser-api"
)

func strPtr(s string) *string { return &s }

func TestMapPage_BasicFields(t *testing.T) {
	lang := "en"
	src := scrapper_model.ScrappedPage{
		Url:      "https://example.com/page",
		Title:    "Test Title",
		Language: &lang,
		Blocks:   []scrapper_model.ContentBlock{},
	}

	got := MapPage(src, "")

	if got.Url != src.Url {
		t.Errorf("Url: got %q, want %q", got.Url, src.Url)
	}
	if got.Title != src.Title {
		t.Errorf("Title: got %q, want %q", got.Title, src.Title)
	}
	if got.Language == nil || *got.Language != lang {
		t.Errorf("Language: got %v, want %q", got.Language, lang)
	}
}

func TestMapPage_LastModified(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	src := scrapper_model.ScrappedPage{
		Url:          "https://example.com/page",
		Title:        "T",
		LastModified: &now,
		Blocks:       []scrapper_model.ContentBlock{},
	}
	got := MapPage(src, "")
	if got.LastModified == nil || !got.LastModified.Equal(now) {
		t.Errorf("LastModified: got %v, want %v", got.LastModified, now)
	}
}

func TestMapPage_NilLastModified(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
	}
	got := MapPage(src, "")
	if got.LastModified != nil {
		t.Errorf("expected nil LastModified, got %v", got.LastModified)
	}
}

func TestMapPage_SourcePageURL_Set(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/doc.pdf",
		Title:  "Doc",
		Blocks: []scrapper_model.ContentBlock{},
	}
	got := MapPage(src, "https://example.com/parent")
	if got.SourcePageUrl == nil || *got.SourcePageUrl != "https://example.com/parent" {
		t.Errorf("SourcePageUrl: got %v, want 'https://example.com/parent'", got.SourcePageUrl)
	}
}

func TestMapPage_SourcePageURL_Empty(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
	}
	got := MapPage(src, "")
	if got.SourcePageUrl != nil {
		t.Errorf("expected nil SourcePageUrl when empty, got %v", got.SourcePageUrl)
	}
}

func TestMapPage_Blocks(t *testing.T) {
	htmlID := "section-1"
	var headingLevel int32 = 2
	src := scrapper_model.ScrappedPage{
		Url:   "https://example.com/page",
		Title: "T",
		Blocks: []scrapper_model.ContentBlock{
			{HtmlId: &htmlID, Type: "heading", HeadingLevel: &headingLevel, Text: "Hello"},
			{Type: "paragraph", Text: "World"},
		},
	}
	got := MapPage(src, "")
	if len(got.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(got.Blocks))
	}
	if got.Blocks[0].Text != "Hello" {
		t.Errorf("block[0].Text = %q, want 'Hello'", got.Blocks[0].Text)
	}
	if got.Blocks[0].HtmlId == nil || *got.Blocks[0].HtmlId != htmlID {
		t.Errorf("block[0].HtmlId = %v, want %q", got.Blocks[0].HtmlId, htmlID)
	}
	if string(got.Blocks[1].Type) != "paragraph" {
		t.Errorf("block[1].Type = %q, want 'paragraph'", got.Blocks[1].Type)
	}
}

func TestMapPage_Blocks_Empty(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
	}
	got := MapPage(src, "")
	if len(got.Blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(got.Blocks))
	}
}

func TestMapPage_Links(t *testing.T) {
	anchor := "Click here"
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
		Links: []scrapper_model.Link{
			{Href: "https://example.com/other", Type: "internal", AnchorText: &anchor},
		},
	}
	got := MapPage(src, "")
	if len(got.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(got.Links))
	}
	if got.Links[0].Href != "https://example.com/other" {
		t.Errorf("link Href = %q", got.Links[0].Href)
	}
	if got.Links[0].AnchorText == nil || *got.Links[0].AnchorText != anchor {
		t.Errorf("link AnchorText = %v, want %q", got.Links[0].AnchorText, anchor)
	}
	if string(got.Links[0].Type) != "internal" {
		t.Errorf("link Type = %q, want 'internal'", got.Links[0].Type)
	}
}

func TestMapPage_Links_Empty(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
	}
	got := MapPage(src, "")
	if len(got.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(got.Links))
	}
}

func TestMapPage_Documents(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
		Documents: []scrapper_model.Document{
			{Url: "https://example.com/doc.pdf", MimeType: "application/pdf"},
		},
	}
	got := MapPage(src, "")
	if len(got.Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(got.Documents))
	}
	if got.Documents[0].Url != "https://example.com/doc.pdf" {
		t.Errorf("document Url = %q", got.Documents[0].Url)
	}
	if got.Documents[0].MimeType != "application/pdf" {
		t.Errorf("document MimeType = %q, want 'application/pdf'", got.Documents[0].MimeType)
	}
}

func TestMapPage_Documents_WithDescriptionSource(t *testing.T) {
	ds := scrapper_model.DocumentDescriptionSourceType("extracted")
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
		Documents: []scrapper_model.Document{
			{Url: "https://example.com/doc.pdf", DescriptionSource: &ds},
		},
	}
	got := MapPage(src, "")
	if got.Documents[0].DescriptionSource == nil {
		t.Fatal("expected DescriptionSource to be set")
	}
	if string(*got.Documents[0].DescriptionSource) != "extracted" {
		t.Errorf("DescriptionSource = %q, want 'extracted'", *got.Documents[0].DescriptionSource)
	}
}

func TestMapPage_Documents_NilDescriptionSource(t *testing.T) {
	src := scrapper_model.ScrappedPage{
		Url:    "https://example.com/page",
		Title:  "T",
		Blocks: []scrapper_model.ContentBlock{},
		Documents: []scrapper_model.Document{
			{Url: "https://example.com/doc.pdf"},
		},
	}
	got := MapPage(src, "")
	if got.Documents[0].DescriptionSource != nil {
		t.Errorf("expected nil DescriptionSource, got %v", got.Documents[0].DescriptionSource)
	}
}
