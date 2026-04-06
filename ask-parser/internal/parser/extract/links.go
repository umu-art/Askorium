package extract

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/core"
	"ask-parser/internal/parser/text/segment"
)


var _ core.Extractor = (*LinkExtractor)(nil)

type LinkExtractor struct{}

func NewLinkExtractor() core.Extractor {
	return &LinkExtractor{}
}

func (e *LinkExtractor) Extract(ctx core.ExtractionContext) core.ExtractionResult {
	pageHost := ""
	if u, err := url.Parse(ctx.PageURL); err == nil {
		pageHost = u.Host
	}

	seen := make(map[string]struct{})
	var links []scrappermodel.Link

	ctx.FullDoc.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		href := strings.TrimSpace(a.AttrOr("href", ""))
		if href == "" ||
			strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "javascript:") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "tel:") ||
			strings.HasPrefix(href, "data:") {
			return
		}

		resolved := resolveURL(href, ctx.PageURL)
		if segment.IsDocumentURL(resolved) {
			return
		}

		normalized := normalizeURL(resolved)
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}

		linkType := classifyLink(normalized, pageHost)
		link := *scrappermodel.NewLink(normalized, linkType, 0)

		anchorText := strings.TrimSpace(a.Text())
		if anchorText != "" {
			link.SetAnchorText(anchorText)
		}

		links = append(links, link)
	})

	return core.ExtractionResult{Links: links}
}

func classifyLink(href, pageHost string) scrappermodel.LinkType {
	u, err := url.Parse(href)
	if err != nil || u.Host == "" {
		return scrappermodel.LINKTYPE_INTERNAL
	}
	linkHost := strings.ToLower(u.Host)
	base := strings.ToLower(pageHost)
	if linkHost == base || strings.HasSuffix(linkHost, "."+base) {
		return scrappermodel.LINKTYPE_INTERNAL
	}
	return scrappermodel.LINKTYPE_EXTERNAL
}

