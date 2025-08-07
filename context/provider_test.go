package context

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/davidleitw/go-agent/session"
	"github.com/davidleitw/go-agent/session/memory"
)

func TestSystemPromptProvider(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())
	provider := NewSystemPromptProvider("You are a helpful assistant.")

	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}

	ctx := contexts[0]
	if ctx.Type != "system" {
		t.Errorf("Expected type 'system', got '%s'", ctx.Type)
	}

	if ctx.Content != "You are a helpful assistant." {
		t.Errorf("Expected content 'You are a helpful assistant.', got '%s'", ctx.Content)
	}

	if ctx.Metadata == nil {
		t.Error("Expected metadata to be initialized")
	}
}

func TestHistoryProvider_EmptyHistory(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())
	provider := NewHistoryProvider(10)

	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 0 {
		t.Errorf("Expected 0 contexts for empty history, got %d", len(contexts))
	}
}

func TestHistoryProvider_MessageEntries(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Add message entries
	sess.AddEntry(session.NewMessageEntry("user", "Hello"))
	time.Sleep(1 * time.Millisecond)
	sess.AddEntry(session.NewMessageEntry("assistant", "Hi there"))
	time.Sleep(1 * time.Millisecond)
	sess.AddEntry(session.NewMessageEntry("system", "System message"))

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 3 {
		t.Errorf("Expected 3 contexts, got %d", len(contexts))
	}

	// Check newest first ordering
	if contexts[0].Type != "system" || contexts[0].Content != "System message" {
		t.Errorf("Expected first context to be system message, got type=%s, content=%s",
			contexts[0].Type, contexts[0].Content)
	}

	if contexts[1].Type != "assistant" || contexts[1].Content != "Hi there" {
		t.Errorf("Expected second context to be assistant message, got type=%s, content=%s",
			contexts[1].Type, contexts[1].Content)
	}

	if contexts[2].Type != "user" || contexts[2].Content != "Hello" {
		t.Errorf("Expected third context to be user message, got type=%s, content=%s",
			contexts[2].Type, contexts[2].Content)
	}

	// Check metadata
	for i, ctx := range contexts {
		if ctx.Metadata["entry_id"] == nil {
			t.Errorf("Context %d missing entry_id in metadata", i)
		}
		if ctx.Metadata["timestamp"] == nil {
			t.Errorf("Context %d missing timestamp in metadata", i)
		}
	}
}

func TestHistoryProvider_ToolCallEntries(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	params := map[string]any{
		"query": "test",
		"limit": 10,
	}
	sess.AddEntry(session.NewToolCallEntry("search", params))

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}

	ctx := contexts[0]
	if ctx.Type != "tool_call" {
		t.Errorf("Expected type 'tool_call', got '%s'", ctx.Type)
	}

	if !strings.Contains(ctx.Content, "Tool: search") {
		t.Errorf("Expected content to contain 'Tool: search', got '%s'", ctx.Content)
	}

	if !strings.Contains(ctx.Content, "Parameters:") {
		t.Errorf("Expected content to contain 'Parameters:', got '%s'", ctx.Content)
	}

	if ctx.Metadata["tool_name"] != "search" {
		t.Errorf("Expected tool_name metadata to be 'search', got '%v'", ctx.Metadata["tool_name"])
	}
}

func TestHistoryProvider_ToolResultEntries(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Test successful tool result
	result := []string{"result1", "result2"}
	sess.AddEntry(session.NewToolResultEntry("search", result, nil))

	// Test failed tool result
	sess.AddEntry(session.NewToolResultEntry("search", nil,
		&testError{msg: "connection failed"}))

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 2 {
		t.Errorf("Expected 2 contexts, got %d", len(contexts))
	}

	// Check failed result (newest first)
	failCtx := contexts[0]
	if failCtx.Type != "tool_result" {
		t.Errorf("Expected type 'tool_result', got '%s'", failCtx.Type)
	}

	if !strings.Contains(failCtx.Content, "Success: false") {
		t.Errorf("Expected content to contain 'Success: false', got '%s'", failCtx.Content)
	}

	if !strings.Contains(failCtx.Content, "Error: connection failed") {
		t.Errorf("Expected content to contain error message, got '%s'", failCtx.Content)
	}

	if failCtx.Metadata["success"] != false {
		t.Errorf("Expected success metadata to be false, got '%v'", failCtx.Metadata["success"])
	}

	// Check successful result
	successCtx := contexts[1]
	if !strings.Contains(successCtx.Content, "Success: true") {
		t.Errorf("Expected content to contain 'Success: true', got '%s'", successCtx.Content)
	}

	if successCtx.Metadata["success"] != true {
		t.Errorf("Expected success metadata to be true, got '%v'", successCtx.Metadata["success"])
	}
}

func TestHistoryProvider_ThinkingEntries(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Create thinking entry manually
	entry := session.Entry{
		ID:        "test-thinking",
		Type:      session.EntryTypeThinking,
		Timestamp: time.Now(),
		Content:   "I need to think about this carefully...",
		Metadata:  make(map[string]any),
	}
	sess.AddEntry(entry)

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}

	ctx := contexts[0]
	if ctx.Type != "thinking" {
		t.Errorf("Expected type 'thinking', got '%s'", ctx.Type)
	}

	if ctx.Content != "I need to think about this carefully..." {
		t.Errorf("Expected thinking content, got '%s'", ctx.Content)
	}
}

func TestHistoryProvider_LimitFunctionality(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Add multiple entries
	for i := 0; i < 5; i++ {
		sess.AddEntry(session.NewMessageEntry("user", "Message "+string(rune('A'+i))))
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Test with limit
	provider := NewHistoryProvider(3)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 3 {
		t.Errorf("Expected 3 contexts with limit=3, got %d", len(contexts))
	}

	// Should get the 3 most recent (E, D, C)
	expectedMessages := []string{"Message E", "Message D", "Message C"}
	for i, expected := range expectedMessages {
		if contexts[i].Content != expected {
			t.Errorf("Context %d: expected '%s', got '%s'", i, expected, contexts[i].Content)
		}
	}

	// Test with no limit (0)
	providerNoLimit := NewHistoryProvider(0)
	allContexts := providerNoLimit.Provide(context.Background(), sess)

	if len(allContexts) != 5 {
		t.Errorf("Expected 5 contexts with no limit, got %d", len(allContexts))
	}
}

func TestHistoryProvider_MixedEntryTypes(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Add mixed entry types
	sess.AddEntry(session.NewMessageEntry("user", "Hello"))
	time.Sleep(1 * time.Millisecond)

	sess.AddEntry(session.NewToolCallEntry("search", map[string]any{"q": "test"}))
	time.Sleep(1 * time.Millisecond)

	sess.AddEntry(session.NewToolResultEntry("search", "found results", nil))
	time.Sleep(1 * time.Millisecond)

	sess.AddEntry(session.NewMessageEntry("assistant", "Found some results"))

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 4 {
		t.Errorf("Expected 4 contexts, got %d", len(contexts))
	}

	// Check types in newest-first order
	expectedTypes := []string{"assistant", "tool_result", "tool_call", "user"}
	for i, expectedType := range expectedTypes {
		if contexts[i].Type != expectedType {
			t.Errorf("Context %d: expected type '%s', got '%s'", i, expectedType, contexts[i].Type)
		}
	}
}

func TestHistoryProvider_MetadataPreservation(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Create entry with custom metadata
	entry := session.Entry{
		ID:        "test-meta",
		Type:      session.EntryTypeMessage,
		Timestamp: time.Now(),
		Content: session.MessageContent{
			Role: "user",
			Text: "Test message",
		},
		Metadata: map[string]any{
			"custom_field": "custom_value",
			"priority":     "high",
		},
	}
	sess.AddEntry(entry)

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}

	ctx := contexts[0]

	// Check original metadata is preserved
	if ctx.Metadata["custom_field"] != "custom_value" {
		t.Errorf("Expected custom_field='custom_value', got '%v'", ctx.Metadata["custom_field"])
	}

	if ctx.Metadata["priority"] != "high" {
		t.Errorf("Expected priority='high', got '%v'", ctx.Metadata["priority"])
	}

	// Check added metadata
	if ctx.Metadata["entry_id"] != "test-meta" {
		t.Errorf("Expected entry_id='test-meta', got '%v'", ctx.Metadata["entry_id"])
	}

	if ctx.Metadata["timestamp"] == nil {
		t.Error("Expected timestamp to be added to metadata")
	}
}

func TestHistoryProvider_FallbackForUnknownTypes(t *testing.T) {
	store := memory.NewStore()
	defer store.Close()

	sess := store.Create(context.Background())

	// Create entry with unknown type
	entry := session.Entry{
		ID:        "test-unknown",
		Type:      session.EntryType("unknown_type"),
		Timestamp: time.Now(),
		Content:   map[string]string{"data": "test"},
		Metadata:  make(map[string]any),
	}
	sess.AddEntry(entry)

	provider := NewHistoryProvider(10)
	contexts := provider.Provide(context.Background(), sess)

	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}

	ctx := contexts[0]
	if ctx.Type != "unknown_type" {
		t.Errorf("Expected type 'unknown_type', got '%s'", ctx.Type)
	}

	// Should be JSON marshaled
	var data map[string]string
	if err := json.Unmarshal([]byte(ctx.Content), &data); err != nil {
		t.Errorf("Expected JSON content, got error: %v", err)
	}

	if data["data"] != "test" {
		t.Errorf("Expected data='test', got '%s'", data["data"])
	}
}

// Helper type for testing errors
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
