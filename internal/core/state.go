package core

import (
	"sync"
	"time"

	"github.com/orchard9/watch-now/internal/monitors"
)

type StateStore struct {
	mu       sync.RWMutex
	results  map[string]*monitors.Result
	history  map[string][]HistoryEntry
	watchers []chan StateUpdate
}

type HistoryEntry struct {
	Result    *monitors.Result
	Timestamp time.Time
}

type StateUpdate struct {
	Name   string
	Result *monitors.Result
}

func NewStateStore() *StateStore {
	return &StateStore{
		results: make(map[string]*monitors.Result),
		history: make(map[string][]HistoryEntry),
	}
}

func (s *StateStore) Update(result *monitors.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store current result
	s.results[result.Name] = result

	// Add to history (keep last 100 entries)
	history := s.history[result.Name]
	history = append(history, HistoryEntry{
		Result:    result,
		Timestamp: time.Now(),
	})
	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	s.history[result.Name] = history

	// Notify watchers
	update := StateUpdate{
		Name:   result.Name,
		Result: result,
	}
	for _, watcher := range s.watchers {
		select {
		case watcher <- update:
		default:
			// Don't block if watcher is not ready
		}
	}
}

func (s *StateStore) Get(name string) *monitors.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results[name]
}

func (s *StateStore) GetAll() map[string]*monitors.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	results := make(map[string]*monitors.Result)
	for k, v := range s.results {
		results[k] = v
	}
	return results
}

func (s *StateStore) Subscribe() <-chan StateUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan StateUpdate, 10)
	s.watchers = append(s.watchers, ch)
	return ch
}

func (s *StateStore) Unsubscribe(ch <-chan StateUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, watcher := range s.watchers {
		if watcher == ch {
			s.watchers = append(s.watchers[:i], s.watchers[i+1:]...)
			close(watcher)
			break
		}
	}
}
