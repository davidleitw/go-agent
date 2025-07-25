# Agent Module

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

The Agent module provides the main interface for creating and using AI agents in the Go Agent Framework. It coordinates between LLM models, tools, session management, and context providers to create powerful, stateful AI interactions.

## Features

- **Simple Agent Interface**: Clean `Agent.Execute()` method for running agent tasks
- **Builder Pattern**: Flexible agent construction with fluent API
- **Session Management**: Stateful conversations with persistent memory and TTL support
- **Tool Integration**: Seamless tool calling and execution
- **Context Providers**: Gather information from various sources automatically
- **Convenience Functions**: Simple patterns for common use cases
- **Extensible Engine**: Pluggable execution engines for different behaviors
- **Performance Optimized**: Pre-cached session options and efficient component management

## Quick Start

### Simple Agent

```go
import (
    "github.com/davidleitw/go-agent/agent"
    "github.com/davidleitw/go-agent/llm/openai"
)

// Create a simple agent
model := openai.New(llm.Config{
    APIKey: "your-api-key",
    Model:  "gpt-4",
})

myAgent := agent.NewSimpleAgent(model)

// Use the agent
response, err := myAgent.Execute(ctx, agent.Request{
    Input: "What is the capital of France?",
})

fmt.Println(response.Output) // "The capital of France is Paris."
```

### Agent with Tools

```go
// Define a custom tool
type WeatherTool struct{}

func (w *WeatherTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "get_weather",
            Description: "Get current weather for a location",
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
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    location := params["location"].(string)
    return fmt.Sprintf("Weather in %s: 22°C, sunny", location), nil
}

// Create agent with tools
weatherTool := &WeatherTool{}
myAgent := agent.NewAgentWithTools(model, weatherTool)

response, _ := myAgent.Execute(ctx, agent.Request{
    Input: "What's the weather in Tokyo?",
})
```

## Builder Pattern

For advanced configurations, use the builder pattern:

```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithMemorySessionStore().
    WithTools(weatherTool, calculatorTool).
    WithSessionHistory(20).
    WithSessionTTL(6*time.Hour).         // Session expires after 6 hours
    WithMaxIterations(5).
    WithTemperature(0.7).
    Build()

if err != nil {
    log.Fatal(err)
}

response, _ := agent.Execute(ctx, agent.Request{
    Input:     "Help me plan a trip to Tokyo",
    SessionID: "user-123", // Optional: use existing session
})
```

## Convenience Functions

### One-shot Chat

```go
// Simple chat without session management
response, err := agent.Chat(ctx, model, "Hello, how are you?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### Conversational Interface

```go
// Multi-turn conversation with automatic session management
conv := agent.NewConversationWithModel(model)

response1, _ := conv.Say(ctx, "Hello!")
response2, _ := conv.Say(ctx, "What did I just say?")
fmt.Println(response2) // Agent remembers previous messages

// Reset conversation
conv.Reset()
```

### Multi-turn without Sessions

```go
// Simple multi-turn without session persistence
mt := agent.NewMultiTurn(model)

response1, _ := mt.Ask(ctx, "What is machine learning?")
response2, _ := mt.Ask(ctx, "Can you give me an example?")

// Get conversation history
history := mt.GetHistory()
```

## API Reference

### Agent Interface

```go
type Agent interface {
    Execute(ctx context.Context, request Request) (*Response, error)
}
```

### Request Structure

```go
type Request struct {
    Input     string            // User input or instruction
    SessionID string            // Optional session ID
}
```

### Response Structure

```go
type Response struct {
    Output    string            // Agent's response
    SessionID string            // Session ID used
    Session   session.Session   // Access to session state
    Metadata  map[string]any    // Additional response data
    Usage     Usage             // Resource usage information
}
```

### Usage Tracking

```go
type Usage struct {
    LLMTokens     TokenUsage    // Language model token usage
    ToolCalls     int           // Number of tool executions
    SessionWrites int           // Session state modifications
}

type TokenUsage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

## Builder Options

### Core Components

```go
builder := agent.NewBuilder()

// Required: Set the language model
builder.WithLLM(model)

// Optional: Set session storage
builder.WithMemorySessionStore()        // In-memory (default)
builder.WithSessionStore(customStore)   // Custom storage

// Optional: Add tools
builder.WithTools(tool1, tool2)
builder.WithToolRegistry(registry)

// Optional: Add context providers
builder.WithContextProviders(provider1, provider2)
builder.WithSessionHistory(20)          // Include conversation history
```

### Configuration Options

```go
// Session management
builder.WithSessionTTL(24*time.Hour)    // Session expiration (default: 24h)

// Execution limits
builder.WithMaxIterations(5)            // Max thinking loops

// LLM parameters
builder.WithTemperature(0.7)            // Response creativity
builder.WithMaxTokens(1000)             // Response length limit
```

## Context Providers

Context providers gather information for the agent:

```go
import (
    agentcontext "github.com/davidleitw/go-agent/context"
)

// System prompt provider
systemProvider := agentcontext.NewSystemPromptProvider("You are a helpful assistant")

// History provider (last 10 messages)
historyProvider := agentcontext.NewHistoryProvider(10)

// Custom user context provider for dynamic information
type UserContextProvider struct {
    userPreferences map[string]any
}

func (p *UserContextProvider) Provide(s session.Session) []agentcontext.Context {
    return []agentcontext.Context{{
        Type:    "user_info",
        Content: fmt.Sprintf("User preferences: %v", p.userPreferences),
    }}
}

userProvider := &UserContextProvider{
    userPreferences: map[string]any{
        "language": "Chinese",
        "location": "Tokyo",
    },
}

agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithContextProviders(systemProvider, historyProvider, userProvider).
    Build()
```

## Session Management

Sessions provide stateful conversations with automatic expiration and metadata tracking.

### Automatic Session Creation

```go
// Agent creates new session automatically with default 24h TTL
response, _ := agent.Execute(ctx, agent.Request{
    Input: "Hello!",
    // SessionID left empty - new session created automatically
})

sessionID := response.SessionID // Use for future interactions

// Session contains metadata and state:
// - Metadata: created_by="agent", agent_version="v1.0"
// - State: initial_input_length, session_start_time, etc.
```

### Explicit Session Management

```go
// Use specific session
response, _ := agent.Execute(ctx, agent.Request{
    Input:     "Continue our conversation", 
    SessionID: "existing-session-id",
})

// Access session state and metadata
session := response.Session
userPrefs := session.Get("user_preferences")      // User-defined state
startTime := session.Get("session_start_time")    // System-added state

// Sessions automatically expire based on TTL
// Expired sessions return ErrSessionNotFound
```

### Custom Session TTL

```go
// Short-lived sessions for temporary interactions
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithSessionTTL(30*time.Minute).    // 30 minutes
    Build()

// Long-lived sessions for persistent conversations
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithSessionTTL(7*24*time.Hour).    // 7 days
    Build()

// No expiration (use with caution)
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithSessionTTL(0).                 // Never expires
    Build()
```

### Long-running Conversations

```go
// Example: Travel planning conversation
conv := agent.NewConversationWithModel(model)

// First interaction establishes context
response1, _ := conv.Say(ctx, "I'm planning a 3-day trip to Tokyo with a budget of 50,000 yen")

// Subsequent interactions automatically maintain context
response2, _ := conv.Say(ctx, "What museums should I visit on day 2?")
// Agent remembers: Tokyo, 3 days, 50,000 yen budget

response3, _ := conv.Say(ctx, "I prefer modern art over traditional")
// Agent now knows: Tokyo, 3 days, budget, modern art preference
```

### Dynamic Context with ContextProviders

```go
// Location-aware agent
type LocationContextProvider struct {
    getCurrentLocation func() string
}

func (p *LocationContextProvider) Provide(s session.Session) []agentcontext.Context {
    location := p.getCurrentLocation()
    return []agentcontext.Context{{
        Type:    "location",
        Content: fmt.Sprintf("User's current location: %s", location),
    }}
}

// Task-aware agent
type TaskContextProvider struct{}

func (p *TaskContextProvider) Provide(s session.Session) []agentcontext.Context {
    currentTask, _ := s.Get("current_task")
    if currentTask != nil {
        return []agentcontext.Context{{
            Type:    "task",
            Content: fmt.Sprintf("Current task: %s", currentTask),
        }}
    }
    return nil
}

agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithContextProviders(
        &LocationContextProvider{getCurrentLocation: getGPSLocation},
        &TaskContextProvider{},
    ).
    Build()

// Usage: Context is automatically injected
response, _ := agent.Execute(ctx, agent.Request{
    Input: "Find nearby restaurants",
    // No need to manually specify location - ContextProvider handles it
})
```

## Error Handling

```go
response, err := agent.Execute(ctx, request)
if err != nil {
    switch {
    case errors.Is(err, agent.ErrInvalidInput):
        log.Println("Invalid input provided")
    case errors.Is(err, agent.ErrSessionNotFound):
        log.Println("Session not found")
    case errors.Is(err, agent.ErrMaxIterationsExceeded):
        log.Println("Agent thinking loop exceeded limit")
    case errors.Is(err, agent.ErrToolExecutionFailed):
        log.Println("Tool execution failed")
    case errors.Is(err, agent.ErrLLMCallFailed):
        log.Println("LLM request failed")
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}

// Check resource usage
if response.Usage.LLMTokens.TotalTokens > 10000 {
    log.Println("High token usage detected")
}
```

## Custom Engine

For advanced use cases, implement a custom execution engine:

```go
type CustomEngine struct {
    // Custom fields
}

func (e *CustomEngine) Execute(ctx context.Context, request agent.Request, config agent.ExecutionConfig) (*agent.Response, error) {
    // Custom execution logic
    // - Handle session management
    // - Gather contexts
    // - Call LLM with tools
    // - Execute tool calls
    // - Return structured response
}

// Use custom engine
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithEngine(&CustomEngine{}).
    Build()
```

## Best Practices

### 1. Resource Management

```go
// Set reasonable limits
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithMaxIterations(3).           // Prevent infinite loops
    WithMaxTokens(500).             // Control costs
    WithTemperature(0.3).           // More deterministic
    Build()

// Monitor usage
response, _ := agent.Execute(ctx, request)
log.Printf("Used %d tokens", response.Usage.LLMTokens.TotalTokens)
```

### 2. Error Handling

```go
// Always handle context cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := agent.Execute(ctx, request)
if err != nil {
    // Handle specific error types
    return
}
```

### 3. Session Management

```go
// Reuse sessions for conversational experiences
sessionID := ""

for {
    input := getUserInput()
    
    response, err := agent.Execute(ctx, agent.Request{
        Input:     input,
        SessionID: sessionID,
    })
    
    if err != nil {
        break
    }
    
    sessionID = response.SessionID // Remember for next interaction
    fmt.Println(response.Output)
}
```

### 4. Tool Design

```go
// Design tools to be idempotent and handle errors gracefully
type SafeTool struct{}

func (t *SafeTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    // Validate inputs
    input, ok := params["input"].(string)
    if !ok {
        return nil, fmt.Errorf("input must be a string")
    }
    
    // Respect context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Perform operation safely
    result, err := safeOperation(input)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    return result, nil
}
```

## Development Status

**Current Status**: Core interfaces, builder pattern, and session management implemented.

**Completed Features**:
- ✅ Agent interface and builder pattern
- ✅ Session management with TTL and metadata
- ✅ Component configuration and caching
- ✅ Context provider framework
- ✅ Tool registry integration
- ✅ Convenience functions and multi-turn conversations
- ✅ Comprehensive test coverage

**Next Steps** (Engine execution logic):
1. Implement context gathering from providers
2. Implement LLM message construction pipeline
3. Implement main execution loop with tool calling
4. Add tool execution orchestration
5. Add usage tracking and error handling

**Testing**: All interface and session management tests pass. Engine execution tests will be added as core logic is implemented.

## Architecture

```
Agent Module
├── Core Interfaces
│   ├── Agent.Execute() - Main entry point
│   └── Engine.Execute() - Core execution logic
├── Builder Pattern
│   ├── Component configuration
│   ├── Session TTL settings
│   └── Performance optimizations
├── ConfiguredEngine (✅ Implemented)
│   ├── Session Management (✅ Complete)
│   ├── Context Gathering (🚧 Framework ready)
│   ├── LLM Orchestration (🚧 Placeholder)
│   └── Tool Execution (🚧 Placeholder)
├── Convenience Functions (✅ Complete)
│   ├── Chat (one-shot)
│   ├── Conversation (stateful)
│   └── MultiTurn (simple)
└── Session Features (✅ Complete)
    ├── Automatic TTL management
    ├── Metadata tracking
    └── State persistence
```

**Legend**: ✅ Complete, 🚧 In Progress, ❌ Not Started

## License

MIT License - See LICENSE file in the project root directory.