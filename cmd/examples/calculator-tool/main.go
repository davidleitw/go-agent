package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/openai"
	"github.com/joho/godotenv"
)

// CalculationResult represents the structured output for mathematical calculations
type CalculationResult struct {
	Expression    string    `json:"expression"`
	Result        float64   `json:"result"`
	Steps         []string  `json:"steps"`
	OperationType string    `json:"operation_type"`
	Timestamp     time.Time `json:"timestamp"`
}

// CalculatorTool implements the agent.Tool interface for mathematical operations
type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "Perform mathematical calculations including basic arithmetic, powers, and square roots"
}

func (t *CalculatorTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The mathematical operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide", "power", "sqrt"},
			},
			"operand1": map[string]interface{}{
				"type":        "number",
				"description": "The first number",
			},
			"operand2": map[string]interface{}{
				"type":        "number",
				"description": "The second number (not required for sqrt)",
			},
		},
		"required": []string{"operation", "operand1"},
	}
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	log.Printf("TOOL: Calculator tool execution started")
	log.Printf("TOOL: Input arguments: %v", args)

	operation, ok := args["operation"].(string)
	if !ok {
		log.Printf("TOOL: ERROR - operation is not a string")
		return nil, fmt.Errorf("operation must be a string")
	}

	operand1, ok := args["operand1"].(float64)
	if !ok {
		// Try to convert from interface{} to float64
		if val, err := convertToFloat64(args["operand1"]); err == nil {
			operand1 = val
		} else {
			log.Printf("TOOL: ERROR - operand1 conversion failed: %v", err)
			return nil, fmt.Errorf("operand1 must be a number")
		}
	}

	var operand2 float64
	var hasOperand2 bool
	if op2, exists := args["operand2"]; exists && op2 != nil {
		if val, ok := op2.(float64); ok {
			operand2 = val
			hasOperand2 = true
		} else if val, err := convertToFloat64(op2); err == nil {
			operand2 = val
			hasOperand2 = true
		}
	}

	log.Printf("TOOL: Operation: %s, Operand1: %f, Operand2: %f (has: %t)", operation, operand1, operand2, hasOperand2)

	var result float64
	var steps []string
	var expression string

	switch operation {
	case "add":
		if !hasOperand2 {
			return nil, fmt.Errorf("addition requires two operands")
		}
		result = operand1 + operand2
		expression = fmt.Sprintf("%.2f + %.2f", operand1, operand2)
		steps = []string{
			fmt.Sprintf("Addition: %.2f + %.2f", operand1, operand2),
			fmt.Sprintf("Result: %.2f", result),
		}

	case "subtract":
		if !hasOperand2 {
			return nil, fmt.Errorf("subtraction requires two operands")
		}
		result = operand1 - operand2
		expression = fmt.Sprintf("%.2f - %.2f", operand1, operand2)
		steps = []string{
			fmt.Sprintf("Subtraction: %.2f - %.2f", operand1, operand2),
			fmt.Sprintf("Result: %.2f", result),
		}

	case "multiply":
		if !hasOperand2 {
			return nil, fmt.Errorf("multiplication requires two operands")
		}
		result = operand1 * operand2
		expression = fmt.Sprintf("%.2f × %.2f", operand1, operand2)
		steps = []string{
			fmt.Sprintf("Multiplication: %.2f × %.2f", operand1, operand2),
			fmt.Sprintf("Result: %.2f", result),
		}

	case "divide":
		if !hasOperand2 {
			return nil, fmt.Errorf("division requires two operands")
		}
		if operand2 == 0 {
			return nil, fmt.Errorf("division by zero is not allowed")
		}
		result = operand1 / operand2
		expression = fmt.Sprintf("%.2f ÷ %.2f", operand1, operand2)
		steps = []string{
			fmt.Sprintf("Division: %.2f ÷ %.2f", operand1, operand2),
			fmt.Sprintf("Result: %.2f", result),
		}

	case "power":
		if !hasOperand2 {
			return nil, fmt.Errorf("power operation requires two operands")
		}
		result = math.Pow(operand1, operand2)
		expression = fmt.Sprintf("%.2f ^ %.2f", operand1, operand2)
		steps = []string{
			fmt.Sprintf("Power: %.2f raised to %.2f", operand1, operand2),
			fmt.Sprintf("Result: %.2f", result),
		}

	case "sqrt":
		if operand1 < 0 {
			return nil, fmt.Errorf("square root of negative number is not supported")
		}
		result = math.Sqrt(operand1)
		expression = fmt.Sprintf("√%.2f", operand1)
		steps = []string{
			fmt.Sprintf("Square root of %.2f", operand1),
			fmt.Sprintf("Result: %.2f", result),
		}

	default:
		log.Printf("TOOL: ERROR - unsupported operation: %s", operation)
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	calculationResult := CalculationResult{
		Expression:    expression,
		Result:        result,
		Steps:         steps,
		OperationType: operation,
		Timestamp:     time.Now(),
	}

	log.Printf("TOOL: Calculation completed successfully")
	log.Printf("TOOL: Expression: %s", expression)
	log.Printf("TOOL: Result: %.2f", result)
	log.Printf("TOOL: Steps: %v", steps)

	return calculationResult, nil
}

func main() {
	fmt.Println("🧮 Calculator Tool Example")
	fmt.Println(strings.Repeat("=", 50))

	// Load environment variables from .env file
	// Try local .env first, then parent directory
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

	// Create calculator tool
	log.Println("🛠️  Creating calculator tool...")
	calculatorTool := &CalculatorTool{}

	// Create OpenAI chat model
	log.Println("📝 Creating OpenAI chat model...")
	chatModel, err := openai.NewChatModel(apiKey, nil)
	if err != nil {
		log.Fatalf("❌ Failed to create OpenAI chat model: %v", err)
	}

	// Create the math assistant with calculator tool
	log.Println("📝 Creating math assistant agent...")
	mathAssistant, err := agent.New(
		agent.WithName("math-assistant"),
		agent.WithDescription("A mathematical assistant that can perform calculations using the calculator tool"),
		agent.WithInstructions(`You are a helpful math assistant. You can perform mathematical calculations using the calculator tool.

When users ask for calculations:
1. Use the calculator tool to perform the computation
2. Explain the result clearly
3. Show the mathematical expression and steps

Available operations:
- add: Addition (requires two numbers)
- subtract: Subtraction (requires two numbers)  
- multiply: Multiplication (requires two numbers)
- divide: Division (requires two numbers, second cannot be zero)
- power: Exponentiation (requires base and exponent)
- sqrt: Square root (requires one positive number)

Always be helpful and provide clear explanations of the calculations.`),
		agent.WithChatModel(chatModel),
		agent.WithModel("gpt-4"),
		agent.WithModelSettings(&agent.ModelSettings{
			Temperature: floatPtr(0.1), // Low temperature for precise calculations
			MaxTokens:   intPtr(1000),
		}),
		agent.WithTools(calculatorTool),
		agent.WithSessionStore(agent.NewInMemorySessionStore()),
		agent.WithDebugLogging(),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create math assistant: %v", err)
	}

	log.Printf("✅ Math assistant '%s' created with calculator tool", mathAssistant.Name())

	// Test calculations
	testCalculations := []string{
		"Calculate 15 + 27",
		"What is 144 divided by 12?",
		"Find the square root of 64",
		"Calculate 2 to the power of 8",
		"What is 125 - 47?",
		"Multiply 13 by 7",
	}

	sessionID := fmt.Sprintf("calculator-%d", time.Now().Unix())
	log.Printf("🆔 Session ID: %s", sessionID)

	ctx := context.Background()

	fmt.Println("\n🧮 Starting calculator demonstrations...")
	fmt.Println(strings.Repeat("=", 50))

	for i, calculation := range testCalculations {
		fmt.Printf("\n🔄 Calculation %d/%d\n", i+1, len(testCalculations))
		fmt.Printf("👤 User: %s\n", calculation)

		// Log the request
		log.Printf("REQUEST[%d]: Processing calculation request", i+1)
		log.Printf("REQUEST[%d]: Input: %s", i+1, calculation)

		// Get agent response
		startTime := time.Now()
		response, structuredOutput, err := mathAssistant.Chat(ctx, sessionID, calculation)
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("❌ ERROR[%d]: Failed to get response: %v", i+1, err)
			continue
		}

		// Log response details
		log.Printf("RESPONSE[%d]: Duration: %v", i+1, duration)
		log.Printf("RESPONSE[%d]: Content length: %d characters", i+1, len(response.Content))
		log.Printf("RESPONSE[%d]: Tool calls: %d", i+1, len(response.ToolCalls))

		// Display the text response
		fmt.Printf("🤖 Assistant: %s\n", response.Content)

		// Show tool calls if any
		if len(response.ToolCalls) > 0 {
			fmt.Println("\n🔧 Tool Calls:")
			for j, toolCall := range response.ToolCalls {
				log.Printf("TOOLCALL[%d.%d]: Function: %s", i+1, j+1, toolCall.Function.Name)
				log.Printf("TOOLCALL[%d.%d]: Arguments: %s", i+1, j+1, toolCall.Function.Arguments)
				
				fmt.Printf("   • Tool: %s\n", toolCall.Function.Name)
				fmt.Printf("   • Arguments: %s\n", toolCall.Function.Arguments)
			}
		}

		// Process structured output if available
		if structuredOutput != nil {
			log.Printf("STRUCTURED[%d]: Received structured output: %T", i+1, structuredOutput)
			fmt.Printf("📊 Structured output available: %T\n", structuredOutput)
		}

		// Check session state
		session, err := mathAssistant.GetSession(ctx, sessionID)
		if err == nil {
			log.Printf("SESSION[%d]: Total messages: %d", i+1, len(session.Messages()))
		}

		// Add delay between calculations
		time.Sleep(2 * time.Second)
	}

	// Display final session summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	session, err := mathAssistant.GetSession(ctx, sessionID)
	if err == nil {
		fmt.Printf("📊 Session Summary:\n")
		fmt.Printf("   • Session ID: %s\n", session.ID())
		fmt.Printf("   • Total messages: %d\n", len(session.Messages()))
		fmt.Printf("   • Calculations performed: %d\n", len(testCalculations))
		fmt.Printf("   • Created at: %s\n", session.CreatedAt().Format("2006-01-02 15:04:05"))
		fmt.Printf("   • Updated at: %s\n", session.UpdatedAt().Format("2006-01-02 15:04:05"))

		log.Printf("SUMMARY: Session %s completed with %d calculations", session.ID(), len(testCalculations))

		// Count tool calls in session
		toolCallCount := 0
		for _, msg := range session.Messages() {
			toolCallCount += len(msg.ToolCalls)
		}
		fmt.Printf("   • Total tool calls: %d\n", toolCallCount)
		log.Printf("SUMMARY: Total tool calls executed: %d", toolCallCount)
	}

	fmt.Println("\n🎉 Calculator tool example finished!")
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }

func convertToFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}