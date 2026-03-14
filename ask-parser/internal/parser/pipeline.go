package parser

import (
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
	"ask-parser/internal/parser/text/selector/scoring"
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
			&prune.VisibilityPruner{},       // DONE (omo-ri) implement [4]: убирает display:none/aria-hidden — если появится текст из скрытых элементов
			&prune.BoilerplateClassPruner{}, // TODO (aidweserd) implement [2]: class/id паттерны (ambox, hatnote, sidebar, banner)
		),
		segment.NewDOMBlockSegmenter(),
		filter.NewChainBlockCleaner(
			filter.NewMinCharLengthFilter(filter.DefaultMinCharLength),
			filter.NewMaxLinkDensityFilter(filter.DefaultMaxLinkDensity),
			// TODO (omo-ri) add [3]: filter.NewMinWordCountFilter      — страховка от однословного мусора; порог 3
			// TODO (aidweserd) add [7]: filter.NewBoilerplateContextFilter — контекстный фильтр (блоки после References/See also)
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
			// TODO (omo-ri) add [5]: &normalize.ValidUTF8Transform{}         — зачистка невалидных UTF-8 последовательностей
			// TODO (aidweserd) add [5]: &normalize.ControlCharStripTransform{} — зачистка управляющих символов
		),
	)

	return &GlobalPipeline{
		analyzer: analyze.NewStubAnalyzer(), // TODO implement [8]: HeuristicPageTypeAnalyzer — после стабилизации baseline фильтров
		extractors: []core.Extractor{
			extract.NewMetadataExtractor(), // TODO add [10]: JSON-LD парсинг (text/meta/jsonld.go) — articleBody, datePublished, author
			extract.NewTextExtractor(textPipeline),
			extract.NewLinkExtractor(),
			extract.NewDocumentExtractor(), // TODO implement [9]: извлечение PDF/DOCX ссылок
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
