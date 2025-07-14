package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
)

// SessionStoreCall records a call to the session store
type SessionStoreCall struct {
	Method    string      // "Save", "Load", "Delete", "List", "Exists"
	SessionID string      
	Session   agent.Session
	Filter    agent.SessionFilter
	Time      time.Time
	Result    interface{} // The result that was returned
	Error     error       // The error that was returned
}

// MockSessionStore provides a comprehensive mock implementation of SessionStore
type MockSessionStore struct {
	mu sync.RWMutex
	
	// In-memory storage
	sessions map[string]agent.Session
	
	// Call tracking
	calls []SessionStoreCall
	
	// Error simulation
	saveError   error
	loadError   error
	deleteError error
	listError   error
	existsError error
	
	// Behavior configuration
	saveDelay   time.Duration
	loadDelay   time.Duration
	deleteDelay time.Duration
	listDelay   time.Duration
	existsDelay time.Duration
}

// NewMockSessionStore creates a new mock session store
func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string]agent.Session),
		calls:    make([]SessionStoreCall, 0),
	}
}

// Save implements SessionStore.Save
func (m *MockSessionStore) Save(ctx context.Context, session agent.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	call := SessionStoreCall{
		Method:    "Save",
		SessionID: session.ID(),
		Session:   session,
		Time:      time.Now(),
	}
	
	// Simulate delay if configured
	if m.saveDelay > 0 {
		time.Sleep(m.saveDelay)
	}
	
	// Check for configured error
	if m.saveError != nil {
		call.Error = m.saveError
		m.calls = append(m.calls, call)
		return m.saveError
	}
	
	// Store the session (create a clone to avoid mutation if possible)
	if cloneable, ok := session.(interface{ Clone() agent.Session }); ok {
		m.sessions[session.ID()] = cloneable.Clone()
	} else {
		// For other implementations, store as-is
		m.sessions[session.ID()] = session
	}
	
	call.Result = nil
	m.calls = append(m.calls, call)
	return nil
}

// Load implements SessionStore.Load
func (m *MockSessionStore) Load(ctx context.Context, sessionID string) (agent.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	call := SessionStoreCall{
		Method:    "Load",
		SessionID: sessionID,
		Time:      time.Now(),
	}
	
	// Simulate delay if configured
	if m.loadDelay > 0 {
		time.Sleep(m.loadDelay)
	}
	
	// Check for configured error
	if m.loadError != nil {
		call.Error = m.loadError
		m.calls = append(m.calls, call)
		return nil, m.loadError
	}
	
	// Check if session exists
	session, exists := m.sessions[sessionID]
	if !exists {
		call.Error = agent.ErrSessionNotFound
		m.calls = append(m.calls, call)
		return nil, agent.ErrSessionNotFound
	}
	
	// Return a clone to prevent external mutation if possible
	var result agent.Session
	if cloneable, ok := session.(interface{ Clone() agent.Session }); ok {
		result = cloneable.Clone()
	} else {
		result = session
	}
	
	call.Result = result
	m.calls = append(m.calls, call)
	return result, nil
}

// Delete implements SessionStore.Delete
func (m *MockSessionStore) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	call := SessionStoreCall{
		Method:    "Delete",
		SessionID: sessionID,
		Time:      time.Now(),
	}
	
	// Simulate delay if configured
	if m.deleteDelay > 0 {
		time.Sleep(m.deleteDelay)
	}
	
	// Check for configured error
	if m.deleteError != nil {
		call.Error = m.deleteError
		m.calls = append(m.calls, call)
		return m.deleteError
	}
	
	// Delete the session
	delete(m.sessions, sessionID)
	
	call.Result = nil
	m.calls = append(m.calls, call)
	return nil
}

// List implements SessionStore.List
func (m *MockSessionStore) List(ctx context.Context, filter agent.SessionFilter) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	call := SessionStoreCall{
		Method: "List",
		Filter: filter,
		Time:   time.Now(),
	}
	
	// Simulate delay if configured
	if m.listDelay > 0 {
		time.Sleep(m.listDelay)
	}
	
	// Check for configured error
	if m.listError != nil {
		call.Error = m.listError
		m.calls = append(m.calls, call)
		return nil, m.listError
	}
	
	// Collect all session IDs (simplified filter implementation)
	var sessionIDs []string
	for id := range m.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	
	// Apply basic filtering
	if filter.IDPrefix != "" {
		filtered := make([]string, 0)
		for _, id := range sessionIDs {
			if len(id) >= len(filter.IDPrefix) && id[:len(filter.IDPrefix)] == filter.IDPrefix {
				filtered = append(filtered, id)
			}
		}
		sessionIDs = filtered
	}
	
	// Apply limit
	if filter.Limit > 0 && len(sessionIDs) > filter.Limit {
		sessionIDs = sessionIDs[:filter.Limit]
	}
	
	call.Result = sessionIDs
	m.calls = append(m.calls, call)
	return sessionIDs, nil
}

// Exists implements SessionStore.Exists
func (m *MockSessionStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	call := SessionStoreCall{
		Method:    "Exists",
		SessionID: sessionID,
		Time:      time.Now(),
	}
	
	// Simulate delay if configured
	if m.existsDelay > 0 {
		time.Sleep(m.existsDelay)
	}
	
	// Check for configured error
	if m.existsError != nil {
		call.Error = m.existsError
		m.calls = append(m.calls, call)
		return false, m.existsError
	}
	
	_, exists := m.sessions[sessionID]
	
	call.Result = exists
	m.calls = append(m.calls, call)
	return exists, nil
}

// Configuration methods

// SetSaveError configures Save to return an error
func (m *MockSessionStore) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveError = err
}

// SetLoadError configures Load to return an error
func (m *MockSessionStore) SetLoadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadError = err
}

// SetDeleteError configures Delete to return an error
func (m *MockSessionStore) SetDeleteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteError = err
}

// SetListError configures List to return an error
func (m *MockSessionStore) SetListError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listError = err
}

// SetExistsError configures Exists to return an error
func (m *MockSessionStore) SetExistsError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.existsError = err
}

// SetSaveDelay adds artificial delay to Save operations
func (m *MockSessionStore) SetSaveDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveDelay = delay
}

// SetLoadDelay adds artificial delay to Load operations
func (m *MockSessionStore) SetLoadDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadDelay = delay
}

// ClearErrors removes all configured errors
func (m *MockSessionStore) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveError = nil
	m.loadError = nil
	m.deleteError = nil
	m.listError = nil
	m.existsError = nil
}

// ClearDelays removes all configured delays
func (m *MockSessionStore) ClearDelays() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveDelay = 0
	m.loadDelay = 0
	m.deleteDelay = 0
	m.listDelay = 0
	m.existsDelay = 0
}

// Inspection methods

// GetCallHistory returns all recorded calls
func (m *MockSessionStore) GetCallHistory() []SessionStoreCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]SessionStoreCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the total number of calls
func (m *MockSessionStore) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.calls)
}

// GetCallCountByMethod returns the number of calls for a specific method
func (m *MockSessionStore) GetCallCountByMethod(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	count := 0
	for _, call := range m.calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// GetStoredSessions returns all currently stored sessions
func (m *MockSessionStore) GetStoredSessions() map[string]agent.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]agent.Session)
	for id, session := range m.sessions {
		result[id] = session
	}
	return result
}

// Reset clears all data and call history
func (m *MockSessionStore) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.sessions = make(map[string]agent.Session)
	m.calls = make([]SessionStoreCall, 0)
	m.ClearErrors()
	m.ClearDelays()
}

// Preload adds sessions to the store for testing
func (m *MockSessionStore) Preload(sessions map[string]agent.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for id, session := range sessions {
		m.sessions[id] = session
	}
}

// Verification helpers

// VerifySaved checks if a session with given ID was saved
func (m *MockSessionStore) VerifySaved(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, call := range m.calls {
		if call.Method == "Save" && call.SessionID == sessionID {
			return true
		}
	}
	return false
}

// VerifyLoaded checks if a session with given ID was loaded
func (m *MockSessionStore) VerifyLoaded(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, call := range m.calls {
		if call.Method == "Load" && call.SessionID == sessionID {
			return true
		}
	}
	return false
}

// VerifyDeleted checks if a session with given ID was deleted
func (m *MockSessionStore) VerifyDeleted(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, call := range m.calls {
		if call.Method == "Delete" && call.SessionID == sessionID {
			return true
		}
	}
	return false
}