// Package main demonstrates basic schema-based information collection.
//
// This example shows how to use the schema package to define expected fields
// and automatically collect missing information from users through natural conversation.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/schema"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔧 Simple Schema Collection Example")
	fmt.Println("Demonstrating basic information collection using schema fields")
	fmt.Println("========================================================")

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

	// Create a simple agent with schema-based collection
	bot, err := agent.New("simple-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("A simple assistant that demonstrates schema-based information collection").
		WithInstructions("You are a helpful assistant that collects user information.").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create agent: %v", err)
	}

	log.Printf("✅ Simple assistant created successfully")

	ctx := context.Background()

	// Test scenarios demonstrating schema collection
	testScenarios := []struct {
		description string
		userInput   string
		schema      []*schema.Field
	}{
		{
			description: "Basic email collection",
			userInput:   "I need some help",
			schema: []*schema.Field{
				schema.Define("email", "Please provide your email address"),
			},
		},
		{
			description: "Multiple required fields",
			userInput:   "I want to report an issue",
			schema: []*schema.Field{
				schema.Define("email", "Please provide your email address"),
				schema.Define("issue", "Please describe your issue"),
			},
		},
		{
			description: "Mix of required and optional fields",
			userInput:   "I need assistance with my account",
			schema: []*schema.Field{
				schema.Define("email", "Please provide your email address"),
				schema.Define("issue", "Please describe the issue you're experiencing"),
				schema.Define("phone", "Contact number for urgent matters").Optional(),
			},
		},
		{
			description: "User provides some information initially",
			userInput:   "My email is user@example.com and I have a billing question",
			schema: []*schema.Field{
				schema.Define("email", "Please provide your email address"),
				schema.Define("question", "Please describe your billing question in detail"),
				schema.Define("account_type", "What type of account do you have?").Optional(),
			},
		},
	}

	// Run each test scenario
	for i, scenario := range testScenarios {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, scenario.description)
		fmt.Printf("👤 User: %s\n", scenario.userInput)

		// Show what fields we're collecting
		fmt.Printf("📋 Expected fields:\n")
		for _, field := range scenario.schema {
			required := "required"
			if !field.Required() {
				required = "optional"
			}
			fmt.Printf("   - %s (%s): %s\n", field.Name(), required, field.Prompt())
		}

		// Use a unique session for each test
		sessionID := fmt.Sprintf("simple-test-%d", i+1)

		// Execute with schema
		response, err := bot.Chat(ctx, scenario.userInput,
			agent.WithSession(sessionID),
			agent.WithSchema(scenario.schema...),
		)
		if err != nil {
			log.Printf("❌ Error in test %d: %v", i+1, err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", response.Message)

		// Show metadata if schema collection occurred
		if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
			if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
				fmt.Printf("📊 Missing fields: %v\n", missingFields)
			}
		}

		// Wait a moment between tests
		fmt.Println("----------------------------------------")
	}

	// Demonstrate conversation continuation
	fmt.Printf("\n🔄 Conversation Continuation Example\n")
	fmt.Printf("Showing how information is remembered across turns\n")
	fmt.Printf("=================================================\n")

	conversationSchema := []*schema.Field{
		schema.Define("name", "What's your name?"),
		schema.Define("email", "Please provide your email address"),
		schema.Define("topic", "What would you like to discuss?"),
	}

	sessionID := "conversation-demo"

	// First turn - partial information
	fmt.Printf("👤 User: Hi, my name is John\n")
	response1, err := bot.Chat(ctx, "Hi, my name is John",
		agent.WithSession(sessionID),
		agent.WithSchema(conversationSchema...),
	)
	if err != nil {
		log.Printf("❌ Error in conversation turn 1: %v", err)
	} else {
		fmt.Printf("🤖 Assistant: %s\n", response1.Message)
	}

	// Second turn - more information
	fmt.Printf("\n👤 User: My email is john@example.com and I want to talk about pricing\n")
	response2, err := bot.Chat(ctx, "My email is john@example.com and I want to talk about pricing",
		agent.WithSession(sessionID),
		agent.WithSchema(conversationSchema...),
	)
	if err != nil {
		log.Printf("❌ Error in conversation turn 2: %v", err)
	} else {
		fmt.Printf("🤖 Assistant: %s\n", response2.Message)
		fmt.Printf("📨 Session messages: %d\n", len(response2.Session.Messages()))
	}

	fmt.Printf("\n========================================================\n")
	fmt.Printf("✅ Simple Schema Collection Example Complete!\n")
	fmt.Printf("🎯 Key Features Demonstrated:\n")
	fmt.Printf("   • Basic field definition with schema.Define()\n")
	fmt.Printf("   • Required vs optional fields\n")
	fmt.Printf("   • Automatic information extraction\n")
	fmt.Printf("   • Natural conversation flow\n")
	fmt.Printf("   • Session-based information persistence\n")
	fmt.Printf("========================================================\n")
}