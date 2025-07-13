package agent

import (
	"context"
	"time"
)

// Message represents a single turn in a conversation
type Message struct {
	// Role identifies the message sender (user, assistant, system, tool)
	Role string `json:"role"`

	// Content is the text content of the message
	Content string `json:"content"`

	// ToolCalls contains tool calls requested by the assistant
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is set when this message is a response to a tool call
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Name is used for tool messages to identify which tool was called
	Name string `json:"name,omitempty"`

	// Timestamp records when this message was created
	Timestamp time.Time `json:"timestamp"`
}

// ToolCall represents a request to call a tool
type ToolCall struct {
	// ID is a unique identifier for this tool call
	ID string `json:"id"`

	// Type is the type of tool call (currently always "function")
	Type string `json:"type"`

	// Function contains the function call details
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction contains the details of a function call
type ToolCallFunction struct {
	// Name is the name of the function to call
	Name string `json:"name"`

	// Arguments is the JSON-encoded arguments for the function
	Arguments string `json:"arguments"`
}

// MessageRole constants for different message types
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// Session represents a conversation's history and state
type Session interface {
	// ID returns the unique identifier for this session
	ID() string

	// Messages returns all messages in the session
	Messages() []Message

	// AddMessage adds a new message to the session
	AddMessage(msg Message)

	// AddUserMessage is a convenience method to add a user message
	AddUserMessage(content string)

	// AddAssistantMessage is a convenience method to add an assistant message
	AddAssistantMessage(content string)

	// AddSystemMessage is a convenience method to add a system message
	AddSystemMessage(content string)

	// AddToolMessage is a convenience method to add a tool response message
	AddToolMessage(toolCallID, toolName, content string)

	// CreatedAt returns when the session was created
	CreatedAt() time.Time

	// UpdatedAt returns when the session was last updated
	UpdatedAt() time.Time

	// Clear removes all messages from the session
	Clear()

	// Clone creates a deep copy of the session
	Clone() Session
}

// SessionStore handles conversation history persistence
type SessionStore interface {
	// Save persists a session
	Save(ctx context.Context, session Session) error

	// Load retrieves a session by ID
	Load(ctx context.Context, sessionID string) (Session, error)

	// Delete removes a session
	Delete(ctx context.Context, sessionID string) error

	// List returns all session IDs, optionally filtered
	List(ctx context.Context, filter SessionFilter) ([]string, error)

	// Exists checks if a session exists
	Exists(ctx context.Context, sessionID string) (bool, error)
}

// SessionFilter provides criteria for filtering sessions
type SessionFilter struct {
	// IDPrefix filters sessions with IDs starting with this prefix
	IDPrefix string

	// CreatedAfter filters sessions created after this time
	CreatedAfter time.Time

	// CreatedBefore filters sessions created before this time
	CreatedBefore time.Time

	// UpdatedAfter filters sessions updated after this time
	UpdatedAfter time.Time

	// UpdatedBefore filters sessions updated before this time
	UpdatedBefore time.Time

	// MinMessages filters sessions with at least this many messages
	MinMessages int

	// MaxMessages filters sessions with at most this many messages
	MaxMessages int

	// Limit limits the number of results (0 = no limit)
	Limit int

	// Offset skips the first N results
	Offset int
}

// Session factory functions are now defined in agent.go to avoid duplication

// NewMessage creates a new message with the given role and content
func NewMessage(role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewUserMessage creates a new user message
func NewUserMessage(content string) Message {
	return NewMessage(RoleUser, content)
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(content string) Message {
	return NewMessage(RoleAssistant, content)
}

// NewSystemMessage creates a new system message
func NewSystemMessage(content string) Message {
	return NewMessage(RoleSystem, content)
}

// NewToolMessage creates a new tool response message
func NewToolMessage(toolCallID, toolName, content string) Message {
	return Message{
		Role:       RoleTool,
		Content:    content,
		ToolCallID: toolCallID,
		Name:       toolName,
		Timestamp:  time.Now(),
	}
}

// NewAssistantMessageWithToolCalls creates an assistant message with tool calls
func NewAssistantMessageWithToolCalls(content string, toolCalls []ToolCall) Message {
	return Message{
		Role:      RoleAssistant,
		Content:   content,
		ToolCalls: toolCalls,
		Timestamp: time.Now(),
	}
}

// HasToolCalls returns true if the message contains tool calls
func (m *Message) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}

// IsUserMessage returns true if this is a user message
func (m *Message) IsUserMessage() bool {
	return m.Role == RoleUser
}

// IsAssistantMessage returns true if this is an assistant message
func (m *Message) IsAssistantMessage() bool {
	return m.Role == RoleAssistant
}

// IsSystemMessage returns true if this is a system message
func (m *Message) IsSystemMessage() bool {
	return m.Role == RoleSystem
}

// IsToolMessage returns true if this is a tool message
func (m *Message) IsToolMessage() bool {
	return m.Role == RoleTool
}

// Validate checks if the message is valid
func (m *Message) Validate() error {
	if m.Role == "" {
		return ErrInvalidInput
	}

	switch m.Role {
	case RoleUser, RoleAssistant, RoleSystem:
		// These roles should have content
		if m.Content == "" && len(m.ToolCalls) == 0 {
			return ErrInvalidInput
		}
	case RoleTool:
		// Tool messages must have tool_call_id and name
		if m.ToolCallID == "" || m.Name == "" {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}

	return nil
}

// Validate checks if the tool call is valid
func (tc *ToolCall) Validate() error {
	if tc.ID == "" {
		return ErrInvalidInput
	}
	if tc.Type == "" {
		tc.Type = "function" // Default type
	}
	if tc.Function.Name == "" {
		return ErrInvalidInput
	}
	return nil
}
