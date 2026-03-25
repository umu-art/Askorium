package segment

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
	"golang.org/x/net/html"

	"ask-parser/internal/parser/text"
)

var _ text.BlockSegmenter = (*EnhancedDOMBlockSegmenter)(nil)

// Extended selector: adds dt/dd (definition lists), figcaption (image captions).
// Tables are handled as whole units via TableSerializer, so td is excluded.
const enhancedBlockSelector = "h1, h2, h3, h4, h5, h6, p, blockquote, pre, dt, dd, figcaption"

// containerTags — elements that may contain direct text without a <p> wrapper.
var containerTags = []string{"section", "article", "main"}

const shortListItemThreshold = 10 // words — li shorter than this gets merged
const shortListThreshold = 0.5    // minimum percentage of short li to start merge

type EnhancedDOMBlockSegmenter struct {
	tableSerializer TableSerializer
}

func NewEnhancedDOMBlockSegmenter(ts TableSerializer) text.BlockSegmenter {
	if ts == nil {
		ts = &CompactTableSerializer{}
	}
	return &EnhancedDOMBlockSegmenter{tableSerializer: ts}
}

func (s *EnhancedDOMBlockSegmenter) Segment(doc *goquery.Selection, _ string) []text.RawBlock {
	var blocks []text.RawBlock
	idx := 0
	collected := make(map[*html.Node]bool)

	emit := func(node *goquery.Selection, rawText, tag string, blockType scrappermodel.ContentBlockType, headingLevel int32) {
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
	}

	// Pass 1: standard block elements (without li — handled separately)
	doc.Find(enhancedBlockSelector).Each(func(_ int, node *goquery.Selection) {
		rawNode := node.Get(0)
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
		blockType, headingLevel := classifyTagEnhanced(tag)
		emit(node, rawText, tag, blockType, headingLevel)
	})

	// Pass 2: lists — merge short lists into one block, otherwise emit individual li
	doc.Find("ul, ol").Each(func(_ int, list *goquery.Selection) {
		listNode := list.Get(0)
		if collected[listNode] {
			return
		}
		for p := listNode.Parent; p != nil; p = p.Parent {
			if collected[p] {
				return
			}
		}

		items := list.Children().Filter("li")
		if items.Length() == 0 {
			return
		}

		listItemCount, shortListItemCount := .0, .0
		items.Each(func(_ int, li *goquery.Selection) {
			listItemCount++
			if wordCount(li.Text()) > shortListItemThreshold {
				shortListItemCount++
			}
		})

		if shortListItemCount/listItemCount >= shortListThreshold {
			// Merge: concatenate all li texts with newline
			var sb strings.Builder
			items.Each(func(i int, li *goquery.Selection) {
				t := strings.TrimSpace(li.Text())
				if t == "" {
					return
				}
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(t)
			})
			if sb.Len() > 0 {
				collected[listNode] = true
				emit(list, sb.String(), goquery.NodeName(list), scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, 0)
			}
		} else {
			// Individual li blocks
			items.Each(func(_ int, li *goquery.Selection) {
				liNode := li.Get(0)
				rawText := li.Text()
				if rawText == "" || collected[liNode] {
					return
				}
				collected[liNode] = true
				emit(li, rawText, "li", scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, 0)
			})
		}
	})

	// Pass 3: tables — serialize via TableSerializer into a single block
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		tableNode := table.Get(0)
		if collected[tableNode] {
			return
		}
		for p := tableNode.Parent; p != nil; p = p.Parent {
			if collected[p] {
				return
			}
		}
		serialized := s.tableSerializer.Serialize(table)
		if serialized == "" {
			return
		}
		collected[tableNode] = true
		// Mark all descendant nodes as collected to prevent div/container passes from re-emitting.
		table.Find("*").Each(func(_ int, child *goquery.Selection) {
			collected[child.Get(0)] = true
		})
		emit(table, serialized, "table", scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0)
	})

	// Pass 4: text-bearing <div>s without block children (same as base)
	doc.Find("div").Each(func(_ int, node *goquery.Selection) {
		rawNode := node.Get(0)
		if collected[rawNode] {
			return
		}
		for p := rawNode.Parent; p != nil; p = p.Parent {
			if collected[p] {
				return
			}
		}
		if node.Find(enhancedBlockSelector+", li, ul, ol, table").Length() > 0 {
			return
		}
		rawText := directTextContent(rawNode)
		if rawText == "" {
			return
		}
		collected[rawNode] = true
		emit(node, rawText, "div", scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0)
	})

	// Pass 5: container elements (section/article/main) with direct text
	// but no block children — catches text outside <p> wrappers.
	for _, tag := range containerTags {
		doc.Find(tag).Each(func(_ int, node *goquery.Selection) {
			rawNode := node.Get(0)
			if collected[rawNode] {
				return
			}
			if node.Find(enhancedBlockSelector+", li, ul, ol, div, table").Length() > 0 {
				return
			}
			rawText := directTextContent(rawNode)
			if rawText == "" {
				return
			}
			collected[rawNode] = true
			emit(node, rawText, tag, scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0)
		})
	}

	// Link blocks and compute position
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

func classifyTagEnhanced(tag string) (scrappermodel.ContentBlockType, int32) {
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
	case "dt":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 0
	case "li", "dd":
		return scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, 0
	default:
		return scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0
	}
}
