package segment

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
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

// directTextContent extracts text from an HTML node, collecting
// text from the node itself and its inline children (spans, links, etc.)
// but not from block-level descendants.
func directTextContent(n *html.Node) string {
	var sb strings.Builder
	collectText(n, &sb)
	return strings.TrimSpace(sb.String())
}

func collectText(n *html.Node, sb *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		} else if c.Type == html.ElementNode {
			collectText(c, sb)
		}
	}
}
