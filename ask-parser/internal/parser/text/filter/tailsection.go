package filter

import (
	"fmt"
	"log/slog"
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

// sectionBreakPatterns are heading texts that signal the end of main content.
// When a heading block in the tail portion of the page matches one of these,
// everything from that block onward is dropped.
var sectionBreakPatterns = []string{
	// News / related content sections (RU)
	"новости",
	"все новости",
	"ещё новости",
	"еще новости",
	"последние новости",
	"читайте также",
	"смотрите также",
	"похожие материалы",
	"похожие статьи",
	"рекомендуем",
	"вам также может быть интересно",

	// Other-programs / other-courses blocks (common pattern)
	"другие программы",
	"другие бакалаврские программы",
	"другие магистерские программы",
	"другие курсы",

	// Related / recommended sections (EN)
	"related articles",
	"related posts",
	"related stories",
	"more stories",
	// "more from" omitted — too broad, can match author byline sections
	"you may also like",
	"recommended for you",
	"also read",
	"further reading",
	"other programs",

	// Trending / popular sections (EN)
	"trending",
	"trending now",
	"trending news",
	"most popular",
	"most read",
	"popular stories",
	"what's hot",

	// Comments sections (EN/RU) — only specific headings, bare "comments"
	// is too broad (e.g. articles about commenting systems)
	"top rated comments",
	"reader comments",
	"user comments",
	"комментарии",
}

const (
	// minTruncDOMPosition: minimum DOM-relative position (RawBlock.Position)
	// to consider a section break. Both this AND minTruncBlockFrac must be
	// satisfied to prevent false positives. Set relatively high because
	// sidebars (Most Popular, Related) are often interleaved with content
	// in the DOM at low positions.
	minTruncDOMPosition = 0.30

	// minTruncBlockFrac: the block must be in the latter portion (by index)
	// of all blocks. DOM Position can be misleading when the noise section
	// has shallow DOM depth, so this acts as a second guard.
	// Note: total block count includes footer/nav blocks that will be filtered
	// later, so the effective content fraction is higher than this number.
	minTruncBlockFrac = 0.20
)

// TailSectionTruncator removes everything from a section-break heading onward.
// It implements BlockCleaner (not BlockFilter) because it needs to see the full
// block list and must run BEFORE per-block filters — the heading markers are
// often short and would be dropped by MinCharLengthFilter before a per-block
// filter could see them.
type TailSectionTruncator struct{}

func (t *TailSectionTruncator) Clean(blocks []text.RawBlock) []text.RawBlock {
	n := len(blocks)
	if n == 0 {
		return blocks
	}

	minIdx := int(float64(n) * minTruncBlockFrac)

	for i, b := range blocks {
		if i < minIdx || b.Position < minTruncDOMPosition {
			continue
		}
		if b.Type != scrappermodel.CONTENTBLOCKTYPE_HEADING {
			continue
		}

		lower := strings.ToLower(strings.TrimSpace(b.Text))
		for _, pattern := range sectionBreakPatterns {
			if lower == pattern || strings.Contains(lower, pattern) {
				if slog.Default().Enabled(nil, slog.LevelDebug) {
					slog.Debug(fmt.Sprintf(
						"tailsection TRUNCATE at [%d] <%s> pos=%.2f idx=%d/%d: %q (matched %q)",
						b.Idx, b.TagName, b.Position, i, n, b.Text, pattern))
				}
				return blocks[:i]
			}
		}
	}
	return blocks
}
