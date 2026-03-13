package core

import (
	"time"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

// PageType classifies the page before extraction begins.
type PageType int

const (
	PageTypeUnknown  PageType = iota
	PageTypeArticle           // article, post, documentation
	PageTypeListing           // catalogue, forum, list
	PageTypeDocument          // PDF, DOCX, etc. (non-HTML)
	PageTypeSPA               // empty without JS, no point in deep parsing
)

// Metadata holds page-level metadata produced by MetadataExtractor.
type Metadata struct {
	Title        string
	Description  *string
	PreviewURL   *string
	IconURL      *string
	Language     *string
	LastModified *time.Time
}

// ExtractionContext is the immutable input shared across all Extractors.
// FullDoc must not be mutated — clone before pruning.
type ExtractionContext struct {
	FullDoc  *goquery.Document
	PageURL  string
	PageType PageType
}

// ExtractionResult is the partial output of a single Extractor.
// Nil / empty fields mean "this extractor did not produce this kind of data".
type ExtractionResult struct {
	Metadata  *Metadata
	Blocks    []scrappermodel.ContentBlock
	Links     []scrappermodel.Link
	Documents []scrappermodel.Document
}
