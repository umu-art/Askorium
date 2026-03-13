package prune

import (
	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*BoilerplateClassPruner)(nil)

// BoilerplateClassPruner is a no-op stub for V1.
// V2: remove elements whose class/id matches ad/banner/sidebar/menu/cookie patterns.
type BoilerplateClassPruner struct{}

func (p *BoilerplateClassPruner) Prune(_ *goquery.Selection) {}
