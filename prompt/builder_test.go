package prompt

import (
	"context"
	"testing"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/session/memory"
)

func TestBuilder_BasicMethods(t *testing.T) {
	template := New().
		System().
		History().
		UserInput().
		Build()

	variables := template.Variables()
	expected := []string{"system", "history", "user_input"}

	if len(variables) != len(expected) {
		t.Fatalf("Expected %d variables, got %d", len(expected), len(variables))
	}

	for i, v := range variables {
		if v != expected[i] {
			t.Errorf("Variable %d: expected %s, got %s", i, expected[i], v)
		}
	}
}

func TestBuilder_CustomProviders(t *testing.T) {
	template := New().
		Provider("custom_context").
		NamedProvider("project_info", "main").
		Build()

	variables := template.Variables()
	expected := []string{"custom_context", "project_info"}

	if len(variables) != len(expected) {
		t.Fatalf("Expected %d variables, got %d", len(expected), len(variables))
	}

	for i, v := range variables {
		if v != expected[i] {
			t.Errorf("Variable %d: expected %s, got %s", i, expected[i], v)
		}
	}
}

func TestBuilder_TextMethods(t *testing.T) {
	template := New().
		Text("Static text").
		Line("Line with newline").
		System().
		Build()

	// Test that the template renders text correctly
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

	// Should have 3 messages: static text, line text, system
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	if messages[0].Content != "Static text" {
		t.Errorf("Static text incorrect: %s", messages[0].Content)
	}

	if messages[1].Content != "Line with newline" {
		t.Errorf("Line text incorrect: %s", messages[1].Content)
	}

	if messages[2].Content != "System prompt" {
		t.Errorf("System content incorrect: %s", messages[2].Content)
	}
}

func TestBuilder_DefaultFlow(t *testing.T) {
	template := New().
		DefaultFlow().
		Build()

	variables := template.Variables()

	// DefaultFlow should include system, history, context_providers, user_input
	expectedVars := []string{"system", "history", "context_providers", "user_input"}

	if len(variables) != len(expectedVars) {
		t.Fatalf("Expected %d variables, got %d", len(expectedVars), len(variables))
	}

	for i, expected := range expectedVars {
		if variables[i] != expected {
			t.Errorf("Variable %d: expected %s, got %s", i, expected, variables[i])
		}
	}
}

func TestBuilder_Separator(t *testing.T) {
	template := New().
		System().
		Separator().
		UserInput().
		Build()

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

	// Should have system, separator, user input
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check separator becomes a system message
	if messages[1].Role != "system" || messages[1].Content != "---" {
		t.Errorf("Separator message incorrect: %+v", messages[1])
	}
}

func TestBuilder_EmptyText(t *testing.T) {
	template := New().
		System().
		Text(""). // Empty text should be ignored
		UserInput().
		Build()

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

	// Should only have system and user input (empty text ignored)
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}
}

func TestBuilder_ComplexTemplate(t *testing.T) {
	template := New().
		Text("You are a helpful assistant.").
		Line("").
		Text("Project Context:").
		Provider("project_info").
		Line("").
		Text("Conversation History:").
		History().
		Line("").
		Text("User Question:").
		UserInput().
		Build()

	// Test with multiple providers
	projectProvider := newMockProvider("project_info", agentcontext.Context{
		Type:    "project_info",
		Content: "Go Agent Framework",
	})

	historyProvider := newMockProvider("history", agentcontext.Context{
		Type:    "history",
		Content: "Previous conversation",
		Metadata: map[string]any{
			"original_role": "user",
		},
	})

	providers := []agentcontext.Provider{projectProvider, historyProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "How does this work?")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should have: helpful assistant text, empty line, project context text, project content, empty line, history text, history, empty line, question text, user input
	// But empty lines (Line("")) are filtered out during parsing, so we get 7 messages
	expectedRoles := []string{"system", "system", "system", "system", "user", "system", "user"}

	if len(messages) != len(expectedRoles) {
		t.Fatalf("Expected %d messages, got %d", len(expectedRoles), len(messages))
	}

	for i, expectedRole := range expectedRoles {
		if messages[i].Role != expectedRole {
			t.Errorf("Message %d role: expected %s, got %s", i, expectedRole, messages[i].Role)
		}
	}
}

// TestPredefinedTemplates has been removed as predefined template constants were deleted
// during cleanup. These templates are now built using the Builder API in DefaultFlow()

func TestParseCustomTemplate(t *testing.T) {
	// Test custom template string parsing functionality
	template := Parse("{{custom}}\n{{user_input}}")
	variables := template.Variables()
	expected := []string{"custom", "user_input"}

	if len(variables) != len(expected) {
		t.Fatalf("Expected %d variables, got %d", len(expected), len(variables))
	}

	for i, expectedVar := range expected {
		if variables[i] != expectedVar {
			t.Errorf("Variable %d: expected %s, got %s", i, expectedVar, variables[i])
		}
	}
}

// TestTemplateBuilder_toBuilder has been removed as toBuilder() method was deleted
// during cleanup. Template modification is now done through the Builder API

// TestBuilder_GenerateTemplateString has been removed as generateTemplateString() method
// was deleted during cleanup. Template string generation is handled internally by Build()

func TestBuilder_ChainedMethods(t *testing.T) {
	// Test that all methods return Builder for chaining
	builder := New()

	result := builder.
		System().
		History().
		UserInput().
		Provider("custom").
		NamedProvider("named", "test").
		Text("text").
		Line("line").
		DefaultFlow().
		Separator()

	// Should still be a Builder
	template := result.Build()
	if template == nil {
		t.Error("Chained methods should return a valid builder")
	}
}

func TestBuilder_EmptyBuilder(t *testing.T) {
	template := New().Build()

	// Empty builder should create valid but empty template
	if template == nil {
		t.Fatal("Empty builder should create valid template")
	}

	variables := template.Variables()
	if len(variables) != 0 {
		t.Errorf("Empty template should have 0 variables, got %d", len(variables))
	}

	// Empty template should render no messages
	messages := renderTemplateForTest(t, template)
	if len(messages) != 0 {
		t.Errorf("Empty template should render 0 messages, got %d", len(messages))
	}
}

// Helper function to render a template for testing
func renderTemplateForTest(t *testing.T, template Template) []interface{} {
	systemProvider := newMockProvider("system", agentcontext.Context{
		Type:    "system",
		Content: "Test system",
	})

	customProvider := newMockProvider("custom", agentcontext.Context{
		Type:    "custom",
		Content: "Test custom",
	})

	providers := []agentcontext.Provider{systemProvider, customProvider}
	store := memory.NewStore()
	session := store.Create(context.Background())

	messages, err := template.Render(context.Background(), providers, session, "Test input")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Convert to []interface{} for easier comparison
	result := make([]interface{}, len(messages))
	for i, msg := range messages {
		result[i] = msg
	}

	return result
}
