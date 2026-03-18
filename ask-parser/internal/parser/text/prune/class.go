package prune

import (
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*BoilerplateClassPruner)(nil)

// noiseTokens is the list of boilerplate class/id keywords.
// "-ad-" is handled separately — hyphen-bounded to avoid matching "add", "address", etc.
var noiseTokens = []string{
	"advert(?:isement)?", "ambox", "banner", "breadcrumb", "combx",
	"comment", "community", "cookie", "disqus", "gdpr", "hatnote",
	"legend", "menu", "modal", "overlay", "popup", "promo", "related",
	"rss", "share", "shoutbox", "sidebar", "skyscraper", "social",
	"sponsor", "widget",
}

// positiveTokens are exact class/id tokens that signal content.
// ^...$ matching prevents "main-menu" from being saved by "main".
var positiveTokens = []string{
	"article", "body", "column", "content", "entry",
	"main", "post", "story", "text",
}

var (
	noiseClassPattern    = regexp.MustCompile(`(?i)(-ad-|\b(?:` + strings.Join(noiseTokens, "|") + `)\b)`)
	positiveTokenPattern = regexp.MustCompile(`^(?i)(?:` + strings.Join(positiveTokens, "|") + `)$`)
)

// BoilerplateClassPruner removes elements whose class/id matches boilerplate patterns,
// unless any token in the same attribute is an exact content keyword.
type BoilerplateClassPruner struct{}

func (p *BoilerplateClassPruner) Prune(doc *goquery.Selection) {
	doc.Find("[class],[id]").Each(func(_ int, s *goquery.Selection) {
		combined := s.AttrOr("class", "") + " " + s.AttrOr("id", "")
		if noiseClassPattern.MatchString(combined) &&
			!slices.ContainsFunc(strings.Fields(combined), positiveTokenPattern.MatchString) {
			s.Remove()
		}
	})
}
