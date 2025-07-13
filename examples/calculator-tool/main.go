package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/joho/godotenv"
)

// Simple calculation result
type CalculationResult struct {
	Expression string   `json:"expression"`
	Result     float64  `json:"result"`
	Steps      []string `json:"steps"`
}

// Create elegant function-based calculator tools
func createCalculatorTools() []agent.Tool {
	// Addition tool
	add := agent.NewTool("add", "Add two numbers", 
		func(a, b float64) CalculationResult {
			result := a + b
			expression := fmt.Sprintf("%.2f + %.2f", a, b)
			steps := []string{
				fmt.Sprintf("Adding %.2f + %.2f", a, b),
				fmt.Sprintf("Result: %.2f", result),
			}
			log.Printf("➕ ADD: %s = %.2f", expression, result)
			return CalculationResult{Expression: expression, Result: result, Steps: steps}
		})

	// Subtraction tool
	subtract := agent.NewTool("subtract", "Subtract two numbers",
		func(a, b float64) CalculationResult {
			result := a - b
			expression := fmt.Sprintf("%.2f - %.2f", a, b)
			steps := []string{
				fmt.Sprintf("Subtracting %.2f - %.2f", a, b),
				fmt.Sprintf("Result: %.2f", result),
			}
			log.Printf("➖ SUBTRACT: %s = %.2f", expression, result)
			return CalculationResult{Expression: expression, Result: result, Steps: steps}
		})

	// Multiplication tool
	multiply := agent.NewTool("multiply", "Multiply two numbers",
		func(a, b float64) CalculationResult {
			result := a * b
			expression := fmt.Sprintf("%.2f × %.2f", a, b)
			steps := []string{
				fmt.Sprintf("Multiplying %.2f × %.2f", a, b),
				fmt.Sprintf("Result: %.2f", result),
			}
			log.Printf("✖️ MULTIPLY: %s = %.2f", expression, result)
			return CalculationResult{Expression: expression, Result: result, Steps: steps}
		})

	// Division tool
	divide := agent.NewTool("divide", "Divide two numbers",
		func(a, b float64) (CalculationResult, error) {
			if b == 0 {
				return CalculationResult{}, fmt.Errorf("division by zero is not allowed")
			}
			result := a / b
			expression := fmt.Sprintf("%.2f ÷ %.2f", a, b)
			steps := []string{
				fmt.Sprintf("Dividing %.2f ÷ %.2f", a, b),
				fmt.Sprintf("Result: %.2f", result),
			}
			log.Printf("➗ DIVIDE: %s = %.2f", expression, result)
			return CalculationResult{Expression: expression, Result: result, Steps: steps}, nil
		})

	// Power tool
	power := agent.NewTool("power", "Calculate power (a^b)",
		func(base, exponent float64) CalculationResult {
			result := math.Pow(base, exponent)
			expression := fmt.Sprintf("%.2f ^ %.2f", base, exponent)
			steps := []string{
				fmt.Sprintf("Calculating %.2f raised to the power of %.2f", base, exponent),
				fmt.Sprintf("Result: %.2f", result),
			}
			log.Printf("🔄 POWER: %s = %.2f", expression, result)
			return CalculationResult{Expression: expression, Result: result, Steps: steps}
		})

	// Square root tool
	sqrt := agent.NewTool("sqrt", "Calculate square root",
		func(number float64) (CalculationResult, error) {
			if number < 0 {
				return CalculationResult{}, fmt.Errorf("square root of negative number is not supported")
			}
			result := math.Sqrt(number)
			expression := fmt.Sprintf("√%.2f", number)
			steps := []string{
				fmt.Sprintf("Calculating square root of %.2f", number),
				fmt.Sprintf("Result: %.2f", result),
			}
			log.Printf("√ SQRT: %s = %.2f", expression, result)
			return CalculationResult{Expression: expression, Result: result, Steps: steps}, nil
		})

	return []agent.Tool{add, subtract, multiply, divide, power, sqrt}
}

func main() {
	fmt.Println("🧮 Calculator Tool Example - Elegant Function-Based Tools")
	fmt.Println(strings.Repeat("=", 70))

	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
			log.Println("Make sure you have set OPENAI_API_KEY environment variable")
			log.Println("Or copy .env.example to .env and add your API key")
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	log.Printf("✅ OpenAI API key loaded")

	// Create elegant function-based calculator tools
	log.Println("🧠 Creating function-based calculator tools...")
	calculatorTools := createCalculatorTools()

	// Create the math assistant with elegant API
	log.Println("🤖 Creating math assistant with simplified API...")
	mathAssistant, err := agent.New("math-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("A mathematical assistant with function-based calculation tools").
		WithInstructions(`You are a helpful math assistant with access to elegant calculation tools.

Available tools:
- add: Add two numbers together
- subtract: Subtract one number from another
- multiply: Multiply two numbers
- divide: Divide one number by another (checks for zero division)
- power: Calculate exponential power (base^exponent)
- sqrt: Calculate square root (checks for negative numbers)

When users ask for calculations:
1. Choose the appropriate tool for the operation
2. Use the tool to perform the calculation
3. Explain the result clearly and show the steps
4. Be friendly and educational in your responses

Always provide clear explanations and show your work!`).
		WithTemperature(0.1). // Low temperature for precise calculations
		WithMaxTokens(1000).
		WithTools(calculatorTools...).
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create math assistant: %v", err)
	}

	log.Printf("✅ Math assistant created with %d elegant calculation tools", len(calculatorTools))

	// Test calculations
	testCalculations := []string{
		"Calculate 15 + 27",
		"What is 144 divided by 12?",
		"Find the square root of 64",
		"Calculate 2 to the power of 8",
		"What is 125 - 47?",
		"Multiply 13 by 7",
	}

	// Session will be automatically managed
	sessionID := fmt.Sprintf("calculator-%d", time.Now().Unix())
	log.Printf("🆔 Session ID: %s", sessionID)

	ctx := context.Background()

	fmt.Println("\n🧮 Starting calculator demonstrations with function-based tools...")
	fmt.Println(strings.Repeat("=", 70))

	for i, calculation := range testCalculations {
		fmt.Printf("\n🔄 Calculation %d/%d\n", i+1, len(testCalculations))
		fmt.Printf("👤 User: %s\n", calculation)

		// Log the request
		log.Printf("REQUEST[%d]: Processing calculation request", i+1)
		log.Printf("REQUEST[%d]: Input: %s", i+1, calculation)

		// Get agent response with simplified API
		startTime := time.Now()
		response, err := mathAssistant.Chat(ctx, calculation,
			agent.WithSession(sessionID))
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("❌ ERROR[%d]: Failed to get response: %v", i+1, err)
			continue
		}

		// Log response details
		log.Printf("RESPONSE[%d]: Duration: %v", i+1, duration)
		log.Printf("RESPONSE[%d]: Content length: %d characters", i+1, len(response.Message))
		if response.Metadata["tools_used"] != nil {
			log.Printf("RESPONSE[%d]: Tools used: %v", i+1, response.Metadata["tools_used"])
		}

		// Display the text response
		fmt.Printf("🤖 Assistant: %s\n", response.Message)

		// Show structured output if available (calculation results)
		if response.Data != nil {
			if calcResult, ok := response.Data.(CalculationResult); ok {
				fmt.Printf("📊 Calculation Details:\n")
				fmt.Printf("   • Expression: %s\n", calcResult.Expression)
				fmt.Printf("   • Result: %.2f\n", calcResult.Result)
				fmt.Printf("   • Steps: %v\n", calcResult.Steps)
				log.Printf("STRUCTURED[%d]: Calculation result: %s = %.2f", i+1, calcResult.Expression, calcResult.Result)
			}
		}

		// Check session state
		log.Printf("SESSION[%d]: Total messages in session: %d", i+1, len(response.Session.Messages()))

		// Add delay between calculations
		time.Sleep(2 * time.Second)
	}

	// Display final session summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("📊 Session Summary:\n")
	fmt.Printf("   • Session ID: %s\n", sessionID)
	fmt.Printf("   • Calculations performed: %d\n", len(testCalculations))
	fmt.Printf("   • Function-based tools available: %d\n", len(calculatorTools))

	log.Printf("SUMMARY: Session %s completed with %d calculations using elegant function-based tools", sessionID, len(testCalculations))

	fmt.Println("\n🎉 Calculator tool example finished!")
	fmt.Println("🎯 This example demonstrated:")
	fmt.Println("   • Elegant function-based tool definitions")
	fmt.Println("   • Automatic type inference and validation")
	fmt.Println("   • Clean separation of concerns (one function per operation)")
	fmt.Println("   • Zero boilerplate tool creation")
	fmt.Println("   • Built-in error handling and safety checks")
	fmt.Println("   • Type-safe function signatures")
}

// This example demonstrates the new elegant tool system:
// - Function-based tools with automatic type inference
// - Clean separation of concerns (one function per operation)
// - Automatic error handling and validation
// - Zero boilerplate tool definitions
// - Type-safe function signatures