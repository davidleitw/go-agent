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

// reservationSchema defines fields needed for restaurant reservation
func reservationSchema() []*schema.Field {
	return []*schema.Field{
		schema.Define("name", "Please provide your name"),
		schema.Define("phone", "Please provide your phone number"),
		schema.Define("date", "What date would you like to reserve?"),
		schema.Define("time", "What time would you prefer?"),
		schema.Define("party_size", "How many people will be dining?"),
	}
}

func main() {
	fmt.Println("🏪 Task Completion Example")
	fmt.Println("=========================")

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

	// Create reservation assistant
	assistant, err := agent.New("reservation-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Restaurant reservation assistant").
		WithInstructions("You are a helpful restaurant reservation assistant. Collect all required information for the reservation.").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create assistant: %v", err)
	}

	ctx := context.Background()
	sessionID := fmt.Sprintf("reservation-%d", time.Now().Unix())

	// Simulate multi-turn conversation
	conversations := []string{
		"I want to make a reservation for dinner",
		"My name is John Smith, phone is 555-1234",
		"Tomorrow at 7pm for 4 people",
	}

	schema := reservationSchema()
	fmt.Printf("📋 Collecting information for reservation:\n")
	for _, field := range schema {
		fmt.Printf("   - %s\n", field.Name())
	}

	for i, userInput := range conversations {
		fmt.Printf("\n🔄 Turn %d\n", i+1)
		fmt.Printf("👤 Customer: %s\n", userInput)

		start := time.Now()
		response, err := assistant.Chat(ctx, userInput,
			agent.WithSession(sessionID),
			agent.WithSchema(schema...),
		)
		if err != nil {
			log.Printf("❌ Error: %v", err)
			continue
		}

		duration := time.Since(start)
		fmt.Printf("🤖 Assistant: %s\n", response.Message)
		fmt.Printf("⏱️  Response time: %.3fs\n", duration.Seconds())

		// Show collected data if available
		if response.Data != nil {
			fmt.Printf("📊 Collected Data: %v\n", response.Data)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n✅ Task Completion Example Finished!")
}