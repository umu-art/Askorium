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
	DefaultClusterGap    = 2
	DefaultFloorQuantile = 0.20 // floor = P20 of non-zero scores
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
	best := bestCluster(clusters)
	out := collectContent(blocks, best)

	if slog.Default().Enabled(nil, slog.LevelDebug) {
		s.logScores(blocks, best, floor)
	}

	return out
}

// computeFloor takes the P20 of non-zero scores.
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

	idx := int(float64(len(nonzero)) * s.floorQuantile)
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

func bestCluster(clusters []cluster) cluster {
	if len(clusters) == 0 {
		return cluster{start: -1, end: -1}
	}
	best := clusters[0]
	for _, c := range clusters[1:] {
		if c.totalScore > best.totalScore {
			best = c
		}
	}
	return best
}

// collectContent extracts the best cluster's blocks, expanding backward
// to include headings that immediately precede the cluster.
func collectContent(blocks []text.RawBlock, best cluster) []text.RawBlock {
	if best.start < 0 {
		return nil
	}

	start := best.start
	for start > 0 && blocks[start-1].Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
		start--
	}

	out := make([]text.RawBlock, 0, best.end-start+1)
	for i := start; i <= best.end; i++ {
		out = append(out, blocks[i])
	}
	return out
}

func (s *ScoredBlockSelector) logScores(blocks []text.RawBlock, best cluster, floor float64) {
	maxScore := 0.0
	for _, b := range blocks {
		if b.Score > maxScore {
			maxScore = b.Score
		}
	}
	if maxScore <= 0 {
		maxScore = 1
	}

	const barWidth = 30

	slog.Debug(fmt.Sprintf("--- Block Scores (floor=%.2f, gap=%d, cluster=[%d..%d]) ---",
		floor, s.maxGap, best.start, best.end))

	keptStart := best.start
	for keptStart > 0 && blocks[keptStart-1].Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
		keptStart--
	}

	for i, b := range blocks {
		preview := strings.ReplaceAll(b.Text, "\n", " ")
		if len([]rune(preview)) > 50 {
			preview = string([]rune(preview)[:50]) + "..."
		}

		barLen := int((b.Score / maxScore) * barWidth)
		if barLen < 0 {
			barLen = 0
		}
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)

		status := "CUT"
		if i >= keptStart && i <= best.end {
			if i < best.start {
				status = "KEEP(h)"
			} else {
				status = "KEEP"
			}
		}

		slog.Debug(fmt.Sprintf("[%3d] %-12s %7.2f |%s| %-7s %s",
			b.Idx, "<"+b.TagName+">", b.Score, bar, status, preview))
	}
}
