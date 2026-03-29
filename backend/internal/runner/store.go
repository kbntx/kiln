package runner

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// RunStore is a thread-safe in-memory store for runs.
type RunStore struct {
	mu   sync.RWMutex
	runs map[string]*Run
}

// NewRunStore creates a new RunStore.
func NewRunStore() *RunStore {
	return &RunStore{
		runs: make(map[string]*Run),
	}
}

// Create generates a UUID for the run, sets CreatedAt, and stores it.
func (s *RunStore) Create(run *Run) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run.ID = newUUID()
	run.CreatedAt = time.Now()
	s.runs[run.ID] = run
}

// Get returns a run by ID, or nil if not found.
func (s *RunStore) Get(id string) *Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runs[id]
}

// Update applies a mutation function to the run under a write lock.
func (s *RunStore) Update(id string, fn func(*Run)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.runs[id]; ok {
		fn(r)
	}
}

// FindWorkDir returns the WorkDir from the most recent discovery run for the given PR.
func (s *RunStore) FindWorkDir(owner, repo string, prNumber int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.runs {
		if r.Owner == owner && r.Repo == repo && r.PRNumber == prNumber && r.WorkDir != "" {
			return r.WorkDir
		}
	}
	return ""
}

// newUUID generates a v4 UUID using crypto/rand.
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
