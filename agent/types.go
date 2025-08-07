package agent

import (
	"context"
	"errors"
	"time"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/prompt"
	"github.com/davidleitw/go-agent/session"
	"github.com/davidleitw/go-agent/tool"
)

// Engine interface defines the core execution engine
type Engine interface {
	// Execute handles the core agent logic with pre-configured components
	Execute(ctx context.Context, request Request) (*Response, error)
}

// HistoryInterceptor allows custom processing of conversation history
type HistoryInterceptor interface {
	// ProcessHistory processes session entries before they are converted to contexts
	// It can filter, compress, summarize, or otherwise modify the history
	ProcessHistory(ctx context.Context, entries []session.Entry, llm llm.Model) ([]session.Entry, error)
}

// EngineConfig provides configuration for engine construction
type EngineConfig struct {
	// Model is the LLM to use
	Model llm.Model

	// SessionStore for session persistence
	SessionStore session.SessionStore

	// ToolRegistry for available tools
	ToolRegistry *tool.Registry

	// ContextProviders for gathering context
	ContextProviders []agentcontext.Provider

	// PromptTemplate for organizing contexts into LLM messages
	PromptTemplate prompt.Template

	// MaxIterations limits agent thinking/tool loops
	MaxIterations int

	// Temperature for LLM calls
	Temperature *float32

	// MaxTokens for LLM responses
	MaxTokens *int

	// SessionTTL sets the default session time-to-live
	SessionTTL time.Duration

	// History configuration
	HistoryLimit int // Number of history entries to include (0 = disabled)

	// HistoryInterceptor for advanced history processing (optional)
	HistoryInterceptor HistoryInterceptor
}

// Common errors
var (
	// ErrInvalidInput indicates the input request is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrSessionNotFound indicates the requested session doesn't exist
	ErrSessionNotFound = errors.New("session not found")

	// ErrMaxIterationsExceeded indicates the agent hit iteration limit
	ErrMaxIterationsExceeded = errors.New("maximum iterations exceeded")

	// ErrToolExecutionFailed indicates a tool call failed
	ErrToolExecutionFailed = errors.New("tool execution failed")

	// ErrLLMCallFailed indicates the LLM request failed
	ErrLLMCallFailed = errors.New("LLM call failed")
)

// IterationResult represents the result of a single agent iteration
type IterationResult struct {
	// Completed indicates if the agent has finished
	Completed bool

	// LLMResponse from this iteration
	LLMResponse *llm.Response

	// ToolResults from executed tools
	ToolResults []ToolResult

	// Usage for this iteration
	Usage Usage

	// Error if iteration failed
	Error error
}

// ToolResult represents the result of executing a tool
type ToolResult struct {
	// Call that was executed
	Call tool.Call

	// Result returned by the tool
	Result any

	// Error if tool execution failed
	Error error
}

