package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/joho/godotenv"
)

// CalculatorTool implements basic mathematical operations
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
				"description": "The operation to perform",
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

func main() {
	fmt.Println("🧮 Calculator Tool Example")
	fmt.Println("===========================")

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	// Create calculator tool
	calcTool := &CalculatorTool{}

	// Create math assistant
	assistant, err := agent.New("math-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Mathematical assistant with calculator tool").
		WithInstructions("You are a helpful math assistant. Use the calculator tool to perform calculations and provide clear explanations.").
		WithTools(calcTool).
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create assistant: %v", err)
	}

	ctx := context.Background()

	// Demo calculations
	calculations := []string{
		"Calculate 15 + 27",
		"What is the square root of 64?",
		"Divide 144 by 12",
	}

	for i, calc := range calculations {
		fmt.Printf("\n🔄 Calculation %d\n", i+1)
		fmt.Printf("👤 User: %s\n", calc)

		response, err := assistant.Chat(ctx, calc,
			agent.WithSession("calculator-demo"))
		if err != nil {
			log.Printf("❌ Error: %v", err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", response.Message)
	}

	fmt.Println("\n✅ Calculator Tool Example Complete!")
}