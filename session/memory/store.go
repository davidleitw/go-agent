package memory

import (
	"sync"
	"time"

	"github.com/davidleitw/go-agent/session"
	"github.com/google/uuid"
)

// Store is an in-memory implementation of SessionStore
type Store struct {
	sessions sync.Map // map[string]*memorySession
	done     chan struct{}
}

// NewStore creates a new in-memory session store
func NewStore() *Store {
	store := &Store{
		done: make(chan struct{}),
	}

	// Start background cleanup routine
	go store.cleanupExpired()
	return store
}

// Create creates a new session with the given options
func (s *Store) Create(opts ...session.CreateOption) session.Session {
	options := session.ApplyOptions(opts...)

	id := options.ID
	if id == "" {
		id = uuid.New().String()
	}

	now := time.Now()
	sess := &memorySession{
		id:        id,
		createdAt: now,
		updatedAt: now,
		state:     make(map[string]any),
		history:   make([]session.Entry, 0),
		metadata:  options.Metadata,
		mu:        sync.RWMutex{},
	}

	if options.TTL > 0 {
		expiresAt := now.Add(options.TTL)
		sess.expiresAt = &expiresAt
	}

	s.sessions.Store(id, sess)
	return sess
}

// Get retrieves a session by ID
func (s *Store) Get(id string) (session.Session, error) {
	value, ok := s.sessions.Load(id)
	if !ok {
		return nil, session.ErrSessionNotFound
	}

	sess := value.(*memorySession)

	// Check if session has expired
	if sess.IsExpired() {
		s.sessions.Delete(id)
		return nil, session.ErrSessionNotFound
	}

	return sess, nil
}

// Save saves a session (no-op for memory store as changes are immediate)
func (s *Store) Save(sess session.Session) error {
	// For memory store, sessions are saved immediately when modified
	// This method exists to satisfy the interface
	return nil
}

// Delete removes a session by ID
func (s *Store) Delete(id string) error {
	s.sessions.Delete(id)
	return nil
}

// DeleteExpired removes all expired sessions
func (s *Store) DeleteExpired() error {
	var toDelete []string

	s.sessions.Range(func(key, value any) bool {
		id := key.(string)
		sess := value.(*memorySession)

		if sess.IsExpired() {
			toDelete = append(toDelete, id)
		}

		return true
	})

	for _, id := range toDelete {
		s.sessions.Delete(id)
	}

	return nil
}

// cleanupExpired runs a background cleanup routine
func (s *Store) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.DeleteExpired()
		case <-s.done:
			return
		}
	}
}

// Close stops the background cleanup routine and releases resources
func (s *Store) Close() error {
	close(s.done)
	return nil
}
