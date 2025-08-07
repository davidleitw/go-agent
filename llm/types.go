package llm

import "github.com/davidleitw/go-agent/tool"

// Request represents a completion request
type Request struct {
	Messages []Message

	// Optional model parameters
	Temperature *float32 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`

	// Optional tool definitions
	Tools []tool.Definition `json:"tools,omitempty"`

	// TODO: Future enhancements
	// - Model override
	// - TopP, StopSequences
	// - ToolChoice strategies
	// - User identification for rate limiting
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"` // system/user/assistant/tool
	Content string `json:"content"`

	// For tool-related messages
	Name       string      `json:"name,omitempty"`         // tool name
	ToolCallID string      `json:"tool_call_id,omitempty"` // for tool responses
	ToolCalls  []tool.Call `json:"tool_calls,omitempty"`   // for assistant messages with tool calls
}

// Response represents model completion response
type Response struct {
	Content   string      `json:"content"`
	ToolCalls []tool.Call `json:"tool_calls,omitempty"`

	// Basic metadata
	Usage        Usage  `json:"usage"`
	FinishReason string `json:"finish_reason"` // stop/length/tool_calls
}

// Usage tracks token consumption
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
