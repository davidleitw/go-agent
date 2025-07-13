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

// WeatherTool simulates weather information retrieval
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
	return "get_weather"
}

func (t *WeatherTool) Description() string {
	return "Get current weather information for a specific location"
}

func (t *WeatherTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The city and country (e.g., 'Tokyo, Japan')",
			},
			"unit": map[string]interface{}{
				"type":        "string",
				"description": "Temperature unit",
				"enum":        []string{"celsius", "fahrenheit"},
				"default":     "celsius",
			},
		},
		"required": []string{"location"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	location, ok := args["location"].(string)
	if !ok {
		return nil, fmt.Errorf("location must be a string")
	}

	unit := "celsius"
	if u, exists := args["unit"].(string); exists {
		unit = u
	}

	// Simulate weather data
	weatherData := map[string]interface{}{
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

func (t *CalculatorTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The mathematical operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide", "power", "sqrt"},
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "First number (or only number for sqrt)",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "Second number (not needed for sqrt)",
			},
		},
		"required": []string{"operation", "a"},
	}
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	response := map[string]interface{}{
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

func (t *TimeTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "Timezone (e.g., 'UTC', 'America/New_York', 'Asia/Tokyo')",
				"default":     "UTC",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Time format",
				"enum":        []string{"iso", "human", "timestamp"},
				"default":     "human",
			},
		},
	}
}

func (t *TimeTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	response := map[string]interface{}{
		"timezone":     timezone,
		"format":       format,
		"time":         timeString,
		"unix_timestamp": now.Unix(),
		"iso_format":   now.Format(time.RFC3339),
		"day_of_week":  now.Weekday().String(),
		"day_of_year":  now.YearDay(),
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

func (t *NotificationTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The notification message to send",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Optional notification title",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Notification priority level",
				"enum":        []string{"low", "normal", "high", "urgent"},
				"default":     "normal",
			},
			"delay_minutes": map[string]interface{}{
				"type":        "number",
				"description": "Delay in minutes before sending (0 for immediate)",
				"default":     0,
			},
		},
		"required": []string{"message"},
	}
}

func (t *NotificationTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	response := map[string]interface{}{
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
	chatModel, err := openai.NewChatModel(apiKey, nil)
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

	// Create multi-tool agent
	log.Println("📝 Creating multi-tool AI assistant...")
	assistant, err := agent.New(
		agent.WithName("multi-tool-assistant"),
		agent.WithDescription("A versatile AI assistant with weather, calculator, time, and notification capabilities"),
		agent.WithInstructions(`You are a helpful assistant with access to multiple tools:

1. Weather Tool: Get weather information for any location
2. Calculator Tool: Perform mathematical calculations
3. Time Tool: Get current time in different timezones
4. Notification Tool: Send notifications and reminders

When users ask questions, determine which tool(s) to use based on their request. You can use multiple tools in sequence if needed. Always provide helpful and detailed responses based on the tool results.

Examples:
- "What's the weather in Tokyo?" → use get_weather
- "Calculate 15 * 23" → use calculate
- "What time is it in London?" → use get_time
- "Remind me about the meeting" → use send_notification
- "What's the weather in Paris and what time is it there?" → use both get_weather and get_time`),
		agent.WithChatModel(chatModel),
		agent.WithModel("gpt-4"),
		agent.WithModelSettings(&agent.ModelSettings{
			Temperature: floatPtr(0.7),
			MaxTokens:   intPtr(2000),
		}),
		agent.WithTools(tools...),
		agent.WithSessionStore(agent.NewInMemorySessionStore()),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create multi-tool assistant: %v", err)
	}

	log.Printf("✅ Multi-tool assistant '%s' created successfully", assistant.Name())

	// Test scenarios that require different tools
	testScenarios := []string{
		"What's the weather like in Tokyo, Japan?",
		"Calculate the square root of 144",
		"What time is it in New York?",
		"Send me a reminder notification about my dentist appointment tomorrow",
		"What's the weather in London and what time is it there?",
		"Calculate 25 * 4 and tell me the weather in San Francisco",
		"Get the current time in Tokyo and set a notification to remind me about the meeting in 30 minutes",
	}

	baseSessionID := fmt.Sprintf("multi-tool-%d", time.Now().Unix())
	log.Printf("🆔 Base Session ID: %s", baseSessionID)

	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🚀 Multi-Tool AI Assistant Demo")
	fmt.Println("Testing various tool combinations...")
	fmt.Println(strings.Repeat("=", 80))

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Test %d/%d\n", i+1, len(testScenarios))
		fmt.Printf("👤 User: %s\n", scenario)

		// Use separate session for each test to avoid cross-contamination
		sessionID := fmt.Sprintf("%s-test-%d", baseSessionID, i+1)

		log.Printf("REQUEST[%d]: Processing user input (Session: %s)", i+1, sessionID)
		start := time.Now()

		response, _, err := assistant.Chat(ctx, sessionID, scenario)
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
	fmt.Println("✅ Multi-tool assistant demo completed!")
	fmt.Println("The assistant successfully used different tools based on context.")
	fmt.Println(strings.Repeat("=", 80))
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }