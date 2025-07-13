package main

import (
	"context"
	"testing"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/schema"
)

// MockChatModel implements ChatModel for testing
type MockChatModel struct {
	responses []string
	callCount int
}

func NewMockChatModel(responses ...string) *MockChatModel {
	return &MockChatModel{
		responses: responses,
		callCount: 0,
	}
}

func (m *MockChatModel) GenerateChatCompletion(ctx context.Context, messages []agent.Message, model string, settings *agent.ModelSettings, tools []agent.Tool) (*agent.Message, error) {
	if m.callCount >= len(m.responses) {
		m.callCount = len(m.responses) - 1
	}
	
	response := &agent.Message{
		Role:    agent.RoleAssistant,
		Content: m.responses[m.callCount],
	}
	
	m.callCount++
	return response, nil
}

func (m *MockChatModel) GetSupportedModels() []string {
	return []string{"mock-model"}
}

func (m *MockChatModel) ValidateModel(model string) error {
	return nil
}

func (m *MockChatModel) GetModelInfo(model string) (*agent.ModelInfo, error) {
	return &agent.ModelInfo{
		ID:               "mock-model",
		Name:             "Mock Model",
		Description:      "Mock model for testing",
		ContextWindow:    4000,
		MaxOutputTokens:  1000,
		SupportsTools:    true,
		SupportsStreaming: false,
		SupportsJSON:     true,
	}, nil
}

func TestSchemaCollection(t *testing.T) {
	// Create mock chat model that returns a request for missing information
	mockModel := NewMockChatModel(
		"Please provide your email address so I can assist you.",
		`{"email": "user@example.com", "issue": null}`,
		"Thank you for providing your email. Please describe your issue in detail.",
	)

	// Create agent with mock model
	bot, err := agent.New("test-agent").
		WithChatModel(mockModel).
		WithDescription("Test agent for schema collection").
		WithInstructions("You are a test assistant.").
		Build()
	if err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	ctx := context.Background()

	// Test schema with required fields
	testSchema := []*schema.Field{
		schema.Define("email", "Please provide your email address"),
		schema.Define("issue", "Please describe your issue"),
	}

	// First request should ask for missing information
	response, err := bot.Chat(ctx, "I need help",
		agent.WithSchema(testSchema...),
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Should get a response asking for information
	if response.Message == "" {
		t.Error("Expected non-empty response message")
	}

	// Check metadata for schema collection
	if schemaCollection, ok := response.Metadata["schema_collection"].(bool); !ok || !schemaCollection {
		t.Error("Expected schema_collection metadata to be true")
	}

	if missingFields, ok := response.Metadata["missing_fields"].([]string); !ok || len(missingFields) == 0 {
		t.Error("Expected missing_fields metadata")
	}
}

func TestOptionalFields(t *testing.T) {
	// Test that optional fields don't block conversation flow
	mockModel := NewMockChatModel(
		`{"email": "user@example.com", "phone": null}`,
		"Thank you! I have your email. How can I help you today?",
	)

	bot, err := agent.New("test-agent").
		WithChatModel(mockModel).
		WithDescription("Test agent").
		WithInstructions("You are helpful.").
		Build()
	if err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	ctx := context.Background()

	// Schema with required and optional fields
	testSchema := []*schema.Field{
		schema.Define("email", "Please provide your email address"),
		schema.Define("phone", "Phone number (optional)").Optional(),
	}

	// Should not ask for optional phone if email is provided
	response, err := bot.Chat(ctx, "My email is user@example.com",
		agent.WithSchema(testSchema...),
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Should proceed with conversation since required field is provided
	if response.Message == "" {
		t.Error("Expected non-empty response message")
	}
}

func TestEmptySchema(t *testing.T) {
	// Test that empty schema doesn't interfere with normal operation
	mockModel := NewMockChatModel("Hello! How can I help you today?")

	bot, err := agent.New("test-agent").
		WithChatModel(mockModel).
		WithDescription("Test agent").
		WithInstructions("You are helpful.").
		Build()
	if err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	ctx := context.Background()

	// Empty schema should not interfere
	response, err := bot.Chat(ctx, "Hello",
		agent.WithSchema(), // Empty schema
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response.Message != "Hello! How can I help you today?" {
		t.Errorf("Expected normal response, got: %s", response.Message)
	}

	// Should not have schema collection metadata
	if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
		t.Error("Did not expect schema_collection metadata for empty schema")
	}
}

func TestSchemaFieldDefinition(t *testing.T) {
	// Test schema field creation and properties
	email := schema.Define("email", "Please provide your email address")
	
	if email.Name() != "email" {
		t.Errorf("Expected field name 'email', got '%s'", email.Name())
	}
	
	if email.Prompt() != "Please provide your email address" {
		t.Errorf("Expected specific prompt, got '%s'", email.Prompt())
	}
	
	if !email.Required() {
		t.Error("Expected field to be required by default")
	}

	// Test optional field
	phone := schema.Define("phone", "Phone number").Optional()
	
	if phone.Required() {
		t.Error("Expected field to be optional after calling Optional()")
	}
}