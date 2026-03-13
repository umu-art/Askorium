package extract

import "ask-parser/internal/parser/core"

var _ core.Extractor = (*DocumentExtractor)(nil)

// DocumentExtractor is a stub for V1.
// V2: extract <a href> links to files (.pdf, .docx, etc.) and <img> elements.
type DocumentExtractor struct{}

func NewDocumentExtractor() core.Extractor {
	return &DocumentExtractor{}
}

func (e *DocumentExtractor) Extract(_ core.ExtractionContext) core.ExtractionResult {
	return core.ExtractionResult{}
}
