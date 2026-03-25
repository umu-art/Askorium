package applib

import "container/heap"

// URLCandidate — элемент frontier с оценкой приоритета.
type URLCandidate struct {
	URL             string
	Depth           int32
	Score           float64 // 0.0 = FIFO-порядок; выше = более приоритетный
	ContentTypeHint string  // "" для HTML-страниц
	SourcePageURL   string  // заполняется для документных ссылок
	AnchorText      string  // текст анкора — используется scorer'ом
}

// Frontier — очередь URL-кандидатов для обхода.
type Frontier interface {
	Push(c URLCandidate)
	Pop() (URLCandidate, bool)
	Len() int
}

// FIFOFrontier — FIFO-очередь (текущее поведение, BFS).
type FIFOFrontier struct {
	items []URLCandidate
}

func NewFIFOFrontier() *FIFOFrontier {
	return &FIFOFrontier{}
}

func (f *FIFOFrontier) Push(c URLCandidate) {
	f.items = append(f.items, c)
}

func (f *FIFOFrontier) Pop() (URLCandidate, bool) {
	if len(f.items) == 0 {
		return URLCandidate{}, false
	}
	c := f.items[0]
	f.items = f.items[1:]
	return c, true
}

func (f *FIFOFrontier) Len() int {
	return len(f.items)
}

// PriorityFrontier — очередь с приоритетом (Best-First Search).
// Кандидаты с более высоким Score обрабатываются первыми.
type PriorityFrontier struct {
	h priorityHeap
}

func NewPriorityFrontier() *PriorityFrontier {
	pf := &PriorityFrontier{}
	heap.Init(&pf.h)
	return pf
}

func (f *PriorityFrontier) Push(c URLCandidate) {
	heap.Push(&f.h, c)
}

func (f *PriorityFrontier) Pop() (URLCandidate, bool) {
	if f.h.Len() == 0 {
		return URLCandidate{}, false
	}
	c := heap.Pop(&f.h).(URLCandidate)
	return c, true
}

func (f *PriorityFrontier) Len() int {
	return f.h.Len()
}

// priorityHeap реализует heap.Interface для URLCandidate (max-heap по Score).
type priorityHeap []URLCandidate

func (h priorityHeap) Len() int            { return len(h) }
func (h priorityHeap) Less(i, j int) bool  { return h[i].Score > h[j].Score } // max-heap
func (h priorityHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }

func (h *priorityHeap) Push(x interface{}) {
	*h = append(*h, x.(URLCandidate))
}

func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	c := old[n-1]
	*h = old[:n-1]
	return c
}
