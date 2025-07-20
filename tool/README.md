# Tool Module

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

The Tool module provides a framework for defining and executing tools that can be used by language models in the Go Agent Framework.

## Features

- **Simple Tool Interface**: Easy to implement custom tools
- **Type-Safe Definitions**: JSON Schema-based parameter definitions
- **Registry Management**: Central registry for tool registration and execution
- **Thread-Safe**: Concurrent access support
- **Extensible**: Ready for future enhancements like validation and MCP support

## Quick Start

```go
import (
    "github.com/davidleitw/go-agent/tool"
)

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
                    "unit": {
                        Type:        "string",
                        Description: "Temperature unit (celsius/fahrenheit)",
                    },
                },
                Required: []string{"location"},
            },
        },
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    location, _ := params["location"].(string)
    unit, _ := params["unit"].(string)
    if unit == "" {
        unit = "celsius"
    }
    
    // Implement weather fetching logic
    return map[string]any{
        "temperature": 22,
        "unit":        unit,
        "description": "Sunny",
    }, nil
}

// Use the tool
registry := tool.NewRegistry()
registry.Register(&WeatherTool{})
```

## API Reference

### Tool Interface

```go
type Tool interface {
    Definition() Definition
    Execute(ctx context.Context, params map[string]any) (any, error)
}
```

### Definition Structure

```go
type Definition struct {
    Type     string   // Always "function" for now
    Function Function
}

type Function struct {
    Name        string
    Description string
    Parameters  Parameters
}
```

### Parameters (JSON Schema Subset)

```go
type Parameters struct {
    Type       string              // "object"
    Properties map[string]Property
    Required   []string
}

type Property struct {
    Type        string // string/number/boolean/array/object
    Description string
}
```

## Registry Usage

### Creating and Managing Registry

```go
// Create registry
registry := tool.NewRegistry()

// Register tools
err := registry.Register(tool1)
err := registry.Register(tool2)

// Get all definitions (for LLM)
definitions := registry.GetDefinitions()

// Get specific tool
weatherTool, exists := registry.Get("get_weather")

// Clear all tools
registry.Clear()
```

### Executing Tools

```go
// Execute a tool call from LLM
call := tool.Call{
    ID: "call_123",
    Function: tool.FunctionCall{
        Name:      "get_weather",
        Arguments: `{"location": "Tokyo", "unit": "celsius"}`,
    },
}

result, err := registry.Execute(ctx, call)
if err != nil {
    log.Printf("Tool execution failed: %v", err)
}
```

## Creating Custom Tools

### Simple Calculator Tool

```go
type CalculatorTool struct{}

func (c *CalculatorTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "calculator",
            Description: "Perform basic arithmetic operations",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "operation": {
                        Type:        "string",
                        Description: "Operation to perform (add/subtract/multiply/divide)",
                    },
                    "a": {
                        Type:        "number",
                        Description: "First number",
                    },
                    "b": {
                        Type:        "number",
                        Description: "Second number",
                    },
                },
                Required: []string{"operation", "a", "b"},
            },
        },
    }
}

func (c *CalculatorTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    op, _ := params["operation"].(string)
    a, _ := params["a"].(float64)
    b, _ := params["b"].(float64)
    
    switch op {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return nil, errors.New("division by zero")
        }
        return a / b, nil
    default:
        return nil, fmt.Errorf("unknown operation: %s", op)
    }
}
```

### Database Query Tool

```go
type DatabaseTool struct {
    db *sql.DB
}

func (d *DatabaseTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "query_database",
            Description: "Query user information from database",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "user_id": {
                        Type:        "string",
                        Description: "User ID to query",
                    },
                },
                Required: []string{"user_id"},
            },
        },
    }
}

func (d *DatabaseTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    userID, _ := params["user_id"].(string)
    
    var user struct {
        ID    string
        Name  string
        Email string
    }
    
    err := d.db.QueryRowContext(ctx, 
        "SELECT id, name, email FROM users WHERE id = ?", 
        userID,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

## Error Handling

Tools should return meaningful errors:

```go
func (t *MyTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    // Validate required parameters
    value, ok := params["required_param"].(string)
    if !ok {
        return nil, fmt.Errorf("required_param must be a string")
    }
    
    // Handle context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Business logic errors
    result, err := doSomething(value)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    return result, nil
}
```

## Thread Safety

The registry is thread-safe for concurrent access:

```go
// Safe to use from multiple goroutines
go func() {
    registry.Register(tool1)
}()

go func() {
    result, _ := registry.Execute(ctx, call)
}()

go func() {
    definitions := registry.GetDefinitions()
}()
```

## Future Enhancements

The following features are planned:

- **Input Validation**: Automatic parameter validation against schema
- **Output Schema**: Optional output validation
- **MCP Support**: Integration with Model Context Protocol
- **Async Execution**: Support for long-running tools
- **Middleware**: Interceptors for logging, metrics, etc.
- **Tool Composition**: Combine multiple tools
- **Rate Limiting**: Per-tool rate limits

## Testing

### Unit Testing Tools

```go
func TestMyTool(t *testing.T) {
    tool := &MyTool{}
    
    // Test definition
    def := tool.Definition()
    if def.Function.Name != "my_tool" {
        t.Errorf("Expected name my_tool, got %s", def.Function.Name)
    }
    
    // Test execution
    params := map[string]any{
        "param1": "value1",
    }
    
    result, err := tool.Execute(context.Background(), params)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    // Verify result
    if result != expectedResult {
        t.Errorf("Expected %v, got %v", expectedResult, result)
    }
}
```

### Testing with Registry

```go
func TestToolIntegration(t *testing.T) {
    registry := tool.NewRegistry()
    registry.Register(&MyTool{})
    
    call := tool.Call{
        ID: "test-call",
        Function: tool.FunctionCall{
            Name:      "my_tool",
            Arguments: `{"param1": "value1"}`,
        },
    }
    
    result, err := registry.Execute(context.Background(), call)
    // Assert results
}
```

## Best Practices

1. **Clear Descriptions**: Write clear, concise tool descriptions for LLM understanding
2. **Parameter Types**: Use appropriate JSON Schema types
3. **Error Messages**: Return helpful error messages
4. **Context Respect**: Always respect context cancellation
5. **Idempotency**: Make tools idempotent when possible
6. **Logging**: Add appropriate logging for debugging
7. **Documentation**: Document expected inputs and outputs

## License

MIT License - See LICENSE file in the project root directory.