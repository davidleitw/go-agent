package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/joho/godotenv"
)

// MultiToolAgent implements a custom Agent with advanced multi-tool coordination
type MultiToolAgent struct {
	name         string
	description  string
	instructions string
	model        string
	settings     *agent.ModelSettings
	tools        []agent.Tool
	chatModel    agent.ChatModel
	sessionStore agent.SessionStore
	maxTurns     int
	toolTimeout  time.Duration

	// Custom fields for multi-tool coordination
	toolUsageStats map[string]int
	lastToolUsed   string
}

// NewMultiToolAgent creates a new multi-tool agent with custom coordination logic
func NewMultiToolAgent(config MultiToolAgentConfig) *MultiToolAgent {
	return &MultiToolAgent{
		name:           config.Name,
		description:    config.Description,
		instructions:   config.Instructions,
		model:          config.Model,
		settings:       config.Settings,
		tools:          config.Tools,
		chatModel:      config.ChatModel,
		sessionStore:   config.SessionStore,
		maxTurns:       config.MaxTurns,
		toolTimeout:    config.ToolTimeout,
		toolUsageStats: make(map[string]int),
	}
}

type MultiToolAgentConfig struct {
	Name         string
	Description  string
	Instructions string
	Model        string
	Settings     *agent.ModelSettings
	Tools        []agent.Tool
	ChatModel    agent.ChatModel
	SessionStore agent.SessionStore
	MaxTurns     int
	ToolTimeout  time.Duration
}

// Agent interface implementation
func (a *MultiToolAgent) Name() string {
	return a.name
}

func (a *MultiToolAgent) Description() string {
	return a.description
}

func (a *MultiToolAgent) GetOutputType() agent.OutputType {
	return nil // No structured output for this agent
}

func (a *MultiToolAgent) GetTools() []agent.Tool {
	return a.tools
}

func (a *MultiToolAgent) GetFlowRules() []agent.FlowRule {
	return nil // No flow rules for this agent
}

// Chat implements sophisticated multi-tool conversation logic
func (a *MultiToolAgent) Chat(ctx context.Context, session agent.Session, userInput string) (*agent.Message, any, error) {
	// Add user message to session
	userMessage := agent.NewUserMessage(userInput)
	session.AddMessage(userMessage)

	// Prepare enhanced instructions with tool usage context
	enhancedInstructions := a.buildEnhancedInstructions()

	// Prepare messages for chat model
	messages := session.Messages()
	systemMessage := agent.NewSystemMessage(enhancedInstructions)
	allMessages := []agent.Message{systemMessage}
	allMessages = append(allMessages, messages...)

	// Execute conversation loop with advanced tool coordination
	for turn := 0; turn < a.maxTurns; turn++ {
		log.Printf("🔄 TURN[%d]: Starting conversation turn", turn+1)

		// Call chat model with current context
		response, err := a.chatModel.GenerateChatCompletion(
			ctx,
			allMessages,
			a.model,
			a.settings,
			a.tools,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate chat completion: %w", err)
		}

		// Add response to session
		session.AddMessage(*response)

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			log.Printf("✅ TURN[%d]: Conversation completed without tool calls", turn+1)
			// Save session
			if err := a.sessionStore.Save(ctx, session); err != nil {
				log.Printf("⚠️  Warning: failed to save session: %v", err)
			}
			return response, nil, nil
		}

		// Execute tool calls with coordination logic
		if err := a.executeToolCalls(ctx, session, response.ToolCalls); err != nil {
			return nil, nil, fmt.Errorf("failed to execute tool calls: %w", err)
		}

		// Update messages for next iteration
		allMessages = []agent.Message{systemMessage}
		allMessages = append(allMessages, session.Messages()...)

		log.Printf("📊 TURN[%d]: Tool usage stats: %v", turn+1, a.toolUsageStats)
	}

	// If we've exhausted max turns, return the last response
	messages = session.Messages()
	if len(messages) > 0 {
		lastMessage := &messages[len(messages)-1]
		if lastMessage.Role == agent.RoleAssistant {
			return lastMessage, nil, fmt.Errorf("reached maximum turns (%d) without completion", a.maxTurns)
		}
	}

	return nil, nil, fmt.Errorf("conversation ended unexpectedly after %d turns", a.maxTurns)
}

// buildEnhancedInstructions creates context-aware instructions
func (a *MultiToolAgent) buildEnhancedInstructions() string {
	base := a.instructions

	// Add tool usage context
	if len(a.toolUsageStats) > 0 {
		base += "\n\nTool Usage Context:\n"
		for toolName, count := range a.toolUsageStats {
			base += fmt.Sprintf("- %s: used %d times\n", toolName, count)
		}
	}

	// Add last tool used context
	if a.lastToolUsed != "" {
		base += fmt.Sprintf("\nLast tool used: %s\n", a.lastToolUsed)
	}

	// Add coordination hints
	base += `
Advanced Tool Coordination Guidelines:
1. Analyze user requests to determine optimal tool combinations
2. Use multiple tools in sequence when beneficial
3. Provide comprehensive responses that integrate tool results
4. Consider tool usage patterns for better coordination
5. Explain your tool selection reasoning when helpful
`

	return base
}

// executeToolCalls handles tool execution with coordination logic
func (a *MultiToolAgent) executeToolCalls(ctx context.Context, session agent.Session, toolCalls []agent.ToolCall) error {
	log.Printf("🔧 Executing %d tool calls", len(toolCalls))

	for i, toolCall := range toolCalls {
		// Find matching tool
		var matchingTool agent.Tool
		for _, tool := range a.tools {
			if tool.Name() == toolCall.Function.Name {
				matchingTool = tool
				break
			}
		}

		if matchingTool == nil {
			log.Printf("❌ Tool '%s' not found", toolCall.Function.Name)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: tool '%s' not found", toolCall.Function.Name),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Parse tool arguments
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Printf("❌ Invalid arguments for tool '%s': %v", toolCall.Function.Name, err)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: invalid arguments - %v", err),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Execute tool with timeout
		toolCtx, cancel := context.WithTimeout(ctx, a.toolTimeout)

		log.Printf("🔧 TOOL[%d/%d]: Executing %s", i+1, len(toolCalls), toolCall.Function.Name)
		start := time.Now()

		result, err := matchingTool.Execute(toolCtx, args)
		cancel()

		duration := time.Since(start)

		if err != nil {
			log.Printf("❌ TOOL[%d/%d]: %s failed in %v: %v", i+1, len(toolCalls), toolCall.Function.Name, duration, err)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: %v", err),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Convert result to JSON string
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Printf("❌ TOOL[%d/%d]: Failed to serialize result for %s: %v", i+1, len(toolCalls), toolCall.Function.Name, err)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: failed to serialize result - %v", err),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Add successful tool result
		toolMsg := agent.NewToolMessage(
			toolCall.ID,
			toolCall.Function.Name,
			string(resultJSON),
		)
		session.AddMessage(toolMsg)

		// Update coordination stats
		a.toolUsageStats[toolCall.Function.Name]++
		a.lastToolUsed = toolCall.Function.Name

		log.Printf("✅ TOOL[%d/%d]: %s completed in %v", i+1, len(toolCalls), toolCall.Function.Name, duration)
	}

	return nil
}

// WeatherTool simulates weather information retrieval
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
	return "get_weather"
}

func (t *WeatherTool) Description() string {
	return "Get current weather information for a specific location"
}

func (t *WeatherTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "The city and country (e.g., 'Tokyo, Japan')",
			},
			"unit": map[string]any{
				"type":        "string",
				"description": "Temperature unit",
				"enum":        []string{"celsius", "fahrenheit"},
				"default":     "celsius",
			},
		},
		"required": []string{"location"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	location, ok := args["location"].(string)
	if !ok {
		return nil, fmt.Errorf("location must be a string")
	}

	unit := "celsius"
	if u, exists := args["unit"].(string); exists {
		unit = u
	}

	// Simulate weather data
	weatherData := map[string]any{
		"location":    location,
		"temperature": getSimulatedTemperature(location, unit),
		"condition":   getSimulatedCondition(location),
		"humidity":    fmt.Sprintf("%d%%", 45+len(location)%30),
		"wind_speed":  fmt.Sprintf("%.1f km/h", float64(5+len(location)%15)),
		"unit":        unit,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
	}

	log.Printf("🌤️  WEATHER[%s]: %s, %s", location, weatherData["temperature"], weatherData["condition"])
	return weatherData, nil
}

func getSimulatedTemperature(location, unit string) string {
	// Simple hash-based simulation for consistent results
	temp := 15 + (len(location)*7)%25
	if unit == "fahrenheit" {
		temp = temp*9/5 + 32
		return fmt.Sprintf("%d°F", temp)
	}
	return fmt.Sprintf("%d°C", temp)
}

func getSimulatedCondition(location string) string {
	conditions := []string{"Sunny", "Cloudy", "Partly Cloudy", "Light Rain", "Clear"}
	return conditions[len(location)%len(conditions)]
}

// CalculatorTool performs mathematical calculations
type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
	return "calculate"
}

func (t *CalculatorTool) Description() string {
	return "Perform mathematical calculations including basic arithmetic, power, and square root"
}

func (t *CalculatorTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"description": "The mathematical operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide", "power", "sqrt"},
			},
			"a": map[string]any{
				"type":        "number",
				"description": "First number (or only number for sqrt)",
			},
			"b": map[string]any{
				"type":        "number",
				"description": "Second number (not needed for sqrt)",
			},
		},
		"required": []string{"operation", "a"},
	}
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	a, ok := args["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("parameter 'a' must be a number")
	}

	var result float64
	var expression string

	switch operation {
	case "add":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("parameter 'b' is required for addition")
		}
		result = a + b
		expression = fmt.Sprintf("%.2f + %.2f", a, b)

	case "subtract":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("parameter 'b' is required for subtraction")
		}
		result = a - b
		expression = fmt.Sprintf("%.2f - %.2f", a, b)

	case "multiply":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("parameter 'b' is required for multiplication")
		}
		result = a * b
		expression = fmt.Sprintf("%.2f × %.2f", a, b)

	case "divide":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("parameter 'b' is required for division")
		}
		if b == 0 {
			return nil, fmt.Errorf("division by zero is not allowed")
		}
		result = a / b
		expression = fmt.Sprintf("%.2f ÷ %.2f", a, b)

	case "power":
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("parameter 'b' is required for power operation")
		}
		result = math.Pow(a, b)
		expression = fmt.Sprintf("%.2f ^ %.2f", a, b)

	case "sqrt":
		if a < 0 {
			return nil, fmt.Errorf("cannot calculate square root of negative number")
		}
		result = math.Sqrt(a)
		expression = fmt.Sprintf("√%.2f", a)

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	response := map[string]any{
		"operation":  operation,
		"expression": expression,
		"result":     result,
		"formatted":  fmt.Sprintf("%s = %.2f", expression, result),
	}

	log.Printf("🧮 CALC[%s]: %s", operation, response["formatted"])
	return response, nil
}

// TimeTool provides current time information
type TimeTool struct{}

func (t *TimeTool) Name() string {
	return "get_time"
}

func (t *TimeTool) Description() string {
	return "Get current time information in various formats and timezones"
}

func (t *TimeTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"timezone": map[string]any{
				"type":        "string",
				"description": "Timezone (e.g., 'UTC', 'America/New_York', 'Asia/Tokyo')",
				"default":     "UTC",
			},
			"format": map[string]any{
				"type":        "string",
				"description": "Time format",
				"enum":        []string{"iso", "human", "timestamp"},
				"default":     "human",
			},
		},
	}
}

func (t *TimeTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	timezone := "UTC"
	if tz, exists := args["timezone"].(string); exists {
		timezone = tz
	}

	format := "human"
	if f, exists := args["format"].(string); exists {
		format = f
	}

	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
		timezone = "UTC"
	}

	now := time.Now().In(loc)

	var timeString string
	switch format {
	case "iso":
		timeString = now.Format(time.RFC3339)
	case "timestamp":
		timeString = strconv.FormatInt(now.Unix(), 10)
	default: // human
		timeString = now.Format("Monday, January 2, 2006 at 3:04:05 PM")
	}

	response := map[string]any{
		"timezone":       timezone,
		"format":         format,
		"time":           timeString,
		"unix_timestamp": now.Unix(),
		"iso_format":     now.Format(time.RFC3339),
		"day_of_week":    now.Weekday().String(),
		"day_of_year":    now.YearDay(),
	}

	log.Printf("⏰ TIME[%s]: %s", timezone, timeString)
	return response, nil
}

// NotificationTool simulates sending notifications
type NotificationTool struct{}

func (t *NotificationTool) Name() string {
	return "send_notification"
}

func (t *NotificationTool) Description() string {
	return "Send a notification message with optional scheduling"
}

func (t *NotificationTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]any{
				"type":        "string",
				"description": "The notification message to send",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Optional notification title",
			},
			"priority": map[string]any{
				"type":        "string",
				"description": "Notification priority level",
				"enum":        []string{"low", "normal", "high", "urgent"},
				"default":     "normal",
			},
			"delay_minutes": map[string]any{
				"type":        "number",
				"description": "Delay in minutes before sending (0 for immediate)",
				"default":     0,
			},
		},
		"required": []string{"message"},
	}
}

func (t *NotificationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message must be a string")
	}

	title := "Notification"
	if t, exists := args["title"].(string); exists && t != "" {
		title = t
	}

	priority := "normal"
	if p, exists := args["priority"].(string); exists {
		priority = p
	}

	delayMinutes := 0.0
	if d, exists := args["delay_minutes"].(float64); exists {
		delayMinutes = d
	}

	// Generate notification ID
	notificationID := fmt.Sprintf("notif_%d", time.Now().Unix())

	// Calculate scheduled time
	scheduledTime := time.Now().Add(time.Duration(delayMinutes) * time.Minute)

	response := map[string]any{
		"notification_id": notificationID,
		"title":           title,
		"message":         message,
		"priority":        priority,
		"status":          "scheduled",
		"scheduled_time":  scheduledTime.Format(time.RFC3339),
		"delay_minutes":   delayMinutes,
	}

	if delayMinutes == 0 {
		response["status"] = "sent_immediately"
		log.Printf("📢 NOTIFICATION[%s]: %s - %s", priority, title, message)
	} else {
		log.Printf("⏲️  NOTIFICATION[scheduled]: %s - %s (in %.0f minutes)", title, message, delayMinutes)
	}

	return response, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Verify OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}
	log.Printf("✅ OpenAI API key loaded (length: %d)", len(apiKey))

	// Create OpenAI chat model
	log.Println("🔧 Creating OpenAI chat model...")
	chatModel, err := agent.NewOpenAIChatModel(apiKey)
	if err != nil {
		log.Fatalf("❌ Failed to create OpenAI chat model: %v", err)
	}

	// Create tools
	tools := []agent.Tool{
		&WeatherTool{},
		&CalculatorTool{},
		&TimeTool{},
		&NotificationTool{},
	}

	// Create custom multi-tool agent
	log.Println("🤖 Creating custom multi-tool AI assistant...")
	assistant := NewMultiToolAgent(MultiToolAgentConfig{
		Name:        "multi-tool-assistant",
		Description: "A versatile AI assistant with advanced multi-tool coordination capabilities",
		Instructions: `You are a sophisticated assistant with access to multiple tools and advanced coordination capabilities:

1. Weather Tool: Get weather information for any location
2. Calculator Tool: Perform mathematical calculations
3. Time Tool: Get current time in different timezones
4. Notification Tool: Send notifications and reminders

Advanced Capabilities:
- Intelligent tool selection based on context
- Multi-tool coordination for complex requests
- Tool usage pattern analysis
- Comprehensive response integration
- Adaptive conversation flow

When users ask questions, analyze their intent and determine the optimal tool combination. Use multiple tools in sequence when beneficial and provide comprehensive responses that integrate all tool results.`,
		Model: "gpt-4",
		Settings: &agent.ModelSettings{
			Temperature: floatPtr(0.7),
			MaxTokens:   intPtr(2000),
		},
		Tools:        tools,
		ChatModel:    chatModel,
		SessionStore: agent.NewInMemorySessionStore(),
		MaxTurns:     5,
		ToolTimeout:  30 * time.Second,
	})

	log.Printf("✅ Multi-tool assistant '%s' created successfully", assistant.Name())

	// Create base session for testing
	baseSessionID := fmt.Sprintf("multi-tool-test-%d", time.Now().Unix())
	log.Printf("🆔 Base Session ID: %s", baseSessionID)

	// Test scenarios that require different tools
	testScenarios := []string{
		"What's the weather like in Tokyo, Japan?",
		"Calculate the result of 25 * 4 + 10",
		"What time is it in London right now?",
		"Send me a notification to call mom at 3 PM",
		"What's the weather in Paris and what time is it there?",
		"Calculate 15% of 200 and tell me the current time in New York",
	}

	ctx := context.Background()

	fmt.Println("\n🚀 Testing custom multi-tool agent with different scenarios...")
	fmt.Println(strings.Repeat("=", 60))

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Test %d/%d\n", i+1, len(testScenarios))
		fmt.Printf("👤 User: %s\n", scenario)

		// Use separate session for each test to avoid cross-contamination
		sessionID := fmt.Sprintf("%s-test-%d", baseSessionID, i+1)
		session := agent.NewSession(sessionID)

		log.Printf("REQUEST[%d]: Processing user input (Session: %s)", i+1, sessionID)
		start := time.Now()

		response, _, err := assistant.Chat(ctx, session, scenario)
		if err != nil {
			log.Printf("❌ ERROR[%d]: %v", i+1, err)
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		duration := time.Since(start)
		log.Printf("RESPONSE[%d]: Duration: %.3fs", i+1, duration.Seconds())

		fmt.Printf("🤖 Assistant: %s\n", response.Content)

		// Add a small delay between requests
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ Custom multi-tool assistant demo completed!")
	fmt.Println("🎯 The assistant successfully demonstrated:")
	fmt.Println("   • Advanced multi-tool coordination")
	fmt.Println("   • Context-aware tool selection")
	fmt.Println("   • Tool usage pattern tracking")
	fmt.Println("   • Comprehensive response integration")
	fmt.Println("   • Adaptive conversation flow")
	fmt.Println(strings.Repeat("=", 80))
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }
