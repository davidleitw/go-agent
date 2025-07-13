package main

import (
	"context"
	"testing"

	"github.com/davidleitw/go-agent/pkg/agent"
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

func TestIntentClassifier(t *testing.T) {
	classifier := NewIntentClassifier()

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "I'm getting a login error",
			expected: "technical_support",
		},
		{
			input:    "I have a billing question about my invoice",
			expected: "billing_inquiry",
		},
		{
			input:    "I need to change my password",
			expected: "account_management",
		},
		{
			input:    "How do I use the analytics feature?",
			expected: "product_inquiry",
		},
		{
			input:    "I want to purchase your enterprise plan",
			expected: "sales_inquiry",
		},
		{
			input:    "Hello, I have some questions",
			expected: "general_inquiry",
		},
	}

	for _, tc := range testCases {
		result := classifier.ClassifyIntent(tc.input)
		if result != tc.expected {
			t.Errorf("Expected intent '%s' for input '%s', got '%s'", tc.expected, tc.input, result)
		}
	}
}

func TestGetSchemaForIntent(t *testing.T) {
	// Test that different intents return appropriate schemas
	testCases := []struct {
		intent           string
		expectedRequired []string
		expectedOptional []string
	}{
		{
			intent:           "technical_support",
			expectedRequired: []string{"email", "error_description", "steps_taken"},
			expectedOptional: []string{"environment", "urgency"},
		},
		{
			intent:           "billing_inquiry",
			expectedRequired: []string{"email", "account_id", "billing_question"},
			expectedOptional: []string{"amount", "transaction_date"},
		},
		{
			intent:           "account_management",
			expectedRequired: []string{"email", "request_type", "reason"},
			expectedOptional: []string{"verification_code"},
		},
		{
			intent:           "sales_inquiry",
			expectedRequired: []string{"email", "company", "team_size", "use_case"},
			expectedOptional: []string{"timeline", "budget"},
		},
	}

	for _, tc := range testCases {
		schema := getSchemaForIntent(tc.intent)

		// Check required fields
		requiredFields := make([]string, 0)
		optionalFields := make([]string, 0)

		for _, field := range schema {
			if field.Required() {
				requiredFields = append(requiredFields, field.Name())
			} else {
				optionalFields = append(optionalFields, field.Name())
			}
		}

		// Verify required fields
		if len(requiredFields) != len(tc.expectedRequired) {
			t.Errorf("Intent %s: expected %d required fields, got %d", tc.intent, len(tc.expectedRequired), len(requiredFields))
		}

		for _, expected := range tc.expectedRequired {
			found := false
			for _, actual := range requiredFields {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Intent %s: missing required field '%s'", tc.intent, expected)
			}
		}

		// Verify optional fields
		if len(optionalFields) != len(tc.expectedOptional) {
			t.Errorf("Intent %s: expected %d optional fields, got %d", tc.intent, len(tc.expectedOptional), len(optionalFields))
		}

		for _, expected := range tc.expectedOptional {
			found := false
			for _, actual := range optionalFields {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Intent %s: missing optional field '%s'", tc.intent, expected)
			}
		}
	}
}

func TestGetWorkflowForIntent(t *testing.T) {
	// Test multi-step workflows
	testCases := []struct {
		intent        string
		expectedSteps int
	}{
		{
			intent:        "technical_support",
			expectedSteps: 3,
		},
		{
			intent:        "sales_inquiry",
			expectedSteps: 3,
		},
		{
			intent:        "general_inquiry",
			expectedSteps: 1,
		},
	}

	for _, tc := range testCases {
		workflow := getWorkflowForIntent(tc.intent)
		if len(workflow) != tc.expectedSteps {
			t.Errorf("Intent %s: expected %d workflow steps, got %d", tc.intent, tc.expectedSteps, len(workflow))
		}

		// Ensure each step has at least one field
		for i, step := range workflow {
			if len(step) == 0 {
				t.Errorf("Intent %s: workflow step %d has no fields", tc.intent, i+1)
			}
		}
	}
}

func TestDynamicSchemaWithBot(t *testing.T) {
	// Test dynamic schema selection with agent
	mockModel := NewMockChatModel(
		`{"email": null, "error_description": null, "steps_taken": null}`,
		"I understand you're having a technical issue. Please provide your email address and describe the error you're experiencing.",
	)

	bot, err := agent.New("test-adaptive-bot").
		WithChatModel(mockModel).
		WithDescription("Test adaptive bot").
		WithInstructions("You are an intelligent assistant.").
		Build()
	if err != nil {
		t.Fatalf("Failed to create test bot: %v", err)
	}

	ctx := context.Background()
	classifier := NewIntentClassifier()

	// Test technical support intent
	userInput := "I'm getting a login error"
	intent := classifier.ClassifyIntent(userInput)
	
	if intent != "technical_support" {
		t.Errorf("Expected technical_support intent, got %s", intent)
	}

	schema := getSchemaForIntent(intent)
	
	response, err := bot.Chat(ctx, userInput,
		agent.WithSchema(schema...),
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

func TestSchemaFieldsForAllIntents(t *testing.T) {
	// Ensure all intents return valid schemas
	intents := []string{
		"technical_support",
		"billing_inquiry", 
		"account_management",
		"product_inquiry",
		"sales_inquiry",
		"general_inquiry",
	}

	for _, intent := range intents {
		schema := getSchemaForIntent(intent)
		
		if len(schema) == 0 {
			t.Errorf("Intent %s returned empty schema", intent)
		}

		// All schemas should have email field
		hasEmail := false
		for _, field := range schema {
			if field.Name() == "email" {
				hasEmail = true
				break
			}
		}
		if !hasEmail {
			t.Errorf("Intent %s schema missing email field", intent)
		}

		// Validate field structure
		for _, field := range schema {
			if field.Name() == "" {
				t.Errorf("Intent %s has field with empty name", intent)
			}
			if field.Prompt() == "" {
				t.Errorf("Intent %s has field with empty prompt", intent)
			}
		}
	}
}

func TestIntentKeywordMatching(t *testing.T) {
	classifier := NewIntentClassifier()

	// Test that multiple keywords increase classification confidence
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "I have a technical error with the login system",
			expected: "technical_support", // Should match multiple keywords
		},
		{
			input:    "billing payment charge refund issue",
			expected: "billing_inquiry", // Multiple billing keywords
		},
		{
			input:    "enterprise sales demo pricing quote",
			expected: "sales_inquiry", // Multiple sales keywords
		},
	}

	for _, tc := range testCases {
		result := classifier.ClassifyIntent(tc.input)
		if result != tc.expected {
			t.Errorf("Expected intent '%s' for input '%s', got '%s'", tc.expected, tc.input, result)
		}
	}
}