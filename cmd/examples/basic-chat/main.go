package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/openai"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🤖 Basic Chat Agent Example")
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

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	log.Printf("✅ OpenAI API key loaded (length: %d)", len(apiKey))

	// Create OpenAI chat model
	log.Println("📝 Creating OpenAI chat model...")
	chatModel, err := openai.NewChatModel(apiKey, nil)
	if err != nil {
		log.Fatalf("❌ Failed to create OpenAI chat model: %v", err)
	}

	// Create an agent with functional options
	log.Println("📝 Creating AI agent...")
	assistant, err := agent.New(
		agent.WithName("helpful-assistant"),
		agent.WithDescription("A helpful AI assistant for general conversations"),
		agent.WithInstructions("You are a helpful, friendly AI assistant. Keep your responses concise and engaging. Always be polite and professional."),
		agent.WithChatModel(chatModel),
		agent.WithModel("gpt-4"),
		agent.WithModelSettings(&agent.ModelSettings{
			Temperature: floatPtr(0.7),
			MaxTokens:   intPtr(1000),
		}),
		agent.WithSessionStore(agent.NewInMemorySessionStore()),
		agent.WithDebugLogging(),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create agent: %v", err)
	}

	log.Printf("✅ Agent '%s' created successfully", assistant.Name())
	log.Printf("📋 Model: %s", assistant.Model())
	log.Printf("📝 Description: %s", assistant.Description())

	// Create a session ID
	sessionID := fmt.Sprintf("basic-chat-%d", time.Now().Unix())
	log.Printf("🆔 Session ID: %s", sessionID)

	// Prepare conversation examples
	conversations := []struct {
		user     string
		expected string
	}{
		{
			user:     "Hello! How are you doing today?",
			expected: "greeting response",
		},
		{
			user:     "What's the weather like?",
			expected: "weather-related response (without actual data)",
		},
		{
			user:     "Can you help me write a simple Python function to add two numbers?",
			expected: "programming help",
		},
	}

	ctx := context.Background()

	fmt.Println("\n💬 Starting conversation...")
	fmt.Println(strings.Repeat("=", 50))

	// Run conversation examples
	for i, conv := range conversations {
		fmt.Printf("\n🔄 Turn %d/%d\n", i+1, len(conversations))
		fmt.Printf("👤 User: %s\n", conv.user)

		// Log the request
		log.Printf("REQUEST[%d]: Sending user input to agent", i+1)
		log.Printf("REQUEST[%d]: Input: %s", i+1, conv.user)

		// Get agent response
		startTime := time.Now()
		response, structuredOutput, err := assistant.Chat(ctx, sessionID, conv.user)
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("❌ ERROR[%d]: Failed to get response: %v", i+1, err)
			continue
		}

		// Log the response details
		log.Printf("RESPONSE[%d]: Duration: %v", i+1, duration)
		log.Printf("RESPONSE[%d]: Role: %s", i+1, response.Role)
		log.Printf("RESPONSE[%d]: Content length: %d characters", i+1, len(response.Content))
		if len(response.ToolCalls) > 0 {
			log.Printf("RESPONSE[%d]: Tool calls: %d", i+1, len(response.ToolCalls))
		}
		if structuredOutput != nil {
			log.Printf("RESPONSE[%d]: Structured output: %T", i+1, structuredOutput)
		}

		// Display response
		fmt.Printf("🤖 Assistant: %s\n", response.Content)

		// Log session state
		session, err := assistant.GetSession(ctx, sessionID)
		if err == nil {
			log.Printf("SESSION[%d]: Total messages: %d", i+1, len(session.Messages()))
		}

		// Add a small delay between requests
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Conversation completed successfully!")

	// Display session summary
	session, err := assistant.GetSession(ctx, sessionID)
	if err == nil {
		fmt.Printf("📊 Session Summary:\n")
		fmt.Printf("   • Session ID: %s\n", session.ID())
		fmt.Printf("   • Total messages: %d\n", len(session.Messages()))
		fmt.Printf("   • Created at: %s\n", session.CreatedAt().Format("2006-01-02 15:04:05"))
		fmt.Printf("   • Updated at: %s\n", session.UpdatedAt().Format("2006-01-02 15:04:05"))

		log.Printf("SUMMARY: Session %s completed with %d messages", session.ID(), len(session.Messages()))
	}

	fmt.Println("\n🎉 Basic chat example finished!")
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }