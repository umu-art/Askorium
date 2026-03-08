package parser

import (
	"net/url"
	"regexp"
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

const minBlockTextLength = 2

var whitespaceRe = regexp.MustCompile(`\s+`)

type resultNormalizer struct{}

func (n *resultNormalizer) Normalize(result *extractResult, _ string) *extractResult {
	result.Blocks = filterShortBlocks(result.Blocks)
	result.Blocks = collapseWhitespace(result.Blocks)
	result.Links = normalizeLinks(result.Links)
	result.Links = deduplicateLinks(result.Links)
	return result
}

func filterShortBlocks(blocks []scrappermodel.ContentBlock) []scrappermodel.ContentBlock {
	filtered := make([]scrappermodel.ContentBlock, 0, len(blocks))
	for _, b := range blocks {
		if len([]rune(b.GetText())) >= minBlockTextLength {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func collapseWhitespace(blocks []scrappermodel.ContentBlock) []scrappermodel.ContentBlock {
	for i := range blocks {
		text := blocks[i].GetText()
		text = whitespaceRe.ReplaceAllString(text, " ")
		text = strings.TrimSpace(text)
		blocks[i].SetText(text)
	}
	return blocks
}

func normalizeLinks(links []scrappermodel.Link) []scrappermodel.Link {
	for i := range links {
		links[i].SetHref(normalizeURL(links[i].GetHref()))
	}
	return links
}

func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Fragment = ""

	q := u.Query()
	toRemove := make([]string, 0)
	for key := range q {
		if strings.HasPrefix(strings.ToLower(key), "utm_") {
			toRemove = append(toRemove, key)
		}
	}
	for _, key := range toRemove {
		q.Del(key)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func deduplicateLinks(links []scrappermodel.Link) []scrappermodel.Link {
	seen := make(map[string]bool)
	deduped := make([]scrappermodel.Link, 0, len(links))
	for _, l := range links {
		href := l.GetHref()
		if !seen[href] {
			seen[href] = true
			deduped = append(deduped, l)
		}
	}
	return deduped
}
