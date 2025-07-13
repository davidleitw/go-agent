package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
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

	// Create an agent with elegant, simple configuration
	log.Println("🤖 Creating AI assistant with simplified API...")
	assistant, err := agent.New("helpful-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("A helpful AI assistant for general conversations").
		WithInstructions("You are a helpful, friendly AI assistant. Keep your responses concise and engaging. Always be polite and professional.").
		WithTemperature(0.7).
		WithMaxTokens(1000).
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create agent: %v", err)
	}

	log.Printf("✅ Agent created successfully with elegant API")

	// Session will be automatically managed
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

		// Get agent response with simplified API
		startTime := time.Now()
		response, err := assistant.Chat(ctx, conv.user, 
			agent.WithSession(sessionID))
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("❌ ERROR[%d]: Failed to get response: %v", i+1, err)
			continue
		}

		// Log the response details
		log.Printf("RESPONSE[%d]: Duration: %v", i+1, duration)
		log.Printf("RESPONSE[%d]: Content length: %d characters", i+1, len(response.Message))
		if response.Data != nil {
			log.Printf("RESPONSE[%d]: Structured output: %T", i+1, response.Data)
		}

		// Display response
		fmt.Printf("🤖 Assistant: %s\n", response.Message)

		// Log session state
		log.Printf("SESSION[%d]: Total messages: %d", i+1, len(response.Session.Messages()))

		// Add a small delay between requests
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Conversation completed successfully!")

	// Get final session state from last response
	fmt.Printf("📊 Session Summary:\n")
	fmt.Printf("   • Session ID: %s\n", sessionID)
	fmt.Printf("   • Conversations completed: %d\n", len(conversations))

	log.Printf("SUMMARY: Session %s completed with %d conversations", sessionID, len(conversations))

	fmt.Println("\n🎉 Basic chat example finished!")
}

// This example demonstrates the new simplified API:
// - One-line agent creation with fluent builder
// - Automatic session management
// - Clean, readable configuration
// - No boilerplate code