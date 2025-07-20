# Context Module

The Context module provides a unified data structure and provider system for managing different types of information that need to be passed to LLMs in the Go Agent Framework. It acts as a bridge between various data sources (sessions, prompts, tools) and the final prompt construction.

## Features

- **Unified Context Structure**: Consistent data format for all types of information
- **Provider Pattern**: Extensible system for different context sources
- **Session Integration**: Automatic conversion from session history to contexts
- **Metadata Support**: Rich metadata preservation and enhancement
- **Type Safety**: Proper handling of different entry types from sessions

## Core Concepts

### Context Structure

```go
type Context struct {
    Type     string         // Context type (e.g., "user", "assistant", "system", "tool_call")
    Content  string         // Text content of the context
    Metadata map[string]any // Additional metadata and information
}
```

### Provider Interface

```go
type Provider interface {
    Provide(s session.Session) []Context
}
```

Providers are responsible for converting data sources into Context objects that can be used by the Agent.

## Built-in Providers

### SystemPromptProvider

Provides system-level instructions and prompts.

```go
provider := context.NewSystemPromptProvider("You are a helpful assistant.")
contexts := provider.Provide(session)
// Returns: [Context{Type: "system", Content: "You are a helpful assistant.", ...}]
```

### HistoryProvider

Converts session history into context objects, handling all entry types.

```go
provider := context.NewHistoryProvider(10) // Limit to 10 most recent entries
contexts := provider.Provide(session)
```

#### Entry Type Conversion

1. **Message Entries** → Context with role-based type
   - `session.MessageContent{Role: "user", Text: "Hello"}` → `Context{Type: "user", Content: "Hello"}`

2. **Tool Call Entries** → Context with formatted tool information
   - `session.ToolCallContent{Tool: "search", Parameters: {...}}` → `Context{Type: "tool_call", Content: "Tool: search\nParameters: {...}"}`

3. **Tool Result Entries** → Context with success/error information
   - Success: `Context{Type: "tool_result", Content: "Tool: search\nSuccess: true\nResult: [...]"}`
   - Error: `Context{Type: "tool_result", Content: "Tool: search\nSuccess: false\nError: connection failed"}`

4. **Thinking Entries** → Context for internal reasoning
   - `Context{Type: "thinking", Content: "I need to consider..."}`

## Usage Examples

### Basic Usage

```go
import (
    "github.com/davidleitw/go-agent/context"
    "github.com/davidleitw/go-agent/session/memory"
)

// Create session with history
store := memory.NewStore()
sess := store.Create()

sess.AddEntry(session.NewMessageEntry("user", "Book a flight to Tokyo"))
sess.AddEntry(session.NewMessageEntry("assistant", "I'll help you book a flight."))

// Get system context
systemProvider := context.NewSystemPromptProvider("You are a travel assistant.")
systemContexts := systemProvider.Provide(sess)

// Get history contexts
historyProvider := context.NewHistoryProvider(5)
historyContexts := historyProvider.Provide(sess)

// Combine contexts for LLM prompt
allContexts := append(systemContexts, historyContexts...)
```

### Advanced Metadata Usage

```go
// HistoryProvider automatically adds metadata
contexts := historyProvider.Provide(sess)
for _, ctx := range contexts {
    entryID := ctx.Metadata["entry_id"].(string)
    timestamp := ctx.Metadata["timestamp"].(time.Time)
    
    if ctx.Type == "tool_call" {
        toolName := ctx.Metadata["tool_name"].(string)
        // Use tool name for special handling
    }
    
    if ctx.Type == "tool_result" {
        success := ctx.Metadata["success"].(bool)
        // Handle based on success status
    }
}
```

## Context Types

### Standard Types
- `"system"` - System prompts and instructions
- `"user"` - User messages and inputs
- `"assistant"` - Assistant responses
- `"tool_call"` - Tool invocation requests
- `"tool_result"` - Tool execution results
- `"thinking"` - Internal reasoning (reserved for future use)

### Custom Types
Providers can define custom types for specialized use cases. Unknown types are handled gracefully with JSON fallback formatting.

## Metadata Fields

### Common Fields (added by HistoryProvider)
- `"entry_id"` - Unique identifier from session entry
- `"timestamp"` - When the entry was created

### Tool-specific Fields
- `"tool_name"` - Name of the tool (for tool_call and tool_result types)
- `"success"` - Boolean success status (for tool_result types)

### Custom Fields
Original metadata from session entries is preserved, allowing for application-specific information.

## Testing

The module includes comprehensive tests covering:

- Empty history handling
- All entry type conversions
- Metadata preservation and enhancement
- Ordering (newest-first from session history)
- Limit functionality
- Mixed entry types
- Fallback for unknown types
- Error handling

Run tests with:
```bash
go test ./context -v
```

## Creating Custom Providers

```go
type CustomProvider struct {
    data string
}

func NewCustomProvider(data string) Provider {
    return &CustomProvider{data: data}
}

func (p *CustomProvider) Provide(s session.Session) []Context {
    // Custom logic to generate contexts
    return []Context{
        {
            Type:     "custom",
            Content:  p.data,
            Metadata: map[string]any{"source": "custom_provider"},
        },
    }
}
```

## Integration with Agent System

*To be implemented: This section will be filled in after the Agent core functionality is complete*

The Context module is designed to integrate seamlessly with the Agent's prompt construction system:

```go
// Expected integration
agent := agent.New(
    agent.WithSystemPrompt("You are helpful"),
    agent.WithContextProviders(
        context.NewHistoryProvider(10),
        // other providers...
    ),
)
```

## Performance Considerations

- **Memory Efficiency**: Contexts are created on-demand, not cached
- **Ordering**: History is already sorted by session, no additional sorting needed
- **JSON Marshaling**: Only performed when necessary for complex data types
- **Metadata Copying**: Shallow copy for performance while preserving original data

## Best Practices

1. **Provider Selection**: Use appropriate providers for your use case
   - SystemPromptProvider for static instructions
   - HistoryProvider for conversation context

2. **Limit Setting**: Set reasonable limits for HistoryProvider to control context size
   - Consider token limits of your target LLM
   - Balance between context richness and performance

3. **Metadata Usage**: Leverage metadata for:
   - Debugging and logging
   - Conditional logic based on entry types
   - Analytics and monitoring

4. **Error Handling**: Providers are designed to be resilient
   - Invalid entries are handled gracefully
   - Type assertion failures fall back to JSON formatting

## Architecture

```
Session History → HistoryProvider → Context[] → Agent → LLM
System Prompt  → SystemPromptProvider → Context[] ↗
Custom Data    → CustomProvider → Context[] ↗
```

The Context module serves as the data transformation layer, converting various input sources into a unified format that the Agent can process and send to LLMs.

## Future Extensions

1. **Additional Providers**:
   - FileProvider (for document context)
   - DatabaseProvider (for dynamic data)
   - APIProvider (for external data sources)

2. **Context Filtering**:
   - Content-based filtering
   - Time-based filtering
   - Relevance scoring

3. **Context Compression**:
   - Smart truncation for large contexts
   - Summarization of older history
   - Importance-based selection

## Contributing

When adding new providers or extending functionality:
- Follow the Provider interface
- Add comprehensive tests
- Update documentation
- Consider metadata standardization
- Ensure graceful error handling

## License

MIT License - See LICENSE file in the project root directory.