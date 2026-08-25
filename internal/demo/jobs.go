package demo

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Job struct {
	ID     string
	Events chan any
	Done   chan struct{}
}

type JobStore struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

func newJobID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *JobStore) Create() *Job {
	job := &Job{
		ID:     newJobID(),
		Events: make(chan any, 32),
		Done:   make(chan struct{}),
	}
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	return job
}

func (s *JobStore) Get(id string) (*Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	return job, ok
}

func (s *JobStore) Delete(id string) {
	s.mu.Lock()
	delete(s.jobs, id)
	s.mu.Unlock()
}
