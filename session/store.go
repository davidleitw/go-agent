package session

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrSessionNotFound = errors.New("session not found")
)

// SessionStore manages session persistence
type SessionStore interface {
	Create(ctx context.Context, opts ...CreateOption) Session
	Get(ctx context.Context, id string) (Session, error)
	Save(ctx context.Context, session Session) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
	Close() error
}
