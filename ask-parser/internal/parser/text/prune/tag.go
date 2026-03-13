package prune

import (
	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*TagPruner)(nil)

var noiseTags = []string{
	// scripting / styling — never contain user-visible text
	"script", "style", "noscript", "template",
	// embedded content — opaque to text extraction
	"iframe", "object", "embed",
	// media — no text content; handled by DocumentExtractor on the full DOM.
	// <track> subtitles are ignored intentionally.
	// <img> has no text nodes; alt is an attribute, not body text.
	"video", "audio", "picture", "source", "canvas", "img",
	// vector / formula markup
	"svg", "math",
	// interactive controls
	"form",
	"button", "select", "input", "textarea",
}

type TagPruner struct{}

func (p *TagPruner) Prune(doc *goquery.Selection) {
	for _, tag := range noiseTags {
		doc.Find(tag).Remove()
	}
}
