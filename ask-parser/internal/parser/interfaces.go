package parser

import (
	"time"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

// Parser converts raw HTML into a structured page.
// This is the only exported interface — used by handler for dependency injection and mocking.
type Parser interface {
	Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error)
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
