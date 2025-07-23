package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/agent"
	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/llm/openai"
	"github.com/davidleitw/go-agent/tool"
)

// ========================================
// TOOLS IMPLEMENTATION
// ========================================

// WeatherTool simulates weather information lookup
type WeatherTool struct{}

func (w *WeatherTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "get_weather",
			Description: "Get current weather information for a specified city",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"city": {
						Type:        "string",
						Description: "The city name to get weather for",
					},
					"units": {
						Type:        "string",
						Description: "Temperature units: celsius or fahrenheit",
					},
				},
				Required: []string{"city"},
			},
		},
	}
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	log.Printf("🌤️  [WeatherTool] Starting weather lookup...")
	log.Printf("🌤️  [WeatherTool] Received params: %+v", params)

	city, ok := params["city"].(string)
	if !ok {
		log.Printf("❌ [WeatherTool] Error: city parameter missing or invalid")
		return nil, fmt.Errorf("city parameter is required and must be a string")
	}

	units, _ := params["units"].(string)
	if units == "" {
		units = "celsius"
		log.Printf("🌤️  [WeatherTool] No units specified, defaulting to celsius")
	}

	log.Printf("🌤️  [WeatherTool] Looking up weather for %s (units: %s)", city, units)

	// Simulate API delay
	time.Sleep(500 * time.Millisecond)

	// Mock weather data based on city
	var temp int
	var condition string
	
	switch strings.ToLower(city) {
	case "tokyo", "東京":
		temp = 22
		condition = "Sunny"
	case "london":
		temp = 15
		condition = "Cloudy"
	case "new york":
		temp = 18
		condition = "Partly cloudy"
	case "taipei", "台北":
		temp = 28
		condition = "Humid and warm"
	case "paris":
		temp = 20
		condition = "Light rain"
	default:
		temp = 25
		condition = "Clear"
	}

	// Convert temperature if needed
	if units == "fahrenheit" {
		temp = int(float64(temp)*9.0/5.0 + 32)
		log.Printf("🌤️  [WeatherTool] Converted temperature to Fahrenheit: %d°F", temp)
	}

	result := map[string]any{
		"city":        city,
		"temperature": temp,
		"condition":   condition,
		"units":       units,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
	}

	log.Printf("✅ [WeatherTool] Successfully retrieved weather: %+v", result)
	
	return result, nil
}

// CalculatorTool performs mathematical calculations
type CalculatorTool struct{}

func (c *CalculatorTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "calculate",
			Description: "Perform mathematical calculations including basic arithmetic and advanced functions",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"expression": {
						Type:        "string",
						Description: "Mathematical expression to evaluate (supports +, -, *, /, ^, sqrt, sin, cos, tan, log)",
					},
				},
				Required: []string{"expression"},
			},
		},
	}
}

func (c *CalculatorTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	log.Printf("🧮 [CalculatorTool] Starting calculation...")
	log.Printf("🧮 [CalculatorTool] Received params: %+v", params)

	expression, ok := params["expression"].(string)
	if !ok {
		log.Printf("❌ [CalculatorTool] Error: expression parameter missing or invalid")
		return nil, fmt.Errorf("expression parameter is required and must be a string")
	}

	log.Printf("🧮 [CalculatorTool] Evaluating expression: %s", expression)

	result, err := c.evaluateExpression(expression)
	if err != nil {
		log.Printf("❌ [CalculatorTool] Calculation failed: %v", err)
		return nil, fmt.Errorf("calculation error: %w", err)
	}

	response := map[string]any{
		"expression": expression,
		"result":     result,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
	}

	log.Printf("✅ [CalculatorTool] Calculation successful: %s = %g", expression, result)
	
	return response, nil
}

func (c *CalculatorTool) evaluateExpression(expr string) (float64, error) {
	// Simple expression evaluator for demo purposes
	expr = strings.ReplaceAll(expr, " ", "")
	
	// Handle special functions first
	if strings.Contains(expr, "sqrt(") {
		return c.handleFunction(expr, "sqrt")
	}
	if strings.Contains(expr, "sin(") {
		return c.handleFunction(expr, "sin")
	}
	if strings.Contains(expr, "cos(") {
		return c.handleFunction(expr, "cos")
	}
	
	// Handle complex expressions with multiple operations
	// For demo purposes, we'll evaluate step by step following order of operations
	return c.evaluateComplexExpression(expr)
}

// evaluateComplexExpression handles expressions with multiple operations
func (c *CalculatorTool) evaluateComplexExpression(expr string) (float64, error) {
	// Simple order of operations: multiplication/division first, then addition/subtraction
	
	// First handle multiplication and division from left to right
	for {
		found := false
		// Find first * or /
		for i, char := range expr {
			if char == '*' || char == '/' {
				// Extract left and right operands
				left, right, err := c.extractOperands(expr, i)
				if err != nil {
					return 0, err
				}
				
				var result float64
				if char == '*' {
					result = left * right
				} else {
					if right == 0 {
						return 0, fmt.Errorf("division by zero")
					}
					result = left / right
				}
				
				// Replace this operation with the result
				expr = c.replaceOperation(expr, i, result, left, right, string(char))
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	
	// Then handle addition and subtraction from left to right
	for {
		found := false
		// Find first + or - (but not at the beginning for negative numbers)
		for i := 1; i < len(expr); i++ {
			char := expr[i]
			if char == '+' || char == '-' {
				// Extract left and right operands
				left, right, err := c.extractOperands(expr, i)
				if err != nil {
					return 0, err
				}
				
				var result float64
				if char == '+' {
					result = left + right
				} else {
					result = left - right
				}
				
				// Replace this operation with the result
				expr = c.replaceOperation(expr, i, result, left, right, string(char))
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	
	// Final result should be a single number
	return strconv.ParseFloat(expr, 64)
}

// extractOperands extracts the left and right operands around an operator at position i
func (c *CalculatorTool) extractOperands(expr string, opPos int) (float64, float64, error) {
	// Find left operand (go backwards until we hit an operator or start)
	leftStart := 0
	for i := opPos - 1; i >= 0; i-- {
		if expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/' {
			leftStart = i + 1
			break
		}
	}
	
	// Find right operand (go forwards until we hit an operator or end)
	rightEnd := len(expr)
	for i := opPos + 1; i < len(expr); i++ {
		if expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/' {
			rightEnd = i
			break
		}
	}
	
	leftStr := expr[leftStart:opPos]
	rightStr := expr[opPos+1:rightEnd]
	
	left, err := strconv.ParseFloat(leftStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid left operand: %s", leftStr)
	}
	
	right, err := strconv.ParseFloat(rightStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid right operand: %s", rightStr)
	}
	
	return left, right, nil
}

// replaceOperation replaces an operation in the expression with its result
func (c *CalculatorTool) replaceOperation(expr string, opPos int, result, left, right float64, op string) string {
	// Find the bounds of the operation
	leftStart := 0
	for i := opPos - 1; i >= 0; i-- {
		if expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/' {
			leftStart = i + 1
			break
		}
	}
	
	rightEnd := len(expr)
	for i := opPos + 1; i < len(expr); i++ {
		if expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/' {
			rightEnd = i
			break
		}
	}
	
	// Replace the operation with the result
	before := expr[:leftStart]
	after := expr[rightEnd:]
	resultStr := strconv.FormatFloat(result, 'f', -1, 64)
	
	return before + resultStr + after
}

func (c *CalculatorTool) handleFunction(expr, funcName string) (float64, error) {
	start := strings.Index(expr, funcName+"(")
	if start == -1 {
		return 0, fmt.Errorf("function %s not found", funcName)
	}
	
	// Extract the argument (simplified - assumes single argument)
	argStart := start + len(funcName) + 1
	argEnd := strings.Index(expr[argStart:], ")")
	if argEnd == -1 {
		return 0, fmt.Errorf("missing closing parenthesis for %s", funcName)
	}
	
	argStr := expr[argStart : argStart+argEnd]
	arg, err := strconv.ParseFloat(argStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid argument for %s: %s", funcName, argStr)
	}
	
	switch funcName {
	case "sqrt":
		return math.Sqrt(arg), nil
	case "sin":
		return math.Sin(arg), nil
	case "cos":
		return math.Cos(arg), nil
	default:
		return 0, fmt.Errorf("unsupported function: %s", funcName)
	}
}

func (c *CalculatorTool) handleBinaryOperation(expr, op string) (float64, error) {
	parts := strings.Split(expr, op)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid expression: %s", expr)
	}
	
	left, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid left operand: %s", parts[0])
	}
	
	right, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid right operand: %s", parts[1])
	}
	
	switch op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	default:
		return 0, fmt.Errorf("unsupported operation: %s", op)
	}
}

// TimeTool provides time-related information
type TimeTool struct{}

func (t *TimeTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "get_time_info",
			Description: "Get current time information for different timezones or perform time calculations",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"action": {
						Type:        "string",
						Description: "Action to perform: current_time, timezone_convert, add_time",
					},
					"timezone": {
						Type:        "string",
						Description: "Timezone (e.g., UTC, Asia/Tokyo, America/New_York)",
					},
					"format": {
						Type:        "string", 
						Description: "Time format (default: 2006-01-02 15:04:05)",
					},
				},
				Required: []string{"action"},
			},
		},
	}
}

func (t *TimeTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	log.Printf("⏰ [TimeTool] Starting time operation...")
	log.Printf("⏰ [TimeTool] Received params: %+v", params)

	action, ok := params["action"].(string)
	if !ok {
		log.Printf("❌ [TimeTool] Error: action parameter missing or invalid")
		return nil, fmt.Errorf("action parameter is required")
	}

	timezone, _ := params["timezone"].(string)
	format, _ := params["format"].(string)
	if format == "" {
		format = "2006-01-02 15:04:05"
	}

	log.Printf("⏰ [TimeTool] Action: %s, Timezone: %s, Format: %s", action, timezone, format)

	var result map[string]any

	switch action {
	case "current_time":
		result = t.getCurrentTime(timezone, format)
	case "timezone_convert":
		result = t.getMultipleTimezones(format)
	default:
		log.Printf("❌ [TimeTool] Unsupported action: %s", action)
		return nil, fmt.Errorf("unsupported action: %s", action)
	}

	log.Printf("✅ [TimeTool] Time operation successful: %+v", result)
	return result, nil
}

func (t *TimeTool) getCurrentTime(timezone, format string) map[string]any {
	now := time.Now()
	
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			log.Printf("⚠️  [TimeTool] Invalid timezone %s, using local time", timezone)
		} else {
			now = now.In(loc)
			log.Printf("⏰ [TimeTool] Converted to timezone: %s", timezone)
		}
	}

	return map[string]any{
		"current_time": now.Format(format),
		"timezone":     now.Location().String(),
		"unix":         now.Unix(),
		"weekday":      now.Weekday().String(),
	}
}

func (t *TimeTool) getMultipleTimezones(format string) map[string]any {
	now := time.Now()
	timezones := []string{"UTC", "Asia/Tokyo", "Europe/London", "America/New_York", "Asia/Taipei"}
	
	times := make(map[string]string)
	for _, tz := range timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			continue
		}
		times[tz] = now.In(loc).Format(format)
	}
	
	return map[string]any{
		"times": times,
		"base_time": now.Format(format),
	}
}

// FileWriteTool writes content to files
type FileWriteTool struct{}

func (f *FileWriteTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "write_file",
			Description: "Write content to a file on the filesystem",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"filename": {
						Type:        "string",
						Description: "Name of the file to write (will be created in current directory)",
					},
					"content": {
						Type:        "string",
						Description: "Content to write to the file",
					},
					"append": {
						Type:        "boolean",
						Description: "Whether to append to existing file (default: false = overwrite)",
					},
				},
				Required: []string{"filename", "content"},
			},
		},
	}
}

func (f *FileWriteTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	log.Printf("📝 [FileWriteTool] Starting file write operation...")
	log.Printf("📝 [FileWriteTool] Received params: %+v", params)

	filename, ok := params["filename"].(string)
	if !ok {
		log.Printf("❌ [FileWriteTool] Error: filename parameter missing or invalid")
		return nil, fmt.Errorf("filename parameter is required")
	}

	content, ok := params["content"].(string)
	if !ok {
		log.Printf("❌ [FileWriteTool] Error: content parameter missing or invalid")
		return nil, fmt.Errorf("content parameter is required")
	}

	append, _ := params["append"].(bool)

	log.Printf("📝 [FileWriteTool] Writing to file: %s (append: %v)", filename, append)
	log.Printf("📝 [FileWriteTool] Content length: %d characters", len(content))

	var file *os.File
	var err error

	if append {
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		log.Printf("📝 [FileWriteTool] Opened file in append mode")
	} else {
		file, err = os.Create(filename)
		log.Printf("📝 [FileWriteTool] Created/opened file in write mode")
	}

	if err != nil {
		log.Printf("❌ [FileWriteTool] Failed to open file: %v", err)
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytesWritten, err := file.WriteString(content)
	if err != nil {
		log.Printf("❌ [FileWriteTool] Failed to write content: %v", err)
		return nil, fmt.Errorf("failed to write content: %w", err)
	}

	result := map[string]any{
		"filename":       filename,
		"bytes_written":  bytesWritten,
		"append_mode":    append,
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}

	log.Printf("✅ [FileWriteTool] Successfully wrote %d bytes to %s", bytesWritten, filename)
	
	return result, nil
}

// ========================================
// ENHANCED LOGGING FOR REAL LLM
// ========================================

// logLLMRequest logs the LLM request details
func logLLMRequest(request llm.Request, callCount int) {
	log.Printf("🤖 [OpenAI] === LLM CALL #%d ===", callCount)
	log.Printf("🤖 [OpenAI] Request contains %d messages", len(request.Messages))
	
	for i, msg := range request.Messages {
		log.Printf("🤖 [OpenAI] Message %d - Role: %s", i+1, msg.Role)
		if len(msg.Content) > 300 {
			log.Printf("🤖 [OpenAI] Content (first 300 chars): %s...", msg.Content[:300])
		} else {
			log.Printf("🤖 [OpenAI] Content: %s", msg.Content)
		}
		
		// Log tool call ID if present
		if msg.ToolCallID != "" {
			log.Printf("🤖 [OpenAI] Tool Call ID: %s", msg.ToolCallID)
		}
	}
	
	if len(request.Tools) > 0 {
		log.Printf("🤖 [OpenAI] Available tools: %d", len(request.Tools))
		for _, tool := range request.Tools {
			log.Printf("🤖 [OpenAI] - Tool: %s - %s", tool.Function.Name, tool.Function.Description)
		}
	}
	
	// Log request parameters
	if request.Temperature != nil {
		log.Printf("🤖 [OpenAI] Temperature: %.2f", *request.Temperature)
	}
	if request.MaxTokens != nil {
		log.Printf("🤖 [OpenAI] Max tokens: %d", *request.MaxTokens)
	}
}

// logLLMResponse logs the LLM response details
func logLLMResponse(response *llm.Response, callCount int) {
	log.Printf("🤖 [OpenAI] Response received for call #%d", callCount)
	log.Printf("🤖 [OpenAI] Content length: %d characters", len(response.Content))
	if len(response.Content) > 200 {
		log.Printf("🤖 [OpenAI] Content (first 200 chars): %s...", response.Content[:200])
	} else if response.Content != "" {
		log.Printf("🤖 [OpenAI] Content: %s", response.Content)
	}
	
	log.Printf("🤖 [OpenAI] Finish reason: %s", response.FinishReason)
	
	if len(response.ToolCalls) > 0 {
		log.Printf("🤖 [OpenAI] Tool calls requested: %d", len(response.ToolCalls))
		for i, tc := range response.ToolCalls {
			log.Printf("🤖 [OpenAI] Tool call %d: %s", i+1, tc.Function.Name)
			log.Printf("🤖 [OpenAI] Arguments: %s", tc.Function.Arguments)
		}
	}
	
	// Log usage stats
	log.Printf("🤖 [OpenAI] Usage - Prompt: %d, Completion: %d, Total: %d tokens",
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		response.Usage.TotalTokens)
	
	log.Printf("🤖 [OpenAI] === END LLM CALL #%d ===\n", callCount)
}

// LoggingModel wraps the OpenAI model to add detailed logging
type LoggingModel struct {
	client    llm.Model
	callCount int
}

func NewLoggingModel(client llm.Model) *LoggingModel {
	return &LoggingModel{
		client:    client,
		callCount: 0,
	}
}

func (m *LoggingModel) Complete(ctx context.Context, request llm.Request) (*llm.Response, error) {
	m.callCount++
	
	// Log the request
	logLLMRequest(request, m.callCount)
	
	// Call the real OpenAI client
	startTime := time.Now()
	response, err := m.client.Complete(ctx, request)
	duration := time.Since(startTime)
	
	if err != nil {
		log.Printf("❌ [OpenAI] Call #%d failed after %v: %v", m.callCount, duration, err)
		return nil, err
	}
	
	log.Printf("🤖 [OpenAI] Call #%d completed in %v", m.callCount, duration)
	
	// Log the response
	logLLMResponse(response, m.callCount)
	
	return response, nil
}

// ========================================
// MAIN APPLICATION
// ========================================

func main() {
	fmt.Println("🚀 Starting Multi-Tool Agent Demo")
	fmt.Println("================================")
	fmt.Println()

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[DEMO] ")
	
	log.Printf("🔧 Initializing agent components...")

	// Create tools
	weatherTool := &WeatherTool{}
	calculatorTool := &CalculatorTool{}
	timeTool := &TimeTool{}
	fileWriteTool := &FileWriteTool{}
	
	log.Printf("✅ Created 4 tools: Weather, Calculator, Time, FileWrite")

	// Create context providers
	systemProvider := agentcontext.NewSystemPromptProvider(
		"You are a helpful AI assistant with access to multiple tools. " +
		"Use the appropriate tools to help users with weather, calculations, time information, and file operations. " +
		"Always explain what you're doing when using tools, and provide helpful, friendly responses.",
	)
	historyProvider := agentcontext.NewHistoryProvider(10)
	
	log.Printf("✅ Created context providers: System + History")

	// Create OpenAI model with API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalf("❌ OPENAI_API_KEY environment variable is required")
	}
	
	openaiClient := openai.New(llm.Config{
		APIKey: apiKey,
		Model:  "gpt-3.5-turbo", // Default to GPT-3.5-turbo
	})
	model := NewLoggingModel(openaiClient)
	log.Printf("✅ Created OpenAI client with logging wrapper")

	// Build agent
	log.Printf("🔨 Building agent with tools and context providers...")
	
	myAgent, err := agent.NewBuilder().
		WithLLM(model).
		WithTools(weatherTool, calculatorTool, timeTool, fileWriteTool).
		WithContextProviders(systemProvider, historyProvider).
		WithMaxIterations(5).
		Build()

	if err != nil {
		log.Fatalf("❌ Failed to create agent: %v", err)
	}

	log.Printf("✅ Agent created successfully!")
	log.Printf("📊 Agent configuration:")
	log.Printf("   - Max iterations: 5")
	log.Printf("   - Tools available: 4")
	log.Printf("   - Context providers: 2")
	fmt.Println()

	// Interactive demo
	fmt.Println("🎯 Multi-Tool Agent is ready!")
	fmt.Println("Try asking about:")
	fmt.Println("  • Weather: 'What's the weather in Tokyo?'")
	fmt.Println("  • Math: 'Calculate 15 + 25' or 'What is sqrt(16)?'")
	fmt.Println("  • Time: 'What time is it?' or 'Show me different timezones'")
	fmt.Println("  • Files: 'Write a file with some content'")
	fmt.Println("  • Type 'quit' to exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("💬 You: ")
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		if strings.ToLower(input) == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		log.Printf("👤 User input received: %s", input)
		log.Printf("🚀 Starting agent execution...")
		
		startTime := time.Now()
		
		// Execute agent
		response, err := myAgent.Execute(context.Background(), agent.Request{
			Input: input,
		})
		
		executionTime := time.Since(startTime)
		
		if err != nil {
			log.Printf("❌ Agent execution failed: %v", err)
			fmt.Printf("❌ Sorry, I encountered an error: %v\n\n", err)
			continue
		}

		log.Printf("✅ Agent execution completed in %v", executionTime)
		log.Printf("📊 Usage stats:")
		log.Printf("   - LLM tokens: %d (prompt: %d, completion: %d)", 
			response.Usage.LLMTokens.TotalTokens,
			response.Usage.LLMTokens.PromptTokens,
			response.Usage.LLMTokens.CompletionTokens)
		log.Printf("   - Tool calls: %d", response.Usage.ToolCalls)
		log.Printf("   - Session writes: %d", response.Usage.SessionWrites)
		
		if response.Metadata != nil {
			if iterations, ok := response.Metadata["total_iterations"]; ok {
				log.Printf("   - Total iterations: %v", iterations)
			}
			if toolsCalled, ok := response.Metadata["tools_called"]; ok {
				log.Printf("   - Tools called in execution: %v", toolsCalled)
			}
		}
		
		fmt.Printf("🤖 Agent: %s\n\n", response.Output)
		
		// Show session info
		if response.SessionID != "" {
			log.Printf("📝 Session ID: %s", response.SessionID)
			if response.Session != nil {
				history := response.Session.GetHistory(5) // Get last 5 entries
				log.Printf("📝 Session has %d history entries", len(history))
			}
		}
		fmt.Println("---")
	}

	log.Printf("🏁 Demo completed. Total LLM calls made: %d", model.callCount)
}