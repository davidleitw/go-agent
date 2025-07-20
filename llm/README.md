# LLM Module

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

The LLM module provides a unified interface for interacting with language models, with initial support for OpenAI's API.

## Features

- **Simple Interface**: Clean `Model` interface with synchronous completion
- **Tool Support**: Built-in support for function calling/tools
- **Type Safety**: Strongly typed requests and responses
- **Extensible**: Easy to add support for other LLM providers
- **Configuration**: Flexible configuration with API key and custom endpoints

## Quick Start

```go
import (
    "github.com/davidleitw/go-agent/llm"
    "github.com/davidleitw/go-agent/llm/openai"
)

// Create an OpenAI client
model := openai.New(llm.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Model:  "gpt-4",
})

// Make a completion request
resp, err := model.Complete(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "system", Content: "You are a helpful assistant"},
        {Role: "user", Content: "What's the weather like?"},
    },
})

fmt.Println(resp.Content)
```

## API Reference

### Model Interface

```go
type Model interface {
    Complete(ctx context.Context, request Request) (*Response, error)
    // TODO: Stream method for streaming completions
}
```

### Request Structure

```go
type Request struct {
    Messages     []Message         // Conversation messages
    Temperature  *float32          // Optional: 0.0-2.0
    MaxTokens    *int             // Optional: Max tokens to generate
    Tools        []tool.Definition // Optional: Available tools
}
```

### Response Structure

```go
type Response struct {
    Content      string      // Generated text
    ToolCalls    []tool.Call // Tool invocations if any
    Usage        Usage       // Token usage statistics
    FinishReason string      // stop/length/tool_calls
}
```

## Using with Tools

```go
// Define a tool
weatherTool := tool.Definition{
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

// Include tools in request
resp, err := model.Complete(ctx, llm.Request{
    Messages: messages,
    Tools:    []tool.Definition{weatherTool},
})

// Handle tool calls
if len(resp.ToolCalls) > 0 {
    for _, call := range resp.ToolCalls {
        // Execute tool and continue conversation
        result := executeWeatherTool(call.Function.Arguments)
        
        // Add tool response to messages
        messages = append(messages, llm.Message{
            Role:       "tool",
            Content:    result,
            Name:       call.Function.Name,
            ToolCallID: call.ID,
        })
    }
}
```

## Configuration

### Basic Configuration

```go
config := llm.Config{
    APIKey: "your-api-key",
    Model:  "gpt-4",
}
```

### Using Custom Endpoint

```go
config := llm.Config{
    APIKey:  "your-api-key",
    Model:   "gpt-4",
    BaseURL: "https://your-proxy.com/v1", // For proxies or custom endpoints
}
```

## Token Usage

Track token consumption for cost management:

```go
resp, _ := model.Complete(ctx, request)

fmt.Printf("Prompt tokens: %d\n", resp.Usage.PromptTokens)
fmt.Printf("Completion tokens: %d\n", resp.Usage.CompletionTokens)
fmt.Printf("Total tokens: %d\n", resp.Usage.TotalTokens)
```

## Error Handling

```go
resp, err := model.Complete(ctx, request)
if err != nil {
    // Handle API errors
    log.Printf("LLM error: %v", err)
    return
}

// Check finish reason
switch resp.FinishReason {
case "stop":
    // Normal completion
case "length":
    // Hit max tokens limit
case "tool_calls":
    // Model wants to use tools
default:
    // Unexpected finish reason
}
```

## Future Enhancements

The following features are planned for future releases:

- **Streaming Support**: Real-time token streaming
- **Additional Parameters**: TopP, stop sequences, frequency penalty
- **Multi-modal Support**: Image inputs
- **Provider Extensions**: Anthropic, Google, and other providers
- **Response Validation**: Schema validation for responses
- **Retry Logic**: Automatic retry with exponential backoff
- **Rate Limiting**: Built-in rate limit handling

## Testing

The module includes comprehensive tests:

```bash
go test ./llm/... -v
```

For integration testing with actual API:

```go
// Use mock in tests
type mockModel struct{}

func (m *mockModel) Complete(ctx context.Context, req llm.Request) (*llm.Response, error) {
    return &llm.Response{
        Content: "Mock response",
        Usage: llm.Usage{
            TotalTokens: 10,
        },
    }, nil
}
```

## Best Practices

1. **API Key Security**: Never hardcode API keys, use environment variables
2. **Context Usage**: Always pass context for cancellation support
3. **Error Handling**: Check both error and finish reason
4. **Token Limits**: Set reasonable MaxTokens to control costs
5. **Temperature**: Use lower temperature (0.0-0.3) for deterministic outputs

## License

MIT License - See LICENSE file in the project root directory.