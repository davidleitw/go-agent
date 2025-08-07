package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/davidleitw/go-agent/session"
)

func TestStore_Create(t *testing.T) {
	store := NewStore()

	// Test creating session with default options
	sess := store.Create(context.Background())
	if sess.ID() == "" {
		t.Error("Expected session to have an ID")
	}

	// Test creating session with custom ID
	customID := "custom-session-id"
	sess2 := store.Create(context.Background(), session.WithID(customID))
	if sess2.ID() != customID {
		t.Errorf("Expected session ID to be %s, got %s", customID, sess2.ID())
	}

	// Test creating session with TTL
	ttl := 1 * time.Hour
	_ = store.Create(context.Background(), session.WithTTL(ttl))
	// Note: ExpiresAt() has been removed from the interface
	// TTL is now handled internally

	// Test creating session with metadata
	_ = store.Create(context.Background(), session.WithMetadata("user_id", "123"))
	// Note: metadata is not directly accessible through Session interface
	// This would need additional methods or inspection through the store
}

func TestStore_GetAndSave(t *testing.T) {
	store := NewStore()

	// Test getting non-existent session
	_, err := store.Get(context.Background(), "non-existent")
	if !errors.Is(err, session.ErrSessionNotFound) {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}

	// Test creating and getting session
	sess := store.Create(context.Background(), session.WithID("test-session"))
	retrieved, err := store.Get(context.Background(), "test-session")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrieved.ID() != "test-session" {
		t.Errorf("Expected session ID test-session, got %s", retrieved.ID())
	}

	// Test save (should be no-op for memory store)
	err = store.Save(context.Background(), sess)
	if err != nil {
		t.Errorf("Expected no error from Save, got %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewStore()

	// Create and delete session
	sess := store.Create(context.Background(), session.WithID("delete-test"))
	err := store.Delete(context.Background(), sess.ID())
	if err != nil {
		t.Errorf("Expected no error from Delete, got %v", err)
	}

	// Verify session is gone
	_, err = store.Get(context.Background(), sess.ID())
	if !errors.Is(err, session.ErrSessionNotFound) {
		t.Errorf("Expected ErrSessionNotFound after deletion, got %v", err)
	}
}

func TestStore_ExpiredSessions(t *testing.T) {
	store := NewStore()

	// Create session with very short TTL
	shortTTL := 1 * time.Millisecond
	sess := store.Create(context.Background(), session.WithID("expired-test"), session.WithTTL(shortTTL))

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should not be able to get expired session
	_, err := store.Get(context.Background(), sess.ID())
	if !errors.Is(err, session.ErrSessionNotFound) {
		t.Errorf("Expected ErrSessionNotFound for expired session, got %v", err)
	}
}

func TestStore_DeleteExpired(t *testing.T) {
	store := NewStore()

	// Create mix of expired and non-expired sessions
	expiredSess := store.Create(context.Background(), session.WithID("expired"), session.WithTTL(1*time.Millisecond))
	validSess := store.Create(context.Background(), session.WithID("valid"), session.WithTTL(1*time.Hour))

	// Wait for one to expire
	time.Sleep(10 * time.Millisecond)

	// Run cleanup
	err := store.DeleteExpired(context.Background())
	if err != nil {
		t.Errorf("Expected no error from DeleteExpired, got %v", err)
	}

	// Expired should be gone
	_, err = store.Get(context.Background(), expiredSess.ID())
	if !errors.Is(err, session.ErrSessionNotFound) {
		t.Errorf("Expected expired session to be deleted")
	}

	// Valid should still exist
	_, err = store.Get(context.Background(), validSess.ID())
	if err != nil {
		t.Errorf("Expected valid session to still exist, got %v", err)
	}
}

func TestSession_StateManagement(t *testing.T) {
	store := NewStore()
	sess := store.Create(context.Background())

	// Test Set and Get
	sess.Set("key1", "value1")
	sess.Set("key2", 42)

	value1, exists1 := sess.Get("key1")
	if !exists1 || value1 != "value1" {
		t.Errorf("Expected key1=value1, got exists=%v, value=%v", exists1, value1)
	}

	value2, exists2 := sess.Get("key2")
	if !exists2 || value2 != 42 {
		t.Errorf("Expected key2=42, got exists=%v, value=%v", exists2, value2)
	}

	// Test non-existent key
	_, exists3 := sess.Get("non-existent")
	if exists3 {
		t.Error("Expected non-existent key to return false")
	}

	// Test Delete
	sess.Delete("key1")
	_, exists4 := sess.Get("key1")
	if exists4 {
		t.Error("Expected deleted key to not exist")
	}
}

func TestSession_HistoryManagement(t *testing.T) {
	store := NewStore()
	sess := store.Create(context.Background())

	// Add entries with different timestamps
	entry1 := session.NewMessageEntry("user", "Hello")
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	entry2 := session.NewMessageEntry("assistant", "Hi there")
	time.Sleep(1 * time.Millisecond)
	entry3 := session.NewToolCallEntry("search", map[string]any{"query": "test"})

	err := sess.AddEntry(entry1)
	if err != nil {
		t.Errorf("Expected no error adding entry, got %v", err)
	}

	err = sess.AddEntry(entry2)
	if err != nil {
		t.Errorf("Expected no error adding entry, got %v", err)
	}

	err = sess.AddEntry(entry3)
	if err != nil {
		t.Errorf("Expected no error adding entry, got %v", err)
	}

	// Test GetHistory with no limit
	history := sess.GetHistory(0)
	if len(history) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(history))
	}

	// Verify newest first ordering
	if !history[0].Timestamp.After(history[1].Timestamp) ||
		!history[1].Timestamp.After(history[2].Timestamp) {
		t.Error("Expected history to be sorted newest first")
	}

	// Test GetHistory with limit
	limitedHistory := sess.GetHistory(2)
	if len(limitedHistory) != 2 {
		t.Errorf("Expected 2 entries with limit, got %d", len(limitedHistory))
	}
}

func TestSession_Timestamps(t *testing.T) {
	store := NewStore()
	createdBefore := time.Now()
	sess := store.Create(context.Background())
	createdAfter := time.Now()

	// Test CreatedAt
	createdAt := sess.CreatedAt()
	if createdAt.Before(createdBefore) || createdAt.After(createdAfter) {
		t.Error("CreatedAt timestamp is not within expected range")
	}

	// Test UpdatedAt initially equals CreatedAt
	updatedAt := sess.UpdatedAt()
	if !updatedAt.Equal(createdAt) {
		t.Error("Initial UpdatedAt should equal CreatedAt")
	}

	// Test UpdatedAt changes after modification
	time.Sleep(1 * time.Millisecond)
	sess.Set("key", "value")
	newUpdatedAt := sess.UpdatedAt()
	if !newUpdatedAt.After(updatedAt) {
		t.Error("UpdatedAt should be updated after modification")
	}

	// Touch() has been removed - UpdatedAt is now automatically updated
}

func TestEntryHelpers(t *testing.T) {
	// Test NewMessageEntry
	msgEntry := session.NewMessageEntry("user", "Hello world")
	if msgEntry.Type != session.EntryTypeMessage {
		t.Errorf("Expected type %s, got %s", session.EntryTypeMessage, msgEntry.Type)
	}

	content, ok := session.GetMessageContent(msgEntry)
	if !ok {
		t.Error("Expected to extract MessageContent")
	}
	if content.Role != "user" || content.Text != "Hello world" {
		t.Errorf("Expected role=user, text='Hello world', got role=%s, text=%s", content.Role, content.Text)
	}

	// Test NewToolCallEntry
	params := map[string]any{"query": "test", "limit": 10}
	toolEntry := session.NewToolCallEntry("search", params)
	if toolEntry.Type != session.EntryTypeToolCall {
		t.Errorf("Expected type %s, got %s", session.EntryTypeToolCall, toolEntry.Type)
	}

	toolContent, ok := session.GetToolCallContent(toolEntry)
	if !ok {
		t.Error("Expected to extract ToolCallContent")
	}
	if toolContent.Tool != "search" {
		t.Errorf("Expected tool=search, got %s", toolContent.Tool)
	}

	// Test NewToolResultEntry with success
	resultEntry := session.NewToolResultEntry("search", []string{"result1", "result2"}, nil)
	if resultEntry.Type != session.EntryTypeToolResult {
		t.Errorf("Expected type %s, got %s", session.EntryTypeToolResult, resultEntry.Type)
	}

	resultContent, ok := session.GetToolResultContent(resultEntry)
	if !ok {
		t.Error("Expected to extract ToolResultContent")
	}
	if !resultContent.Success || resultContent.Error != "" {
		t.Errorf("Expected success=true, error='', got success=%v, error=%s", resultContent.Success, resultContent.Error)
	}

	// Test NewToolResultEntry with error
	testErr := errors.New("tool failed")
	errorEntry := session.NewToolResultEntry("search", nil, testErr)
	errorContent, ok := session.GetToolResultContent(errorEntry)
	if !ok {
		t.Error("Expected to extract ToolResultContent")
	}
	if errorContent.Success || errorContent.Error != "tool failed" {
		t.Errorf("Expected success=false, error='tool failed', got success=%v, error=%s", errorContent.Success, errorContent.Error)
	}
}

func TestStore_Close(t *testing.T) {
	store := NewStore()

	// Create a session
	sess := store.Create(context.Background(), session.WithID("test-close"))

	// Close the store
	err := store.Close()
	if err != nil {
		t.Errorf("Expected no error from Close, got %v", err)
	}

	// Store should still be usable after close (just no background cleanup)
	retrieved, err := store.Get(context.Background(), sess.ID())
	if err != nil {
		t.Errorf("Expected to retrieve session after Close, got %v", err)
	}
	if retrieved.ID() != "test-close" {
		t.Errorf("Expected session ID test-close, got %s", retrieved.ID())
	}
}

func TestConcurrency(t *testing.T) {
	store := NewStore()
	defer store.Close() // Clean up
	sess := store.Create(context.Background())

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Test concurrent state operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				sess.Set(key, value)
				retrievedValue, exists := sess.Get(key)
				if exists && retrievedValue != value {
					t.Errorf("Concurrent access issue: expected %s, got %v", value, retrievedValue)
				}
			}
		}(i)
	}

	// Test concurrent history operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				entry := session.NewMessageEntry("user", fmt.Sprintf("msg_%d_%d", id, j))
				err := sess.AddEntry(entry)
				if err != nil {
					t.Errorf("Error adding entry: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	history := sess.GetHistory(0)
	expectedHistoryLength := numGoroutines * numOperations
	if len(history) != expectedHistoryLength {
		t.Errorf("Expected %d history entries, got %d", expectedHistoryLength, len(history))
	}
}
