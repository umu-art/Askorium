package normalize

import (
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.TextTransform = (*TrimTransform)(nil)

type TrimTransform struct{}

func (t *TrimTransform) Transform(s string) string {
	return strings.TrimSpace(s)
}
