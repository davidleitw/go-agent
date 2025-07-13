package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/conditions"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🧪 AskAI and OrElse API Test")
	fmt.Println("Testing LLM-enhanced responses...")
	fmt.Println("==========================================")

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

	// Create a simple agent with Ask and AskAI examples
	testAgent, err := agent.New("ask-ai-tester").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Testing Ask and AskAI functionality").
		WithInstructions(`You are a helpful assistant testing different response mechanisms.
		
When you receive contextual prompts, respond appropriately based on the conversation context.`).

		// Simple Ask example - direct response
		When(conditions.Contains("hello")).
			Ask("Hello! I'm here to help you test the Ask and AskAI functionality.").
			Build().

		// AskAI example - LLM-enhanced response with context
		When(conditions.Contains("help")).
			AskAI("Based on the conversation context, provide a helpful and contextual response about what assistance you can offer. Consider the user's previous messages and current situation.").
			OrElse("I'm here to help! What would you like assistance with?").
			Build().

		// Another AskAI example for testing context awareness
		When(conditions.Count(3)).
			AskAI("We've been chatting for a while. Based on our conversation history, provide a summary of what we've discussed and suggest what we could explore next.").
			OrElse("We've been having a good conversation! Is there anything specific you'd like to continue discussing?").
			Build().

		Build()

	if err != nil {
		log.Fatalf("❌ Failed to create test agent: %v", err)
	}

	log.Printf("✅ Test agent created successfully")

	// Test scenarios
	testScenarios := []struct {
		input       string
		description string
	}{
		{
			input:       "hello there",
			description: "Should trigger Ask (direct response)",
		},
		{
			input:       "I need some help with this system",
			description: "Should trigger AskAI (LLM-enhanced response)",
		},
		{
			input:       "What can you do exactly?",
			description: "Should trigger message count AskAI after 3rd message",
		},
		{
			input:       "Tell me more about your capabilities",
			description: "Continue conversation to test context",
		},
	}

	sessionID := fmt.Sprintf("ask-ai-test-%d", time.Now().Unix())
	ctx := context.Background()

	fmt.Println("\n🚀 Testing Ask vs AskAI Behavior")
	fmt.Println("==========================================")

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Test %d/%d: %s\n", i+1, len(testScenarios), scenario.description)
		fmt.Printf("👤 User: %s\n", scenario.input)

		start := time.Now()

		response, err := testAgent.Chat(ctx, scenario.input, 
			agent.WithSession(sessionID))
		if err != nil {
			log.Printf("❌ ERROR[%d]: %v", i+1, err)
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		duration := time.Since(start)
		log.Printf("RESPONSE[%d]: Duration: %.3fs", i+1, duration.Seconds())

		fmt.Printf("🤖 Assistant: %s\n", response.Message)
		fmt.Printf("📈 Messages in session: %d\n", len(response.Session.Messages()))

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n==========================================")
	fmt.Println("✅ Ask and AskAI Test Completed!")
	fmt.Println("🎯 Expected behaviors:")
	fmt.Println("   • Ask: Direct text responses without LLM call")
	fmt.Println("   • AskAI: Context-aware LLM-enhanced responses")
	fmt.Println("   • OrElse: Fallback responses when AskAI fails")
	fmt.Println("   • Context consideration in AskAI responses")
	fmt.Println("==========================================")
}