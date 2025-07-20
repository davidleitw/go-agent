package agent

import (
	"context"

	"github.com/davidleitw/go-agent/session"
)

// Agent represents the main agent interface
type Agent interface {
	// Execute runs the agent with the given request
	Execute(ctx context.Context, request Request) (*Response, error)
}

// Request represents a request to the agent
type Request struct {
	// Input is the user input or instruction
	Input string
	
	// SessionID is optional - if empty, agent creates new session
	SessionID string
}

// Response represents the agent's response
type Response struct {
	// Output is the agent's response content
	Output string
	
	// SessionID of the session used for this interaction
	SessionID string
	
	// Session provides access to the updated session
	Session session.Session
	
	// Metadata contains additional response information
	Metadata map[string]any
	
	// Usage contains token and resource usage information
	Usage Usage
}

// Usage represents resource usage information
type Usage struct {
	// LLMTokens tracks language model token usage
	LLMTokens TokenUsage
	
	// ToolCalls tracks number of tool executions
	ToolCalls int
	
	// SessionWrites tracks session state modifications
	SessionWrites int
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	// PromptTokens used for input
	PromptTokens int
	
	// CompletionTokens generated in response
	CompletionTokens int
	
	// TotalTokens is the sum of prompt and completion tokens
	TotalTokens int
}