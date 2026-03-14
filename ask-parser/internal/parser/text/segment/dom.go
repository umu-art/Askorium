package segment

import (
	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
	"golang.org/x/net/html"

	"ask-parser/internal/parser/text"
)

var _ text.BlockSegmenter = (*DOMBlockSegmenter)(nil)

const blockSelector = "h1, h2, h3, h4, h5, h6, p, li, blockquote, pre, td"

type DOMBlockSegmenter struct{}

func NewDOMBlockSegmenter() text.BlockSegmenter {
	return &DOMBlockSegmenter{}
}

func (s *DOMBlockSegmenter) Segment(doc *goquery.Selection, _ string) []text.RawBlock {
	var blocks []text.RawBlock
	idx := 0
	collected := make(map[*html.Node]bool)

	// Pass 1: standard block-level elements
	doc.Find(blockSelector).Each(func(_ int, node *goquery.Selection) {
		rawNode := node.Get(0)

		// Skip if an ancestor is already collected (prevents double-counting)
		for p := rawNode.Parent; p != nil; p = p.Parent {
			if collected[p] {
				return
			}
		}

		rawText := node.Text()
		if rawText == "" {
			return
		}

		collected[rawNode] = true

		tag := goquery.NodeName(node)
		blockType, headingLevel := classifyTag(tag)

		blocks = append(blocks, text.RawBlock{
			Idx:          idx,
			HTMLId:       htmlIdAttr(node),
			TagName:      tag,
			Text:         rawText,
			WordCount:    wordCount(rawText),
			LinkDensity:  linkDensity(node),
			DOMDepth:     domDepth(node),
			AncestorTags: ancestorTags(node),
			Type:         blockType,
			HeadingLevel: headingLevel,
		})
		idx++
	})

	// Pass 2: text-bearing <div>s that have no block-level children
	doc.Find("div").Each(func(_ int, node *goquery.Selection) {
		rawNode := node.Get(0)
		if collected[rawNode] {
			return
		}
		// Skip if an ancestor is already collected
		for p := rawNode.Parent; p != nil; p = p.Parent {
			if collected[p] {
				return
			}
		}
		// Skip divs that contain block-level children
		if node.Find(blockSelector).Length() > 0 {
			return
		}
		rawText := directTextContent(rawNode)
		if rawText == "" {
			return
		}

		collected[rawNode] = true

		blocks = append(blocks, text.RawBlock{
			Idx:          idx,
			HTMLId:       htmlIdAttr(node),
			TagName:      "div",
			Text:         rawText,
			WordCount:    wordCount(rawText),
			LinkDensity:  linkDensity(node),
			DOMDepth:     domDepth(node),
			AncestorTags: ancestorTags(node),
			Type:         scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH,
			HeadingLevel: 0,
		})
		idx++
	})

	n := len(blocks)
	for i := range blocks {
		blocks[i].Position = float64(i) / float64(max(n, 1))
		if i > 0 {
			blocks[i].PrevBlock = &blocks[i-1]
		}
		if i < n-1 {
			blocks[i].NextBlock = &blocks[i+1]
		}
	}

	return blocks
}

func classifyTag(tag string) (scrappermodel.ContentBlockType, int32) {
	switch tag {
	case "h1":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 1
	case "h2":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 2
	case "h3":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 3
	case "h4":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 4
	case "h5":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 5
	case "h6":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 6
	case "li":
		return scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, 0
	default:
		return scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0
	}
}

func htmlIdAttr(node *goquery.Selection) string {
	if id, exists := node.Attr("id"); exists && id != "" {
		return id
	}
	return ""
}

// ancestorTags collects semantic tag names by walking up the DOM.
var semanticTags = map[string]bool{
	"main": true, "article": true, "section": true,
	"aside": true, "nav": true, "footer": true, "header": true,
}

func ancestorTags(node *goquery.Selection) []string {
	var tags []string
	node.Parents().Each(func(_ int, parent *goquery.Selection) {
		tag := goquery.NodeName(parent)
		if semanticTags[tag] {
			tags = append(tags, tag)
		}
	})
	return tags
}

func domDepth(node *goquery.Selection) int {
	depth := 0
	node.Parents().Each(func(_ int, _ *goquery.Selection) {
		depth++
	})
	return depth
}
