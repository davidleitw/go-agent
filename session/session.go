package session

import "time"

// Session manages conversation state and history
type Session interface {
	// Basic information
	ID() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	// State management (key-value)
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string)

	// History management
	AddEntry(entry Entry) error
	GetHistory(limit int) []Entry
}
