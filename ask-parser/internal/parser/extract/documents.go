package extract

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/core"
	"ask-parser/internal/parser/text/segment"
)

var _ core.Extractor = (*DocumentExtractor)(nil)

type DocumentExtractor struct{}

func NewDocumentExtractor() core.Extractor {
	return &DocumentExtractor{}
}

func (e *DocumentExtractor) Extract(ctx core.ExtractionContext) core.ExtractionResult {
	seen := make(map[string]struct{})
	var docs []scrappermodel.Document

	// Pass 1: <a href> → документные файлы и изображения-ссылки
	ctx.FullDoc.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		href := strings.TrimSpace(a.AttrOr("href", ""))
		if href == "" {
			return
		}
		resolved := resolveURL(href, ctx.PageURL)
		if !segment.IsDocumentURL(resolved) && !segment.IsImageURL(resolved) {
			return
		}
		if _, ok := seen[resolved]; ok {
			return
		}
		seen[resolved] = struct{}{}

		// Текст ссылки; если пустой — пробуем alt вложенного <img>
		anchorText := strings.TrimSpace(a.Text())
		if anchorText == "" {
			anchorText = strings.TrimSpace(a.Find("img").AttrOr("alt", ""))
		}
		docs = append(docs, segment.NewDocumentFromAnchor(resolved, anchorText))
	})

	// Pass 2: <img src> — inline-изображения, не охваченные Pass 1
	ctx.FullDoc.Find("img[src]").Each(func(_ int, img *goquery.Selection) {
		src := strings.TrimSpace(img.AttrOr("src", ""))
		if src == "" {
			return
		}
		resolved := resolveURL(src, ctx.PageURL)
		if !segment.IsImageURL(resolved) {
			return
		}
		if _, ok := seen[resolved]; ok {
			return
		}
		// Пропустить если родитель <a> уже собран в Pass 1
		if parentHref := resolveURL(img.Parent().AttrOr("href", ""), ctx.PageURL); parentHref != "" {
			if _, ok := seen[parentHref]; ok {
				return
			}
		}
		seen[resolved] = struct{}{}
		docs = append(docs, segment.NewDocumentFromImg(resolved, img.AttrOr("alt", "")))
	})

	return core.ExtractionResult{Documents: docs}
}
