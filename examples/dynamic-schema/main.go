// Package main demonstrates dynamic schema selection based on user keywords.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/schema"
	"github.com/joho/godotenv"
)

// getSchemaForKeywords selects appropriate schema based on simple keyword matching
func getSchemaForKeywords(input string) []*schema.Field {
	inputLower := strings.ToLower(input)
	
	if strings.Contains(inputLower, "technical") || strings.Contains(inputLower, "error") || strings.Contains(inputLower, "bug") {
		return []*schema.Field{
			schema.Define("email", "Please provide your email for follow-up"),
			schema.Define("error_description", "Please describe the error you're experiencing"),
			schema.Define("browser", "What browser are you using?").Optional(),
		}
	}
	
	if strings.Contains(inputLower, "billing") || strings.Contains(inputLower, "payment") || strings.Contains(inputLower, "charge") {
		return []*schema.Field{
			schema.Define("email", "Please provide your account email"),
			schema.Define("account_id", "What is your account ID?"),
			schema.Define("billing_question", "Please describe your billing question"),
		}
	}
	
	// Default schema for general inquiries
	return []*schema.Field{
		schema.Define("email", "Please provide your email address"),
		schema.Define("topic", "What would you like to know about?"),
	}
}

func main() {
	fmt.Println("🎯 Dynamic Schema Selection Example")
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

	// Create assistant with schema capabilities
	assistant, err := agent.New("schema-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Assistant that adapts information collection based on user input").
		WithInstructions("You are a helpful assistant. Collect information based on the provided schema.").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create assistant: %v", err)
	}

	ctx := context.Background()

	// Test scenarios for dynamic schema selection
	scenarios := []string{
		"I'm getting a technical error when logging in",
		"I have a billing question about my invoice", 
		"Hello, I have some general questions",
	}

	for i, userInput := range scenarios {
		sessionID := fmt.Sprintf("schema-demo-%d", i+1)
		
		fmt.Printf("\n📝 Scenario %d\n", i+1)
		fmt.Printf("👤 User: %s\n", userInput)

		// Get appropriate schema based on keywords
		selectedSchema := getSchemaForKeywords(userInput)
		fmt.Printf("📋 Selected Schema (%d fields):\n", len(selectedSchema))
		for _, field := range selectedSchema {
			required := "required"
			if !field.Required() {
				required = "optional"
			}
			fmt.Printf("   - %s (%s)\n", field.Name(), required)
		}

		// Execute with dynamically selected schema
		response, err := assistant.Chat(ctx, userInput,
			agent.WithSession(sessionID),
			agent.WithSchema(selectedSchema...),
		)
		if err != nil {
			log.Printf("❌ Error: %v", err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", response.Message)
		
		// Show collected data if available
		if response.Data != nil {
			fmt.Printf("📊 Collected Data: %v\n", response.Data)
		}
		
		fmt.Println("-----------------------------------")
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n✅ Dynamic Schema Selection Example Complete!")
}