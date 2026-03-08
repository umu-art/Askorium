package parser

import (
	"time"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

// Parser converts raw HTML into a structured page.
// This is the only exported interface — used by handler for dependency injection and mocking.
type Parser interface {
	Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error)
}

// --- internal pipeline stages ---

// sanitizer removes noise nodes from the DOM before extraction.
type sanitizer interface {
	Sanitize(root *goquery.Selection)
}

// extractor pulls structured data from a sanitized DOM.
type extractor interface {
	Extract(doc *goquery.Document, pageURL string) *extractResult
}

// normalizer cleans and deduplicates extraction results.
type normalizer interface {
	Normalize(result *extractResult, pageURL string) *extractResult
}

// --- internal data types ---

type metadata struct {
	Title        string
	Description  *string
	PreviewURL   *string
	IconURL      *string
	Language     *string
	LastModified *time.Time
}

type extractResult struct {
	Meta      metadata
	Blocks    []scrappermodel.ContentBlock
	Links     []scrappermodel.Link
	Documents []scrappermodel.Document
}
