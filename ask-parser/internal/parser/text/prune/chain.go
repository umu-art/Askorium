package prune

import (
	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*ChainPruner)(nil)

type ChainPruner struct {
	pruners []text.DOMPruner
}

func NewChainPruner(pruners ...text.DOMPruner) text.DOMPruner {
	return &ChainPruner{pruners: pruners}
}

func (p *ChainPruner) Prune(doc *goquery.Selection) {
	for _, pruner := range p.pruners {
		pruner.Prune(doc)
	}
}
