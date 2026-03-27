package selector

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

var _ text.BlockSelector = (*ScoredBlockSelector)(nil)

const (
	DefaultClusterGap    = 5
	DefaultFloorQuantile = 0.15 // base floor = P15 of non-zero scores
	DefaultMergeDistance = 8    // max gap between clusters to consider merging
	DefaultMergeMinRatio = 0.3  // gap avg score must be >= floor * this to merge
	DefaultKeepRatio     = 0.3  // keep clusters with totalScore >= best * this

	// Adaptive floor: on large pages, P15 is too low (noise teasers pull it
	// down), letting them bridge gaps and form one giant cluster.
	// effective_quantile = base + slope * min(1, blockCount / scale)
	adaptiveFloorSlope = 0.10 // max quantile bump
	adaptiveFloorScale = 50.0 // blocks at which full bump is reached

)

type cluster struct {
	start, end int
	totalScore float64
}

type ScoredBlockSelector struct {
	strategies    []text.ScoringStrategy
	maxGap        int
	floorQuantile float64
}

func NewScoredBlockSelector(maxGap int, strategies ...text.ScoringStrategy) *ScoredBlockSelector {
	return &ScoredBlockSelector{
		strategies:    strategies,
		maxGap:        maxGap,
		floorQuantile: DefaultFloorQuantile,
	}
}

func (s *ScoredBlockSelector) Select(blocks []text.RawBlock) []text.RawBlock {
	if len(blocks) == 0 {
		return blocks
	}

	for _, strategy := range s.strategies {
		blocks = strategy.Score(blocks)
	}

	floor := s.computeFloor(blocks)
	clusters := s.findClusters(blocks, floor)
	clusters = mergeClusters(clusters, blocks, floor)
	kept := selectClusters(clusters)

	if slog.Default().Enabled(nil, slog.LevelDebug) {
		s.logScores(blocks, kept, floor)
	}

	return collectMultiContent(blocks, kept)
}

// computeFloor returns the adaptive floor score.
// Base quantile is floorQuantile (0.15). On pages with many blocks, noise
// teasers inflate the block count and pull P15 too low, letting them bridge
// gaps and form one giant cluster. The quantile scales up with block count:
//
//	effective = base + adaptiveSlope * min(1.0, nonzeroCount / adaptiveScale)
//
// Examples: 10 blocks → ~P17, 50 blocks → P25, 100+ blocks → P25.
// Non-zero filtering avoids stopword-gated blocks (score=0) from
// dragging the floor down to zero and making it useless.
func (s *ScoredBlockSelector) computeFloor(blocks []text.RawBlock) float64 {
	var nonzero []float64
	for _, b := range blocks {
		if b.Score > 0 {
			nonzero = append(nonzero, b.Score)
		}
	}
	if len(nonzero) == 0 {
		return 0
	}
	sort.Float64s(nonzero)

	// Adaptive quantile: raise floor on large pages
	ratio := float64(len(nonzero)) / adaptiveFloorScale
	if ratio > 1.0 {
		ratio = 1.0
	}
	quantile := s.floorQuantile + adaptiveFloorSlope*ratio

	idx := int(float64(len(nonzero)) * quantile)
	if idx >= len(nonzero) {
		idx = len(nonzero) - 1
	}
	return nonzero[idx]
}

func (s *ScoredBlockSelector) findClusters(blocks []text.RawBlock, floor float64) []cluster {
	var clusters []cluster
	n := len(blocks)
	i := 0

	for i < n {
		if blocks[i].Score < floor {
			i++
			continue
		}

		c := cluster{start: i, end: i, totalScore: blocks[i].Score}
		i++
		gap := 0

		for i < n {
			if blocks[i].Score >= floor {
				// Include gap blocks in total score
				for g := i - gap; g < i; g++ {
					c.totalScore += blocks[g].Score
				}
				c.end = i
				c.totalScore += blocks[i].Score
				gap = 0
				i++
			} else {
				gap++
				if gap > s.maxGap {
					break
				}
				i++
			}
		}

		clusters = append(clusters, c)
	}

	return clusters
}

// mergeClusters combines adjacent clusters when the gap between them is small
// and the gap blocks have reasonable scores (avg >= floor * DefaultMergeMinRatio).
func mergeClusters(clusters []cluster, blocks []text.RawBlock, floor float64) []cluster {
	if len(clusters) <= 1 {
		return clusters
	}

	minGapScore := floor * DefaultMergeMinRatio
	merged := []cluster{clusters[0]}

	for _, next := range clusters[1:] {
		prev := &merged[len(merged)-1]
		gapLen := next.start - prev.end - 1

		if gapLen <= DefaultMergeDistance {
			// Compute average score of gap blocks
			gapScore := 0.0
			for g := prev.end + 1; g < next.start; g++ {
				gapScore += blocks[g].Score
			}
			gapAvg := 0.0
			if gapLen > 0 {
				gapAvg = gapScore / float64(gapLen)
			}

			if gapAvg >= minGapScore || gapLen == 0 {
				// Merge: extend prev to cover gap + next
				prev.end = next.end
				prev.totalScore += gapScore + next.totalScore
				continue
			}
		}
		merged = append(merged, next)
	}

	return merged
}

// selectClusters keeps all clusters whose totalScore is at least
// DefaultKeepRatio of the best cluster's score.
func selectClusters(clusters []cluster) []cluster {
	if len(clusters) == 0 {
		return nil
	}

	bestScore := 0.0
	for _, c := range clusters {
		if c.totalScore > bestScore {
			bestScore = c.totalScore
		}
	}

	threshold := bestScore * DefaultKeepRatio
	var kept []cluster
	for _, c := range clusters {
		if c.totalScore >= threshold {
			kept = append(kept, c)
		}
	}
	return kept
}

// collectMultiContent extracts blocks from all kept clusters, expanding backward
// to include headings that immediately precede each cluster.
func collectMultiContent(blocks []text.RawBlock, kept []cluster) []text.RawBlock {
	if len(kept) == 0 {
		return nil
	}

	// Build a set of included block indices
	include := make([]bool, len(blocks))
	for _, c := range kept {
		start := c.start
		for start > 0 && blocks[start-1].Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
			start--
		}
		for i := start; i <= c.end; i++ {
			include[i] = true
		}
	}

	var out []text.RawBlock
	for i, b := range blocks {
		if include[i] {
			out = append(out, b)
		}
	}
	return out
}

func (s *ScoredBlockSelector) logScores(blocks []text.RawBlock, kept []cluster, floor float64) {
	maxScore := 0.0
	for _, b := range blocks {
		if b.Score > maxScore {
			maxScore = b.Score
		}
	}
	if maxScore <= 0 {
		maxScore = 1
	}

	// Build include set for KEEP status (same logic as collectMultiContent)
	include := make([]bool, len(blocks))
	headingExpanded := make([]bool, len(blocks))
	for _, c := range kept {
		start := c.start
		for start > 0 && blocks[start-1].Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
			start--
		}
		for i := start; i <= c.end; i++ {
			include[i] = true
			if i < c.start {
				headingExpanded[i] = true
			}
		}
	}

	const barWidth = 30

	// Format cluster ranges for header
	var clusterStrs []string
	for _, c := range kept {
		clusterStrs = append(clusterStrs, fmt.Sprintf("[%d..%d]", c.start, c.end))
	}
	slog.Debug(fmt.Sprintf("--- Block Scores (floor=%.2f, gap=%d, clusters=%s) ---",
		floor, s.maxGap, strings.Join(clusterStrs, " ")))

	for i, b := range blocks {
		preview := strings.ReplaceAll(b.Text, "\n", " ")
		if len([]rune(preview)) > 50 {
			preview = string([]rune(preview)[:50]) + "..."
		}

		barLen := int((b.Score / maxScore) * barWidth)
		barLen = max(barLen, 0)
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)

		status := "CUT"
		if include[i] {
			if headingExpanded[i] {
				status = "KEEP(h)"
			} else {
				status = "KEEP"
			}
		}

		slog.Debug(fmt.Sprintf("[%3d] %-12s %7.2f |%s| %-7s %s",
			b.Idx, "<"+b.TagName+">", b.Score, bar, status, preview))
	}
}
