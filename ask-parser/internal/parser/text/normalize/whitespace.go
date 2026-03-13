package normalize

import (
	"regexp"

	"ask-parser/internal/parser/text"
)

var _ text.TextTransform = (*WhitespaceCollapseTransform)(nil)

var whitespaceRe = regexp.MustCompile(`\s+`)

type WhitespaceCollapseTransform struct{}

func (t *WhitespaceCollapseTransform) Transform(s string) string {
	return whitespaceRe.ReplaceAllString(s, " ")
}
