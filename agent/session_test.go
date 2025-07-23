package agent

import (
	"context"
	"testing"
	"time"

	"github.com/davidleitw/go-agent/session/memory"
)

func TestHandleSession_CreateNew(t *testing.T) {
	model := &MockModel{}
	store := memory.NewStore()
	
	config := EngineConfig{
		Model:        model,
		SessionStore: store,
		SessionTTL:   time.Hour,
	}
	
	engineInterface, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test creating new session
	request := Request{Input: "Hello"}
	
	concreteEngine := engineInterface.(*engine)
	sess, err := concreteEngine.handleSession(context.Background(), request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if sess == nil {
		t.Fatal("Expected session to be created")
	}
	
	// Verify session ID was generated
	if sess.ID() == "" {
		t.Error("Expected session ID to be generated")
	}
	
	// Verify dynamic state was set
	length, exists := sess.Get("initial_input_length")
	if !exists {
		t.Error("Expected initial_input_length to be set")
	}
	if length != 5 { // len("Hello")
		t.Errorf("Expected initial_input_length to be 5, got %v", length)
	}
	
	startTime, exists := sess.Get("session_start_time")
	if !exists {
		t.Error("Expected session_start_time to be set")
	}
	if startTime == "" {
		t.Error("Expected session_start_time to be non-empty")
	}
}

func TestHandleSession_LoadExisting(t *testing.T) {
	model := &MockModel{}
	store := memory.NewStore()
	
	config := EngineConfig{
		Model:        model,
		SessionStore: store,
	}
	
	engineInterface, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Create an existing session
	existingSession := store.Create(context.Background())
	existingSession.Set("test_key", "test_value")
	
	// Test loading existing session
	request := Request{
		Input:     "Continue",
		SessionID: existingSession.ID(),
	}
	
	concreteEngine := engineInterface.(*engine)
	sess, err := concreteEngine.handleSession(context.Background(), request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if sess == nil {
		t.Fatal("Expected session to be loaded")
	}
	
	// Verify it's the same session
	if sess.ID() != existingSession.ID() {
		t.Errorf("Expected session ID %s, got %s", existingSession.ID(), sess.ID())
	}
	
	// Verify existing data is preserved
	value, exists := sess.Get("test_key")
	if !exists {
		t.Error("Expected test_key to exist")
	}
	if value != "test_value" {
		t.Errorf("Expected test_value, got %v", value)
	}
}

func TestHandleSession_NotFound(t *testing.T) {
	model := &MockModel{}
	store := memory.NewStore()
	
	config := EngineConfig{
		Model:        model,
		SessionStore: store,
	}
	
	engineInterface, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test with non-existent session ID
	request := Request{
		Input:     "Hello",
		SessionID: "non-existent-id",
	}
	
	concreteEngine := engineInterface.(*engine)
	sess, err := concreteEngine.handleSession(context.Background(), request)
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
	
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
	
	if sess != nil {
		t.Error("Expected session to be nil")
	}
}

func TestConfiguredEngine_SessionTTLDefaults(t *testing.T) {
	model := &MockModel{}
	
	// Test with no TTL specified
	config := EngineConfig{Model: model}
	
	engineInterface, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Verify default TTL was set
	expectedTTL := 24 * time.Hour
	engineImpl := engineInterface.(*engine)
	if engineImpl.sessionTTL != expectedTTL {
		t.Errorf("Expected default TTL %v, got %v", expectedTTL, engineImpl.sessionTTL)
	}
	
	// Test with custom TTL
	customTTL := 2 * time.Hour
	configWithTTL := EngineConfig{
		Model:      model,
		SessionTTL: customTTL,
	}
	
	engineWithTTLInterface, err := NewEngine(configWithTTL)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	engineWithTTLImpl := engineWithTTLInterface.(*engine)
	if engineWithTTLImpl.sessionTTL != customTTL {
		t.Errorf("Expected custom TTL %v, got %v", customTTL, engineWithTTLImpl.sessionTTL)
	}
}

func TestBuilder_WithSessionTTL(t *testing.T) {
	model := &MockModel{}
	customTTL := 6 * time.Hour
	
	agent, err := NewBuilder().
		WithLLM(model).
		WithSessionTTL(customTTL).
		Build()
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if agent == nil {
		t.Fatal("Expected agent to be non-nil")
	}
	
	// Verify the engine has the correct TTL
	builtAgent := agent.(*BuiltAgent)
	engine := builtAgent.engine.(*engine)
	
	if engine.sessionTTL != customTTL {
		t.Errorf("Expected TTL %v, got %v", customTTL, engine.sessionTTL)
	}
}