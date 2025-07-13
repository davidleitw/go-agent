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
		ID:                "mock-model",
		Name:              "Mock Model",
		Description:       "Mock model for testing",
		ContextWindow:     4000,
		MaxOutputTokens:   1000,
		SupportsTools:     true,
		SupportsStreaming: false,
		SupportsJSON:      true,
	}, nil
}

func TestSupportSchema(t *testing.T) {
	// Test that support schema has the expected required fields
	schema := supportSchema()

	expectedRequired := []string{"email", "issue_category", "description"}
	expectedOptional := []string{"order_id", "urgency", "previous_contact"}

	requiredCount := 0
	optionalCount := 0

	for _, field := range schema {
		if field.Required() {
			requiredCount++
			found := false
			for _, expected := range expectedRequired {
				if field.Name() == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unexpected required field: %s", field.Name())
			}
		} else {
			optionalCount++
			found := false
			for _, expected := range expectedOptional {
				if field.Name() == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unexpected optional field: %s", field.Name())
			}
		}
	}

	if requiredCount != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), requiredCount)
	}

	if optionalCount != len(expectedOptional) {
		t.Errorf("Expected %d optional fields, got %d", len(expectedOptional), optionalCount)
	}
}

func TestBillingSchema(t *testing.T) {
	// Test billing schema structure
	schema := billingSchema()

	expectedRequired := []string{"email", "account_number", "billing_question"}
	expectedOptional := []string{"amount_disputed", "payment_method", "billing_period"}

	requiredFields := make([]string, 0)
	optionalFields := make([]string, 0)

	for _, field := range schema {
		if field.Required() {
			requiredFields = append(requiredFields, field.Name())
		} else {
			optionalFields = append(optionalFields, field.Name())
		}
	}

	if len(requiredFields) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(requiredFields))
	}

	if len(optionalFields) != len(expectedOptional) {
		t.Errorf("Expected %d optional fields, got %d", len(expectedOptional), len(optionalFields))
	}
}

func TestTechnicalSchema(t *testing.T) {
	// Test technical schema structure
	schema := technicalSchema()

	expectedRequired := []string{"email", "error_message", "steps_taken"}
	expectedOptional := []string{"browser", "device_type", "operating_system"}

	requiredFields := make([]string, 0)
	optionalFields := make([]string, 0)

	for _, field := range schema {
		if field.Required() {
			requiredFields = append(requiredFields, field.Name())
		} else {
			optionalFields = append(optionalFields, field.Name())
		}
	}

	if len(requiredFields) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(requiredFields))
	}

	if len(optionalFields) != len(expectedOptional) {
		t.Errorf("Expected %d optional fields, got %d", len(expectedOptional), len(optionalFields))
	}
}

func TestSupportBotWithPartialInformation(t *testing.T) {
	// Test support bot handling partial information
	mockModel := NewMockChatModel(
		`{"email": "customer@example.com", "issue_category": "billing", "description": null}`,
		"Thank you for providing your email and indicating this is a billing issue. Please describe your billing question in detail.",
	)

	bot, err := agent.New("test-support-bot").
		WithChatModel(mockModel).
		WithDescription("Test support bot").
		WithInstructions("You are a professional customer support assistant.").
		Build()
	if err != nil {
		t.Fatalf("Failed to create test support bot: %v", err)
	}

	ctx := context.Background()

	// Test with billing schema and partial information
	response, err := bot.Chat(ctx, "Hi, my email is customer@example.com and I have a billing issue",
		agent.WithSchema(billingSchema()...),
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Should ask for missing information
	if response.Message == "" {
		t.Error("Expected non-empty response message")
	}

	// Should have schema collection metadata
	if schemaCollection, ok := response.Metadata["schema_collection"].(bool); !ok || !schemaCollection {
		t.Error("Expected schema_collection metadata to be true")
	}
}

func TestSupportBotWithCompleteInformation(t *testing.T) {
	// Test support bot when all required information is provided
	mockModel := NewMockChatModel(
		`{"email": "customer@example.com", "issue_category": "technical", "description": "Login error"}`,
		"Thank you for providing all the necessary information. I'll help you resolve this login error.",
	)

	bot, err := agent.New("test-support-bot").
		WithChatModel(mockModel).
		WithDescription("Test support bot").
		WithInstructions("You are a professional customer support assistant.").
		Build()
	if err != nil {
		t.Fatalf("Failed to create test support bot: %v", err)
	}

	ctx := context.Background()

	// Test with complete information
	response, err := bot.Chat(ctx, "My email is customer@example.com, I have a technical issue with a login error",
		agent.WithSchema(supportSchema()...),
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Should proceed with normal conversation
	if response.Message == "" {
		t.Error("Expected non-empty response message")
	}

	// Should not have schema collection metadata (all info collected)
	if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
		t.Error("Did not expect schema_collection metadata when all info is provided")
	}
}

func TestDifferentSchemaTypes(t *testing.T) {
	// Test that different schema types have appropriate fields
	schemas := map[string][]*schema.Field{
		"support":   supportSchema(),
		"billing":   billingSchema(),
		"technical": technicalSchema(),
	}

	// All schemas should have email as required field
	for schemaType, fields := range schemas {
		hasEmail := false
		for _, field := range fields {
			if field.Name() == "email" && field.Required() {
				hasEmail = true
				break
			}
		}
		if !hasEmail {
			t.Errorf("Schema %s should have required email field", schemaType)
		}
	}

	// Billing schema should have account_number
	billingFields := schemas["billing"]
	hasAccountNumber := false
	for _, field := range billingFields {
		if field.Name() == "account_number" && field.Required() {
			hasAccountNumber = true
			break
		}
	}
	if !hasAccountNumber {
		t.Error("Billing schema should have required account_number field")
	}

	// Technical schema should have error_message
	technicalFields := schemas["technical"]
	hasErrorMessage := false
	for _, field := range technicalFields {
		if field.Name() == "error_message" && field.Required() {
			hasErrorMessage = true
			break
		}
	}
	if !hasErrorMessage {
		t.Error("Technical schema should have required error_message field")
	}
}