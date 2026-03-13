package segment

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func wordCount(text string) int {
	return len(strings.Fields(text))
}

func linkDensity(node *goquery.Selection) float64 {
	total := wordCount(node.Text())
	if total == 0 {
		return 0
	}
	return float64(wordCount(node.Find("a").Text())) / float64(total)
}
