package extract

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/core"
)

var _ core.Extractor = (*MetadataExtractor)(nil)

type MetadataExtractor struct{}

func NewMetadataExtractor() core.Extractor {
	return &MetadataExtractor{}
}

func (e *MetadataExtractor) Extract(ctx core.ExtractionContext) core.ExtractionResult {
	m := extractMetadata(ctx.FullDoc)
	return core.ExtractionResult{Metadata: &m}
}

func extractMetadata(doc *goquery.Document) core.Metadata {
	return core.Metadata{
		Title:        getTitle(doc),
		Description:  getDescription(doc),
		PreviewURL:   getPreviewURL(doc),
		IconURL:      getIconURL(doc),
		Language:     getLanguage(doc),
		LastModified: getLastModified(doc),
	}
}

func getTitle(doc *goquery.Document) string {
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", ""))
	}
	if title == "" {
		title = "Untitled"
	}
	return title
}

func getDescription(doc *goquery.Document) *string {
	desc := strings.TrimSpace(doc.Find(`meta[name="description"]`).AttrOr("content", ""))
	if desc == "" {
		desc = strings.TrimSpace(doc.Find(`meta[property="og:description"]`).AttrOr("content", ""))
	}
	return strPtr(desc)
}

func getPreviewURL(doc *goquery.Document) *string {
	return strPtr(strings.TrimSpace(doc.Find(`meta[property="og:image"]`).AttrOr("content", "")))
}

func getIconURL(doc *goquery.Document) *string {
	url := strings.TrimSpace(doc.Find(`link[rel="icon"]`).AttrOr("href", ""))
	if url == "" {
		url = strings.TrimSpace(doc.Find(`link[rel="shortcut icon"]`).AttrOr("href", ""))
	}
	return strPtr(url)
}

func getLanguage(doc *goquery.Document) *string {
	return strPtr(strings.TrimSpace(doc.Find("html").AttrOr("lang", "")))
}

func getLastModified(doc *goquery.Document) *time.Time {
	s := strings.TrimSpace(doc.Find(`meta[property="article:modified_time"]`).AttrOr("content", ""))
	if s == "" {
		s = strings.TrimSpace(doc.Find(`meta[http-equiv="last-modified"]`).AttrOr("content", ""))
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
