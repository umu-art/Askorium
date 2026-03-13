package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*BoilerplateContextFilter)(nil)

// BoilerplateContextFilter is a no-op stub for V1 (always passes).
// V2: use prev/next block context to filter boilerplate (Boilerpipe-style).
type BoilerplateContextFilter struct{}

func (f *BoilerplateContextFilter) Keep(_ text.RawBlock) bool { return true }
