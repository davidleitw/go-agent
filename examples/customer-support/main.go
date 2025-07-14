// Package main demonstrates a customer support bot with schema-based information collection.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/schema"
	"github.com/joho/godotenv"
)

// supportSchema defines basic information required for support tickets
func supportSchema() []*schema.Field {
	return []*schema.Field{
		schema.Define("email", "Please provide your email address"),
		schema.Define("issue_type", "What type of issue? (technical/billing/general)"),
		schema.Define("description", "Please describe your issue"),
		schema.Define("urgency", "How urgent is this? (low/medium/high)").Optional(),
	}
}

func main() {
	fmt.Println("🎧 Customer Support Bot Example")
	fmt.Println("===============================")

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

	// Create customer support agent
	supportBot, err := agent.New("support-bot").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Customer support assistant").
		WithInstructions("You are a helpful customer support agent. Collect information to help resolve issues.").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create support bot: %v", err)
	}

	ctx := context.Background()

	// Demo customer support interaction
	userInput := "I'm having trouble with my account"
	fmt.Printf("\n👤 Customer: %s\n", userInput)

	// Show schema fields
	schema := supportSchema()
	fmt.Printf("📋 Information to collect:\n")
	for _, field := range schema {
		required := "required"
		if !field.Required() {
			required = "optional"
		}
		fmt.Printf("   - %s (%s)\n", field.Name(), required)
	}

	start := time.Now()
	response, err := supportBot.Chat(ctx, userInput,
		agent.WithSession("support-demo"),
		agent.WithSchema(schema...),
	)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("🤖 Support Agent: %s\n", response.Message)
	fmt.Printf("⏱️  Response time: %.3fs\n", duration.Seconds())

	// Show collected data if available
	if response.Data != nil {
		fmt.Printf("📊 Collected Data: %v\n", response.Data)
	}

	fmt.Println("\n✅ Customer Support Example Complete!")
}