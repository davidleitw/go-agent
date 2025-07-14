package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
)

// ChatModelCall records a call to the chat model
type ChatModelCall struct {
	Messages    []agent.Message
	Model       string
	Settings    *agent.ModelSettings
	Tools       []agent.Tool
	Time        time.Time
	ContextData map[string]interface{} // Additional context for testing
}

// MockChatModel provides a comprehensive mock implementation of ChatModel
type MockChatModel struct {
	mu sync.Mutex
	
	// Response configuration
	responses   []agent.Message
	responseIdx int
	
	// Error simulation
	shouldError bool
	errorMsg    string
	errorAfter  int // Error after N calls
	callCount   int
	
	// Call history tracking
	calls []ChatModelCall
	
	// Advanced behavior
	responseDelay time.Duration
	customHandler func(messages []agent.Message, model string, settings *agent.ModelSettings, tools []agent.Tool) (*agent.Message, error)
}

// NewMockChatModel creates a new mock chat model with sensible defaults
func NewMockChatModel() *MockChatModel {
	return &MockChatModel{
		responses:   make([]agent.Message, 0),
		calls:       make([]ChatModelCall, 0),
		responseIdx: 0,
	}
}

// GenerateChatCompletion implements the ChatModel interface
func (m *MockChatModel) GenerateChatCompletion(
	ctx context.Context,
	messages []agent.Message,
	model string,
	settings *agent.ModelSettings,
	tools []agent.Tool,
) (*agent.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Record the call
	call := ChatModelCall{
		Messages: make([]agent.Message, len(messages)),
		Model:    model,
		Settings: settings,
		Tools:    make([]agent.Tool, len(tools)),
		Time:     time.Now(),
	}
	
	// Deep copy messages and tools to prevent mutation
	copy(call.Messages, messages)
	copy(call.Tools, tools)
	
	m.calls = append(m.calls, call)
	m.callCount++
	
	// Simulate response delay if configured
	if m.responseDelay > 0 {
		time.Sleep(m.responseDelay)
	}
	
	// Check if we should return an error
	if m.shouldError || (m.errorAfter > 0 && m.callCount >= m.errorAfter) {
		errorMsg := m.errorMsg
		if errorMsg == "" {
			errorMsg = "mock error"
		}
		return nil, fmt.Errorf(errorMsg)
	}
	
	// Use custom handler if provided
	if m.customHandler != nil {
		return m.customHandler(messages, model, settings, tools)
	}
	
	// Return predefined response
	if len(m.responses) == 0 {
		// Default response if none configured
		return &agent.Message{
			Role:      agent.RoleAssistant,
			Content:   "Mock response",
			Timestamp: time.Now(),
		}, nil
	}
	
	// Cycle through responses or return the last one
	if m.responseIdx < len(m.responses) {
		response := m.responses[m.responseIdx]
		m.responseIdx++
		// Ensure timestamp is set
		if response.Timestamp.IsZero() {
			response.Timestamp = time.Now()
		}
		return &response, nil
	}
	
	// Return last response for subsequent calls
	lastResponse := m.responses[len(m.responses)-1]
	if lastResponse.Timestamp.IsZero() {
		lastResponse.Timestamp = time.Now()
	}
	return &lastResponse, nil
}

// Configuration methods

// SetResponse sets a single response that will be returned for all calls
func (m *MockChatModel) SetResponse(response agent.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = []agent.Message{response}
	m.responseIdx = 0
}

// SetResponses sets multiple responses that will be returned in sequence
func (m *MockChatModel) SetResponses(responses []agent.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = make([]agent.Message, len(responses))
	copy(m.responses, responses)
	m.responseIdx = 0
}

// AddResponse appends a response to the response queue
func (m *MockChatModel) AddResponse(response agent.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = append(m.responses, response)
}

// SetError configures the mock to return an error
func (m *MockChatModel) SetError(err string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = true
	m.errorMsg = err
}

// SetErrorAfter configures the mock to return an error after N calls
func (m *MockChatModel) SetErrorAfter(callCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorAfter = callCount
}

// ClearError removes error configuration
func (m *MockChatModel) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = false
	m.errorMsg = ""
	m.errorAfter = 0
}

// SetResponseDelay adds artificial delay to responses (useful for testing timeouts)
func (m *MockChatModel) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// SetCustomHandler allows custom response logic
func (m *MockChatModel) SetCustomHandler(handler func(messages []agent.Message, model string, settings *agent.ModelSettings, tools []agent.Tool) (*agent.Message, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customHandler = handler
}

// Inspection methods

// GetCallHistory returns all recorded calls
func (m *MockChatModel) GetCallHistory() []ChatModelCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Return a copy to prevent external mutation
	calls := make([]ChatModelCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the number of calls made
func (m *MockChatModel) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// GetLastCall returns the most recent call, or nil if no calls have been made
func (m *MockChatModel) GetLastCall() *ChatModelCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if len(m.calls) == 0 {
		return nil
	}
	
	lastCall := m.calls[len(m.calls)-1]
	return &lastCall
}

// Reset clears all recorded calls and resets state
func (m *MockChatModel) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = make([]ChatModelCall, 0)
	m.callCount = 0
	m.responseIdx = 0
	m.shouldError = false
	m.errorMsg = ""
	m.errorAfter = 0
	m.responseDelay = 0
	m.customHandler = nil
}

// Verification helpers

// VerifyCalledWith checks if the mock was called with specific parameters
func (m *MockChatModel) VerifyCalledWith(model string, messageContains string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, call := range m.calls {
		if call.Model == model {
			for _, msg := range call.Messages {
				if fmt.Sprintf("%s", msg.Content) == messageContains {
					return true
				}
			}
		}
	}
	return false
}

// VerifyCallOrder checks if calls were made in a specific order
func (m *MockChatModel) VerifyCallOrder(expectedModels []string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if len(m.calls) < len(expectedModels) {
		return false
	}
	
	for i, expectedModel := range expectedModels {
		if m.calls[i].Model != expectedModel {
			return false
		}
	}
	return true
}

// VerifyToolsProvided checks if specific tools were provided in any call
func (m *MockChatModel) VerifyToolsProvided(toolNames []string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, call := range m.calls {
		providedTools := make(map[string]bool)
		for _, tool := range call.Tools {
			providedTools[tool.Name()] = true
		}
		
		// Check if all expected tools were provided
		allFound := true
		for _, toolName := range toolNames {
			if !providedTools[toolName] {
				allFound = false
				break
			}
		}
		
		if allFound {
			return true
		}
	}
	return false
}

// GetSupportedModels returns mock model list
func (m *MockChatModel) GetSupportedModels() []string {
	return []string{"mock-model", "test-model", "gpt-4", "gpt-3.5-turbo"}
}

// ValidateModel validates model name (always returns nil for mock)
func (m *MockChatModel) ValidateModel(model string) error {
	return nil
}

// GetModelInfo returns mock model information
func (m *MockChatModel) GetModelInfo(model string) (*agent.ModelInfo, error) {
	return &agent.ModelInfo{
		ID:          model,
		Name:        "Mock Model",
		Description: "A mock model for testing",
		Provider:    "mock",
	}, nil
}