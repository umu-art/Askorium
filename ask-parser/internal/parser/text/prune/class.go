package prune

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

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

// strongPositiveTokens are content signals strong enough to protect on their own.
var strongPositiveTokens = []string{
	"article", "content", "entry", "post", "story",
}

// weakPositiveTokens need at least two hits (or one strong) to protect.
var weakPositiveTokens = []string{
	"body", "column", "main", "text",
}

var (
	noiseClassPattern     = regexp.MustCompile(`(?i)(-ad-|\b(?:` + strings.Join(noiseTokens, "|") + `)\b)`)
	strongPositivePattern = regexp.MustCompile(`(?i)\b(?:` + strings.Join(strongPositiveTokens, "|") + `)\b`)
	weakPositivePattern   = regexp.MustCompile(`(?i)\b(?:` + strings.Join(weakPositiveTokens, "|") + `)\b`)
)

// BoilerplateClassPruner removes elements whose class/id matches boilerplate patterns,
// unless protected by content keywords (one strong token or two+ weak tokens).
type BoilerplateClassPruner struct{}

func (p *BoilerplateClassPruner) Prune(doc *goquery.Selection) {
	// Track nodes whose class/id carried positive protection so that
	// descendants inherit the protection even when they lack their own
	// positive tokens.  This prevents false deletions when the same noise
	// class (e.g. "promo-section") is reused at multiple nesting levels
	// but only the outermost one also carries a content marker.
	shielded := make(map[*html.Node]bool)

	doc.Find("[class],[id]").Each(func(_ int, s *goquery.Selection) {
		node := s.Get(0)

		// Never remove structural root elements — their classes are
		// CMS/template state tokens (e.g. "menu_right", "sidebar-second"),
		// not actual noise containers.
		if node.Data == "html" || node.Data == "body" {
			return
		}

		// If any ancestor is already shielded, this node is safe.
		for p := node.Parent; p != nil; p = p.Parent {
			if shielded[p] {
				return
			}
		}

		combined := s.AttrOr("class", "") + " " + s.AttrOr("id", "")
		if !noiseClassPattern.MatchString(combined) {
			return
		}
		if isPositiveProtected(combined) {
			shielded[node] = true
			return
		}
		// Do not remove elements that contain content landmarks or
		// substantial text paragraphs — a real ad/sidebar/widget never
		// wraps <article>, <main>, or 3+ paragraphs of content.
		if hasContentDescendants(s) {
			shielded[node] = true
			return
		}
		s.Remove()
	})
}

// hasContentDescendants returns true if s contains semantic content landmarks
// (<article>, <main>) or descendant elements whose class/id carries positive
// content signals. This prevents false deletions where a CSS framework reuses
// noise tokens at the wrapper level (e.g. "elementor-widget-wrap") while
// inner elements carry content markers (e.g. "elementor-widget-theme-post-content").
func hasContentDescendants(s *goquery.Selection) bool {
	if s.Find("article, main").Length() > 0 {
		return true
	}
	found := false
	s.Find("[class],[id]").EachWithBreak(func(_ int, child *goquery.Selection) bool {
		combined := child.AttrOr("class", "") + " " + child.AttrOr("id", "")
		if isPositiveProtected(combined) {
			found = true
			return false
		}
		return true
	})
	return found
}

// isPositiveProtected returns true if the combined class/id string contains
// at least one strong positive keyword, or two or more weak positive keywords.
// Uses \b word boundary matching (symmetric with noise detection).
func isPositiveProtected(combined string) bool {
	if strongPositivePattern.MatchString(combined) {
		return true
	}
	weakHits := len(weakPositivePattern.FindAllStringIndex(combined, -1))
	return weakHits >= 2
}
