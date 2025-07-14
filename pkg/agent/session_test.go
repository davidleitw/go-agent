package agent

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSession_Basic(t *testing.T) {
	session := NewSession("test-session")
	
	// Test ID
	if session.ID() != "test-session" {
		t.Errorf("ID() = %v, want %v", session.ID(), "test-session")
	}
	
	// Test empty messages
	messages := session.Messages()
	if len(messages) != 0 {
		t.Errorf("Messages() length = %v, want 0", len(messages))
	}
	
	// Test adding message
	msg := session.AddMessage(RoleUser, "Hello")
	if msg.Role != RoleUser {
		t.Errorf("AddMessage() role = %v, want %v", msg.Role, RoleUser)
	}
	if msg.Content != "Hello" {
		t.Errorf("AddMessage() content = %v, want %v", msg.Content, "Hello")
	}
	if msg.Timestamp.IsZero() {
		t.Error("AddMessage() timestamp should not be zero")
	}
	
	// Test messages count
	messages = session.Messages()
	if len(messages) != 1 {
		t.Errorf("Messages() length = %v, want 1", len(messages))
	}
	
	// Test message content
	if messages[0].Role != RoleUser {
		t.Errorf("Messages()[0].Role = %v, want %v", messages[0].Role, RoleUser)
	}
	if messages[0].Content != "Hello" {
		t.Errorf("Messages()[0].Content = %v, want %v", messages[0].Content, "Hello")
	}
}

func TestSession_MultipleMessages(t *testing.T) {
	session := NewSession("multi-message-test")
	
	// Add multiple messages
	session.AddMessage(RoleUser, "Hello")
	session.AddMessage(RoleAssistant, "Hi there!")
	session.AddMessage(RoleSystem, "System message")
	session.AddMessage(RoleTool, "Tool result")
	
	messages := session.Messages()
	if len(messages) != 4 {
		t.Errorf("Messages() length = %v, want 4", len(messages))
	}
	
	// Check message order and content
	expectedMessages := []struct {
		role    string
		content string
	}{
		{RoleUser, "Hello"},
		{RoleAssistant, "Hi there!"},
		{RoleSystem, "System message"},
		{RoleTool, "Tool result"},
	}
	
	for i, expected := range expectedMessages {
		if messages[i].Role != expected.role {
			t.Errorf("Messages()[%d].Role = %v, want %v", i, messages[i].Role, expected.role)
		}
		if messages[i].Content != expected.content {
			t.Errorf("Messages()[%d].Content = %v, want %v", i, messages[i].Content, expected.content)
		}
	}
}

func TestSession_DataStorage(t *testing.T) {
	session := NewSession("data-test")
	
	// Test setting and getting data
	session.SetData("key1", "value1")
	session.SetData("key2", 42)
	session.SetData("key3", map[string]string{"nested": "value"})
	
	// Test getting data
	value1 := session.GetData("key1")
	if value1 != "value1" {
		t.Errorf("GetData(\"key1\") = %v, want %v", value1, "value1")
	}
	
	value2 := session.GetData("key2")
	if value2 != 42 {
		t.Errorf("GetData(\"key2\") = %v, want %v", value2, 42)
	}
	
	value3 := session.GetData("key3")
	if nestedMap, ok := value3.(map[string]string); ok {
		if nestedMap["nested"] != "value" {
			t.Errorf("GetData(\"key3\")[\"nested\"] = %v, want %v", nestedMap["nested"], "value")
		}
	} else {
		t.Errorf("GetData(\"key3\") type = %T, want map[string]string", value3)
	}
	
	// Test getting non-existent data
	nilValue := session.GetData("non-existent")
	if nilValue != nil {
		t.Errorf("GetData(\"non-existent\") = %v, want nil", nilValue)
	}
	
	// Test overwriting data
	session.SetData("key1", "new-value")
	newValue := session.GetData("key1")
	if newValue != "new-value" {
		t.Errorf("GetData(\"key1\") after overwrite = %v, want %v", newValue, "new-value")
	}
}

func TestSession_ConcurrentAccess(t *testing.T) {
	session := NewSession("concurrent-test")
	const numGoroutines = 10
	const messagesPerGoroutine = 5
	
	var wg sync.WaitGroup
	
	// Add messages concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				content := fmt.Sprintf("Message %d-%d", id, j)
				session.AddMessage(RoleUser, content)
				
				// Also test data operations
				key := fmt.Sprintf("data-%d-%d", id, j)
				session.SetData(key, content)
			}
		}(i)
	}
	
	// Read messages concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				_ = session.Messages()
				_ = session.GetData("some-key")
			}
		}()
	}
	
	wg.Wait()
	
	// Verify final state
	messages := session.Messages()
	expectedCount := numGoroutines * messagesPerGoroutine
	if len(messages) != expectedCount {
		t.Errorf("Messages() length = %v, want %v", len(messages), expectedCount)
	}
	
	// All messages should have the correct role
	for i, msg := range messages {
		if msg.Role != RoleUser {
			t.Errorf("Messages()[%d].Role = %v, want %v", i, msg.Role, RoleUser)
		}
		if msg.Content == "" {
			t.Errorf("Messages()[%d].Content is empty", i)
		}
	}
}

func TestSession_MessageImmutability(t *testing.T) {
	session := NewSession("immutable-test")
	
	// Add a message
	session.AddMessage(RoleUser, "Original message")
	
	// Get messages and try to modify
	messages := session.Messages()
	if len(messages) != 1 {
		t.Fatalf("Messages() length = %v, want 1", len(messages))
	}
	
	originalContent := messages[0].Content
	
	// Try to modify the returned message
	messages[0].Content = "Modified message"
	
	// Get messages again and verify original is unchanged
	messagesAgain := session.Messages()
	if messagesAgain[0].Content != originalContent {
		t.Errorf("Message content was modified externally: got %v, want %v", 
			messagesAgain[0].Content, originalContent)
	}
}

func TestSessionAdvanced_ComplexMessage(t *testing.T) {
	session := NewSession("advanced-test")
	
	// Test that sessionImpl implements SessionAdvanced
	advSession, ok := session.(SessionAdvanced)
	if !ok {
		t.Fatal("NewSession() should return a SessionAdvanced implementation")
	}
	
	// Create a complex message with tool calls
	complexMessage := Message{
		Role:    RoleAssistant,
		Content: "I need to call a tool",
		ToolCalls: []ToolCall{
			{
				ID:   "call_123",
				Type: "function",
				Function: ToolCallFunction{
					Name:      "calculator",
					Arguments: `{"operation": "add", "a": 1, "b": 2}`,
				},
			},
		},
		Timestamp: time.Now(),
	}
	
	// Add complex message
	advSession.AddComplexMessage(complexMessage)
	
	// Verify message was added correctly
	messages := session.Messages()
	if len(messages) != 1 {
		t.Errorf("Messages() length = %v, want 1", len(messages))
	}
	
	retrievedMsg := messages[0]
	if retrievedMsg.Role != RoleAssistant {
		t.Errorf("Message role = %v, want %v", retrievedMsg.Role, RoleAssistant)
	}
	if retrievedMsg.Content != "I need to call a tool" {
		t.Errorf("Message content = %v, want %v", retrievedMsg.Content, "I need to call a tool")
	}
	if len(retrievedMsg.ToolCalls) != 1 {
		t.Errorf("ToolCalls length = %v, want 1", len(retrievedMsg.ToolCalls))
	}
	if retrievedMsg.ToolCalls[0].ID != "call_123" {
		t.Errorf("ToolCall ID = %v, want %v", retrievedMsg.ToolCalls[0].ID, "call_123")
	}
}

func TestSession_HelperMethods(t *testing.T) {
	session := NewSession("helper-test")
	// Cast to internal interface for testing helper methods
	type internalSession interface {
		Session
		MessageCount() int
		LastMessage() *Message
		GetMessagesByRole(role string) []Message
		ClearMessages()
		Clone() Session
	}
	sessionImpl := session.(internalSession)
	
	// Test empty session helpers
	if sessionImpl.MessageCount() != 0 {
		t.Errorf("MessageCount() = %v, want 0", sessionImpl.MessageCount())
	}
	
	lastMsg := sessionImpl.LastMessage()
	if lastMsg != nil {
		t.Errorf("LastMessage() = %v, want nil", lastMsg)
	}
	
	userMessages := sessionImpl.GetMessagesByRole(RoleUser)
	if len(userMessages) != 0 {
		t.Errorf("GetMessagesByRole(RoleUser) length = %v, want 0", len(userMessages))
	}
	
	// Add some messages
	session.AddMessage(RoleUser, "User message 1")
	session.AddMessage(RoleAssistant, "Assistant response")
	session.AddMessage(RoleUser, "User message 2")
	session.AddMessage(RoleSystem, "System message")
	
	// Test MessageCount
	if sessionImpl.MessageCount() != 4 {
		t.Errorf("MessageCount() = %v, want 4", sessionImpl.MessageCount())
	}
	
	// Test LastMessage
	lastMsg = sessionImpl.LastMessage()
	if lastMsg == nil {
		t.Fatal("LastMessage() should not be nil")
	}
	if lastMsg.Role != RoleSystem {
		t.Errorf("LastMessage().Role = %v, want %v", lastMsg.Role, RoleSystem)
	}
	if lastMsg.Content != "System message" {
		t.Errorf("LastMessage().Content = %v, want %v", lastMsg.Content, "System message")
	}
	
	// Test GetMessagesByRole
	userMessages = sessionImpl.GetMessagesByRole(RoleUser)
	if len(userMessages) != 2 {
		t.Errorf("GetMessagesByRole(RoleUser) length = %v, want 2", len(userMessages))
	}
	
	assistantMessages := sessionImpl.GetMessagesByRole(RoleAssistant)
	if len(assistantMessages) != 1 {
		t.Errorf("GetMessagesByRole(RoleAssistant) length = %v, want 1", len(assistantMessages))
	}
	
	systemMessages := sessionImpl.GetMessagesByRole(RoleSystem)
	if len(systemMessages) != 1 {
		t.Errorf("GetMessagesByRole(RoleSystem) length = %v, want 1", len(systemMessages))
	}
	
	toolMessages := sessionImpl.GetMessagesByRole(RoleTool)
	if len(toolMessages) != 0 {
		t.Errorf("GetMessagesByRole(RoleTool) length = %v, want 0", len(toolMessages))
	}
}

func TestSession_Clone(t *testing.T) {
	session := NewSession("clone-test")
	// Cast to internal interface for testing clone method
	type cloneableSession interface {
		Session
		Clone() Session
	}
	sessionImpl := session.(cloneableSession)
	
	// Add some data
	session.AddMessage(RoleUser, "Test message")
	session.SetData("key1", "value1")
	session.SetData("key2", 42)
	
	// Clone the session
	cloned := sessionImpl.Clone()
	
	// Verify clone has same data
	if cloned.ID() != session.ID() {
		t.Errorf("Cloned session ID = %v, want %v", cloned.ID(), session.ID())
	}
	
	if len(cloned.Messages()) != len(session.Messages()) {
		t.Errorf("Cloned messages count = %v, want %v", len(cloned.Messages()), len(session.Messages()))
	}
	
	if cloned.GetData("key1") != "value1" {
		t.Errorf("Cloned data key1 = %v, want %v", cloned.GetData("key1"), "value1")
	}
	
	if cloned.GetData("key2") != 42 {
		t.Errorf("Cloned data key2 = %v, want %v", cloned.GetData("key2"), 42)
	}
	
	// Verify independence - modify original
	session.AddMessage(RoleAssistant, "New message")
	session.SetData("key1", "modified")
	
	// Clone should be unchanged
	if len(cloned.Messages()) != 1 {
		t.Errorf("Cloned messages count after original modification = %v, want 1", len(cloned.Messages()))
	}
	
	if cloned.GetData("key1") != "value1" {
		t.Errorf("Cloned data key1 after original modification = %v, want %v", cloned.GetData("key1"), "value1")
	}
	
	// Verify independence - modify clone
	if clearable, ok := cloned.(interface{ ClearMessages() }); ok {
		clearable.ClearMessages()
	}
	cloned.SetData("key2", 999)
	
	// Original should be unchanged
	if len(session.Messages()) != 2 {
		t.Errorf("Original messages count after clone modification = %v, want 2", len(session.Messages()))
	}
	
	if session.GetData("key2") != 42 {
		t.Errorf("Original data key2 after clone modification = %v, want 42", session.GetData("key2"))
	}
}

func TestSession_ClearMessages(t *testing.T) {
	session := NewSession("clear-test")
	
	// Cast to internal interface for testing clear functionality
	type clearableSession interface {
		Session
		ClearMessages()
	}
	sessionImpl := session.(clearableSession)
	
	// Add some messages
	session.AddMessage(RoleUser, "Message 1")
	session.AddMessage(RoleAssistant, "Message 2")
	session.AddMessage(RoleSystem, "Message 3")
	
	// Verify messages exist
	if len(session.Messages()) != 3 {
		t.Errorf("Messages count before clear = %v, want 3", len(session.Messages()))
	}
	
	// Clear messages
	sessionImpl.ClearMessages()
	
	// Verify messages are cleared
	if len(session.Messages()) != 0 {
		t.Errorf("Messages count after clear = %v, want 0", len(session.Messages()))
	}
	
	// Verify data is preserved
	session.SetData("preserved", "data")
	sessionImpl.ClearMessages()
	if session.GetData("preserved") != "data" {
		t.Errorf("Data should be preserved after ClearMessages")
	}
}