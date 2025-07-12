package agent

import (
	"context"
	"fmt"
	"strings"
)

// ChatModel abstracts interaction with any chat-oriented large language model provider.
// Implementations should handle provider-specific details like authentication,
// rate limiting, error handling, and data format conversion.
type ChatModel interface {
	// GenerateChatCompletion sends a chat request to the LLM and returns its response
	GenerateChatCompletion(
		ctx context.Context,
		messages []Message,
		model string,
		settings *ModelSettings,
		tools []Tool,
	) (*Message, error)

	// GetSupportedModels returns a list of model identifiers supported by this provider
	GetSupportedModels() []string

	// ValidateModel checks if a model identifier is supported
	ValidateModel(model string) error

	// GetModelInfo returns information about a specific model
	GetModelInfo(model string) (*ModelInfo, error)
}

// ModelInfo contains metadata about a specific model
type ModelInfo struct {
	// ID is the unique identifier for the model
	ID string `json:"id"`

	// Name is the human-readable name of the model
	Name string `json:"name"`

	// Description provides details about the model's capabilities
	Description string `json:"description"`

	// ContextWindow is the maximum number of tokens the model can process
	ContextWindow int `json:"context_window"`

	// MaxOutputTokens is the maximum number of tokens the model can generate
	MaxOutputTokens int `json:"max_output_tokens"`

	// SupportsTools indicates if the model supports function calling
	SupportsTools bool `json:"supports_tools"`

	// SupportsStreaming indicates if the model supports streaming responses
	SupportsStreaming bool `json:"supports_streaming"`

	// SupportsJSON indicates if the model supports JSON mode
	SupportsJSON bool `json:"supports_json"`

	// Pricing contains cost information for the model
	Pricing *ModelPricing `json:"pricing,omitempty"`

	// Provider is the name of the model provider
	Provider string `json:"provider"`

	// Version is the version of the model
	Version string `json:"version,omitempty"`
}

// ModelPricing contains cost information for a model
type ModelPricing struct {
	// InputTokenPrice is the cost per 1000 input tokens
	InputTokenPrice float64 `json:"input_token_price"`

	// OutputTokenPrice is the cost per 1000 output tokens
	OutputTokenPrice float64 `json:"output_token_price"`

	// Currency is the currency code (e.g., "USD")
	Currency string `json:"currency"`
}

// The ChatModel creation functions have been moved to agent.go
// to provide a cleaner functional options API.
// 
// Use these patterns instead:
//
// For OpenAI:
//   agent, err := agent.New(
//       agent.WithOpenAI("your-api-key"),
//       // ... other options
//   )
//
// For custom ChatModel:
//   agent, err := agent.New(
//       agent.WithChatModel(customChatModel),
//       // ... other options
//   )

// Common provider constants
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGoogle    = "google"
	ProviderMicrosoft = "microsoft"
	ProviderMeta      = "meta"
	ProviderCohere    = "cohere"
	ProviderMistral   = "mistral"
)

// Common model identifiers
const (
	// OpenAI models
	ModelGPT4       = "gpt-4"
	ModelGPT4Turbo  = "gpt-4-turbo"
	ModelGPT4o      = "gpt-4o"
	ModelGPT4oMini  = "gpt-4o-mini"
	ModelGPT35Turbo = "gpt-3.5-turbo"

	// Anthropic models
	ModelClaude3Opus    = "claude-3-opus-20240229"
	ModelClaude3Sonnet  = "claude-3-sonnet-20240229"
	ModelClaude3Haiku   = "claude-3-haiku-20240307"
	ModelClaude35Sonnet = "claude-3-5-sonnet-20241022"
)

// GetProviderFromModel attempts to determine the provider from a model name
func GetProviderFromModel(model string) string {
	switch {
	case strings.Contains(model, "gpt"):
		return ProviderOpenAI
	case strings.Contains(model, "claude"):
		return ProviderAnthropic
	case strings.Contains(model, "gemini"):
		return ProviderGoogle
	case strings.Contains(model, "command"):
		return ProviderCohere
	case strings.Contains(model, "mistral"):
		return ProviderMistral
	default:
		return ""
	}
}

// ValidateProviderModel checks if a model is valid for a specific provider
func ValidateProviderModel(provider, model string) error {
	expectedProvider := GetProviderFromModel(model)
	if expectedProvider != "" && expectedProvider != provider {
		return fmt.Errorf("model %s is not compatible with provider %s", model, provider)
	}
	return nil
}

// MockChatModel provides a simple mock implementation for testing
type MockChatModel struct {
	SupportedModels []string
	Response        *Message
	Error           error
	CallCount       int
	LastRequest     *MockChatRequest
}

// MockChatRequest captures the parameters of the last request
type MockChatRequest struct {
	Messages []Message
	Model    string
	Settings *ModelSettings
	Tools    []Tool
}

// NewMockChatModel creates a new mock chat model
func NewMockChatModel() *MockChatModel {
	response := NewAssistantMessage("Mock response")
	return &MockChatModel{
		SupportedModels: []string{"mock-model", "test-model"},
		Response:        &response,
	}
}

// GenerateChatCompletion implements ChatModel interface
func (m *MockChatModel) GenerateChatCompletion(
	ctx context.Context,
	messages []Message,
	model string,
	settings *ModelSettings,
	tools []Tool,
) (*Message, error) {
	m.CallCount++
	m.LastRequest = &MockChatRequest{
		Messages: messages,
		Model:    model,
		Settings: settings,
		Tools:    tools,
	}

	if m.Error != nil {
		return nil, m.Error
	}

	return m.Response, nil
}

// GetSupportedModels implements ChatModel interface
func (m *MockChatModel) GetSupportedModels() []string {
	return m.SupportedModels
}

// ValidateModel implements ChatModel interface
func (m *MockChatModel) ValidateModel(model string) error {
	for _, supported := range m.SupportedModels {
		if supported == model {
			return nil
		}
	}
	return fmt.Errorf("model %s is not supported", model)
}

// GetModelInfo implements ChatModel interface
func (m *MockChatModel) GetModelInfo(model string) (*ModelInfo, error) {
	if err := m.ValidateModel(model); err != nil {
		return nil, err
	}

	return &ModelInfo{
		ID:              model,
		Name:            "Mock Model",
		Description:     "A mock model for testing",
		ContextWindow:   8192,
		MaxOutputTokens: 4096,
		SupportsTools:   true,
		SupportsJSON:    true,
		Provider:        "mock",
	}, nil
}

// WithResponse sets the mock response
func (m *MockChatModel) WithResponse(response *Message) *MockChatModel {
	m.Response = response
	return m
}

// WithError sets the mock error
func (m *MockChatModel) WithError(err error) *MockChatModel {
	m.Error = err
	return m
}

// WithSupportedModels sets the supported models
func (m *MockChatModel) WithSupportedModels(models ...string) *MockChatModel {
	m.SupportedModels = models
	return m
}

// Reset clears the call count and last request
func (m *MockChatModel) Reset() {
	m.CallCount = 0
	m.LastRequest = nil
	m.Error = nil
}
