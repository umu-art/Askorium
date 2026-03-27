package merge

import (
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

// QAPairingRule merges a question block (short, ends with ?) with its following answer.
type QAPairingRule struct{}

func (QAPairingRule) Apply(blocks []text.RawBlock, cfg MergeConfig) []text.RawBlock {
	var out []text.RawBlock
	n := len(blocks)

	for i := 0; i < n; i++ {
		b := blocks[i]

		if i+1 < n && isQuestion(b, cfg.MaxQuestionWords) && !isQuestion(blocks[i+1], cfg.MaxQuestionWords) {
			// Merge question + answer
			answer := blocks[i+1]
			b.Text = b.Text + "\n" + answer.Text
			b.WordCount = wordCount(b.Text)
			b.Type = scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH
			i++ // skip the answer block
		}

		out = append(out, b)
	}
	return out
}

func isQuestion(b text.RawBlock, maxWords int) bool {
	if b.WordCount > maxWords {
		return false
	}
	trimmed := strings.TrimRight(b.Text, " \t\n")
	return strings.HasSuffix(trimmed, "?") || strings.HasSuffix(trimmed, "？")
}
