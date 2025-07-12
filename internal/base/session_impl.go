package base

import (
	"sync"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
)

// sessionImpl is the concrete implementation of the Session interface
type sessionImpl struct {
	id        string
	messages  []agent.Message
	createdAt time.Time
	updatedAt time.Time
	mutex     sync.RWMutex
}

// NewSession creates a new Session implementation with the given ID
func NewSession(id string) agent.Session {
	now := time.Now()
	return &sessionImpl{
		id:        id,
		messages:  make([]agent.Message, 0),
		createdAt: now,
		updatedAt: now,
	}
}

// ID returns the unique identifier for this session
func (s *sessionImpl) ID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.id
}

// Messages returns all messages in the session
func (s *sessionImpl) Messages() []agent.Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	result := make([]agent.Message, len(s.messages))
	copy(result, s.messages)
	return result
}

// AddMessage adds a new message to the session
func (s *sessionImpl) AddMessage(msg agent.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Set timestamp if not provided
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	
	s.messages = append(s.messages, msg)
	s.updatedAt = time.Now()
}

// AddUserMessage is a convenience method to add a user message
func (s *sessionImpl) AddUserMessage(content string) {
	s.AddMessage(agent.Message{
		Role:      agent.RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddAssistantMessage is a convenience method to add an assistant message
func (s *sessionImpl) AddAssistantMessage(content string) {
	s.AddMessage(agent.Message{
		Role:      agent.RoleAssistant,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddSystemMessage is a convenience method to add a system message
func (s *sessionImpl) AddSystemMessage(content string) {
	s.AddMessage(agent.Message{
		Role:      agent.RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddToolMessage is a convenience method to add a tool response message
func (s *sessionImpl) AddToolMessage(toolCallID, toolName, content string) {
	s.AddMessage(agent.Message{
		Role:       agent.RoleTool,
		Content:    content,
		ToolCallID: toolCallID,
		Name:       toolName,
		Timestamp:  time.Now(),
	})
}

// CreatedAt returns when the session was created
func (s *sessionImpl) CreatedAt() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.createdAt
}

// UpdatedAt returns when the session was last updated
func (s *sessionImpl) UpdatedAt() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.updatedAt
}

// Clear removes all messages from the session
func (s *sessionImpl) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.messages = make([]agent.Message, 0)
	s.updatedAt = time.Now()
}

// Clone creates a deep copy of the session
func (s *sessionImpl) Clone() agent.Session {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	clone := &sessionImpl{
		id:        s.id,
		messages:  make([]agent.Message, len(s.messages)),
		createdAt: s.createdAt,
		updatedAt: s.updatedAt,
	}
	
	copy(clone.messages, s.messages)
	return clone
}

// ToOpenAIMessages converts the session messages to OpenAI format
// This is a helper method for OpenAI integration
func (s *sessionImpl) ToOpenAIMessages() []OpenAIMessage {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	messages := make([]OpenAIMessage, 0, len(s.messages))
	for _, msg := range s.messages {
		openaiMsg := OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		
		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			openaiMsg.ToolCalls = make([]OpenAIToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				openaiMsg.ToolCalls[i] = OpenAIToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: OpenAIFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
		
		// Set tool call ID if this is a tool message
		if msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}
		
		// Set name for tool messages
		if msg.Name != "" {
			openaiMsg.Name = msg.Name
		}
		
		messages = append(messages, openaiMsg)
	}
	
	return messages
}

// OpenAI-specific types for conversion
type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type OpenAIToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

type OpenAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}