package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/joho/godotenv"
)

// WeatherTool provides weather information for any location
type WeatherTool struct{}

func (w *WeatherTool) Name() string {
	return "get_weather"
}

func (w *WeatherTool) Description() string {
	return "Get current weather information for a specified location"
}

func (w *WeatherTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "The city or location to get weather for",
			},
		},
		"required": []string{"location"},
	}
}

func (w *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	location, ok := args["location"].(string)
	if !ok {
		return nil, fmt.Errorf("location must be a string")
	}

	// Simulate weather API call
	return map[string]any{
		"location":    location,
		"temperature": "22°C",
		"condition":   "Sunny",
		"humidity":    "65%",
		"wind":        "5 km/h",
	}, nil
}

// CalculatorTool performs basic mathematical operations
type CalculatorTool struct{}

func (c *CalculatorTool) Name() string {
	return "calculator"
}

func (c *CalculatorTool) Description() string {
	return "Perform basic mathematical calculations"
}

func (c *CalculatorTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"description": "The mathematical operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide", "sqrt"},
			},
			"a": map[string]any{
				"type":        "number",
				"description": "First number",
			},
			"b": map[string]any{
				"type":        "number",
				"description": "Second number (not required for sqrt)",
			},
		},
		"required": []string{"operation", "a"},
	}
}

func (c *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	a, ok := args["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("a must be a number")
	}

	switch operation {
	case "add":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("b must be a number for addition")
		}
		return a + b, nil

	case "subtract":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("b must be a number for subtraction")
		}
		return a - b, nil

	case "multiply":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("b must be a number for multiplication")
		}
		return a * b, nil

	case "divide":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("b must be a number for division")
		}
		if b == 0 {
			return nil, fmt.Errorf("cannot divide by zero")
		}
		return a / b, nil

	case "sqrt":
		if a < 0 {
			return nil, fmt.Errorf("cannot calculate square root of negative number")
		}
		return math.Sqrt(a), nil

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// TimeTool provides current time information
type TimeTool struct{}

func (t *TimeTool) Name() string {
	return "get_time"
}

func (t *TimeTool) Description() string {
	return "Get current time and date information"
}

func (t *TimeTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"timezone": map[string]any{
				"type":        "string",
				"description": "Timezone to get time for (optional, defaults to UTC)",
			},
		},
	}
}

func (t *TimeTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	timezone := "UTC"
	if tz, ok := args["timezone"].(string); ok && tz != "" {
		timezone = tz
	}

	now := time.Now()
	if timezone != "UTC" {
		// For demo purposes, just offset by a few hours for common timezones
		switch timezone {
		case "EST", "America/New_York":
			now = now.Add(-5 * time.Hour)
		case "PST", "America/Los_Angeles":
			now = now.Add(-8 * time.Hour)
		case "JST", "Asia/Tokyo":
			now = now.Add(9 * time.Hour)
		}
	}

	return map[string]any{
		"current_time": now.Format("2006-01-02 15:04:05"),
		"timezone":     timezone,
		"day_of_week":  now.Weekday().String(),
		"unix_timestamp": now.Unix(),
	}, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	fmt.Println("🤖 Multi-Tool Agent Example")
	fmt.Println("=========================")
	fmt.Println("This example demonstrates an agent that can use multiple tools:")
	fmt.Println("- Weather information")
	fmt.Println("- Mathematical calculations") 
	fmt.Println("- Current time and date")
	fmt.Println()

	// Create tools
	weatherTool := &WeatherTool{}
	calculatorTool := &CalculatorTool{}
	timeTool := &TimeTool{}

	// Create agent with multiple tools using the elegant builder API
	assistant, err := agent.New("multi-tool-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("An AI assistant with access to weather, calculator, and time tools").
		WithInstructions(`You are a helpful assistant with access to multiple tools:
- get_weather: Get weather information for any location
- calculator: Perform mathematical calculations  
- get_time: Get current time and date information

Use these tools when users ask questions that require this information. 
Be helpful and provide clear, complete responses.`).
		WithTools(weatherTool, calculatorTool, timeTool).
		Build()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Demonstrate different tool usage scenarios
	demonstrations := []string{
		"What's the weather like in Tokyo?",
		"Calculate the square root of 144",
		"What time is it in New York?",
		"If I have 15 apples and give away 7, how many do I have left?",
		"Can you tell me the weather in Paris and what time it is there?",
	}

	ctx := context.Background()
	sessionID := fmt.Sprintf("multi-tool-demo-%d", time.Now().Unix())

	for i, userInput := range demonstrations {
		fmt.Printf("👤 User: %s\n", userInput)
		
		response, err := assistant.Chat(ctx, userInput, agent.WithSession(sessionID))
		if err != nil {
			log.Printf("❌ Error: %v", err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", response.Message)
		
		if i < len(demonstrations)-1 {
			fmt.Println()
			time.Sleep(1 * time.Second) // Small pause for readability
		}
	}

	fmt.Println()
	fmt.Println("✅ Multi-tool demonstration completed!")
}