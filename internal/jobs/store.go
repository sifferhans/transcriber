package jobs

import (
	"context"
	"sort"
	"sync"
)

// Store is a goroutine-safe in-memory job store. Reads return copies.
type Store struct {
	mu      sync.RWMutex
	jobs    map[string]*Job
	cancels map[string]context.CancelFunc
}

func NewStore() *Store {
	return &Store{
		jobs:    map[string]*Job{},
		cancels: map[string]context.CancelFunc{},
	}
}

func (s *Store) Create(j Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := j
	s.jobs[j.ID] = &cp
}

func (s *Store) Get(id string) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return Job{}, false
	}
	return *j, true
}

// Update applies fn under the write lock; returns false if the job does not exist.
func (s *Store) Update(id string, fn func(*Job)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return false
	}
	fn(j)
	return true
}

func (s *Store) List() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, *j)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// SetCancel registers a cancel func so a later DELETE can kill the subprocess.
func (s *Store) SetCancel(id string, cancel context.CancelFunc) {
	s.mu.Lock()
	s.cancels[id] = cancel
	s.mu.Unlock()
}

func (s *Store) ClearCancel(id string) {
	s.mu.Lock()
	delete(s.cancels, id)
	s.mu.Unlock()
}

// Cancel invokes the registered cancel func; returns false if none is registered.
func (s *Store) Cancel(id string) bool {
	s.mu.Lock()
	cancel := s.cancels[id]
	s.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}
