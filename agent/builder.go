package agent

import (
	"context"
	"time"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/prompt"
	"github.com/davidleitw/go-agent/session"
	"github.com/davidleitw/go-agent/session/memory"
	"github.com/davidleitw/go-agent/tool"
)

// Builder provides a fluent interface for constructing agents
type Builder struct {
	config EngineConfig
}

// NewBuilder creates a new agent builder with sensible defaults
func NewBuilder() *Builder {
	return &Builder{
		config: EngineConfig{
			MaxIterations: 5,
			Temperature:   nil, // Use model default
			MaxTokens:     nil, // Use model default
		},
	}
}

// WithLLM sets the language model for the agent
func (b *Builder) WithLLM(model llm.Model) *Builder {
	b.config.Model = model
	return b
}

// WithSessionStore sets the session storage backend
func (b *Builder) WithSessionStore(store session.SessionStore) *Builder {
	b.config.SessionStore = store
	return b
}

// WithMemorySessionStore sets up in-memory session storage (for development/testing)
func (b *Builder) WithMemorySessionStore() *Builder {
	store := memory.NewStore()
	b.config.SessionStore = store
	return b
}

// WithTools sets up tool registry and registers provided tools
func (b *Builder) WithTools(tools ...tool.Tool) *Builder {
	if b.config.ToolRegistry == nil {
		b.config.ToolRegistry = tool.NewRegistry()
	}

	for _, t := range tools {
		b.config.ToolRegistry.Register(t)
	}
	return b
}

// WithToolRegistry sets the tool registry directly
func (b *Builder) WithToolRegistry(registry *tool.Registry) *Builder {
	b.config.ToolRegistry = registry
	return b
}

// WithContextProviders adds context providers for gathering information
func (b *Builder) WithContextProviders(providers ...agentcontext.Provider) *Builder {
	b.config.ContextProviders = append(b.config.ContextProviders, providers...)
	return b
}

// WithPromptTemplate sets a custom prompt template for organizing contexts
// Accepts: string template, prompt.Template, or prompt.Builder
func (b *Builder) WithPromptTemplate(template interface{}) *Builder {
	switch t := template.(type) {
	case string:
		b.config.PromptTemplate = prompt.Parse(t)
	case prompt.Template:
		b.config.PromptTemplate = t
	case prompt.Builder:
		b.config.PromptTemplate = t.Build()
	default:
		// Silently ignore invalid types - will use default template
	}
	return b
}

// WithSessionHistory adds a history provider that includes session conversation history
// Deprecated: Use WithHistoryLimit instead for better performance and control
func (b *Builder) WithSessionHistory(limit int) *Builder {
	historyProvider := agentcontext.NewHistoryProvider(limit)
	b.config.ContextProviders = append(b.config.ContextProviders, historyProvider)
	return b
}

// WithHistoryLimit sets the number of history entries to include (0 = disabled)
func (b *Builder) WithHistoryLimit(limit int) *Builder {
	b.config.HistoryLimit = limit
	return b
}

// WithHistoryInterceptor sets a custom history processor for advanced features
func (b *Builder) WithHistoryInterceptor(interceptor HistoryInterceptor) *Builder {
	b.config.HistoryInterceptor = interceptor
	return b
}

// WithMaxIterations sets the maximum number of thinking iterations
func (b *Builder) WithMaxIterations(max int) *Builder {
	b.config.MaxIterations = max
	return b
}

// WithTemperature sets the LLM temperature for response generation
func (b *Builder) WithTemperature(temp float32) *Builder {
	b.config.Temperature = &temp
	return b
}

// WithMaxTokens sets the maximum tokens for LLM responses
func (b *Builder) WithMaxTokens(tokens int) *Builder {
	b.config.MaxTokens = &tokens
	return b
}

// WithSessionTTL sets the session time-to-live duration
func (b *Builder) WithSessionTTL(ttl time.Duration) *Builder {
	b.config.SessionTTL = ttl
	return b
}

// WithEngine is not needed in the new design since engine is built from config

// Build constructs the final agent instance
func (b *Builder) Build() (Agent, error) {
	// Set default prompt template if none provided
	if b.config.PromptTemplate == nil {
		b.config.PromptTemplate = prompt.New().DefaultFlow().Build()
	}

	// Build the configured engine
	engine, err := NewEngine(b.config)
	if err != nil {
		return nil, err
	}

	return &BuiltAgent{
		engine: engine,
	}, nil
}

// BuiltAgent implements the Agent interface using the configured engine
type BuiltAgent struct {
	engine Engine
}

// Execute runs the agent using the configured engine
func (a *BuiltAgent) Execute(ctx context.Context, request Request) (*Response, error) {
	return a.engine.Execute(ctx, request)
}

// Quick builder functions for common patterns

// NewSimpleAgent creates a basic agent with just an LLM
func NewSimpleAgent(model llm.Model) (Agent, error) {
	return NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		Build()
}

// NewAgentWithTools creates an agent with LLM and tools
func NewAgentWithTools(model llm.Model, tools ...tool.Tool) (Agent, error) {
	return NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		WithTools(tools...).
		Build()
}

// NewConversationalAgent creates an agent that maintains conversation history
func NewConversationalAgent(model llm.Model, historyLimit int) (Agent, error) {
	return NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		WithHistoryLimit(historyLimit).
		Build()
}

// NewFullAgent creates a fully-featured agent with all capabilities
func NewFullAgent(model llm.Model, tools []tool.Tool, historyLimit int) (Agent, error) {
	builder := NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		WithHistoryLimit(historyLimit)

	if len(tools) > 0 {
		builder = builder.WithTools(tools...)
	}

	return builder.Build()
}
