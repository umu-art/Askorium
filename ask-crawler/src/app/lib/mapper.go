package applib

import (
	crawler_model "github.com/omo-ri/askorium/go-crawler-api"
	scrapper_model "github.com/omo-ri/askorium/go-parser-api"
)

func MapPage(p scrapper_model.ScrappedPage) crawler_model.ScrappedPage {
	page := crawler_model.ScrappedPage{
		Url:         p.Url,
		Title:       p.Title,
		Description: p.Description,
		PreviewUrl:  p.PreviewUrl,
		IconUrl:     p.IconUrl,
		Language:    p.Language,
	}

	if p.LastModified != nil {
		page.LastModified = p.LastModified
	}

	for _, b := range p.Blocks {
		page.Blocks = append(page.Blocks, crawler_model.ContentBlock{
			HtmlId:       b.HtmlId,
			Type:         crawler_model.ContentBlockType(b.Type),
			HeadingLevel: b.HeadingLevel,
			Text:         b.Text,
		})
	}

	for _, l := range p.Links {
		page.Links = append(page.Links, crawler_model.Link{
			Href:           l.Href,
			Type:           crawler_model.LinkType(l.Type),
			AnchorText:     l.AnchorText,
			ContextBlockId: l.ContextBlockId,
			Position:       l.Position,
			Snippet:        l.Snippet,
		})
	}

	for _, d := range p.Documents {
		var descSource *crawler_model.DocumentDescriptionSourceType
		if d.DescriptionSource != nil {
			src := crawler_model.DocumentDescriptionSourceType(*d.DescriptionSource)
			descSource = &src
		}
		page.Documents = append(page.Documents, crawler_model.Document{
			Url:               d.Url,
			MimeType:          d.MimeType,
			SizeBytes:         d.SizeBytes,
			ExtractedText:     d.ExtractedText,
			Description:       d.Description,
			DescriptionSource: descSource,
		})
	}

	return page
}
