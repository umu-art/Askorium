package applib

import (
	"strings"
)

// LinkScorer присваивает кандидату оценку приоритета от 0.0 до 1.0.
type LinkScorer interface {
	Score(c URLCandidate) float64
}

// NullScorer всегда возвращает 0.5 — используется с FIFOFrontier,
// чтобы порядок не менялся.
type NullScorer struct{}

func (NullScorer) Score(_ URLCandidate) float64 { return 0.5 }

// --- URLPatternScorer -------------------------------------------------------

var valuablePathKeywords = []string{
	"docs", "doc", "guide", "tutorial", "api", "reference", "ref",
	"blog", "article", "articles", "post", "posts", "news",
	"learn", "learning", "wiki", "faq", "help",
}

var noisyPathKeywords = []string{
	"tag", "tags", "category", "categories", "author", "authors",
	"page", "feed", "rss", "search", "archive", "archives",
	"comment", "comments", "reply", "replies",
}

// URLPatternScorer оценивает URL по его структуре: глубина пути,
// наличие ценных/шумных ключевых слов в сегментах.
type URLPatternScorer struct{}

func (URLPatternScorer) Score(c URLCandidate) float64 {
	raw := strings.ToLower(c.URL)

	path := raw
	if idx := strings.Index(raw, "://"); idx >= 0 {
		rest := raw[idx+3:]
		if slash := strings.Index(rest, "/"); slash >= 0 {
			path = rest[slash:]
		} else {
			path = "/"
		}
	}
	if idx := strings.IndexAny(path, "?#"); idx >= 0 {
		path = path[:idx]
	}

	segments := splitPathSegments(path)
	depth := len(segments)

	// Базовая оценка по глубине: глубокие URL менее ценны
	depthScore := 1.0 - float64(depth)*0.12
	if depthScore < 0.1 {
		depthScore = 0.1
	}

	// Бонус/штраф за ключевые слова в пути
	keywordBonus := 0.0
	for _, seg := range segments {
		for _, kw := range valuablePathKeywords {
			if strings.Contains(seg, kw) {
				keywordBonus = 0.2
				break
			}
		}
		for _, kw := range noisyPathKeywords {
			if strings.Contains(seg, kw) {
				keywordBonus = -0.15
				break
			}
		}
	}

	score := depthScore + keywordBonus
	return clamp(score)
}

func splitPathSegments(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// --- AnchorTextScorer -------------------------------------------------------

// skipAnchorTexts — навигационные шаблоны без смысловой нагрузки.
var skipAnchorTexts = map[string]bool{
	"read more": true, "click here": true, "learn more": true,
	"here": true, "link": true, "this": true,
	"next": true, "previous": true, "prev": true,
	"home": true, "back": true, "top": true,
	"→": true, "»": true, "«": true, "←": true, "#": true,
	"edit": true, "share": true, "print": true, "subscribe": true,
	"more": true, "details": true, "view": true, "see": true,
}

// AnchorTextScorer оценивает ссылку по тексту её анкора.
// Описательные анкоры (2–8 слов) получают высокую оценку;
// навигационные шаблоны — низкую.
type AnchorTextScorer struct{}

func (AnchorTextScorer) Score(c URLCandidate) float64 {
	anchor := strings.TrimSpace(strings.ToLower(c.AnchorText))
	if anchor == "" || len(anchor) < 2 {
		return 0.3 // нет анкора — нейтральная оценка
	}
	if skipAnchorTexts[anchor] {
		return 0.1
	}
	words := strings.Fields(anchor)
	wc := len(words)
	switch {
	case wc >= 2 && wc <= 8:
		return 1.0 // описательный анкор
	case wc == 1:
		return 0.6
	default: // > 8 слов — возможно заголовок
		return 0.7
	}
}

// --- CompositeScorer --------------------------------------------------------

// CompositeScorer объединяет несколько scorer'ов с весами.
// Итоговый score = сумма (weight × scorer.Score(c)), нормализованная к [0,1].
type CompositeScorer struct {
	entries []weightedScorer
}

type weightedScorer struct {
	scorer LinkScorer
	weight float64
}

// NewCompositeScorer создаёт взвешенный composite scorer.
// Веса не обязаны суммироваться в 1 — нормализация выполняется автоматически.
func NewCompositeScorer(entries ...weightedScorer) *CompositeScorer {
	return &CompositeScorer{entries: entries}
}

// DefaultCompositeScorer — 0.5×URL + 0.5×Anchor.
func DefaultCompositeScorer() *CompositeScorer {
	return NewCompositeScorer(
		weightedScorer{URLPatternScorer{}, 0.5},
		weightedScorer{AnchorTextScorer{}, 0.5},
	)
}

func (cs *CompositeScorer) Score(c URLCandidate) float64 {
	if len(cs.entries) == 0 {
		return 0.5
	}
	totalWeight := 0.0
	weightedSum := 0.0
	for _, e := range cs.entries {
		weightedSum += e.weight * e.scorer.Score(c)
		totalWeight += e.weight
	}
	if totalWeight == 0 {
		return 0.5
	}
	return clamp(weightedSum / totalWeight)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
