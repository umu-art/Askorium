package applib

import "context"

// VisitedStore — хранилище посещённых URL.
type VisitedStore interface {
	Contains(url string) bool
	Add(url string)
}

// FrontierFactory создаёт Frontier и VisitedStore для одной задачи.
type FrontierFactory interface {
	NewFrontier(taskID string, priority bool) Frontier
	NewVisitedStore(taskID string) VisitedStore
	Cleanup(ctx context.Context, taskID string)
}

type MemVisitedStore struct {
	m map[string]bool
}

func NewMemVisitedStore() *MemVisitedStore {
	return &MemVisitedStore{m: make(map[string]bool)}
}

func (s *MemVisitedStore) Contains(url string) bool { return s.m[url] }
func (s *MemVisitedStore) Add(url string)           { s.m[url] = true }

type MemFrontierFactory struct{}

func NewMemFrontierFactory() *MemFrontierFactory { return &MemFrontierFactory{} }

func (f *MemFrontierFactory) NewFrontier(_ string, priority bool) Frontier {
	if priority {
		return NewPriorityFrontier()
	}
	return NewFIFOFrontier()
}

func (f *MemFrontierFactory) NewVisitedStore(_ string) VisitedStore {
	return NewMemVisitedStore()
}

func (f *MemFrontierFactory) Cleanup(_ context.Context, _ string) {}
