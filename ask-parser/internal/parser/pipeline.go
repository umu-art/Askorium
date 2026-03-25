package parser

import (
	"ask-parser/internal/parser/text/selector/scoring"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/analyze"
	"ask-parser/internal/parser/assemble"
	"ask-parser/internal/parser/core"
	"ask-parser/internal/parser/extract"
	"ask-parser/internal/parser/text"
	"ask-parser/internal/parser/text/filter"
	"ask-parser/internal/parser/text/normalize"
	"ask-parser/internal/parser/text/prune"
	"ask-parser/internal/parser/text/segment"
	"ask-parser/internal/parser/text/selector"
)

type GlobalPipeline struct {
	analyzer   core.PageTypeAnalyzer
	extractors []core.Extractor
	assembler  core.PageAssembler
}

func NewParser() Parser {
	textPipeline := text.NewDefaultTextPipeline(
		prune.NewChainPruner(
			&prune.TagPruner{},
			&prune.SemanticNoisePruner{},
			&prune.VisibilityPruner{},
			&prune.BoilerplateClassPruner{},
		),
		segment.NewEnhancedDOMBlockSegmenter(&segment.CompactTableSerializer{}),
		filter.NewChainBlockCleaner(
			filter.NewMinCharLengthFilter(filter.DefaultMinCharLength),
			filter.NewMaxLinkDensityFilter(filter.DefaultMaxLinkDensity),
			filter.NewMinWordCountFilter(filter.DefaultMinWordCount),
			filter.NewBoilerplateContextFilter(),
		),
		selector.NewScoredBlockSelector(
			selector.DefaultClusterGap,
			scoring.NewSemanticMarkupStrategy(),
			scoring.NewReadabilityScoringStrategy(),
			scoring.NewTextDensityStrategy(),
		),
		normalize.NewChainTextNormalizer(
			&normalize.WhitespaceCollapseTransform{},
			&normalize.TrimTransform{},
			&normalize.ValidUTF8Transform{},
			&normalize.ControlCharStripTransform{},
		),
	)

	return &GlobalPipeline{
		analyzer: analyze.NewStubAnalyzer(), // TODO implement [8]: HeuristicPageTypeAnalyzer — после стабилизации baseline фильтров
		extractors: []core.Extractor{
			extract.NewMetadataExtractor(), // TODO add [10]: JSON-LD парсинг (text/meta/jsonld.go) — articleBody, datePublished, author
			extract.NewTextExtractor(textPipeline),
			extract.NewLinkExtractor(),
			extract.NewDocumentExtractor(),
		},
		assembler: assemble.NewDefaultAssembler(),
	}
}

// NewDocumentParser returns a simplified parser for pre-converted document HTML
// (PDF, DOCX, XLSX). Boilerplate filters are omitted — documents have no navigation.
func NewDocumentParser() Parser {
	textPipeline := text.NewDefaultTextPipeline(
		prune.NewChainPruner(
			&prune.TagPruner{},
			&prune.VisibilityPruner{},
		),
		segment.NewDOMBlockSegmenter(),
		filter.NewChainBlockCleaner(
			filter.NewMinCharLengthFilter(filter.DefaultMinCharLength),
		),
		&selector.PassThroughSelector{},
		normalize.NewChainTextNormalizer(
			&normalize.WhitespaceCollapseTransform{},
			&normalize.TrimTransform{},
			&normalize.ValidUTF8Transform{},
			&normalize.ControlCharStripTransform{},
		),
	)

	return &GlobalPipeline{
		analyzer: analyze.NewStubAnalyzer(),
		extractors: []core.Extractor{
			extract.NewMetadataExtractor(),
			extract.NewTextExtractor(textPipeline),
		},
		assembler: assemble.NewDefaultAssembler(),
	}
}

func (p *GlobalPipeline) Parse(rawHTML, pageURL string) (*scrappermodel.ScrappedPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	pt := p.analyzer.Analyze(doc, pageURL)
	ctx := core.ExtractionContext{FullDoc: doc, PageURL: pageURL, PageType: pt}

	results := make([]core.ExtractionResult, 0, len(p.extractors))
	for _, ext := range p.extractors {
		results = append(results, ext.Extract(ctx))
	}

	return p.assembler.Assemble(pageURL, results), nil
}
