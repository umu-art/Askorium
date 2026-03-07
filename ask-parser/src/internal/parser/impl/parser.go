package impl

import (
	"ask-parser/internal/parser"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

type htmlParser struct {
	metadata parser.MetadataExtractor
	content  parser.ContentExtractor
}

func NewParser() parser.Parser {
	return &htmlParser{
		metadata: NewMetadataExtractor(),
		content:  NewContentExtractor(),
	}
}

func (p *htmlParser) Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	meta := p.metadata.Extract(doc)
	blocks, links, documents := p.content.Extract(doc, pageURL)

	page := scrappermodel.NewScrappedPage(pageURL, meta.Title, blocks)
	page.Description = meta.Description
	page.PreviewUrl = meta.PreviewURL
	page.IconUrl = meta.IconURL
	page.Language = meta.Language
	page.LastModified = meta.LastModified

	if len(links) > 0 {
		page.SetLinks(links)
	}
	if len(documents) > 0 {
		page.SetDocuments(documents)
	}

	return page, nil
}
