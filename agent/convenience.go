package agent

import (
	"context"

	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/tool"
)

// Chat provides a simple interface for one-off agent interactions
func Chat(ctx context.Context, model llm.Model, input string) (string, error) {
	agent := NewSimpleAgent(model)
	
	response, err := agent.Execute(ctx, Request{
		Input: input,
	})
	if err != nil {
		return "", err
	}
	
	return response.Output, nil
}

// ChatWithTools provides simple interface for agent interactions with tools
func ChatWithTools(ctx context.Context, model llm.Model, input string, tools ...tool.Tool) (string, error) {
	agent := NewAgentWithTools(model, tools...)
	
	response, err := agent.Execute(ctx, Request{
		Input: input,
	})
	if err != nil {
		return "", err
	}
	
	return response.Output, nil
}

// Conversation manages a multi-turn conversation with session persistence
type Conversation struct {
	agent     Agent
	sessionID string
}

// NewConversation creates a new conversation with the given agent
func NewConversation(agent Agent) *Conversation {
	return &Conversation{
		agent:     agent,
		sessionID: "", // Will be set after first interaction
	}
}

// NewConversationWithModel creates a conversation with a simple agent
func NewConversationWithModel(model llm.Model) *Conversation {
	agent := NewConversationalAgent(model, 10) // 10 message history limit
	return NewConversation(agent)
}

// Say sends a message and returns the response, maintaining conversation context
func (c *Conversation) Say(ctx context.Context, input string) (string, error) {
	response, err := c.agent.Execute(ctx, Request{
		Input:     input,
		SessionID: c.sessionID,
	})
	if err != nil {
		return "", err
	}
	
	// Update session ID if it was empty
	if c.sessionID == "" {
		c.sessionID = response.SessionID
	}
	
	return response.Output, nil
}

// GetSessionID returns the current session ID
func (c *Conversation) GetSessionID() string {
	return c.sessionID
}

// Reset clears the conversation history by starting a new session
func (c *Conversation) Reset() {
	c.sessionID = ""
}

// QuickResponse provides the simplest possible interface - just model and input
func QuickResponse(model llm.Model, input string) string {
	response, err := Chat(context.Background(), model, input)
	if err != nil {
		return "Error: " + err.Error()
	}
	return response
}

// MultiTurn helper for conducting multi-turn conversations without session management
type MultiTurn struct {
	model    llm.Model
	messages []llm.Message
}

// NewMultiTurn creates a new multi-turn conversation helper
func NewMultiTurn(model llm.Model) *MultiTurn {
	return &MultiTurn{
		model: model,
		messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant."},
		},
	}
}

// NewMultiTurnWithSystem creates a multi-turn conversation with custom system message
func NewMultiTurnWithSystem(model llm.Model, systemMessage string) *MultiTurn {
	return &MultiTurn{
		model: model,
		messages: []llm.Message{
			{Role: "system", Content: systemMessage},
		},
	}
}

// Ask adds a user message and gets the assistant's response
func (mt *MultiTurn) Ask(ctx context.Context, input string) (string, error) {
	// Add user message
	mt.messages = append(mt.messages, llm.Message{
		Role:    "user",
		Content: input,
	})
	
	// Get response from model
	response, err := mt.model.Complete(ctx, llm.Request{
		Messages: mt.messages,
	})
	if err != nil {
		return "", err
	}
	
	// Add assistant response to history
	mt.messages = append(mt.messages, llm.Message{
		Role:    "assistant",
		Content: response.Content,
	})
	
	return response.Content, nil
}

// GetHistory returns the conversation history
func (mt *MultiTurn) GetHistory() []llm.Message {
	return mt.messages
}

// Clear resets the conversation but keeps the system message
func (mt *MultiTurn) Clear() {
	if len(mt.messages) > 0 && mt.messages[0].Role == "system" {
		mt.messages = mt.messages[:1] // Keep only system message
	} else {
		mt.messages = []llm.Message{}
	}
}