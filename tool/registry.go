package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Registry manages tool registration and execution
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def := tool.Definition()
	if def.Function.Name == "" {
		return fmt.Errorf("tool must have a name")
	}

	if _, exists := r.tools[def.Function.Name]; exists {
		return fmt.Errorf("tool %s already registered", def.Function.Name)
	}

	r.tools[def.Function.Name] = tool
	return nil
}

// Execute runs a tool by name with given parameters
func (r *Registry) Execute(ctx context.Context, call Call) (any, error) {
	r.mu.RLock()
	tool, exists := r.tools[call.Function.Name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", call.Function.Name)
	}

	// Parse arguments from JSON string
	var params map[string]any
	if err := json.Unmarshal([]byte(call.Function.Arguments), &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute the tool
	result, err := tool.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// TODO: Future enhancements
	// - Validate output against schema if defined
	// - Add execution metrics/logging
	// - Support middleware/interceptors

	return result, nil
}

// GetDefinitions returns all registered tool definitions
func (r *Registry) GetDefinitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]Definition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, tool.Definition())
	}
	return definitions
}

// Get returns a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

// Clear removes all registered tools
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]Tool)
}
