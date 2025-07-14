package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ToolCall records a call to the tool
type ToolCall struct {
	Args    map[string]any
	Time    time.Time
	Result  any
	Error   error
	Context context.Context
}

// MockTool provides a comprehensive mock implementation of Tool
type MockTool struct {
	mu sync.Mutex
	
	// Tool configuration
	name        string
	description string
	schema      map[string]any
	
	// Response configuration  
	result      any
	shouldError bool
	errorMsg    string
	errorAfter  int
	
	// Call tracking
	calls     []ToolCall
	callCount int
	
	// Behavior configuration
	executeDelay  time.Duration
	customHandler func(ctx context.Context, args map[string]any) (any, error)
}

// NewMockTool creates a new mock tool
func NewMockTool(name, description string) *MockTool {
	return &MockTool{
		name:        name,
		description: description,
		schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{
					"type":        "string",
					"description": "Input parameter",
				},
			},
			"required": []string{"input"},
		},
		calls:  make([]ToolCall, 0),
		result: "mock result",
	}
}

// NewMockToolWithSchema creates a mock tool with custom schema
func NewMockToolWithSchema(name, description string, schema map[string]any) *MockTool {
	return &MockTool{
		name:        name,
		description: description,
		schema:      schema,
		calls:       make([]ToolCall, 0),
		result:      "mock result",
	}
}

// Name implements Tool.Name
func (m *MockTool) Name() string {
	return m.name
}

// Description implements Tool.Description
func (m *MockTool) Description() string {
	return m.description
}

// Schema implements Tool.Schema
func (m *MockTool) Schema() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Return a copy to prevent external modification
	schemaCopy := make(map[string]any)
	for k, v := range m.schema {
		schemaCopy[k] = v
	}
	return schemaCopy
}

// Execute implements Tool.Execute
func (m *MockTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	call := ToolCall{
		Args:    make(map[string]any),
		Time:    time.Now(),
		Context: ctx,
	}
	
	// Copy args to prevent external modification
	for k, v := range args {
		call.Args[k] = v
	}
	
	m.callCount++
	
	// Simulate execution delay if configured
	if m.executeDelay > 0 {
		time.Sleep(m.executeDelay)
	}
	
	// Check if we should return an error
	if m.shouldError || (m.errorAfter > 0 && m.callCount >= m.errorAfter) {
		errorMsg := m.errorMsg
		if errorMsg == "" {
			errorMsg = "mock tool error"
		}
		err := fmt.Errorf(errorMsg)
		call.Error = err
		m.calls = append(m.calls, call)
		return nil, err
	}
	
	// Use custom handler if provided
	if m.customHandler != nil {
		result, err := m.customHandler(ctx, args)
		call.Result = result
		call.Error = err
		m.calls = append(m.calls, call)
		return result, err
	}
	
	// Return configured result
	call.Result = m.result
	m.calls = append(m.calls, call)
	return m.result, nil
}

// Configuration methods

// SetResult configures the result to return
func (m *MockTool) SetResult(result any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.result = result
}

// SetError configures the tool to return an error
func (m *MockTool) SetError(err string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = true
	m.errorMsg = err
}

// SetErrorAfter configures the tool to return an error after N calls
func (m *MockTool) SetErrorAfter(callCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorAfter = callCount
}

// ClearError removes error configuration
func (m *MockTool) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = false
	m.errorMsg = ""
	m.errorAfter = 0
}

// SetExecuteDelay adds artificial delay to Execute operations
func (m *MockTool) SetExecuteDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeDelay = delay
}

// SetCustomHandler allows custom execution logic
func (m *MockTool) SetCustomHandler(handler func(ctx context.Context, args map[string]any) (any, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customHandler = handler
}

// SetSchema updates the tool schema
func (m *MockTool) SetSchema(schema map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schema = schema
}

// Inspection methods

// GetCallHistory returns all recorded calls
func (m *MockTool) GetCallHistory() []ToolCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	calls := make([]ToolCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the number of calls made
func (m *MockTool) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// GetLastCall returns the most recent call, or nil if no calls have been made
func (m *MockTool) GetLastCall() *ToolCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if len(m.calls) == 0 {
		return nil
	}
	
	lastCall := m.calls[len(m.calls)-1]
	return &lastCall
}

// Reset clears all recorded calls and resets state
func (m *MockTool) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = make([]ToolCall, 0)
	m.callCount = 0
	m.shouldError = false
	m.errorMsg = ""
	m.errorAfter = 0
	m.executeDelay = 0
	m.customHandler = nil
}

// Verification helpers

// VerifyCalledWith checks if the tool was called with specific arguments
func (m *MockTool) VerifyCalledWith(args map[string]any) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, call := range m.calls {
		if len(call.Args) != len(args) {
			continue
		}
		
		match := true
		for k, expectedV := range args {
			actualV, exists := call.Args[k]
			if !exists || actualV != expectedV {
				match = false
				break
			}
		}
		
		if match {
			return true
		}
	}
	return false
}

// VerifyCalledWithArg checks if the tool was called with a specific argument
func (m *MockTool) VerifyCalledWithArg(key string, value any) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, call := range m.calls {
		if actualValue, exists := call.Args[key]; exists && actualValue == value {
			return true
		}
	}
	return false
}

// WasCalled returns true if the tool was called at least once
func (m *MockTool) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount > 0
}

// WasCalledTimes returns true if the tool was called exactly N times
func (m *MockTool) WasCalledTimes(expectedCount int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount == expectedCount
}

// GetArgsFromCall returns the arguments from a specific call (0-indexed)
func (m *MockTool) GetArgsFromCall(callIndex int) (map[string]any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if callIndex < 0 || callIndex >= len(m.calls) {
		return nil, fmt.Errorf("call index %d out of range (0-%d)", callIndex, len(m.calls)-1)
	}
	
	// Return a copy to prevent external modification
	args := make(map[string]any)
	for k, v := range m.calls[callIndex].Args {
		args[k] = v
	}
	return args, nil
}

// GetResultFromCall returns the result from a specific call (0-indexed)
func (m *MockTool) GetResultFromCall(callIndex int) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if callIndex < 0 || callIndex >= len(m.calls) {
		return nil, fmt.Errorf("call index %d out of range (0-%d)", callIndex, len(m.calls)-1)
	}
	
	return m.calls[callIndex].Result, m.calls[callIndex].Error
}

// Helper functions for creating common mock tools

// NewCalculatorTool creates a mock calculator tool for testing
func NewCalculatorTool() *MockTool {
	tool := NewMockToolWithSchema("calculator", "Performs basic arithmetic operations", map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"description": "The operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": map[string]any{
				"type":        "number",
				"description": "First number",
			},
			"b": map[string]any{
				"type":        "number",
				"description": "Second number",
			},
		},
		"required": []string{"operation", "a", "b"},
	})
	
	// Set custom handler for realistic calculator behavior
	tool.SetCustomHandler(func(ctx context.Context, args map[string]any) (any, error) {
		operation, ok := args["operation"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid operation type")
		}
		
		a, ok := args["a"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid first number")
		}
		
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid second number")
		}
		
		switch operation {
		case "add":
			return a + b, nil
		case "subtract":
			return a - b, nil
		case "multiply":
			return a * b, nil
		case "divide":
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return a / b, nil
		default:
			return nil, fmt.Errorf("unknown operation: %s", operation)
		}
	})
	
	return tool
}

// NewWeatherTool creates a mock weather tool for testing
func NewWeatherTool() *MockTool {
	tool := NewMockToolWithSchema("get_weather", "Gets weather information for a location", map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "The location to get weather for",
			},
			"units": map[string]any{
				"type":        "string",
				"description": "Temperature units",
				"enum":        []string{"celsius", "fahrenheit"},
			},
		},
		"required": []string{"location"},
	})
	
	// Set custom handler for realistic weather responses
	tool.SetCustomHandler(func(ctx context.Context, args map[string]any) (any, error) {
		location, ok := args["location"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid location")
		}
		
		units := "celsius"
		if u, exists := args["units"]; exists {
			if uStr, ok := u.(string); ok {
				units = uStr
			}
		}
		
		temp := "22°C"
		if units == "fahrenheit" {
			temp = "72°F"
		}
		
		return map[string]any{
			"location":    location,
			"temperature": temp,
			"condition":   "Sunny",
			"humidity":    "60%",
			"wind":        "5 mph",
		}, nil
	})
	
	return tool
}