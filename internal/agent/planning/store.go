package planning

import (
	"sync"
	"time"
)

const planTTL = 15 * time.Minute

type Store struct {
	mu    sync.Mutex
	plans map[string]Plan
}

func NewStore() *Store {
	return &Store{plans: map[string]Plan{}}
}

func (s *Store) Put(plan Plan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(time.Now())
	s.plans[plan.ID] = plan
}

func (s *Store) Get(id string) (Plan, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(time.Now())
	plan, ok := s.plans[id]
	return plan, ok
}

func (s *Store) Take(id string) (Plan, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(time.Now())
	plan, ok := s.plans[id]
	if ok {
		delete(s.plans, id)
	}
	return plan, ok
}

func (s *Store) pruneLocked(now time.Time) {
	for id, plan := range s.plans {
		if now.Sub(plan.CreatedAt) > planTTL {
			delete(s.plans, id)
		}
	}
}
