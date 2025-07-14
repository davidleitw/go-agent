package mocks

import (
	"context"
	"fmt"
	"testing"

	"github.com/davidleitw/go-agent/pkg/agent"
)

func TestMockChatModel_Basic(t *testing.T) {
	mock := NewMockChatModel()
	
	// Test default response
	response, err := mock.GenerateChatCompletion(
		context.Background(),
		[]agent.Message{{Role: "user", Content: "Hello"}},
		"test-model",
		nil,
		nil,
	)
	
	if err != nil {
		t.Errorf("GenerateChatCompletion() error = %v, want nil", err)
	}
	
	if response == nil {
		t.Fatal("GenerateChatCompletion() response is nil")
	}
	
	if response.Role != agent.RoleAssistant {
		t.Errorf("Response role = %v, want %v", response.Role, agent.RoleAssistant)
	}
	
	if response.Content != "Mock response" {
		t.Errorf("Response content = %v, want %v", response.Content, "Mock response")
	}
	
	// Verify call was recorded
	if mock.GetCallCount() != 1 {
		t.Errorf("Call count = %v, want 1", mock.GetCallCount())
	}
	
	lastCall := mock.GetLastCall()
	if lastCall == nil {
		t.Fatal("Last call should not be nil")
	}
	
	if lastCall.Model != "test-model" {
		t.Errorf("Last call model = %v, want %v", lastCall.Model, "test-model")
	}
}

func TestMockChatModel_CustomResponses(t *testing.T) {
	mock := NewMockChatModel()
	
	// Set custom response
	customResponse := agent.Message{
		Role:    agent.RoleAssistant,
		Content: "Custom response",
	}
	mock.SetResponse(customResponse)
	
	response, err := mock.GenerateChatCompletion(
		context.Background(),
		[]agent.Message{{Role: "user", Content: "Hello"}},
		"test-model",
		nil,
		nil,
	)
	
	if err != nil {
		t.Errorf("GenerateChatCompletion() error = %v, want nil", err)
	}
	
	if response.Content != "Custom response" {
		t.Errorf("Response content = %v, want %v", response.Content, "Custom response")
	}
}

func TestMockChatModel_MultipleResponses(t *testing.T) {
	mock := NewMockChatModel()
	
	// Set multiple responses
	responses := []agent.Message{
		{Role: agent.RoleAssistant, Content: "Response 1"},
		{Role: agent.RoleAssistant, Content: "Response 2"},
		{Role: agent.RoleAssistant, Content: "Response 3"},
	}
	mock.SetResponses(responses)
	
	// Test sequential responses
	for i, expected := range responses {
		response, err := mock.GenerateChatCompletion(
			context.Background(),
			[]agent.Message{{Role: "user", Content: fmt.Sprintf("Message %d", i+1)}},
			"test-model",
			nil,
			nil,
		)
		
		if err != nil {
			t.Errorf("GenerateChatCompletion() call %d error = %v, want nil", i+1, err)
		}
		
		if response.Content != expected.Content {
			t.Errorf("Response %d content = %v, want %v", i+1, response.Content, expected.Content)
		}
	}
	
	// Test that last response is repeated
	response, err := mock.GenerateChatCompletion(
		context.Background(),
		[]agent.Message{{Role: "user", Content: "Extra message"}},
		"test-model",
		nil,
		nil,
	)
	
	if err != nil {
		t.Errorf("GenerateChatCompletion() extra call error = %v, want nil", err)
	}
	
	if response.Content != "Response 3" {
		t.Errorf("Extra response content = %v, want %v", response.Content, "Response 3")
	}
}

func TestMockChatModel_Errors(t *testing.T) {
	mock := NewMockChatModel()
	
	// Test error configuration
	mock.SetError("Test error")
	
	response, err := mock.GenerateChatCompletion(
		context.Background(),
		[]agent.Message{{Role: "user", Content: "Hello"}},
		"test-model",
		nil,
		nil,
	)
	
	if err == nil {
		t.Error("GenerateChatCompletion() error = nil, want error")
	}
	
	if err.Error() != "Test error" {
		t.Errorf("Error message = %v, want %v", err.Error(), "Test error")
	}
	
	if response != nil {
		t.Errorf("Response = %v, want nil", response)
	}
	
	// Test error after N calls
	mock.Reset()
	mock.SetErrorAfter(2)
	mock.SetResponse(agent.Message{Role: agent.RoleAssistant, Content: "OK"})
	
	// First call should succeed
	_, err = mock.GenerateChatCompletion(context.Background(), nil, "test", nil, nil)
	if err != nil {
		t.Errorf("First call error = %v, want nil", err)
	}
	
	// Second call should fail
	_, err = mock.GenerateChatCompletion(context.Background(), nil, "test", nil, nil)
	if err == nil {
		t.Error("Second call should have failed")
	}
}

func TestMockChatModel_CustomHandler(t *testing.T) {
	mock := NewMockChatModel()
	
	// Set custom handler
	mock.SetCustomHandler(func(messages []agent.Message, model string, settings *agent.ModelSettings, tools []agent.Tool) (*agent.Message, error) {
		if len(messages) == 0 {
			return nil, fmt.Errorf("no messages provided")
		}
		
		lastMessage := messages[len(messages)-1]
		return &agent.Message{
			Role:    agent.RoleAssistant,
			Content: fmt.Sprintf("Echo: %s", lastMessage.Content),
		}, nil
	})
	
	response, err := mock.GenerateChatCompletion(
		context.Background(),
		[]agent.Message{{Role: "user", Content: "Hello World"}},
		"test-model",
		nil,
		nil,
	)
	
	if err != nil {
		t.Errorf("GenerateChatCompletion() error = %v, want nil", err)
	}
	
	if response.Content != "Echo: Hello World" {
		t.Errorf("Response content = %v, want %v", response.Content, "Echo: Hello World")
	}
	
	// Test custom handler error
	_, err = mock.GenerateChatCompletion(context.Background(), []agent.Message{}, "test", nil, nil)
	if err == nil {
		t.Error("Custom handler should have returned error for empty messages")
	}
}

func TestMockChatModel_Verification(t *testing.T) {
	mock := NewMockChatModel()
	
	// Make some calls
	mock.GenerateChatCompletion(context.Background(), 
		[]agent.Message{{Role: "user", Content: "test message"}}, 
		"gpt-4", nil, nil)
	
	mock.GenerateChatCompletion(context.Background(), 
		[]agent.Message{{Role: "user", Content: "another message"}}, 
		"gpt-3.5", nil, nil)
	
	// Test verification methods
	if !mock.VerifyCalledWith("gpt-4", "test message") {
		t.Error("VerifyCalledWith should return true for gpt-4 with test message")
	}
	
	if mock.VerifyCalledWith("gpt-4", "wrong message") {
		t.Error("VerifyCalledWith should return false for wrong message")
	}
	
	if !mock.VerifyCallOrder([]string{"gpt-4", "gpt-3.5"}) {
		t.Error("VerifyCallOrder should return true for correct order")
	}
	
	if mock.VerifyCallOrder([]string{"gpt-3.5", "gpt-4"}) {
		t.Error("VerifyCallOrder should return false for wrong order")
	}
}

func TestMockSessionStore_Basic(t *testing.T) {
	store := NewMockSessionStore()
	ctx := context.Background()
	
	// Create a test session
	session := agent.NewSession("test-session")
	session.AddMessage(agent.RoleUser, "Hello")
	
	// Test Save
	err := store.Save(ctx, session)
	if err != nil {
		t.Errorf("Save() error = %v, want nil", err)
	}
	
	// Verify call was recorded
	if store.GetCallCount() != 1 {
		t.Errorf("Call count = %v, want 1", store.GetCallCount())
	}
	
	if store.GetCallCountByMethod("Save") != 1 {
		t.Errorf("Save call count = %v, want 1", store.GetCallCountByMethod("Save"))
	}
	
	// Test Load
	loadedSession, err := store.Load(ctx, "test-session")
	if err != nil {
		t.Errorf("Load() error = %v, want nil", err)
	}
	
	if loadedSession == nil {
		t.Fatal("Loaded session should not be nil")
	}
	
	if loadedSession.ID() != "test-session" {
		t.Errorf("Loaded session ID = %v, want %v", loadedSession.ID(), "test-session")
	}
	
	// Test Exists
	exists, err := store.Exists(ctx, "test-session")
	if err != nil {
		t.Errorf("Exists() error = %v, want nil", err)
	}
	
	if !exists {
		t.Error("Exists() = false, want true")
	}
	
	// Test non-existent session
	_, err = store.Load(ctx, "non-existent")
	if err != agent.ErrSessionNotFound {
		t.Errorf("Load() non-existent error = %v, want %v", err, agent.ErrSessionNotFound)
	}
	
	exists, err = store.Exists(ctx, "non-existent")
	if err != nil {
		t.Errorf("Exists() non-existent error = %v, want nil", err)
	}
	
	if exists {
		t.Error("Exists() non-existent = true, want false")
	}
}

func TestMockSessionStore_Errors(t *testing.T) {
	store := NewMockSessionStore()
	ctx := context.Background()
	
	session := agent.NewSession("test")
	
	// Test Save error
	testError := fmt.Errorf("save failed")
	store.SetSaveError(testError)
	
	err := store.Save(ctx, session)
	if err != testError {
		t.Errorf("Save() error = %v, want %v", err, testError)
	}
	
	// Test Load error
	store.ClearErrors()
	loadError := fmt.Errorf("load failed")
	store.SetLoadError(loadError)
	
	_, err = store.Load(ctx, "test")
	if err != loadError {
		t.Errorf("Load() error = %v, want %v", err, loadError)
	}
}

func TestMockTool_Basic(t *testing.T) {
	tool := NewMockTool("test-tool", "A test tool")
	
	// Test basic properties
	if tool.Name() != "test-tool" {
		t.Errorf("Name() = %v, want %v", tool.Name(), "test-tool")
	}
	
	if tool.Description() != "A test tool" {
		t.Errorf("Description() = %v, want %v", tool.Description(), "A test tool")
	}
	
	schema := tool.Schema()
	if schema == nil {
		t.Error("Schema() should not be nil")
	}
	
	// Test execution
	args := map[string]any{"input": "test"}
	result, err := tool.Execute(context.Background(), args)
	
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	
	if result != "mock result" {
		t.Errorf("Execute() result = %v, want %v", result, "mock result")
	}
	
	// Test call tracking
	if tool.GetCallCount() != 1 {
		t.Errorf("Call count = %v, want 1", tool.GetCallCount())
	}
	
	if !tool.WasCalled() {
		t.Error("WasCalled() = false, want true")
	}
	
	if !tool.WasCalledTimes(1) {
		t.Error("WasCalledTimes(1) = false, want true")
	}
}

func TestMockTool_CustomResult(t *testing.T) {
	tool := NewMockTool("test-tool", "A test tool")
	
	// Set custom result
	customResult := map[string]string{"key": "value"}
	tool.SetResult(customResult)
	
	result, err := tool.Execute(context.Background(), map[string]any{"input": "test"})
	
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	
	resultMap, ok := result.(map[string]string)
	if !ok {
		t.Errorf("Result type = %T, want map[string]string", result)
	}
	
	if resultMap["key"] != "value" {
		t.Errorf("Result[\"key\"] = %v, want %v", resultMap["key"], "value")
	}
}

func TestMockTool_Errors(t *testing.T) {
	tool := NewMockTool("test-tool", "A test tool")
	
	// Test error configuration
	tool.SetError("Tool failed")
	
	result, err := tool.Execute(context.Background(), map[string]any{"input": "test"})
	
	if err == nil {
		t.Error("Execute() error = nil, want error")
	}
	
	if err.Error() != "Tool failed" {
		t.Errorf("Error message = %v, want %v", err.Error(), "Tool failed")
	}
	
	if result != nil {
		t.Errorf("Result = %v, want nil", result)
	}
}

func TestMockTool_Verification(t *testing.T) {
	tool := NewMockTool("test-tool", "A test tool")
	
	// Make some calls
	tool.Execute(context.Background(), map[string]any{"param1": "value1", "param2": 42})
	tool.Execute(context.Background(), map[string]any{"param1": "value2"})
	
	// Test verification methods
	if !tool.VerifyCalledWith(map[string]any{"param1": "value1", "param2": 42}) {
		t.Error("VerifyCalledWith should return true for first call args")
	}
	
	if !tool.VerifyCalledWithArg("param1", "value2") {
		t.Error("VerifyCalledWithArg should return true for param1=value2")
	}
	
	if tool.VerifyCalledWithArg("param1", "nonexistent") {
		t.Error("VerifyCalledWithArg should return false for nonexistent value")
	}
	
	// Test getting args from specific call
	args, err := tool.GetArgsFromCall(0)
	if err != nil {
		t.Errorf("GetArgsFromCall(0) error = %v, want nil", err)
	}
	
	if args["param1"] != "value1" {
		t.Errorf("GetArgsFromCall(0) param1 = %v, want %v", args["param1"], "value1")
	}
	
	if args["param2"] != 42 {
		t.Errorf("GetArgsFromCall(0) param2 = %v, want %v", args["param2"], 42)
	}
}

func TestCalculatorTool(t *testing.T) {
	calc := NewCalculatorTool()
	
	// Test addition
	result, err := calc.Execute(context.Background(), map[string]any{
		"operation": "add",
		"a":         float64(5),
		"b":         float64(3),
	})
	
	if err != nil {
		t.Errorf("Calculator add error = %v, want nil", err)
	}
	
	if result != float64(8) {
		t.Errorf("Calculator add result = %v, want %v", result, float64(8))
	}
	
	// Test division by zero
	_, err = calc.Execute(context.Background(), map[string]any{
		"operation": "divide",
		"a":         float64(5),
		"b":         float64(0),
	})
	
	if err == nil {
		t.Error("Calculator divide by zero should return error")
	}
}

func TestWeatherTool(t *testing.T) {
	weather := NewWeatherTool()
	
	// Test weather query
	result, err := weather.Execute(context.Background(), map[string]any{
		"location": "New York",
		"units":    "fahrenheit",
	})
	
	if err != nil {
		t.Errorf("Weather tool error = %v, want nil", err)
	}
	
	weatherMap, ok := result.(map[string]any)
	if !ok {
		t.Errorf("Weather result type = %T, want map[string]any", result)
	}
	
	if weatherMap["location"] != "New York" {
		t.Errorf("Weather location = %v, want %v", weatherMap["location"], "New York")
	}
	
	if weatherMap["temperature"] != "72°F" {
		t.Errorf("Weather temperature = %v, want %v", weatherMap["temperature"], "72°F")
	}
}