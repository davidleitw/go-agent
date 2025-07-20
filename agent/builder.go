package agent

import (
	"context"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
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

// WithSessionHistory adds a history provider that includes session conversation history
func (b *Builder) WithSessionHistory(limit int) *Builder {
	historyProvider := agentcontext.NewHistoryProvider(limit)
	b.config.ContextProviders = append(b.config.ContextProviders, historyProvider)
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

// WithEngine is not needed in the new design since engine is built from config

// Build constructs the final agent instance
func (b *Builder) Build() (Agent, error) {
	// Build the configured engine
	engine, err := NewConfiguredEngine(b.config)
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
func NewSimpleAgent(model llm.Model) Agent {
	agent, _ := NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		Build()
	return agent
}

// NewAgentWithTools creates an agent with LLM and tools
func NewAgentWithTools(model llm.Model, tools ...tool.Tool) Agent {
	agent, _ := NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		WithTools(tools...).
		Build()
	return agent
}

// NewConversationalAgent creates an agent that maintains conversation history
func NewConversationalAgent(model llm.Model, historyLimit int) Agent {
	agent, _ := NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		WithSessionHistory(historyLimit).
		Build()
	return agent
}

// NewFullAgent creates a fully-featured agent with all capabilities
func NewFullAgent(model llm.Model, tools []tool.Tool, historyLimit int) Agent {
	builder := NewBuilder().
		WithLLM(model).
		WithMemorySessionStore().
		WithSessionHistory(historyLimit)
	
	if len(tools) > 0 {
		builder = builder.WithTools(tools...)
	}
	
	agent, _ := builder.Build()
	return agent
}