package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/sashabaranov/go-openai"
)

// ChatModel implements the agent.ChatModel interface for OpenAI
type ChatModel struct {
	client       *openai.Client
	baseURL      string
	organization string
	timeout      time.Duration
	retryCount   int
}

// Config holds the configuration for creating an OpenAI ChatModel
type Config struct {
	BaseURL      string
	Organization string
	Timeout      time.Duration
	RetryCount   int
}

// NewChatModel creates a new OpenAI ChatModel implementation
func NewChatModel(apiKey string, config *Config) (agent.ChatModel, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key cannot be empty")
	}

	if config == nil {
		config = &Config{
			Timeout:    30 * time.Second,
			RetryCount: 3,
		}
	}

	clientConfig := openai.DefaultConfig(apiKey)

	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	if config.Organization != "" {
		clientConfig.OrgID = config.Organization
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &ChatModel{
		client:       client,
		baseURL:      config.BaseURL,
		organization: config.Organization,
		timeout:      config.Timeout,
		retryCount:   config.RetryCount,
	}, nil
}

// GenerateChatCompletion implements agent.ChatModel interface
func (o *ChatModel) GenerateChatCompletion(ctx context.Context, messages []agent.Message, model string, settings *agent.ModelSettings, tools []agent.Tool) (*agent.Message, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Handle tool calls if present
		if len(msg.ToolCalls) > 0 {
			var toolCalls []openai.ToolCall
			for _, tc := range msg.ToolCalls {
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
			openaiMessages[i].ToolCalls = toolCalls
		}

		// Handle tool call ID
		if msg.ToolCallID != "" {
			openaiMessages[i].ToolCallID = msg.ToolCallID
		}
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
	}

	// Apply model settings
	if settings != nil {
		if settings.Temperature != nil {
			req.Temperature = float32(*settings.Temperature)
		}
		if settings.MaxTokens != nil {
			req.MaxTokens = *settings.MaxTokens
		}
		if settings.TopP != nil {
			req.TopP = float32(*settings.TopP)
		}
		if settings.FrequencyPenalty != nil {
			req.FrequencyPenalty = float32(*settings.FrequencyPenalty)
		}
		if settings.PresencePenalty != nil {
			req.PresencePenalty = float32(*settings.PresencePenalty)
		}
		if len(settings.Stop) > 0 {
			req.Stop = settings.Stop
		}
	}

	// Convert tools to OpenAI format
	if len(tools) > 0 {
		var openaiTools []openai.Tool
		for _, tool := range tools {
			openaiTool := openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.Schema(),
				},
			}
			openaiTools = append(openaiTools, openaiTool)
		}
		req.Tools = openaiTools
		req.ToolChoice = "auto"
	}

	// Make the request
	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := resp.Choices[0]

	// Convert response back to our format
	response := &agent.Message{
		Role:      choice.Message.Role,
		Content:   choice.Message.Content,
		Timestamp: time.Now(),
	}

	// Handle tool calls in response
	if len(choice.Message.ToolCalls) > 0 {
		var toolCalls []agent.ToolCall
		for _, tc := range choice.Message.ToolCalls {
			toolCall := agent.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: agent.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
			toolCalls = append(toolCalls, toolCall)
		}
		response.ToolCalls = toolCalls
	}

	return response, nil
}

// GetSupportedModels implements agent.ChatModel interface
func (o *ChatModel) GetSupportedModels() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
	}
}

// ValidateModel implements agent.ChatModel interface
func (o *ChatModel) ValidateModel(model string) error {
	supportedModels := o.GetSupportedModels()
	for _, supported := range supportedModels {
		if supported == model {
			return nil
		}
	}
	return fmt.Errorf("model %s is not supported by OpenAI provider", model)
}

// GetModelInfo implements agent.ChatModel interface
func (o *ChatModel) GetModelInfo(model string) (*agent.ModelInfo, error) {
	if err := o.ValidateModel(model); err != nil {
		return nil, err
	}

	// Return model info based on the model name
	switch model {
	case "gpt-4":
		return &agent.ModelInfo{
			ID:              "gpt-4",
			Name:            "GPT-4",
			Description:     "Most capable GPT-4 model",
			ContextWindow:   8192,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        "openai",
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.03,
				OutputTokenPrice: 0.06,
				Currency:         "USD",
			},
		}, nil
	case "gpt-4-turbo":
		return &agent.ModelInfo{
			ID:              "gpt-4-turbo",
			Name:            "GPT-4 Turbo",
			Description:     "GPT-4 Turbo with 128k context",
			ContextWindow:   128000,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        "openai",
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.01,
				OutputTokenPrice: 0.03,
				Currency:         "USD",
			},
		}, nil
	case "gpt-4o":
		return &agent.ModelInfo{
			ID:              "gpt-4o",
			Name:            "GPT-4o",
			Description:     "High-intelligence flagship model for complex, multi-step tasks",
			ContextWindow:   128000,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        "openai",
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.005,
				OutputTokenPrice: 0.015,
				Currency:         "USD",
			},
		}, nil
	case "gpt-4o-mini":
		return &agent.ModelInfo{
			ID:              "gpt-4o-mini",
			Name:            "GPT-4o mini",
			Description:     "Affordable and intelligent small model for fast, lightweight tasks",
			ContextWindow:   128000,
			MaxOutputTokens: 16384,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        "openai",
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.00015,
				OutputTokenPrice: 0.0006,
				Currency:         "USD",
			},
		}, nil
	case "gpt-3.5-turbo":
		return &agent.ModelInfo{
			ID:              "gpt-3.5-turbo",
			Name:            "GPT-3.5 Turbo",
			Description:     "Fast, inexpensive model for simple tasks",
			ContextWindow:   16385,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        "openai",
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.0005,
				OutputTokenPrice: 0.0015,
				Currency:         "USD",
			},
		}, nil
	default:
		return &agent.ModelInfo{
			ID:              model,
			Name:            model,
			Description:     "OpenAI model",
			ContextWindow:   8192,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        "openai",
		}, nil
	}
}
