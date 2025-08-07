package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// mockTool implements Tool interface for testing
type mockTool struct {
	name        string
	description string
	execute     func(ctx context.Context, params map[string]any) (any, error)
}

func (m *mockTool) Definition() Definition {
	return Definition{
		Type: "function",
		Function: Function{
			Name:        m.name,
			Description: m.description,
			Parameters: Parameters{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
	}
}

func (m *mockTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	if m.execute != nil {
		return m.execute(ctx, params)
	}
	return "mock result", nil
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	tool1 := &mockTool{name: "test_tool", description: "A test tool"}
	err := registry.Register(tool1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = registry.Register(tool1)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test registration without name
	toolNoName := &mockTool{name: "", description: "No name"}
	err = registry.Register(toolNoName)
	if err == nil {
		t.Error("Expected error for tool without name")
	}
}

func TestRegistry_Execute(t *testing.T) {
	registry := NewRegistry()

	// Register a test tool
	executed := false
	tool := &mockTool{
		name:        "calculator",
		description: "Performs calculations",
		execute: func(ctx context.Context, params map[string]any) (any, error) {
			executed = true
			if op, ok := params["operation"].(string); ok && op == "add" {
				return 42, nil
			}
			return nil, errors.New("invalid operation")
		},
	}
	registry.Register(tool)

	// Test successful execution
	call := Call{
		ID: "test-call-1",
		Function: FunctionCall{
			Name:      "calculator",
			Arguments: `{"operation": "add"}`,
		},
	}

	result, err := registry.Execute(context.Background(), call)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !executed {
		t.Error("Tool was not executed")
	}
	if result != 42 {
		t.Errorf("Expected result 42, got %v", result)
	}

	// Test tool not found
	notFoundCall := Call{
		ID: "test-call-2",
		Function: FunctionCall{
			Name:      "non_existent",
			Arguments: `{}`,
		},
	}
	_, err = registry.Execute(context.Background(), notFoundCall)
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}

	// Test invalid JSON arguments
	invalidCall := Call{
		ID: "test-call-3",
		Function: FunctionCall{
			Name:      "calculator",
			Arguments: `{invalid json}`,
		},
	}
	_, err = registry.Execute(context.Background(), invalidCall)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestRegistry_GetDefinitions(t *testing.T) {
	registry := NewRegistry()

	// Register multiple tools
	tool1 := &mockTool{name: "tool1", description: "First tool"}
	tool2 := &mockTool{name: "tool2", description: "Second tool"}

	registry.Register(tool1)
	registry.Register(tool2)

	definitions := registry.GetDefinitions()
	if len(definitions) != 2 {
		t.Errorf("Expected 2 definitions, got %d", len(definitions))
	}

	// Check that both tools are in the definitions
	names := make(map[string]bool)
	for _, def := range definitions {
		names[def.Function.Name] = true
	}

	if !names["tool1"] || !names["tool2"] {
		t.Error("Not all tools found in definitions")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	tool := &mockTool{name: "test_tool", description: "Test"}
	registry.Register(tool)

	// Test getting existing tool
	retrieved, exists := registry.Get("test_tool")
	if !exists {
		t.Error("Expected tool to exist")
	}
	if retrieved == nil {
		t.Error("Expected non-nil tool")
	}

	// Test getting non-existent tool
	_, exists = registry.Get("non_existent")
	if exists {
		t.Error("Expected tool to not exist")
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	// Register some tools
	registry.Register(&mockTool{name: "tool1", description: "Tool 1"})
	registry.Register(&mockTool{name: "tool2", description: "Tool 2"})

	// Verify tools exist
	if len(registry.GetDefinitions()) != 2 {
		t.Error("Expected 2 tools before clear")
	}

	// Clear registry
	registry.Clear()

	// Verify no tools remain
	if len(registry.GetDefinitions()) != 0 {
		t.Error("Expected 0 tools after clear")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Register initial tool
	registry.Register(&mockTool{name: "concurrent_tool", description: "Test"})

	// Run concurrent operations
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			tool := &mockTool{
				name:        "dynamic_tool",
				description: "Dynamic",
			}
			registry.Register(tool)
			registry.Clear()
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			registry.GetDefinitions()
			registry.Get("concurrent_tool")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we get here without deadlock or panic, concurrent access is safe
}

func TestCall_ArgumentsParsing(t *testing.T) {
	// Test that Arguments field is properly handled as JSON
	call := Call{
		ID: "test",
		Function: FunctionCall{
			Name:      "test_func",
			Arguments: `{"key": "value", "number": 42}`,
		},
	}

	var params map[string]any
	err := json.Unmarshal([]byte(call.Function.Arguments), &params)
	if err != nil {
		t.Errorf("Failed to parse arguments: %v", err)
	}

	if params["key"] != "value" {
		t.Errorf("Expected key='value', got %v", params["key"])
	}

	// JSON numbers are float64 by default
	if num, ok := params["number"].(float64); !ok || num != 42 {
		t.Errorf("Expected number=42, got %v", params["number"])
	}
}
