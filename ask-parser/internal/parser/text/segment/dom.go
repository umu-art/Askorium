package segment

import (
	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

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

	doc.Find(blockSelector).Each(func(_ int, node *goquery.Selection) {
		rawText := node.Text()
		if rawText == "" {
			return
		}

		tag := goquery.NodeName(node)
		blockType, headingLevel := classifyTag(tag)
		htmlId := htmlIdAttr(node)
		depth := domDepth(node)

		blocks = append(blocks, text.RawBlock{
			Idx:          idx,
			HTMLId:       htmlId,
			TagName:      tag,
			Text:         rawText,
			WordCount:    wordCount(rawText),
			LinkDensity:  linkDensity(node),
			DOMDepth:     depth,
			Type:         blockType,
			HeadingLevel: headingLevel,
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

func domDepth(node *goquery.Selection) int {
	depth := 0
	node.Parents().Each(func(_ int, _ *goquery.Selection) {
		depth++
	})
	return depth
}
