package applib

import (
	"testing"
)

func TestFIFOFrontier_Order(t *testing.T) {
	f := NewFIFOFrontier()
	urls := []string{"a", "b", "c"}
	for _, u := range urls {
		f.Push(URLCandidate{URL: u})
	}
	for _, want := range urls {
		got, ok := f.Pop()
		if !ok {
			t.Fatal("Pop returned false, expected true")
		}
		if got.URL != want {
			t.Errorf("FIFO order: got %q, want %q", got.URL, want)
		}
	}
	if f.Len() != 0 {
		t.Errorf("expected empty frontier, got Len=%d", f.Len())
	}
}

func TestFIFOFrontier_PopEmpty(t *testing.T) {
	f := NewFIFOFrontier()
	_, ok := f.Pop()
	if ok {
		t.Error("Pop on empty frontier should return false")
	}
}

func TestPriorityFrontier_Order(t *testing.T) {
	f := NewPriorityFrontier()
	f.Push(URLCandidate{URL: "low", Score: 0.1})
	f.Push(URLCandidate{URL: "high", Score: 0.9})
	f.Push(URLCandidate{URL: "mid", Score: 0.5})

	order := []string{"high", "mid", "low"}
	for _, want := range order {
		got, ok := f.Pop()
		if !ok {
			t.Fatal("Pop returned false, expected true")
		}
		if got.URL != want {
			t.Errorf("priority order: got %q, want %q", got.URL, want)
		}
	}
}

func TestPriorityFrontier_PopEmpty(t *testing.T) {
	f := NewPriorityFrontier()
	_, ok := f.Pop()
	if ok {
		t.Error("Pop on empty priority frontier should return false")
	}
}

func TestPriorityFrontier_Len(t *testing.T) {
	f := NewPriorityFrontier()
	if f.Len() != 0 {
		t.Errorf("expected Len=0, got %d", f.Len())
	}
	f.Push(URLCandidate{URL: "x", Score: 0.5})
	if f.Len() != 1 {
		t.Errorf("expected Len=1, got %d", f.Len())
	}
	f.Pop()
	if f.Len() != 0 {
		t.Errorf("expected Len=0 after Pop, got %d", f.Len())
	}
}
