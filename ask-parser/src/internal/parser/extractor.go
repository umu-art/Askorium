package parser

import (
	"time"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

type Metadata struct {
	Title        string
	Description  *string
	PreviewURL   *string
	IconURL      *string
	Language     *string
	LastModified *time.Time
}

type MetadataExtractor interface {
	Extract(doc *goquery.Document) *Metadata
}

type ContentExtractor interface {
	Extract(doc *goquery.Document, pageURL string) ([]scrappermodel.ContentBlock, []scrappermodel.Link, []scrappermodel.Document)
}
