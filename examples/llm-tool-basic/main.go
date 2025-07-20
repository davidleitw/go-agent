package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/llm/openai"
	"github.com/davidleitw/go-agent/tool"
)

// WeatherTool implements a simple weather tool
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

	// Mock weather data
	return map[string]any{
		"location":    location,
		"temperature": 22,
		"unit":        unit,
		"description": "Sunny",
		"humidity":    60,
	}, nil
}

// CalculatorTool implements basic arithmetic
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
						Description: "Operation (add/subtract/multiply/divide)",
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
			return nil, fmt.Errorf("division by zero")
		}
		return a / b, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
}

func main() {
	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create LLM client
	model := openai.New(llm.Config{
		APIKey: apiKey,
		Model:  "gpt-3.5-turbo",
	})

	// Create tool registry
	registry := tool.NewRegistry()

	// Register tools
	registry.Register(&WeatherTool{})
	registry.Register(&CalculatorTool{})

	// Get tool definitions for LLM
	tools := registry.GetDefinitions()

	fmt.Println("=== LLM + Tool Integration Demo ===")

	// Example 1: Simple conversation
	fmt.Println("1. Simple conversation:")
	resp1, err := model.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello! What can you help me with?"},
		},
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Assistant: %s\n", resp1.Content)
		fmt.Printf("Tokens used: %d\n\n", resp1.Usage.TotalTokens)
	}

	// Example 2: Using tools (mock - requires actual API call)
	fmt.Println("2. Tool integration example:")
	fmt.Println("Request: What's the weather in Tokyo and calculate 15 + 27?")

	// In a real scenario, this would be the actual LLM response
	// For demonstration, we'll simulate tool calls
	mockToolCalls := []tool.Call{
		{
			ID: "call_1",
			Function: tool.FunctionCall{
				Name:      "get_weather",
				Arguments: `{"location": "Tokyo", "unit": "celsius"}`,
			},
		},
		{
			ID: "call_2",
			Function: tool.FunctionCall{
				Name:      "calculator",
				Arguments: `{"operation": "add", "a": 15, "b": 27}`,
			},
		},
	}

	// Execute tools
	fmt.Println("\nExecuting tools:")
	for _, call := range mockToolCalls {
		result, err := registry.Execute(context.Background(), call)
		if err != nil {
			fmt.Printf("Error executing %s: %v\n", call.Function.Name, err)
			continue
		}

		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("Tool: %s\nResult: %s\n\n", call.Function.Name, string(resultJSON))
	}

	// Example 3: Show available tools
	fmt.Println("3. Available tools:")
	for _, toolDef := range tools {
		fmt.Printf("- %s: %s\n", toolDef.Function.Name, toolDef.Function.Description)

		// Show parameters
		for paramName, param := range toolDef.Function.Parameters.Properties {
			required := ""
			for _, req := range toolDef.Function.Parameters.Required {
				if req == paramName {
					required = " (required)"
					break
				}
			}
			fmt.Printf("  * %s (%s): %s%s\n", paramName, param.Type, param.Description, required)
		}
		fmt.Println()
	}

	fmt.Println("Demo completed! 🎉")
	fmt.Println("\nNote: For actual tool calls, the LLM would decide when and how to use tools.")
	fmt.Println("This demo shows the basic integration between LLM and Tool modules.")
}
