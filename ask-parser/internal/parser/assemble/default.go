package assemble

import (
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/core"
)

var _ core.PageAssembler = (*DefaultPageAssembler)(nil)

type DefaultPageAssembler struct{}

func NewDefaultAssembler() core.PageAssembler {
	return &DefaultPageAssembler{}
}

func (a *DefaultPageAssembler) Assemble(pageURL string, results []core.ExtractionResult) *scrappermodel.ScrappedPage {
	var meta *core.Metadata
	var blocks []scrappermodel.ContentBlock
	var links []scrappermodel.Link
	var documents []scrappermodel.Document

	for _, r := range results {
		if r.Metadata != nil {
			meta = r.Metadata
		}
		if len(r.Blocks) > 0 {
			blocks = r.Blocks
		}
		if len(r.Links) > 0 {
			links = r.Links
		}
		if len(r.Documents) > 0 {
			documents = r.Documents
		}
	}

	title := "Untitled"
	if meta != nil && meta.Title != "" {
		title = meta.Title
	}

	page := scrappermodel.NewScrappedPage(pageURL, title, blocks)
	if meta != nil {
		page.Description = meta.Description
		page.PreviewUrl = meta.PreviewURL
		page.IconUrl = meta.IconURL
		page.Language = meta.Language
		page.LastModified = meta.LastModified
	}
	if len(links) > 0 {
		page.SetLinks(links)
	}
	if len(documents) > 0 {
		page.SetDocuments(documents)
	}

	return page
}
