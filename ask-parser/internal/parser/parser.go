package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

type htmlParser struct {
	sanitizer  sanitizer
	extractor  extractor
	normalizer normalizer
}

func NewParser() Parser {
	return &htmlParser{
		sanitizer:  &domSanitizer{},
		extractor:  &contentExtractor{},
		normalizer: &resultNormalizer{},
	}
}

func (p *htmlParser) Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	p.sanitizer.Sanitize(doc.Find("body"))
	result := p.extractor.Extract(doc, pageURL)
	result = p.normalizer.Normalize(result, pageURL)

	return buildPage(result, pageURL), nil
}

func buildPage(r *extractResult, pageURL string) *scrappermodel.ScrappedPage {
	page := scrappermodel.NewScrappedPage(pageURL, r.Meta.Title, r.Blocks)
	page.Description = r.Meta.Description
	page.PreviewUrl = r.Meta.PreviewURL
	page.IconUrl = r.Meta.IconURL
	page.Language = r.Meta.Language
	page.LastModified = r.Meta.LastModified

	if len(r.Links) > 0 {
		page.SetLinks(r.Links)
	}
	if len(r.Documents) > 0 {
		page.SetDocuments(r.Documents)
	}

	return page
}
