package analyze

import (
	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/core"
)

var _ core.PageTypeAnalyzer = (*StubAnalyzer)(nil)

// StubAnalyzer is a stub for V1 — always returns PageTypeUnknown.
// V2: classify by <main>/<article>, ul/li density, empty <body>, URL extension.
type StubAnalyzer struct{}

func NewStubAnalyzer() core.PageTypeAnalyzer {
	return &StubAnalyzer{}
}

func (a *StubAnalyzer) Analyze(_ *goquery.Document, _ string) core.PageType {
	return core.PageTypeUnknown
}
