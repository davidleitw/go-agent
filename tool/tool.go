package tool

import "context"

// Tool defines the interface for executable tools
type Tool interface {
	// Definition returns the tool's definition for LLM
	Definition() Definition

	// Execute runs the tool with given parameters
	Execute(ctx context.Context, params map[string]any) (any, error)
}

// TODO: Future enhancements
// - Add validation for input parameters
// - Support for output schema validation (optional)
// - Async execution support
// - Tool middleware/interceptors
