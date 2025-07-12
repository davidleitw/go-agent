package storage

import (
	"context"
	"sync"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
)

// inMemoryStore implements SessionStore using in-memory storage
type inMemoryStore struct {
	sessions map[string]agent.Session
	mutex    sync.RWMutex
}

// NewInMemory creates a new in-memory SessionStore implementation
func NewInMemory() agent.SessionStore {
	return &inMemoryStore{
		sessions: make(map[string]agent.Session),
	}
}

// Save persists a session in memory
func (s *inMemoryStore) Save(ctx context.Context, session agent.Session) error {
	if session == nil {
		return agent.ErrInvalidSession
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Clone the session to prevent external modification
	s.sessions[session.ID()] = session.Clone()
	return nil
}

// Load retrieves a session from memory by ID
func (s *inMemoryStore) Load(ctx context.Context, sessionID string) (agent.Session, error) {
	if sessionID == "" {
		return nil, agent.ErrInvalidSessionID
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, agent.ErrSessionNotFound
	}
	
	// Return a clone to prevent external modification
	return session.Clone(), nil
}

// Delete removes a session from memory
func (s *inMemoryStore) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return agent.ErrInvalidSessionID
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.sessions[sessionID]; !exists {
		return agent.ErrSessionNotFound
	}
	
	delete(s.sessions, sessionID)
	return nil
}

// List returns all session IDs, optionally filtered
func (s *inMemoryStore) List(ctx context.Context, filter agent.SessionFilter) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var sessionIDs []string
	
	for id, session := range s.sessions {
		if s.matchesFilter(session, filter) {
			sessionIDs = append(sessionIDs, id)
		}
	}
	
	return sessionIDs, nil
}

// Exists checks if a session exists
func (s *inMemoryStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, agent.ErrInvalidSessionID
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	_, exists := s.sessions[sessionID]
	return exists, nil
}

// Count returns the total number of sessions
func (s *inMemoryStore) Count(ctx context.Context) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return len(s.sessions), nil
}

// Clear removes all sessions from memory
func (s *inMemoryStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.sessions = make(map[string]agent.Session)
	return nil
}

// GetSessions returns all sessions, optionally filtered
func (s *inMemoryStore) GetSessions(ctx context.Context, filter agent.SessionFilter) ([]agent.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var sessions []agent.Session
	
	for _, session := range s.sessions {
		if s.matchesFilter(session, filter) {
			// Return clone to prevent external modification
			sessions = append(sessions, session.Clone())
		}
	}
	
	return sessions, nil
}

// GetSessionsSince returns sessions created or updated since the given time
func (s *inMemoryStore) GetSessionsSince(ctx context.Context, since time.Time) ([]agent.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var sessions []agent.Session
	
	for _, session := range s.sessions {
		if session.CreatedAt().After(since) || session.UpdatedAt().After(since) {
			sessions = append(sessions, session.Clone())
		}
	}
	
	return sessions, nil
}

// Helper methods

func (s *inMemoryStore) matchesFilter(session agent.Session, filter agent.SessionFilter) bool {
	// Filter by prefix
	if filter.IDPrefix != "" && !strings.HasPrefix(session.ID(), filter.IDPrefix) {
		return false
	}
	
	// Filter by created after
	if !filter.CreatedAfter.IsZero() && session.CreatedAt().Before(filter.CreatedAfter) {
		return false
	}
	
	// Filter by created before
	if !filter.CreatedBefore.IsZero() && session.CreatedAt().After(filter.CreatedBefore) {
		return false
	}
	
	// Filter by updated after
	if !filter.UpdatedAfter.IsZero() && session.UpdatedAt().Before(filter.UpdatedAfter) {
		return false
	}
	
	// Filter by updated before
	if !filter.UpdatedBefore.IsZero() && session.UpdatedAt().After(filter.UpdatedBefore) {
		return false
	}
	
	// Filter by minimum message count
	if filter.MinMessages > 0 && len(session.Messages()) < filter.MinMessages {
		return false
	}
	
	// Filter by maximum message count
	if filter.MaxMessages > 0 && len(session.Messages()) > filter.MaxMessages {
		return false
	}
	
	return true
}