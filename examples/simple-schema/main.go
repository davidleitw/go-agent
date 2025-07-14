// Package main demonstrates basic schema-based information collection.
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
	fmt.Println("===================================")

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

	// Create assistant
	assistant, err := agent.New("schema-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Assistant that demonstrates schema-based information collection").
		WithInstructions("You are a helpful assistant that collects user information based on the provided schema.").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create assistant: %v", err)
	}

	ctx := context.Background()

	// Define schema for contact information
	contactSchema := []*schema.Field{
		schema.Define("name", "What's your name?"),
		schema.Define("email", "Please provide your email address"),
		schema.Define("phone", "Contact number").Optional(),
	}

	fmt.Printf("\n📋 Schema fields:\n")
	for _, field := range contactSchema {
		required := "required"
		if !field.Required() {
			required = "optional"
		}
		fmt.Printf("   - %s (%s)\n", field.Name(), required)
	}

	// Demonstrate schema collection
	userInput := "Hi, I need some help"
	fmt.Printf("\n👤 User: %s\n", userInput)

	response, err := assistant.Chat(ctx, userInput,
		agent.WithSession("schema-demo"),
		agent.WithSchema(contactSchema...),
	)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("🤖 Assistant: %s\n", response.Message)

	// Show collected data if available
	if response.Data != nil {
		fmt.Printf("📊 Collected Data: %v\n", response.Data)
	}

	fmt.Println("\n✅ Simple Schema Example Complete!")
}