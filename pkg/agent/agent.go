package agent

import (
	"context"
	"fmt"
	"time"
)

// Agent represents an AI agent that can have conversations and execute tools.
type Agent interface {
	// Configuration methods
	Name() string
	Description() string
	Instructions() string
	Model() string
	ModelSettings() *ModelSettings
	Tools() []Tool
	OutputType() OutputType

	// Execution methods
	Chat(ctx context.Context, sessionID string, userInput string, options ...ChatOption) (*Message, interface{}, error)
	ChatWithSession(ctx context.Context, session Session, userInput string, options ...ChatOption) (*Message, interface{}, error)

	// Session management
	GetSession(ctx context.Context, sessionID string) (Session, error)
	CreateSession(sessionID string) Session
	SaveSession(ctx context.Context, session Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	ListSessions(ctx context.Context, filter SessionFilter) ([]string, error)
}

// Tool defines an external capability that an agent can use.
type Tool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// ChatOption allows customizing individual chat calls
type ChatOption func(*chatOptions)

type chatOptions struct {
	maxTurns                  *int
	modelSettings             *ModelSettings
	additionalTools           []Tool
	systemMessage             string
	clearHistory              bool
	returnIntermediateResults bool
}

// WithChatMaxTurns limits the number of tool execution rounds
func WithChatMaxTurns(maxTurns int) ChatOption {
	return func(o *chatOptions) {
		o.maxTurns = &maxTurns
	}
}

// WithChatModelSettings overrides the agent's default model settings
func WithChatModelSettings(settings *ModelSettings) ChatOption {
	return func(o *chatOptions) {
		o.modelSettings = settings
	}
}

// WithAdditionalTools adds extra tools for this conversation only
func WithAdditionalTools(tools ...Tool) ChatOption {
	return func(o *chatOptions) {
		o.additionalTools = append(o.additionalTools, tools...)
	}
}

// WithSystemMessage overrides the agent's instructions
func WithSystemMessage(message string) ChatOption {
	return func(o *chatOptions) {
		o.systemMessage = message
	}
}

// WithClearHistory starts a fresh conversation
func WithClearHistory() ChatOption {
	return func(o *chatOptions) {
		o.clearHistory = true
	}
}

// WithIntermediateResults returns intermediate tool call results
func WithIntermediateResults() ChatOption {
	return func(o *chatOptions) {
		o.returnIntermediateResults = true
	}
}

// AgentOption configures an Agent during creation
type AgentOption func(*AgentConfig) error

// AgentConfig holds the configuration for creating an Agent
type AgentConfig struct {
	Name          string
	Description   string
	Instructions  string
	Model         string
	ModelSettings *ModelSettings
	Tools         []Tool
	OutputType    OutputType
	FlowRules     []FlowRule

	// Runtime dependencies
	ChatModel    ChatModel
	SessionStore SessionStore
	MaxTurns     int
	ToolTimeout  time.Duration
	DebugLogging bool
}

// New creates a new Agent with the given options
func New(options ...AgentOption) (Agent, error) {
	config := &AgentConfig{
		Model:       "gpt-4",          // Default model
		MaxTurns:    10,               // Default max turns
		ToolTimeout: 30 * time.Second, // Default tool timeout
	}

	for _, option := range options {
		if err := option(config); err != nil {
			return nil, fmt.Errorf("agent configuration error: %w", err)
		}
	}

	if err := validateAgentConfig(config); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	return newSimpleAgent(config), nil
}

// WithName sets the agent's unique identifier
func WithName(name string) AgentOption {
	return func(config *AgentConfig) error {
		if name == "" {
			return fmt.Errorf("agent name cannot be empty")
		}
		config.Name = name
		return nil
	}
}

// WithDescription sets the agent's description
func WithDescription(description string) AgentOption {
	return func(config *AgentConfig) error {
		config.Description = description
		return nil
	}
}

// WithInstructions sets the agent's system instructions/prompt
func WithInstructions(instructions string) AgentOption {
	return func(config *AgentConfig) error {
		config.Instructions = instructions
		return nil
	}
}

// WithModel sets the LLM model to use
func WithModel(model string) AgentOption {
	return func(config *AgentConfig) error {
		if model == "" {
			return fmt.Errorf("model cannot be empty")
		}
		config.Model = model
		return nil
	}
}

// WithModelSettings sets the model configuration parameters
func WithModelSettings(settings *ModelSettings) AgentOption {
	return func(config *AgentConfig) error {
		if settings != nil {
			if err := settings.Validate(); err != nil {
				return fmt.Errorf("invalid model settings: %w", err)
			}
		}
		config.ModelSettings = settings
		return nil
	}
}

// WithTools adds tools to the agent
func WithTools(tools ...Tool) AgentOption {
	return func(config *AgentConfig) error {
		// Validate tools
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			if tool == nil {
				return fmt.Errorf("tool cannot be nil")
			}
			name := tool.Name()
			if name == "" {
				return fmt.Errorf("tool name cannot be empty")
			}
			if toolNames[name] {
				return fmt.Errorf("duplicate tool name: %s", name)
			}
			toolNames[name] = true
		}

		config.Tools = append(config.Tools, tools...)
		return nil
	}
}

// WithOutputType sets the expected structured output format
func WithOutputType(outputType OutputType) AgentOption {
	return func(config *AgentConfig) error {
		config.OutputType = outputType
		return nil
	}
}

// WithStructuredOutput creates an OutputType from a struct example
func WithStructuredOutput(example interface{}) AgentOption {
	return func(config *AgentConfig) error {
		// Create simple OutputType from struct example
		outputType := &simpleOutputType{example: example}
		config.OutputType = outputType
		return nil
	}
}

// WithFlowRules adds flow rules for dynamic behavior
func WithFlowRules(rules ...FlowRule) AgentOption {
	return func(config *AgentConfig) error {
		config.FlowRules = append(config.FlowRules, rules...)
		return nil
	}
}

// WithChatModel sets a custom ChatModel implementation
func WithChatModel(chatModel ChatModel) AgentOption {
	return func(config *AgentConfig) error {
		if chatModel == nil {
			return fmt.Errorf("chat model cannot be nil")
		}
		config.ChatModel = chatModel
		return nil
	}
}

// WithSessionStore sets the session storage backend
func WithSessionStore(store SessionStore) AgentOption {
	return func(config *AgentConfig) error {
		if store == nil {
			return fmt.Errorf("session store cannot be nil")
		}
		config.SessionStore = store
		return nil
	}
}

// WithMaxTurns sets the maximum number of tool execution rounds
func WithMaxTurns(maxTurns int) AgentOption {
	return func(config *AgentConfig) error {
		if maxTurns <= 0 {
			return fmt.Errorf("max turns must be positive")
		}
		config.MaxTurns = maxTurns
		return nil
	}
}

// WithToolTimeout sets the timeout for tool execution
func WithToolTimeout(timeout time.Duration) AgentOption {
	return func(config *AgentConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("tool timeout must be positive")
		}
		config.ToolTimeout = timeout
		return nil
	}
}

// WithDebugLogging enables debug logging for the agent
func WithDebugLogging() AgentOption {
	return func(config *AgentConfig) error {
		config.DebugLogging = true
		return nil
	}
}

// validateAgentConfig validates the agent configuration
func validateAgentConfig(config *AgentConfig) error {
	if config.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if config.ChatModel == nil {
		return fmt.Errorf("chat model is required (use WithChatModel)")
	}

	if config.SessionStore == nil {
		return fmt.Errorf("session store is required (use WithSessionStore)")
	}

	return nil
}

// NewInMemorySessionStore creates a new in-memory SessionStore implementation
func NewInMemorySessionStore() SessionStore {
	return newInMemoryStore()
}
