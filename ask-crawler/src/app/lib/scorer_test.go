package applib

import (
	"testing"
)

func TestURLPatternScorer(t *testing.T) {
	s := URLPatternScorer{}

	// Абсолютные границы
	bounds := []struct {
		url      string
		minScore float64
		maxScore float64
	}{
		// очень глубокий URL без ценных ключевых слов — низкий score
		{"https://example.com/a/b/c/d/e/f", 0.0, 0.4},
		// корневая страница — высокий score
		{"https://example.com/", 0.8, 1.0},
		// ценное ключевое слово docs
		{"https://example.com/docs", 0.8, 1.0},
	}
	for _, tc := range bounds {
		got := s.Score(URLCandidate{URL: tc.url})
		if got < tc.minScore || got > tc.maxScore {
			t.Errorf("URLPatternScorer(%q) = %.2f, want [%.2f, %.2f]", tc.url, got, tc.minScore, tc.maxScore)
		}
	}

	// Относительный порядок: ценные страницы должны обгонять шумные
	scoreValuable := s.Score(URLCandidate{URL: "https://example.com/docs/api"})
	scoreNoisy := s.Score(URLCandidate{URL: "https://example.com/category/author"})
	if scoreNoisy >= scoreValuable {
		t.Errorf("noisy URL (%.2f) should score less than valuable URL (%.2f)", scoreNoisy, scoreValuable)
	}
}

func TestAnchorTextScorer(t *testing.T) {
	s := AnchorTextScorer{}

	cases := []struct {
		anchor   string
		minScore float64
		maxScore float64
	}{
		{"", 0.2, 0.4},                         // нет анкора — нейтральный
		{"read more", 0.0, 0.2},                 // навигационный шаблон
		{"next", 0.0, 0.2},                      // навигационный
		{"Introduction to Machine Learning", 0.6, 1.0}, // описательный
		{"API Reference Guide", 0.9, 1.0},       // 3 слова — отлично
		{"Docs", 0.5, 0.7},                      // одно слово
	}

	for _, tc := range cases {
		got := s.Score(URLCandidate{AnchorText: tc.anchor})
		if got < tc.minScore || got > tc.maxScore {
			t.Errorf("AnchorTextScorer(%q) = %.2f, want [%.2f, %.2f]", tc.anchor, got, tc.minScore, tc.maxScore)
		}
	}
}

func TestNullScorer(t *testing.T) {
	s := NullScorer{}
	got := s.Score(URLCandidate{URL: "https://example.com/anything"})
	if got != 0.5 {
		t.Errorf("NullScorer.Score = %.2f, want 0.5", got)
	}
}

func TestCompositeScorer_Bounds(t *testing.T) {
	s := DefaultCompositeScorer()
	urls := []string{
		"https://example.com/",
		"https://example.com/docs/api/reference",
		"https://example.com/tag/category/author/page/100",
	}
	anchors := []string{"", "read more", "Complete API Reference Guide"}
	for _, u := range urls {
		for _, a := range anchors {
			score := s.Score(URLCandidate{URL: u, AnchorText: a})
			if score < 0 || score > 1 {
				t.Errorf("CompositeScorer score out of [0,1]: url=%q anchor=%q score=%.3f", u, a, score)
			}
		}
	}
}
