package openai

import (
	"testing"

	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/tool"
	openai "github.com/sashabaranov/go-openai"
)

func TestClient_toOpenAIRequest(t *testing.T) {
	client := &Client{model: "gpt-4"}

	// Test basic request conversion
	temp := float32(0.7)
	maxTokens := 100
	req := llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	openaiReq := client.toOpenAIRequest(req)

	// Check model
	if openaiReq.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", openaiReq.Model)
	}

	// Check messages
	if len(openaiReq.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(openaiReq.Messages))
	}
	if openaiReq.Messages[0].Role != "system" {
		t.Errorf("Expected first message role system, got %s", openaiReq.Messages[0].Role)
	}
	if openaiReq.Messages[1].Content != "Hello" {
		t.Errorf("Expected second message content Hello, got %s", openaiReq.Messages[1].Content)
	}

	// Check parameters
	if openaiReq.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", openaiReq.Temperature)
	}
	if openaiReq.MaxTokens != 100 {
		t.Errorf("Expected max tokens 100, got %d", openaiReq.MaxTokens)
	}
}

func TestClient_toOpenAIRequest_WithTools(t *testing.T) {
	client := &Client{model: "gpt-4"}

	req := llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "What's the weather?"},
		},
		Tools: []tool.Definition{
			{
				Type: "function",
				Function: tool.Function{
					Name:        "get_weather",
					Description: "Get current weather",
					Parameters: tool.Parameters{
						Type: "object",
						Properties: map[string]tool.Property{
							"location": {
								Type:        "string",
								Description: "City name",
							},
						},
						Required: []string{"location"},
					},
				},
			},
		},
	}

	openaiReq := client.toOpenAIRequest(req)

	// Check tools conversion
	if len(openaiReq.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(openaiReq.Tools))
	}

	tool := openaiReq.Tools[0]
	if tool.Type != openai.ToolTypeFunction {
		t.Errorf("Expected tool type function, got %v", tool.Type)
	}
	if tool.Function.Name != "get_weather" {
		t.Errorf("Expected function name get_weather, got %s", tool.Function.Name)
	}
}

func TestClient_fromOpenAIResponse(t *testing.T) {
	client := &Client{}

	// Test basic response conversion
	openaiResp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				FinishReason: openai.FinishReasonStop,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	resp := client.fromOpenAIResponse(openaiResp)

	// Check content
	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("Expected content 'Hello! How can I help you?', got %s", resp.Content)
	}

	// Check finish reason
	if resp.FinishReason != "stop" {
		t.Errorf("Expected finish reason stop, got %s", resp.FinishReason)
	}

	// Check usage
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 5 {
		t.Errorf("Expected completion tokens 5, got %d", resp.Usage.CompletionTokens)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("Expected total tokens 15, got %d", resp.Usage.TotalTokens)
	}
}

func TestClient_fromOpenAIResponse_WithToolCalls(t *testing.T) {
	client := &Client{}

	openaiResp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_123",
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "Tokyo"}`,
							},
						},
					},
				},
				FinishReason: openai.FinishReasonToolCalls,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     20,
			CompletionTokens: 10,
			TotalTokens:      30,
		},
	}

	resp := client.fromOpenAIResponse(openaiResp)

	// Check tool calls
	if len(resp.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(resp.ToolCalls))
	}

	toolCall := resp.ToolCalls[0]
	if toolCall.ID != "call_123" {
		t.Errorf("Expected tool call ID call_123, got %s", toolCall.ID)
	}
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name get_weather, got %s", toolCall.Function.Name)
	}
	if toolCall.Function.Arguments != `{"location": "Tokyo"}` {
		t.Errorf("Expected arguments {\"location\": \"Tokyo\"}, got %s", toolCall.Function.Arguments)
	}

	// Check finish reason
	if resp.FinishReason != "tool_calls" {
		t.Errorf("Expected finish reason tool_calls, got %s", resp.FinishReason)
	}
}

func TestClient_fromOpenAIResponse_EmptyChoices(t *testing.T) {
	client := &Client{}

	// Test response with no choices
	openaiResp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{},
	}

	resp := client.fromOpenAIResponse(openaiResp)

	// Should return error response
	if resp.Content != "" {
		t.Errorf("Expected empty content, got %s", resp.Content)
	}
	if resp.FinishReason != "error" {
		t.Errorf("Expected finish reason error, got %s", resp.FinishReason)
	}
}

func TestNew(t *testing.T) {
	// Test creating client with basic config
	config := llm.Config{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	client := New(config)
	if client.model != "gpt-3.5-turbo" {
		t.Errorf("Expected model gpt-3.5-turbo, got %s", client.model)
	}
	if client.client == nil {
		t.Error("Expected non-nil OpenAI client")
	}

	// Test with custom base URL
	configWithURL := llm.Config{
		APIKey:  "test-key",
		Model:   "gpt-4",
		BaseURL: "https://custom.openai.com",
	}

	clientWithURL := New(configWithURL)
	if clientWithURL.model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", clientWithURL.model)
	}
}
