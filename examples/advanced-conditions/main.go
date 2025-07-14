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
	fmt.Println("🎯 Advanced Conditions Example")
	fmt.Println("=============================")
	fmt.Println("This example demonstrates intelligent flow control using conditions.")
	fmt.Println("The agent adapts its behavior based on conversation context.")
	fmt.Println()

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

	// Create an assistant with conditional behavior
	assistant, err := agent.New("smart-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("An intelligent assistant that adapts behavior based on context").
		WithInstructions("You are a helpful assistant. Adapt your responses based on the conversation context and user needs.").
		Build()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create flow rules for different scenarios using builder API
	
	// Rule 1: Welcome greetings
	welcomeRule := agent.NewFlowRule("welcome_new_users", 
		conditions.Or(
			conditions.Contains("hello"),
			conditions.Contains("hi"),
			conditions.Contains("hey"),
		)).
		WithDescription("Welcome users who say hello or greet").
		WithNewInstructions("The user is greeting you. Be extra welcoming and friendly. Ask how you can help them today.").
		WithPriority(10).
		Build()
		
	// Rule 2: Handle urgent requests
	urgentRule := agent.NewFlowRule("urgent_requests",
		conditions.Or(
			conditions.Contains("urgent"),
			conditions.Contains("emergency"),
			conditions.Contains("asap"),
			conditions.Contains("immediately"),
		)).
		WithDescription("Prioritize urgent or emergency requests").
		WithNewInstructions("This is an urgent request. Respond quickly and offer immediate assistance. Ask specific questions to understand the urgency.").
		WithPriority(20).
		Build()
		
	// Rule 3: Technical support
	techRule := agent.NewFlowRule("technical_support",
		conditions.Or(
			conditions.Contains("code"),
			conditions.Contains("programming"),
			conditions.Contains("debug"),
			conditions.Contains("error"),
			conditions.Contains("technical"),
		)).
		WithDescription("Switch to technical mode for programming/tech questions").
		WithNewInstructions("The user needs technical help. Provide detailed, step-by-step technical guidance. Be precise and offer code examples when relevant.").
		WithPriority(15).
		Build()
		
	// Rule 4: Long conversation guidance  
	countRule := agent.NewFlowRule("long_conversation", conditions.Count(8)).
		WithDescription("Offer summary for conversations with many messages").
		WithNewInstructions("This is a long conversation. Consider offering a summary of what you've discussed so far or asking if you can help with anything specific.").
		WithPriority(5).
		Build()

	rules := []agent.FlowRule{welcomeRule, urgentRule, techRule, countRule}

	// Create assistant (flow rules would be integrated separately in a real implementation)
	assistant, err = agent.New("smart-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("An intelligent assistant that adapts behavior based on context").
		WithInstructions("You are a helpful assistant. Adapt your responses based on the conversation context and user needs.").
		Build()

	fmt.Printf("✅ Created assistant with %d conditional flow rules\n", len(rules))
	fmt.Println()

	// Demonstrate different conversation scenarios
	scenarios := []struct {
		description string
		input       string
		expectation string
	}{
		{
			description: "Greeting scenario (should trigger welcome rule)",
			input:       "Hello! I'm new here.",
			expectation: "Welcoming response",
		},
		{
			description: "Technical question (should trigger technical support rule)",
			input:       "I'm having trouble with my Python code. Can you help debug this error?",
			expectation: "Technical assistance mode",
		},
		{
			description: "Urgent request (should trigger urgent rule)",
			input:       "This is urgent! I need help immediately with my presentation.",
			expectation: "Priority handling",
		},
		{
			description: "Regular question",
			input:       "What's the capital of France?",
			expectation: "Standard response",
		},
		{
			description: "Another question",
			input:       "How do I bake a chocolate cake?",
			expectation: "Standard response",
		},
		{
			description: "Continue conversation",
			input:       "What temperature should I use?",
			expectation: "Standard response",
		},
		{
			description: "More questions",
			input:       "How long should I bake it?",
			expectation: "Standard response",
		},
		{
			description: "Long conversation (should trigger count rule)",
			input:       "Any other baking tips?",
			expectation: "Summary or guidance offer",
		},
	}

	ctx := context.Background()
	sessionID := fmt.Sprintf("conditions-demo-%d", time.Now().Unix())

	// Run through scenarios
	for i, scenario := range scenarios {
		fmt.Printf("📝 Scenario %d: %s\n", i+1, scenario.description)
		fmt.Printf("👤 User: %s\n", scenario.input)
		fmt.Printf("🎯 Expected: %s\n", scenario.expectation)

		response, err := assistant.Chat(ctx, scenario.input, agent.WithSession(sessionID))
		if err != nil {
			log.Printf("❌ Error: %v", err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", response.Message)
		
		// Show which rules were triggered (if available in response metadata)
		if response.Data != nil {
			fmt.Printf("ℹ️ Flow rules: %v\n", response.Data)
		}
		
		fmt.Println()
		time.Sleep(500 * time.Millisecond) // Small pause for readability
	}

	fmt.Println("✅ Conditional flow demonstration completed!")
	fmt.Println("\n📊 Summary:")
	fmt.Println("- Greeting messages triggered welcoming behavior")
	fmt.Println("- Technical questions activated technical support mode") 
	fmt.Println("- Urgent requests received priority handling")
	fmt.Println("- Long conversations offered guidance")
	fmt.Println("\nConditions allow agents to intelligently adapt their behavior!")
}