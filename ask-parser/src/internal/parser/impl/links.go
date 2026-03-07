package impl

import (
	"net/url"
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

const snippetWindow = 50

func resolveURL(href string, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

func classifyLink(href string, pageHost string) scrappermodel.LinkType {
	u, err := url.Parse(href)
	if err != nil || u.Host == "" {
		return scrappermodel.LINKTYPE_INTERNAL
	}
	if strings.EqualFold(u.Host, pageHost) {
		return scrappermodel.LINKTYPE_INTERNAL
	}
	return scrappermodel.LINKTYPE_EXTERNAL
}

func calcPosition(blockText string, anchorText string) int32 {
	idx := strings.Index(blockText, anchorText)
	if idx < 0 {
		return 0
	}
	return int32(idx)
}

func makeSnippet(blockText string, position int32, window int) string {
	pos := int(position)
	start := pos - window
	if start < 0 {
		start = 0
	}
	end := pos + window
	if end > len(blockText) {
		end = len(blockText)
	}
	return blockText[start:end]
}
