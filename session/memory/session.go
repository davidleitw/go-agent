package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/davidleitw/go-agent/session"
)

// memorySession is an in-memory implementation of Session
type memorySession struct {
	id        string
	createdAt time.Time
	updatedAt time.Time
	expiresAt *time.Time
	state     map[string]any
	history   []session.Entry
	metadata  map[string]string
	mu        sync.RWMutex
}

// ID returns the session ID
func (s *memorySession) ID() string {
	return s.id
}

// CreatedAt returns when the session was created
func (s *memorySession) CreatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.createdAt
}

// UpdatedAt returns when the session was last updated
func (s *memorySession) UpdatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.updatedAt
}

// Get retrieves a value from the session state
func (s *memorySession) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.state[key]
	return value, exists
}

// Set stores a value in the session state
func (s *memorySession) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = value
	s.updatedAt = time.Now()
}

// Delete removes a value from the session state
func (s *memorySession) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.state, key)
	s.updatedAt = time.Now()
}

// AddEntry adds an entry to the session history
func (s *memorySession) AddEntry(entry session.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, entry)
	s.updatedAt = time.Now()
	return nil
}

// GetHistory returns the session history, sorted by timestamp (newest first)
func (s *memorySession) GetHistory(limit int) []session.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of the history slice
	history := make([]session.Entry, len(s.history))
	copy(history, s.history)

	// Sort by timestamp (newest first)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.After(history[j].Timestamp)
	})

	// Apply limit if specified
	if limit > 0 && limit < len(history) {
		history = history[:limit]
	}

	return history
}

// IsExpired checks if the session has expired
func (s *memorySession) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.expiresAt == nil {
		return false
	}

	return time.Now().After(*s.expiresAt)
}
