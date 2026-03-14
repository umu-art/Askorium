# Block Selector Algorithm

## Overview

The selector takes scored blocks from the segmenter/filter pipeline and determines which blocks constitute the page's main content. It uses a **score-then-cluster** approach: three scoring strategies assign a composite score to each block, then a clustering algorithm finds the densest contiguous region of high-scoring blocks.

```
blocks → [Semantic] → [Readability] → [Density] → computeFloor → findClusters → bestCluster → output
                  scoring (additive)                         clustering
```

The final score for each block is the **sum** of contributions from all three strategies.

---

## Scoring Strategies

### 1. SemanticMarkupStrategy

**Source:** Readability.js id/class heuristics + positional observation from Boilerpipe.

Scores based on three DOM signals that survive the pruning stage (nav/aside/footer/header are already removed by `SemanticNoisePruner` before blocks reach the selector):

| Signal | How it works | Range |
|--------|-------------|-------|
| **Ancestor boost** | Block inside `<article>` or `<main>` → +3.0; inside `<section>` → +1.0. Takes the highest boost among all ancestors. | [0, +3.0] |
| **HTMLId patterns** | Matches the block's `id` attribute against two pattern lists. Negative patterns checked first (e.g. `sidebar`, `comment`, `social` → -2.0). Positive patterns (e.g. `article`, `content`, `post` → +2.0). | [-2.0, +2.0] |
| **Position** | Blocks in the first 5% or last 15% of the document get -1.5. Catches boilerplate in non-semantic `<div>` wrappers that the pruner can't target. | [-1.5, 0] |

**Total range per block: [-3.5, +5.0]**

### 2. ReadabilityScoringStrategy

**Source:** Readability.js / Arc90 algorithm.

Scores based on surface text features that distinguish prose from non-prose:

| Signal | Formula | Range |
|--------|---------|-------|
| **Tag weight** | `p/blockquote/pre` → +1.0, `li` → +0.5, `td` → -0.5, `h1-h6` → -1.0 | [-1.0, +1.0] |
| **Comma count** | `count(, ＋ ， ＋ 、) × 1.0` — prose has commas, nav/boilerplate doesn't. Supports ASCII, fullwidth (CJK), and ideographic commas. | [0, +N] |
| **Length bonus** | `min(len(text) / 100, 3.0)` — longer blocks are more likely body text. Uses char count (not word count) to match Readability.js behavior and work better with CJK text. | [0, +3.0] |
| **Propagation** | After base scoring, each block receives 50% of its neighbors' base scores (snapshot-based, no feedback loops). A high-scoring block boosts its neighbors. | varies |

**Typical range per block: [-1.0, ~+8.0] before propagation**

### 3. TextDensityStrategy

**Source:** Boilerpipe TDQ (Kohlschütter 2010) + JusText stopword gating (Pomikálek 2011).

Two-phase scoring:

**Phase 1 — Per-block TDQ:**

| Condition | Score |
|-----------|-------|
| WordCount = 0 or no letter-bearing tokens | **-2.0** (active noise penalty) |
| Stopword density < 10% (non-natural language: phone numbers, dates, nav labels) | **-2.0** (active noise penalty) |
| Passes gate | `(wordCount / numLines) × (1 - linkDensity)` where `numLines = len(text)/80 + 1` |

The **negative penalty** (-2.0) is critical — previous versions returned 0, which made all scores positive and rendered the floor/clustering mechanism useless.

**Phase 2 — Neighbor context (Boilerpipe's key insight):**

> The most discriminative feature pair is `(current_density, previous_density)`.

Each block gets a bonus = `max(prev_density, next_density) × 0.5` (only from positive-density neighbors). This captures transition sentences — short blocks that would score low on their own but sit between high-density paragraphs.

**Typical range per block: [-2.0, ~+15.0]**

---

## Clustering (scored.go)

After all strategies have scored, the selector does NOT use a fixed threshold. Instead:

### Step 1: Dynamic Floor

```
floor = P20 of non-zero scores
```

- Only considers scores > 0 (excludes blocks that got -2.0 from density gate)
- P20 (20th percentile) means roughly the bottom 20% of "real" blocks become below-floor
- Adapts to each page: a page with uniformly high scores gets a high floor; a page with mixed content gets a lower floor

### Step 2: Find Clusters

Scans blocks sequentially, grouping contiguous runs of above-floor blocks:

```
[above] [above] [below] [above] [above] [above] [below] [below] [below] [above]
|---- cluster 1 (gap=1) ----|   |---- cluster 2 ----|     gap=3 > maxGap    |-- c3 --|
```

- **maxGap = 2**: up to 2 consecutive below-floor blocks are tolerated within a cluster (bridging short dips like transition sentences)
- Gap blocks' scores are included in the cluster's total score
- Each cluster tracks `{start, end, totalScore}`

### Step 3: Best Cluster

The cluster with the highest `totalScore` wins. This naturally selects the longest, densest content region.

### Step 4: Heading Expansion

The winning cluster's boundaries are expanded backward to include any immediately preceding headings (they introduce the content but score low due to tag weight -1.0 and short length).

```
[h2] [h3] [p] [p] [p] [p] [li] [li]
 ↑    ↑   |--- best cluster ---|
 expanded backward
```

---

## Score Composition Example

A typical body paragraph inside `<article>`:

```
Semantic:     +3.0  (article ancestor)
              +0.0  (no id)
              +0.0  (position 0.30, not in head/tail)
Readability:  +1.0  (p tag)
              +2.0  (2 commas)
              +2.5  (250 chars)
              +2.0  (propagation from neighbors)
Density:      +5.5  (TDQ: high word density, low link density)
              +2.0  (neighbor boost)
──────────────────
Total:       +18.0
```

A noise block (phone number in a generic div):

```
Semantic:     +0.0  (no semantic ancestor)
              +0.0  (no id)
              -1.5  (position 0.92, in tail)
Readability:  +0.0  (div tag, not in tagWeights)
              +0.0  (no commas)
              +0.1  (10 chars)
              +0.0  (low-scoring neighbors)
Density:      -2.0  (stopword density 0.00 → noise penalty)
              +0.0  (no positive neighbor)
──────────────────
Total:        -3.4  → well below any floor → excluded from clusters
```

---

## Debug Visualization

When `LOG_LEVEL=debug`, every block is logged with its score and cluster membership:

```
--- Block Scores (floor=4.50, gap=2, cluster=[3..12]) ---
[  0] <div>          -3.40 |░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░| CUT     +7 (495) 22-22...
[  1] <h2>            2.10 |████░░░░░░░░░░░░░░░░░░░░░░░░░░| CUT     О программе
[  2] <h2>            2.30 |████░░░░░░░░░░░░░░░░░░░░░░░░░░| CUT     Описание
[  3] <h2>            5.20 |██████████░░░░░░░░░░░░░░░░░░░░| KEEP(h) Учебный план
[  4] <p>            18.00 |██████████████████████████████| KEEP    Программа бакалавриата...
[  5] <p>            14.20 |███████████████████████░░░░░░░| KEEP    Студенты изучают...
```

- `KEEP` — block is inside the best cluster
- `KEEP(h)` — heading expanded into the cluster
- `CUT` — excluded

---

## Pipeline Context: What Happens Before the Selector

| Stage | Component | What it removes | Why the selector doesn't redo it |
|-------|-----------|----------------|----------------------------------|
| Prune | `TagPruner` | `<script>`, `<style>`, `<noscript>` etc. | Gone from DOM |
| Prune | `SemanticNoisePruner` | `<nav>`, `<aside>`, `<footer>`, `<header>`, ARIA roles | Gone from DOM |
| Prune | `VisibilityPruner` | `display:none`, `aria-hidden` | Gone from DOM |
| Filter | `MinCharLengthFilter` | Blocks < N chars | Hard cut, no scoring needed |
| Filter | `MaxLinkDensityFilter` | Blocks with linkDensity > 0.5 | Hard cut; density strategy still uses linkDensity as a coefficient for blocks ≤ 0.5 |
