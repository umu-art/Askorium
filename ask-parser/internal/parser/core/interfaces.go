package core

import (
	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

// PageTypeAnalyzer classifies the page by DOM and URL.
// Must not mutate the document.
type PageTypeAnalyzer interface {
	Analyze(doc *goquery.Document, pageURL string) PageType
}

// Extractor pulls one category of data from the shared ExtractionContext.
type Extractor interface {
	Extract(ctx ExtractionContext) ExtractionResult
}

// PageAssembler merges all ExtractionResults into the final ScrappedPage.
type PageAssembler interface {
	Assemble(pageURL string, results []ExtractionResult) *scrappermodel.ScrappedPage
}
