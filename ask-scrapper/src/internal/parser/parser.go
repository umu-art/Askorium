package parser

import (
	"ask-scrapper/internal/parser/lib"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

type Parser interface {
	Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error)
}

type htmlParser struct{}

func NewParser() Parser {
	return &htmlParser{}
}

func (p *htmlParser) Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	meta := lib.ExtractMetadata(doc)
	page := scrappermodel.NewScrappedPage(pageURL, meta.Title, []scrappermodel.ContentBlock{})
	page.Description = meta.Description
	page.PreviewUrl = meta.PreviewURL
	page.IconUrl = meta.IconURL
	page.Language = meta.Language
	page.LastModified = meta.LastModified
	return page, nil
}
