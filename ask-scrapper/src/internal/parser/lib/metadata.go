package lib

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Metadata struct {
	Title        string
	Description  *string
	PreviewURL   *string
	IconURL      *string
	Language     *string
	LastModified *time.Time
}

func ExtractMetadata(doc *goquery.Document) *Metadata {
	return &Metadata{
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
	previewURL := strings.TrimSpace(doc.Find(`meta[property="og:image"]`).AttrOr("content", ""))
	return strPtr(previewURL)
}

func getIconURL(doc *goquery.Document) *string {
	iconURL := strings.TrimSpace(doc.Find(`link[rel="icon"]`).AttrOr("href", ""))
	if iconURL == "" {
		iconURL = strings.TrimSpace(doc.Find(`link[rel="shortcut icon"]`).AttrOr("href", ""))
	}
	return strPtr(iconURL)
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
