package openai

import (
	"context"
	"fmt"

	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/tool"
	openai "github.com/sashabaranov/go-openai"
)

// Client implements llm.Model using OpenAI API
type Client struct {
	client *openai.Client
	model  string
}

// New creates a new OpenAI client
func New(config llm.Config) *Client {
	openaiConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		openaiConfig.BaseURL = config.BaseURL
	}

	return &Client{
		client: openai.NewClientWithConfig(openaiConfig),
		model:  config.Model,
	}
}

// Complete performs a synchronous completion
func (c *Client) Complete(ctx context.Context, request llm.Request) (*llm.Response, error) {
	// Convert our request to OpenAI format
	openaiReq := c.toOpenAIRequest(request)

	// Make the API call
	resp, err := c.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("openai completion failed: %w", err)
	}

	// Convert response back to our format
	return c.fromOpenAIResponse(resp), nil
}

// toOpenAIRequest converts our request format to OpenAI format
func (c *Client) toOpenAIRequest(req llm.Request) openai.ChatCompletionRequest {
	openaiReq := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: make([]openai.ChatCompletionMessage, len(req.Messages)),
	}

	// Convert messages
	for i, msg := range req.Messages {
		openaiReq.Messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Handle tool response messages
		if msg.ToolCallID != "" {
			openaiReq.Messages[i].ToolCallID = msg.ToolCallID
		}
	}

	// Set optional parameters
	if req.Temperature != nil {
		openaiReq.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		openaiReq.MaxTokens = *req.MaxTokens
	}

	// Convert tools
	if len(req.Tools) > 0 {
		openaiReq.Tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			openaiReq.Tools[i] = c.toOpenAITool(tool)
		}
	}

	return openaiReq
}

// toOpenAITool converts our tool definition to OpenAI format
func (c *Client) toOpenAITool(def tool.Definition) openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        def.Function.Name,
			Description: def.Function.Description,
			Parameters:  c.toOpenAIParameters(def.Function.Parameters),
		},
	}
}

// toOpenAIParameters converts our parameters to OpenAI format
func (c *Client) toOpenAIParameters(params tool.Parameters) any {
	// OpenAI expects the parameters as a JSON schema object
	schema := map[string]any{
		"type":       params.Type,
		"properties": make(map[string]any),
	}

	// Convert properties
	properties := schema["properties"].(map[string]any)
	for name, prop := range params.Properties {
		properties[name] = map[string]any{
			"type":        prop.Type,
			"description": prop.Description,
		}
	}

	// Add required fields if any
	if len(params.Required) > 0 {
		schema["required"] = params.Required
	}

	return schema
}

// fromOpenAIResponse converts OpenAI response to our format
func (c *Client) fromOpenAIResponse(resp openai.ChatCompletionResponse) *llm.Response {
	if len(resp.Choices) == 0 {
		return &llm.Response{
			Content:      "",
			FinishReason: "error",
		}
	}

	choice := resp.Choices[0]
	response := &llm.Response{
		Content:      choice.Message.Content,
		FinishReason: string(choice.FinishReason),
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Convert tool calls if any
	if len(choice.Message.ToolCalls) > 0 {
		response.ToolCalls = make([]tool.Call, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			response.ToolCalls[i] = tool.Call{
				ID: tc.ID,
				Function: tool.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return response
}

// TODO: Future implementation
// - Stream method for streaming completions
// - Better error handling with retries
// - Response validation
// - Logging and metrics