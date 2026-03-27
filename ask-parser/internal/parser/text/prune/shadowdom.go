package prune

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*ShadowDOMUnwrapPruner)(nil)

// ShadowDOMUnwrapPruner unwraps <template shadowrootmode> elements,
// replacing them with their children. Declarative Shadow DOM uses these
// templates to hold rendered content, but goquery treats <template> as
// inert, making the content invisible to downstream extraction.
type ShadowDOMUnwrapPruner struct{}

func (p *ShadowDOMUnwrapPruner) Prune(doc *goquery.Selection) {
	doc.Find("template[shadowrootmode]").Each(func(_ int, s *goquery.Selection) {
		node := s.Get(0)
		parent := node.Parent
		if parent == nil {
			return
		}
		// Collect children first, then detach and re-insert before <template>.
		var children []*html.Node
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			children = append(children, child)
		}
		for _, child := range children {
			node.RemoveChild(child)
			parent.InsertBefore(child, node)
		}
		parent.RemoveChild(node)
	})
}
