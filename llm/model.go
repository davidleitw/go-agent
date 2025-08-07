package llm

import "context"

// Model defines the core interface for language models
type Model interface {
	// Complete performs a synchronous completion
	Complete(ctx context.Context, request Request) (*Response, error)

	// TODO: Future implementation for streaming
	// Stream(ctx context.Context, request Request) (<-chan StreamEvent, error)
}
