package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/core"
)

type GlobalPipeline struct {
	analyzer   core.PageTypeAnalyzer
	extractors []core.Extractor
	assembler  core.PageAssembler
}

func NewParser() Parser {
	// TODO: wire up components in subsequent steps
	return &GlobalPipeline{}
}

func (p *GlobalPipeline) Parse(rawHTML, pageURL string) (*scrappermodel.ScrappedPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var pt core.PageType
	if p.analyzer != nil {
		pt = p.analyzer.Analyze(doc, pageURL)
	}

	ctx := core.ExtractionContext{FullDoc: doc, PageURL: pageURL, PageType: pt}

	results := make([]core.ExtractionResult, 0, len(p.extractors))
	for _, ext := range p.extractors {
		results = append(results, ext.Extract(ctx))
	}

	if p.assembler != nil {
		return p.assembler.Assemble(pageURL, results), nil
	}

	return scrappermodel.NewScrappedPage(pageURL, "Untitled", nil), nil
}
