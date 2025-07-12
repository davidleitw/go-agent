package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/davidleitw/go-agent/pkg/agent"
)

// openAIChatModel implements the ChatModel interface for OpenAI
type openAIChatModel struct {
	client       *openai.Client
	baseURL      string
	organization string
	timeout      time.Duration
	retryCount   int
}

// OpenAIConfig holds the configuration for creating an OpenAI ChatModel
type OpenAIConfig struct {
	BaseURL      string
	Organization string
	Timeout      time.Duration
	RetryCount   int
}

// NewOpenAIChatModel creates a new OpenAI ChatModel implementation
func NewOpenAIChatModel(apiKey string, config *OpenAIConfig) (agent.ChatModel, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key cannot be empty")
	}
	
	if config == nil {
		config = &OpenAIConfig{
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
	
	return &openAIChatModel{
		client:       client,
		baseURL:      config.BaseURL,
		organization: config.Organization,
		timeout:      config.Timeout,
		retryCount:   config.RetryCount,
	}, nil
}

// GenerateChatCompletion sends a chat request to OpenAI and returns the response
func (o *openAIChatModel) GenerateChatCompletion(
	ctx context.Context,
	messages []agent.Message,
	model string,
	settings *agent.ModelSettings,
	tools []agent.Tool,
) (*agent.Message, error) {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()
	
	// Convert messages to OpenAI format
	openaiMessages := o.convertMessages(messages)
	
	// Create chat completion request
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
		req.Tools = o.convertTools(tools)
		req.ToolChoice = "auto"
	}
	
	// Make the API call with retries
	var resp openai.ChatCompletionResponse
	var err error
	
	for attempt := 0; attempt <= o.retryCount; attempt++ {
		resp, err = o.client.CreateChatCompletion(timeoutCtx, req)
		if err == nil {
			break
		}
		
		// If this is the last attempt, don't wait
		if attempt == o.retryCount {
			break
		}
		
		// Wait before retrying (exponential backoff)
		waitTime := time.Duration(attempt+1) * time.Second
		select {
		case <-timeoutCtx.Done():
			return nil, timeoutCtx.Err()
		case <-time.After(waitTime):
			// Continue to next attempt
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed after %d attempts: %w", o.retryCount+1, err)
	}
	
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}
	
	choice := resp.Choices[0]
	
	// Convert response to agent.Message
	message := &agent.Message{
		Role:      agent.RoleAssistant,
		Content:   choice.Message.Content,
		Timestamp: time.Now(),
	}
	
	// Convert tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		message.ToolCalls = make([]agent.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			message.ToolCalls[i] = agent.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: agent.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}
	
	return message, nil
}

// GetSupportedModels returns a list of OpenAI model identifiers
func (o *openAIChatModel) GetSupportedModels() []string {
	return []string{
		agent.ModelGPT4,
		agent.ModelGPT4Turbo,
		agent.ModelGPT4o,
		agent.ModelGPT4oMini,
		agent.ModelGPT35Turbo,
		"gpt-4-0125-preview",
		"gpt-4-1106-preview",
		"gpt-4-vision-preview",
		"gpt-3.5-turbo-16k",
		"gpt-3.5-turbo-1106",
	}
}

// ValidateModel checks if a model identifier is supported by OpenAI
func (o *openAIChatModel) ValidateModel(model string) error {
	supported := o.GetSupportedModels()
	for _, m := range supported {
		if m == model {
			return nil
		}
	}
	return fmt.Errorf("model %s is not supported by OpenAI provider", model)
}

// GetModelInfo returns information about a specific OpenAI model
func (o *openAIChatModel) GetModelInfo(model string) (*agent.ModelInfo, error) {
	if err := o.ValidateModel(model); err != nil {
		return nil, err
	}
	
	// Model info for common OpenAI models
	modelInfoMap := map[string]*agent.ModelInfo{
		agent.ModelGPT4: {
			ID:              agent.ModelGPT4,
			Name:            "GPT-4",
			Description:     "Most capable GPT-4 model, able to do complex tasks with greater accuracy",
			ContextWindow:   8192,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        agent.ProviderOpenAI,
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.03,  // $0.03 per 1K tokens
				OutputTokenPrice: 0.06,  // $0.06 per 1K tokens
				Currency:         "USD",
			},
		},
		agent.ModelGPT4Turbo: {
			ID:              agent.ModelGPT4Turbo,
			Name:            "GPT-4 Turbo",
			Description:     "Latest GPT-4 model with improved performance and lower cost",
			ContextWindow:   128000,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        agent.ProviderOpenAI,
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.01,  // $0.01 per 1K tokens
				OutputTokenPrice: 0.03,  // $0.03 per 1K tokens
				Currency:         "USD",
			},
		},
		agent.ModelGPT4o: {
			ID:              agent.ModelGPT4o,
			Name:            "GPT-4o",
			Description:     "Flagship model that's faster and cheaper than GPT-4 Turbo",
			ContextWindow:   128000,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        agent.ProviderOpenAI,
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.005, // $0.005 per 1K tokens
				OutputTokenPrice: 0.015, // $0.015 per 1K tokens
				Currency:         "USD",
			},
		},
		agent.ModelGPT4oMini: {
			ID:              agent.ModelGPT4oMini,
			Name:            "GPT-4o mini",
			Description:     "Affordable and intelligent small model for fast, lightweight tasks",
			ContextWindow:   128000,
			MaxOutputTokens: 16384,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        agent.ProviderOpenAI,
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.00015, // $0.00015 per 1K tokens
				OutputTokenPrice: 0.0006,  // $0.0006 per 1K tokens
				Currency:         "USD",
			},
		},
		agent.ModelGPT35Turbo: {
			ID:              agent.ModelGPT35Turbo,
			Name:            "GPT-3.5 Turbo",
			Description:     "Fast, inexpensive model for simple tasks",
			ContextWindow:   16385,
			MaxOutputTokens: 4096,
			SupportsTools:   true,
			SupportsJSON:    true,
			Provider:        agent.ProviderOpenAI,
			Pricing: &agent.ModelPricing{
				InputTokenPrice:  0.0015, // $0.0015 per 1K tokens
				OutputTokenPrice: 0.002,  // $0.002 per 1K tokens
				Currency:         "USD",
			},
		},
	}
	
	if info, exists := modelInfoMap[model]; exists {
		return info, nil
	}
	
	// Return generic info for unknown models
	return &agent.ModelInfo{
		ID:              model,
		Name:            model,
		Description:     "OpenAI model",
		ContextWindow:   8192,  // Conservative default
		MaxOutputTokens: 4096,  // Conservative default
		SupportsTools:   true,
		SupportsJSON:    true,
		Provider:        agent.ProviderOpenAI,
	}, nil
}

// Helper methods

func (o *openAIChatModel) convertMessages(messages []agent.Message) []openai.ChatCompletionMessage {
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	
	for i, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		
		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			openaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				openaiMsg.ToolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
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
		
		openaiMessages[i] = openaiMsg
	}
	
	return openaiMessages
}

func (o *openAIChatModel) convertTools(tools []agent.Tool) []openai.Tool {
	openaiTools := make([]openai.Tool, len(tools))
	
	for i, tool := range tools {
		openaiTools[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Schema(),
			},
		}
	}
	
	return openaiTools
}