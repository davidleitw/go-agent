package prompt

import (
	"context"
	"testing"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/session"
	"github.com/davidleitw/go-agent/session/memory"
)

// Mock provider for testing
type mockProvider struct {
	providerType string
	contexts     []agentcontext.Context
}

func newMockProvider(providerType string, contexts ...agentcontext.Context) *mockProvider {
	return &mockProvider{
		providerType: providerType,
		contexts:     contexts,
	}
}

func (p *mockProvider) Type() string {
	return p.providerType
}

func (p *mockProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
	return p.contexts
}

// Mock named provider for testing
type mockNamedProvider struct {
	mockProvider
	name string
}

func newMockNamedProvider(providerType, name string, contexts ...agentcontext.Context) *mockNamedProvider {
	return &mockNamedProvider{
		mockProvider: mockProvider{
			providerType: providerType,
			contexts:     contexts,
		},
		name: name,
	}
}

func (p *mockNamedProvider) Name() string {
	return p.name
}

func TestTemplateRender_BasicVariables(t *testing.T) {
	template := Parse("{{system}}\n{{history}}\n{{user_input}}")

	systemProvider := newMockProvider("system", agentcontext.Context{
		Type:    "system",
		Content: "You are a helpful assistant",
	})

	historyProvider := newMockProvider("history", agentcontext.Context{
		Type:    "history",
		Content: "Previous conversation",
		Metadata: map[string]any{
			"original_role": "user",
		},
	})

	providers := []agentcontext.Provider{systemProvider, historyProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "Hello, how are you?")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check system message
	if messages[0].Role != "system" || messages[0].Content != "You are a helpful assistant" {
		t.Errorf("System message incorrect: %+v", messages[0])
	}

	// Check history message
	if messages[1].Role != "user" || messages[1].Content != "Previous conversation" {
		t.Errorf("History message incorrect: %+v", messages[1])
	}

	// Check user input message
	if messages[2].Role != "user" || messages[2].Content != "Hello, how are you?" {
		t.Errorf("User input message incorrect: %+v", messages[2])
	}
}

func TestTemplateRender_CustomProviders(t *testing.T) {
	template := Parse("{{system}}\n{{custom_context}}\n{{user_input}}")

	systemProvider := newMockProvider("system", agentcontext.Context{
		Type:    "system",
		Content: "System prompt",
	})

	customProvider := newMockProvider("custom_context", agentcontext.Context{
		Type:    "custom_context",
		Content: "Custom context information",
	})

	providers := []agentcontext.Provider{systemProvider, customProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "User question")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check custom context is rendered as system message
	if messages[1].Role != "system" || messages[1].Content != "Custom context information" {
		t.Errorf("Custom context message incorrect: %+v", messages[1])
	}
}

func TestTemplateRender_NamedProviders(t *testing.T) {
	template := Parse("{{project_info:main}}\n{{project_info:secondary}}\n{{user_input}}")

	mainProvider := newMockNamedProvider("project_info", "main", agentcontext.Context{
		Type:    "project_info",
		Content: "Main project information",
	})

	secondaryProvider := newMockNamedProvider("project_info", "secondary", agentcontext.Context{
		Type:    "project_info",
		Content: "Secondary project information",
	})

	// Add a provider without the correct name (should be ignored)
	ignoredProvider := newMockNamedProvider("project_info", "ignored", agentcontext.Context{
		Type:    "project_info",
		Content: "Should be ignored",
	})

	providers := []agentcontext.Provider{mainProvider, secondaryProvider, ignoredProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "User question")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check main provider content
	if messages[0].Content != "Main project information" {
		t.Errorf("Main provider content incorrect: %s", messages[0].Content)
	}

	// Check secondary provider content
	if messages[1].Content != "Secondary project information" {
		t.Errorf("Secondary provider content incorrect: %s", messages[1].Content)
	}
}

func TestTemplateRender_EmptyUserInput(t *testing.T) {
	template := Parse("{{system}}\n{{user_input}}")

	systemProvider := newMockProvider("system", agentcontext.Context{
		Type:    "system",
		Content: "System prompt",
	})

	providers := []agentcontext.Provider{systemProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should only have system message, user_input should be skipped when empty
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected system message, got role: %s", messages[0].Role)
	}
}

func TestTemplateRender_StaticText(t *testing.T) {
	template := Parse("Instructions: Be helpful\n{{system}}\n{{user_input}}")

	systemProvider := newMockProvider("system", agentcontext.Context{
		Type:    "system",
		Content: "System prompt",
	})

	providers := []agentcontext.Provider{systemProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "Hello")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check static text becomes system message
	if messages[0].Role != "system" || messages[0].Content != "Instructions: Be helpful" {
		t.Errorf("Static text message incorrect: %+v", messages[0])
	}
}

func TestTemplateRender_MultipleContextsFromProvider(t *testing.T) {
	template := Parse("{{system}}\n{{user_input}}")

	systemProvider := newMockProvider("system",
		agentcontext.Context{
			Type:    "system",
			Content: "First system context",
		},
		agentcontext.Context{
			Type:    "system",
			Content: "Second system context",
		},
	)

	providers := []agentcontext.Provider{systemProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "Hello")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Multiple system contexts should be combined
	expectedContent := "First system context\n\nSecond system context"
	if messages[0].Content != expectedContent {
		t.Errorf("Combined system content incorrect: %s", messages[0].Content)
	}
}

func TestTemplateRender_HistoryWithOriginalRoles(t *testing.T) {
	template := Parse("{{history}}\n{{user_input}}")

	historyProvider := newMockProvider("history",
		agentcontext.Context{
			Type:    "history",
			Content: "User message",
			Metadata: map[string]any{
				"original_role": "user",
			},
		},
		agentcontext.Context{
			Type:    "history",
			Content: "Assistant response",
			Metadata: map[string]any{
				"original_role": "assistant",
			},
		},
	)

	providers := []agentcontext.Provider{historyProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "New question")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check original roles are preserved
	if messages[0].Role != "user" || messages[0].Content != "User message" {
		t.Errorf("History user message incorrect: %+v", messages[0])
	}

	if messages[1].Role != "assistant" || messages[1].Content != "Assistant response" {
		t.Errorf("History assistant message incorrect: %+v", messages[1])
	}
}

func TestTemplateRender_NoMatchingProviders(t *testing.T) {
	template := Parse("{{nonexistent}}\n{{user_input}}")

	systemProvider := newMockProvider("system", agentcontext.Context{
		Type:    "system",
		Content: "System prompt",
	})

	providers := []agentcontext.Provider{systemProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "Hello")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should only have user input, nonexistent variable produces no messages
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Content != "Hello" {
		t.Errorf("User input message incorrect: %+v", messages[0])
	}
}

func TestTemplateVariables(t *testing.T) {
	template := Parse("{{system}}\n{{history}}\n{{custom}}\n{{user_input}}")

	variables := template.Variables()
	expected := []string{"system", "history", "custom", "user_input"}

	if len(variables) != len(expected) {
		t.Fatalf("Expected %d variables, got %d", len(expected), len(variables))
	}

	for i, expectedVar := range expected {
		if variables[i] != expectedVar {
			t.Errorf("Variable %d: expected %s, got %s", i, expectedVar, variables[i])
		}
	}
}

func TestTemplateString(t *testing.T) {
	original := "{{system}}\n{{history}}\n{{user_input}}"
	template := Parse(original)

	if template.String() != original {
		t.Errorf("Template string mismatch: expected %s, got %s", original, template.String())
	}
}

func TestTemplateExplain(t *testing.T) {
	template := Parse("Static text\n{{system}}\n{{custom:named}}\n{{user_input}}")

	explanation := template.Explain()

	// Check that explanation contains key information
	if !containsString(explanation, "Static text") {
		t.Error("Explanation should mention static text")
	}

	if !containsString(explanation, "system") {
		t.Error("Explanation should mention system variable")
	}

	if !containsString(explanation, "custom:named") {
		t.Error("Explanation should mention named variable")
	}

	if !containsString(explanation, "user_input") {
		t.Error("Explanation should mention user_input variable")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestRenderContexts_SystemMessages(t *testing.T) {
	template := &promptTemplate{}

	contexts := []agentcontext.Context{
		{Type: "system", Content: "First system message"},
		{Type: "system", Content: "Second system message"},
	}

	messages := template.renderSystemContexts(contexts)

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	expected := "First system message\n\nSecond system message"
	if messages[0].Content != expected {
		t.Errorf("Combined content incorrect: %s", messages[0].Content)
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected system role, got %s", messages[0].Role)
	}
}

func TestRenderContexts_HistoryMessages(t *testing.T) {
	template := &promptTemplate{}

	contexts := []agentcontext.Context{
		{
			Type:     "history",
			Content:  "User message",
			Metadata: map[string]any{"original_role": "user"},
		},
		{
			Type:     "history",
			Content:  "Assistant message",
			Metadata: map[string]any{"original_role": "assistant"},
		},
		{
			Type:     "history",
			Content:  "Message without role",
			Metadata: map[string]any{},
		},
	}

	messages := template.renderHistoryContexts(contexts)

	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check roles are preserved from metadata
	if messages[0].Role != "user" {
		t.Errorf("Expected user role, got %s", messages[0].Role)
	}

	if messages[1].Role != "assistant" {
		t.Errorf("Expected assistant role, got %s", messages[1].Role)
	}

	// Check default role when no original_role in metadata
	if messages[2].Role != "user" {
		t.Errorf("Expected default user role, got %s", messages[2].Role)
	}
}

// TestRenderContexts_UserInputMessages has been removed as renderUserInputContexts() method
// was deleted during cleanup. User input is now handled directly in the main Render() method

func TestRenderContexts_EmptyContexts(t *testing.T) {
	template := &promptTemplate{}

	// Test empty contexts return nil
	messages := template.renderSystemContexts([]agentcontext.Context{})
	if messages != nil {
		t.Error("Expected nil for empty contexts")
	}

	// Test contexts with only empty content - should return nil because empty strings are filtered out
	emptyContexts := []agentcontext.Context{
		{Type: "system", Content: ""},
		{Type: "system", Content: ""},
	}

	messages = template.renderSystemContexts(emptyContexts)
	if messages != nil {
		t.Error("Expected nil for contexts with empty content")
	}

	// Test contexts with whitespace content - should create a message with just the whitespace
	whitespaceContexts := []agentcontext.Context{
		{Type: "system", Content: "   "},
	}

	messages = template.renderSystemContexts(whitespaceContexts)
	if len(messages) != 1 {
		t.Errorf("Expected 1 message for contexts with whitespace content, got %d", len(messages))
	}

	if messages[0].Content != "   " {
		t.Errorf("Expected whitespace content, got: '%s'", messages[0].Content)
	}
}
