package session

import "errors"

// Common errors
var (
	ErrSessionNotFound = errors.New("session not found")
)

// SessionStore manages session persistence
type SessionStore interface {
	Create(opts ...CreateOption) Session
	Get(id string) (Session, error)
	Save(session Session) error
	Delete(id string) error
	DeleteExpired() error
	Close() error
}