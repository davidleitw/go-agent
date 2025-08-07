# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent Mascot" width="300"/>
  
  [![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)
</div>

> ⚠️ **Framework Under Active Development**: This framework is currently in the planning and development phase. The roadmap, interfaces, and implementation details are subject to significant changes. Once the interfaces stabilize, we plan to build long-running agents to verify reliability and performance for production use.

A clean yet feature-complete AI agent framework for Go. We designed this framework to be easy to get started with while maintaining high extensibility, allowing you to quickly integrate AI agent capabilities into your Go projects.

## Why go-agent?

While there are several excellent agent frameworks available, we wanted to create something with a focus on simplicity and Go-idiomatic design. Our design philosophy is "Context is Everything" + **Easy to Start, Easy to Scale**:

**Easy to Start:**
- Get going with just one `Execute()` method
- Clear module responsibilities - no need to understand the entire framework
- Rich examples and documentation that you can follow immediately

**Highly Extensible:**
- Modular design - use only what you need
- Clear interface definitions make custom implementations easy
- Open Provider pattern allows integration with any data source

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/davidleitw/go-agent/agent"
    "github.com/davidleitw/go-agent/llm"
    "github.com/davidleitw/go-agent/llm/openai"
)

func main() {
    // Create LLM model
    model := openai.New(llm.Config{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    
    // Create simple agent (now returns error)
    myAgent, err := agent.NewSimpleAgent(model)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start conversation
    response, err := myAgent.Execute(context.Background(), agent.Request{
        Input: "Plan a 3-day trip to Tokyo for me",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Output)
    fmt.Printf("Used %d tokens\n", response.Usage.LLMTokens.TotalTokens)
}
```

## Framework Architecture

We break down complex AI agent functionality into several independent but well-coordinated modules:

```
┌─────────────┐    ┌─────────────────────────────────────┐    ┌─────────────┐
│ User Input  │───▶│           Agent.Execute()            │───▶│   Response  │
└─────────────┘    └─────────────────┬───────────────────┘    └─────────────┘
                                     │
                        ┌────────────▼────────────┐
                        │  Step 1: Session Mgmt   │
                        │    (handleSession)      │
                        └────────────┬────────────┘
                                     │
                        ┌────────────▼────────────┐
                        │ Step 2: Context Gather  │
                        │   (gatherContexts)      │
                        └────────────┬────────────┘
                                     │
               ┌─────────────────────┼─────────────────────┐
               │                     │                     │
        ┌──────▼──────┐    ┌─────────▼──────┐    ┌─────────▼──────┐
        │System Prompt│    │    History     │    │    Custom      │
        │  Provider   │    │  Management    │    │  Providers     │
        └─────────────┘    └────────────────┘    └────────────────┘
                                     │
                        ┌────────────▼────────────┐
                        │ Step 3: Execute Loop    │
                        │  (executeIterations)    │
                        │                         │
                        │  ┌─────────────────┐    │
                        │  │ Build Messages  │    │
                        │  └─────────┬───────┘    │
                        │            │            │
                        │  ┌─────────▼───────┐    │
                        │  │  LLM Call       │◄───┼──── Tool Registry
                        │  └─────────┬───────┘    │
                        │            │            │
                        │  ┌─────────▼───────┐    │
                        │  │ Tool Execution  │    │
                        │  └─────────┬───────┘    │
                        │            │            │
                        │        Iterate until    │
                        │        completion       │
                        └─────────────────────────┘
                                     │
                              ┌──────▼──────┐
                              │   Session   │
                              │   Storage   │
                              │ (TTL mgmt)  │
                              └─────────────┘
```

### Context Provider System - Our Unique Approach

What makes go-agent special is our **unified Context management system**. Instead of simple string concatenation, we treat context as structured data that flows through the entire system.

**The Provider Pattern:**
Different providers contribute different types of context information, all unified into a consistent format that LLMs can understand:

```go
import (
    "context"
    "fmt"
    
    agentcontext "github.com/davidleitw/go-agent/context"
    "github.com/davidleitw/go-agent/session"
)

// System instructions  
systemProvider := agentcontext.NewSystemPromptProvider("You are a helpful assistant")

// History management is now built-in via WithHistoryLimit()
// No need for a separate HistoryProvider

// Custom provider that reads from session state
type TaskContextProvider struct{}

func (p *TaskContextProvider) Type() string {
    return "task_context"
}

func (p *TaskContextProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    // Read current task from session state
    if task, exists := s.Get("current_task"); exists {
        return []agentcontext.Context{{
            Type:    "task_context",
            Content: fmt.Sprintf("Current task: %s", task),
            Metadata: map[string]any{
                "source": "session_state",
                "key":    "current_task",
            },
        }}
    }
    return nil
}

// This is how it works in practice:
session.Set("current_task", "Planning Tokyo trip")
session.AddEntry(session.NewMessageEntry("user", "What's the weather like?"))
session.AddEntry(session.NewToolCallEntry("weather", map[string]any{"city": "Tokyo"}))
session.AddEntry(session.NewToolResultEntry("weather", "22°C, sunny", nil))

// When engine gathers contexts, it automatically converts session entries to contexts:
// - Message entries → role-specific contexts (user/assistant/system)
// - Tool call entries → "Tool: weather\nParameters: {city: Tokyo}"
// - Tool result entries → "Tool: weather\nSuccess: true\nResult: 22°C, sunny"  
// - TaskContextProvider reads session.Get("current_task") → "Current task: Planning Tokyo trip"

agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(10).  // Built-in history management
    WithContextProviders(systemProvider, &TaskContextProvider{}).
    Build()
if err != nil {
    log.Fatal(err)
}
```

**Key Benefits:**
- **Automatic History Management**: Session conversations are automatically converted to context
- **Rich Metadata**: Every context piece includes metadata for debugging and analytics
- **TTL Integration**: Context providers work seamlessly with session expiration
- **Extensible**: Easy to add new context sources (databases, APIs, files, etc.)

This approach makes "Context is Everything" not just a philosophy, but a practical implementation that scales from simple chatbots to complex multi-modal agents.

### Context vs Session - Key Concept Clarification

It's important to understand the distinction between these two core concepts:

**Context** = Information ingredients (short-lived, stateless)
- Assembled fresh for each execution
- Used to build LLM prompts
- Examples: system instructions, recent messages, current user preferences

**Session** = State container (persistent, stateful)  
- Persists across multiple executions
- Stores conversation history and variables
- Examples: user settings, conversation history, TTL management

Here's how contexts are dynamically assembled for each request:

```
┌─ Step 1: Session Management ─────────────────────────────────────────┐
│ 🚀 User Input: "What's the best time to visit Tokyo?"               │
│ 💾 Session Lookup: Load session "user-123"                          │
│ Found: current_task="Planning Tokyo trip", 3 previous messages      │
└─────────────────────────────────────────────────────────────────────┘
                                   │
┌─ Step 2: Context Assembly ───────────────────────────────────────────┐
│ ⚡ Gather from all providers:                                        │
│                                                                      │
│ 🎯 System Provider →                                                 │
│   Context: "You are a helpful travel assistant."                    │
│                                                                      │
│ 📋 Task Provider (from session state) →                             │
│   Context: "Current task: Planning Tokyo trip"                      │
│                                                                      │
│ 📜 History (from session entries) →                                 │
│   Context: "user: I want to plan a Tokyo trip"                      │
│   Context: "assistant: Great! I'd love to help you plan."           │
│   Context: "user: My budget is $3000"                               │
│                                                                      │
│ 🔗 Result: 5 contexts ready for LLM                                 │
└─────────────────────────────────────────────────────────────────────┘
                                   │
┌─ Step 3: LLM Prompt Construction ────────────────────────────────────┐
│ 🤖 Combined into LLM messages:                                      │
│                                                                      │
│ [                                                                    │
│   {role: "system", content: "You are a helpful travel assistant."}  │
│   {role: "system", content: "Current task: Planning Tokyo trip"}    │
│   {role: "user", content: "I want to plan a Tokyo trip"}           │
│   {role: "assistant", content: "Great! I'd love to help you plan."} │
│   {role: "user", content: "My budget is $3000"}                     │
│   {role: "user", content: "What's the best time to visit Tokyo?"}   │
│ ]                                                                    │
│                                                                      │
│ 💬 LLM Response: "The best time to visit Tokyo is..."               │
└─────────────────────────────────────────────────────────────────────┘
                                   │
┌─ Step 4: Session Update ─────────────────────────────────────────────┐
│ 💾 Save to session history:                                         │
│   - New user message: "What's the best time to visit Tokyo?"        │
│   - New assistant response: "The best time to visit Tokyo is..."    │
│ 🔄 Session now has 5 total messages for next interaction            │
└─────────────────────────────────────────────────────────────────────┘
```

The beauty is that **Context** is assembled fresh each time from the persistent **Session** state, ensuring both consistency and flexibility.

### Designing Effective Context Providers

Context Providers are the heart of our framework's flexibility. They determine what information your agent has access to and how it understands the conversation. Let's explore different patterns and real-world scenarios:

**1. Static Context Providers**
These provide consistent information regardless of session state:

```go
// Assumes imports:
// import (
//     "context"
//     agentcontext "github.com/davidleitw/go-agent/context"
//     "github.com/davidleitw/go-agent/session"
// )

// System role definition
type RoleProvider struct {
    role string
}

func (p *RoleProvider) Type() string {
    return "system"
}

func (p *RoleProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    return []agentcontext.Context{{
        Type: "system",
        Content: p.role,
        Metadata: map[string]any{"priority": "high"},
    }}
}

// Usage: Customer service agent
roleProvider := &RoleProvider{
    role: "You are a friendly customer service agent. Always acknowledge the customer's concerns and provide solutions.",
}
```

**2. Dynamic Session-Based Providers**
These adapt based on session state and history:

```go
// Assumes same imports as above

// User preference provider
type UserPreferenceProvider struct {
    userDB UserDatabase
}

func (p *UserPreferenceProvider) Type() string {
    return "user_preferences"
}

func (p *UserPreferenceProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    userID, exists := s.Get("user_id")
    if !exists {
        return nil // No user context yet
    }
    
    prefs := p.userDB.GetPreferences(userID.(string))
    return []agentcontext.Context{{
        Type: "user_preferences",
        Content: fmt.Sprintf("User prefers: language=%s, style=%s, expertise=%s",
            prefs.Language, prefs.CommunicationStyle, prefs.ExpertiseLevel),
    }}
}
```

**3. Conditional Providers**
These provide different contexts based on conditions:

```go
// Assumes same imports as above

// Business hours provider
type BusinessHoursProvider struct {
    timezone string
}

func (p *BusinessHoursProvider) Type() string {
    return "business_hours"
}

func (p *BusinessHoursProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    loc, _ := time.LoadLocation(p.timezone)
    now := time.Now().In(loc)
    hour := now.Hour()
    
    if hour >= 9 && hour < 17 {
        return []agentcontext.Context{{
            Type: "availability",
            Content: "During business hours. Can offer immediate assistance and schedule calls.",
        }}
    }
    
    return []agentcontext.Context{{
        Type: "availability", 
        Content: "Outside business hours. Can still help but callbacks will be scheduled for next business day.",
    }}
}
```

**4. External Data Providers**
These fetch real-time information from external sources:

```go
// Assumes same imports as above

// Weather context provider for travel agent
type WeatherProvider struct {
    weatherAPI WeatherService
}

func (p *WeatherProvider) Type() string {
    return "environment_data"
}

func (p *WeatherProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    destination, exists := s.Get("travel_destination")
    if !exists {
        return nil
    }
    
    weather := p.weatherAPI.GetCurrent(ctx, destination.(string))
    return []agentcontext.Context{{
        Type: "environment_data",
        Content: fmt.Sprintf("Current weather in %s: %s, %d°C", 
            destination, weather.Condition, weather.Temperature),
        Metadata: map[string]any{
            "source": "weather_api",
            "timestamp": time.Now(),
        },
    }}
}
```

**5. Conversation Stage Providers**
These track and provide context about where you are in a workflow:

```go
// Assumes same imports as above

// Sales funnel stage provider
type SalesFunnelProvider struct{}

func (p *SalesFunnelProvider) Type() string {
    return "sales_guidance"
}

func (p *SalesFunnelProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    history := s.GetHistory(20)
    
    // Analyze conversation to determine stage
    stage := p.analyzeStage(history)
    
    stageGuidance := map[string]string{
        "discovery": "Focus on understanding needs. Ask open-ended questions.",
        "qualification": "Determine budget and decision-making process.",
        "proposal": "Present solutions that match their stated needs.",
        "closing": "Address objections and guide toward decision.",
    }
    
    return []agentcontext.Context{{
        Type: "sales_guidance",
        Content: fmt.Sprintf("Current stage: %s. %s", stage, stageGuidance[stage]),
    }}
}
```

**Real-World Scenarios:**

**Customer Support Agent:**
```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(20).
    WithContextProviders(
        &RoleProvider{role: "Customer support specialist"},
        &UserPreferenceProvider{userDB: db},
        &TicketInfoProvider{ticketSystem: tickets},
        &BusinessHoursProvider{timezone: "America/New_York"},
        &SentimentProvider{}, // Monitors conversation tone
    ).
    Build()
if err != nil {
    log.Fatal(err)
}
```

**Technical Documentation Assistant:**
```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(15).
    WithContextProviders(
        &RoleProvider{role: "Technical documentation expert"},
        &CodeContextProvider{}, // Analyzes code snippets in conversation
        &VersionProvider{docDB: docs}, // Provides version-specific information
        &ExpertiseProvider{}, // Adjusts explanations based on user level
    ).
    Build()
if err != nil {
    log.Fatal(err)
}
```

**E-commerce Shopping Assistant:**
```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(10).
    WithContextProviders(
        &RoleProvider{role: "Personal shopping assistant"},
        &CartProvider{cartService: carts}, // Current cart contents
        &ProductProvider{catalog: products}, // Product recommendations
        &PriceAlertProvider{}, // Deals and discounts
        &OrderHistoryProvider{orderDB: orders},
    ).
    Build()
if err != nil {
    log.Fatal(err)
}
```

The power of Context Providers is that they separate concerns - each provider focuses on one aspect of context, making your system modular, testable, and easy to extend. You can mix and match providers to create agents perfectly suited to your use case!

## Prompt Template System - Flexible Message Organization

One of the most powerful features of go-agent is our **Prompt Template System**. This bridges the gap between collected contexts and the final LLM messages, giving you complete control over how information is organized and presented to your AI models.

### The Challenge We Solve

Traditional frameworks often hardcode how contexts become prompts:
```
Context Providers → [Hardcoded Logic] → LLM Messages
```

Our approach gives you flexible control:
```
Context Providers → [Your Template] → LLM Messages
```

### Three Ways to Define Templates

**1. Use Default Template (Zero Configuration)**
```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(10).  // Add history management
    WithContextProviders(systemProvider, customProvider).
    Build() // Uses sensible default template automatically
if err != nil {
    log.Fatal(err)
}
```

**2. Custom String Templates**
```go
customTemplate := `You are an expert {{role}} assistant.

Project Context:
{{project_info}}

Recent Conversation:
{{history}}

User Question: {{user_input}}`

agent, err := agent.NewBuilder().
    WithLLM(model).
    WithPromptTemplate(customTemplate).
    WithHistoryLimit(10).
    WithContextProviders(roleProvider, projectProvider).
    Build()
if err != nil {
    log.Fatal(err)
}
```

**3. Fluent Builder API (Maximum Flexibility)**
```go
template := prompt.New().
    System().
    Text("Project Information:").
    Provider("project_info").
    Line("").
    Text("Conversation History:").
    History().
    Line("").
    Text("Current Question:").
    UserInput().
    Build()

agent, err := agent.NewBuilder().
    WithLLM(model).
    WithPromptTemplate(template).
    WithHistoryLimit(10).
    WithContextProviders(projectProvider).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Template Variable System

Templates use `{{variable}}` syntax to reference different types of context:

**Built-in Variables:**
- `{{system}}` - System prompts and instructions
- `{{history}}` - Conversation history with preserved roles
- `{{user_input}}` - Current user message
- `{{context_providers}}` - All custom provider contexts

**Custom Variables:**
- `{{project_info}}` - Maps to providers with `Type() == "project_info"`
- `{{user_preferences}}` - Maps to providers with `Type() == "user_preferences"`
- `{{weather_data}}` - Maps to providers with `Type() == "weather_data"`

**Named References:**
```go
// Use specific named providers
template := `Main project: {{project_info:main}}
Secondary project: {{project_info:secondary}}`
```

### Real-World Template Examples

**Code Review Assistant:**
```go
template := `You are an expert Go code reviewer.

Coding Standards:
{{coding_standards}}

Project Context:
{{project_info}}

Previous Review Comments:
{{history}}

Please review this code:
{{user_input}}`
```

**Customer Support Agent:**
```go
template := `You are a helpful customer support agent.

Customer Information:
{{customer_profile}}

Current Ticket Context:
{{ticket_info}}

Company Policies:
{{support_policies}}

Support History:
{{history}}

Current Issue: {{user_input}}`
```

**Technical Documentation Assistant:**
```go
template := prompt.New().
    Text("You are a technical documentation expert.").
    Line("").
    Text("Library Information:").
    Provider("library_context").
    Text("Code Examples:").
    Provider("code_examples").
    Text("User Question:").
    UserInput().
    Build()
```

### Advanced Features

**Conditional Content:**
Templates automatically handle missing providers - if a provider doesn't exist, that section is simply skipped.

**Role Preservation:**
History contexts maintain original message roles (user/assistant/system) for natural conversation flow.

**Metadata Access:**
Templates can access provider metadata for advanced scenarios.

### Template Best Practices

1. **Start Simple**: Use default template, then customize as needed
2. **Logical Flow**: Order information from general to specific
3. **Clear Separators**: Use text sections to create clear boundaries
4. **Descriptive Variables**: Use meaningful provider type names
5. **Test Different Approaches**: String templates for simplicity, Builder for complexity

### Integration Example

See our [Academic Research Assistant Example](./examples/academic-research-assistant/) for a comprehensive demonstration of the template system in action, including:
- Multiple template approaches with workflow-specific prompts
- Custom context providers for research contexts
- 6 specialized research tools (ArXiv search, paper analysis, citation management, etc.)
- Session management with history and research note organization

### Template System Benefits

- **Flexibility**: Choose the right approach for your complexity level
- **Maintainability**: Templates are easy to read and modify
- **Reusability**: Share templates across different agents
- **Type Safety**: Provider type system ensures correct mapping
- **Performance**: Efficient rendering with minimal overhead

This template system makes go-agent truly adaptable to any use case while maintaining our core philosophy that "Context is Everything" - now with complete control over how that context is organized and presented!

### [Agent Module](./agent/) - Core Controller
This is the brain of the framework, coordinating all other modules. Provides a simple `Execute()` interface and flexible Builder pattern for easy configuration.

**Key Features:**
- Clean Agent interface - one method does everything
- Builder pattern makes configuration intuitive
- Automatic session management - no state worries
- Built-in convenience functions for common patterns

### [Session Module](./session/) - Memory Management
Handles conversation state and history. Supports TTL auto-expiration, concurrent safety, and complete JSON serialization.

**Key Features:**
- Key-Value state storage for any data type
- Unified history format supporting multiple conversation types
- Automatic TTL management with expired session cleanup
- Thread-safe for multi-goroutine usage

### [Context Module](./context/) - Information Aggregation
This module's job is to package information from various sources (conversation history, system prompts, external data, etc.) into a unified format that LLMs can understand.

**Key Features:**
- Unified Context data structure
- Extensible Provider system
- Automatic Session history to Context conversion
- Rich Metadata support

### [Tool Module](./tool/) - Tool Integration
Enables your AI agents to call external functions like database queries, API calls, calculations, etc.

**Key Features:**
- Simple Tool interface - easy to implement custom tools
- JSON Schema-based parameter definitions
- Thread-safe tool registry
- Complete error handling mechanisms

### [LLM Module](./llm/) - Language Model Interface
Provides unified language model interface. Currently supports OpenAI, with plans to expand to other providers.

**Key Features:**
- Clear Model interface
- Built-in tool calling support
- Complete token usage tracking
- Support for custom endpoints and proxies

## History Management

The go-agent framework provides intelligent conversation history management that can scale from simple use cases to sophisticated Claude Code-level implementations.

### Basic Usage

Enable history tracking with a simple limit:

```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(20).  // Keep last 20 conversation turns
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Advanced History Processing

For complex scenarios requiring compression, filtering, or intelligent summarization, implement the `HistoryInterceptor` interface:

```go
type HistoryInterceptor interface {
    ProcessHistory(ctx context.Context, entries []session.Entry, llm llm.Model) ([]session.Entry, error)
}
```

### Claude Code-Level Implementation Example

Here's how to implement sophisticated history management similar to Claude Code:

```go
type AdvancedHistoryCompressor struct {
    maxTokens        int
    recentLimit      int
    compressionRatio float32
}

func (c *AdvancedHistoryCompressor) ProcessHistory(ctx context.Context, entries []session.Entry, llm llm.Model) ([]session.Entry, error) {
    if len(entries) <= c.recentLimit {
        return entries, nil
    }

    // 1. Preserve recent conversations
    recent := entries[len(entries)-c.recentLimit:]
    older := entries[:len(entries)-c.recentLimit]

    // 2. Identify important entries
    important := c.filterImportant(older)
    
    // 3. Generate compressed summary using LLM
    summary, err := c.generateSummary(ctx, older, llm)
    if err != nil {
        return entries, nil // Fallback to original on error
    }

    // 4. Combine summary + important entries + recent
    result := []session.Entry{summary}
    result = append(result, important...)
    result = append(result, recent...)
    
    return result, nil
}

func (c *AdvancedHistoryCompressor) generateSummary(ctx context.Context, entries []session.Entry, llm llm.Model) (session.Entry, error) {
    // Build compression prompt
    historyText := c.formatEntriesForSummary(entries)
    
    response, err := llm.Complete(ctx, llm.Request{
        Messages: []llm.Message{
            {
                Role: "system", 
                Content: "You are a conversation summarizer. Preserve key information, decisions, and context.",
            },
            {
                Role: "user",
                Content: fmt.Sprintf("Summarize this conversation history:\n\n%s", historyText),
            },
        },
    })
    
    if err != nil {
        return session.Entry{}, err
    }
    
    // Return as system message entry
    return session.NewMessageEntry("system", 
        fmt.Sprintf("[Compressed History Summary]\n%s", response.Content)), nil
}

func (c *AdvancedHistoryCompressor) filterImportant(entries []session.Entry) []session.Entry {
    var important []session.Entry
    
    for _, entry := range entries {
        // Custom importance scoring logic
        if c.isImportant(entry) {
            important = append(important, entry)
        }
    }
    
    return important
}

func (c *AdvancedHistoryCompressor) isImportant(entry session.Entry) bool {
    // Example importance criteria:
    // - Error messages
    // - Successful tool executions with valuable results
    // - User preferences or settings
    // - Key decisions or confirmations
    
    if entry.Type == session.EntryTypeToolResult {
        if content, ok := session.GetToolResultContent(entry); ok {
            return !content.Success || c.hasValueableResult(content.Result)
        }
    }
    
    // Check for error keywords, preferences, etc.
    return false
}

// Usage
compressor := &AdvancedHistoryCompressor{
    maxTokens:        4000,
    recentLimit:      10,
    compressionRatio: 0.3,
}

agent, err := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(100).
    WithHistoryInterceptor(compressor).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Key Features

**Intelligent Compression:**
- LLM-powered summarization
- Importance-based entry preservation
- Token limit management
- Configurable compression ratios

**Context Awareness:**
- Automatic history notices in system prompts
- Maintains conversation continuity
- Preserves critical information

**Performance Optimized:**
- Internal history processing (no ContextProvider overhead)
- Asynchronous processing capability
- Efficient entry conversion

**Extensible Design:**
- Simple interface for custom implementations
- Access to full LLM capabilities for processing
- Integration with session metadata

### System Prompt Integration

When history is processed, the system automatically informs the LLM:

```
Note on Conversation History:
The conversation history provided may have been compressed or summarized to save space.
Key information and context have been preserved, but some details might be condensed.
Please use this history as reference for maintaining conversation continuity and context.
```

This approach enables building sophisticated conversation agents that can maintain context across long interactions while managing token costs and processing efficiency.

## Current Development Status

**Ready to Use:**
- Complete module interface design and implementation
- Session management with TTL support
- Context provider system
- Tool registration and execution framework
- Agent core execution engine with tool orchestration
- Prompt template system for flexible message organization
- History management with compression support
- OpenAI integration
- Rich test coverage

**In Active Development:**
- More LLM provider support (Anthropic, Google, etc.)
- Streaming response support
- More built-in tools and examples
- Performance optimizations

**Future Plans (Roadmap subject to change):**
- Redis/Database Session storage
- Asynchronous tool execution
- Advanced Context management features
- MCP (Model Context Protocol) tool integration
- Long-running agent reliability testing
- Production deployment patterns

## Design Philosophy

### "Context is Everything"
We believe the core of AI agents is context management. Whether it's conversation history, user preferences, external data, or tool execution results, everything needs to be provided to LLMs in a consistent way.

We're planning to organize talks and compile resources about Context Engineering to help the community better understand this approach.

## Contributing

This project is under active development, and we welcome all forms of participation:

**Interface Design Discussion (Most Important!):**
- Think some interface design isn't intuitive enough?
- Have better API design ideas?
- Feel some functionality abstraction levels are wrong?
- Want a module to provide different usage patterns?

We deeply believe good interface design is key to framework success - any friends with interface ideas are very welcome to discuss!

**Feature Suggestions:**
- What new features would you like to see?
- What usage difficulties have you encountered?
- What real-world scenarios haven't we considered?

**Code Contributions:**
- Implement new LLM providers
- Build more practical tools
- Improve performance and stability
- Add more tests and examples

**Documentation and Examples:**
- Write usage tutorials
- Create real-world application examples
- Translate documentation

Feel free to open Issues for discussion or submit PRs directly. We're happy to work together to make this framework better.

## Getting Started

1. **Check Module Documentation**: Each folder has detailed READMEs - suggest starting with [Agent Module](./agent/)
2. **Run Tests**: `go test ./...` to see if everything works
3. **Join Discussion**: Open Issues with questions or ideas

## License

MIT License

---

Looking forward to seeing what interesting things you build with this framework!