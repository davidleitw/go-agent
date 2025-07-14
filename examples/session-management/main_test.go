package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/agent/mocks"
)

// TestBasicSessionOperations tests the basic session functionality
func TestBasicSessionOperations(t *testing.T) {
	session := agent.NewSession("test-session")
	
	// Test session ID
	if session.ID() != "test-session" {
		t.Errorf("Expected session ID 'test-session', got %s", session.ID())
	}
	
	// Test empty session
	if len(session.Messages()) != 0 {
		t.Errorf("Expected empty session, got %d messages", len(session.Messages()))
	}
	
	// Test adding messages
	userMsg := session.AddMessage(agent.RoleUser, "Test message")
	if userMsg.Role != agent.RoleUser {
		t.Errorf("Expected user role, got %s", userMsg.Role)
	}
	if userMsg.Content != "Test message" {
		t.Errorf("Expected 'Test message', got %s", userMsg.Content)
	}
	
	// Test message count
	if len(session.Messages()) != 1 {
		t.Errorf("Expected 1 message, got %d", len(session.Messages()))
	}
}

// TestSessionDataStorage tests session data storage and retrieval
func TestSessionDataStorage(t *testing.T) {
	session := agent.NewSession("data-test")
	
	// Test storing and retrieving different data types
	session.SetData("string_key", "string_value")
	session.SetData("int_key", 42)
	session.SetData("bool_key", true)
	session.SetData("time_key", time.Now())
	
	// Test string data
	if val := session.GetData("string_key"); val != "string_value" {
		t.Errorf("Expected 'string_value', got %v", val)
	}
	
	// Test int data
	if val := session.GetData("int_key"); val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}
	
	// Test bool data
	if val := session.GetData("bool_key"); val != true {
		t.Errorf("Expected true, got %v", val)
	}
	
	// Test non-existent key
	if val := session.GetData("non_existent"); val != nil {
		t.Errorf("Expected nil for non-existent key, got %v", val)
	}
	
	// Test data overwriting
	session.SetData("string_key", "new_value")
	if val := session.GetData("string_key"); val != "new_value" {
		t.Errorf("Expected 'new_value', got %v", val)
	}
}

// TestSessionCloning tests session cloning functionality
func TestSessionCloning(t *testing.T) {
	original := agent.NewSession("original")
	original.AddMessage(agent.RoleUser, "Original message")
	original.SetData("test_key", "test_value")
	
	// Type assert to cloneable interface
	type cloneableSession interface {
		agent.Session
		Clone() agent.Session
	}
	
	cloneable, ok := original.(cloneableSession)
	if !ok {
		t.Skip("Session implementation doesn't support cloning")
	}
	
	cloned := cloneable.Clone()
	
	// Test that clone has same initial data
	if cloned.ID() != original.ID() {
		t.Errorf("Cloned session should have same ID")
	}
	
	if len(cloned.Messages()) != len(original.Messages()) {
		t.Errorf("Cloned session should have same message count")
	}
	
	if cloned.GetData("test_key") != "test_value" {
		t.Errorf("Cloned session should have same data")
	}
	
	// Test independence: modify original
	original.AddMessage(agent.RoleAssistant, "New message in original")
	original.SetData("test_key", "modified_value")
	
	// Clone should remain unchanged
	if len(cloned.Messages()) != 1 {
		t.Errorf("Clone should not be affected by original modifications")
	}
	
	if cloned.GetData("test_key") != "test_value" {
		t.Errorf("Clone data should not be affected by original modifications")
	}
}

// TestAgentIntegration tests session integration with agents
func TestAgentIntegration(t *testing.T) {
	// Create mock chat model
	mockChat := mocks.NewMockChatModel()
	mockChat.SetResponse(agent.Message{
		Role:      agent.RoleAssistant,
		Content:   "Mock response from agent",
		Timestamp: time.Now(),
	})
	
	// Create test agent
	testAgent, err := agent.NewBasicAgent(agent.BasicAgentConfig{
		Name:      "test-agent",
		ChatModel: mockChat,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	
	// Create session
	session := agent.NewSession("agent-test")
	session.SetData("test_context", "integration_test")
	
	// Test chat interaction
	ctx := context.Background()
	response, _, err := testAgent.Chat(ctx, session, "Test input")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	// Verify response
	if response.Role != agent.RoleAssistant {
		t.Errorf("Expected assistant response, got %s", response.Role)
	}
	
	if response.Content != "Mock response from agent" {
		t.Errorf("Expected mock response, got %s", response.Content)
	}
	
	// Verify session state
	messages := session.Messages()
	if len(messages) < 2 {
		t.Errorf("Expected at least 2 messages after chat, got %d", len(messages))
	}
	
	// Verify context data is preserved
	if val := session.GetData("test_context"); val != "integration_test" {
		t.Errorf("Session data should be preserved during chat")
	}
}

// TestConcurrentAccess tests thread-safe session operations
func TestConcurrentAccess(t *testing.T) {
	session := agent.NewSession("concurrent-test")
	session.SetData("counter", 0)
	
	const numGoroutines = 10
	const operationsPerGoroutine = 5
	
	done := make(chan bool, numGoroutines)
	
	// Start concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				// Add messages
				session.AddMessage(agent.RoleUser, "Concurrent message")
				
				// Update counter
				current := session.GetData("counter")
				if counter, ok := current.(int); ok {
					session.SetData("counter", counter+1)
				}
			}
		}(i)
	}
	
	// Wait for completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify final state
	expectedMessages := numGoroutines * operationsPerGoroutine
	if len(session.Messages()) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(session.Messages()))
	}
	
	// Counter should be updated (exact value may vary due to race conditions)
	finalCounter := session.GetData("counter")
	if counter, ok := finalCounter.(int); !ok || counter <= 0 {
		t.Errorf("Expected positive counter value, got %v", finalCounter)
	}
}

// TestSessionWithComplexData tests session with complex data structures
func TestSessionWithComplexData(t *testing.T) {
	session := agent.NewSession("complex-data-test")
	
	// Store complex data structures
	complexData := map[string]interface{}{
		"user": map[string]string{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"settings": []string{"dark_mode", "notifications"},
		"metadata": map[string]interface{}{
			"version": 1.2,
			"enabled": true,
		},
	}
	
	session.SetData("complex_data", complexData)
	
	// Retrieve and verify complex data
	retrieved := session.GetData("complex_data")
	if retrieved == nil {
		t.Fatal("Complex data should not be nil")
	}
	
	// Type assert and verify structure
	data, ok := retrieved.(map[string]interface{})
	if !ok {
		t.Fatal("Retrieved data should be a map")
	}
	
	// Verify nested user data
	user, ok := data["user"].(map[string]string)
	if !ok {
		t.Fatal("User data should be a map[string]string")
	}
	
	if user["name"] != "Alice" {
		t.Errorf("Expected user name 'Alice', got %s", user["name"])
	}
	
	// Verify array data
	settings, ok := data["settings"].([]string)
	if !ok {
		t.Fatal("Settings should be a []string")
	}
	
	if len(settings) != 2 {
		t.Errorf("Expected 2 settings, got %d", len(settings))
	}
}

// TestSessionMessageTypes tests different message types
func TestSessionMessageTypes(t *testing.T) {
	session := agent.NewSession("message-types-test")
	
	// Add different types of messages
	userMsg := session.AddMessage(agent.RoleUser, "User message")
	assistantMsg := session.AddMessage(agent.RoleAssistant, "Assistant response")
	systemMsg := session.AddMessage(agent.RoleSystem, "System notification")
	toolMsg := session.AddMessage(agent.RoleTool, "Tool result")
	
	// Verify message properties
	messages := []struct {
		msg      agent.Message
		expected string
	}{
		{userMsg, agent.RoleUser},
		{assistantMsg, agent.RoleAssistant},
		{systemMsg, agent.RoleSystem},
		{toolMsg, agent.RoleTool},
	}
	
	for i, test := range messages {
		if test.msg.Role != test.expected {
			t.Errorf("Message %d: expected role %s, got %s", i, test.expected, test.msg.Role)
		}
		
		if test.msg.Timestamp.IsZero() {
			t.Errorf("Message %d: timestamp should not be zero", i)
		}
		
		if strings.TrimSpace(test.msg.Content) == "" {
			t.Errorf("Message %d: content should not be empty", i)
		}
	}
	
	// Verify session contains all messages
	allMessages := session.Messages()
	if len(allMessages) != 4 {
		t.Errorf("Expected 4 messages in session, got %d", len(allMessages))
	}
}