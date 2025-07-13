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

// ReservationStatus represents the structured output for reservation tracking
type ReservationStatus struct {
	MissingFields  []string          `json:"missing_fields"`
	CollectedInfo  map[string]string `json:"collected_info"`
	CompletionFlag bool              `json:"completion_flag"`
	Message        string            `json:"message"`
	NextStep       string            `json:"next_step,omitempty"`
}

func main() {
	fmt.Println("🏪 Task Completion Example - Restaurant Reservation")
	fmt.Println(strings.Repeat("=", 60))

	// Load environment variables from .env file
	// Try local .env first, then parent directory
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
			log.Println("Make sure you have set OPENAI_API_KEY environment variable")
			log.Println("Or copy .env.example to .env and add your API key")
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	log.Printf("✅ OpenAI API key loaded")

	// Create the reservation agent with structured output
	log.Println("📝 Creating reservation agent with structured output...")
	
	// Create OpenAI chat model
	log.Println("📝 Creating OpenAI chat model...")
	chatModel, err := agent.NewOpenAIChatModel(apiKey)
	if err != nil {
		log.Fatalf("❌ Failed to create OpenAI chat model: %v", err)
	}

	reservationAgent, err := agent.NewBasicAgent(agent.BasicAgentConfig{
		Name:        "reservation-assistant",
		Description: "A restaurant reservation assistant that collects required information",
		Instructions: `You are a restaurant reservation assistant. Your job is to collect the following required information for a reservation:
1. Customer name (name)
2. Phone number (phone)  
3. Date (date)
4. Time (time)
5. Number of people (party_size)

You must respond with JSON in the exact format specified. Track which fields are missing and which have been collected. Set completion_flag to true ONLY when ALL required fields are collected and valid.

Be friendly and professional. If information is missing or unclear, ask for clarification.

When all information is collected, confirm the reservation details.`,
		Model: "gpt-4o-mini",
		ModelSettings: &agent.ModelSettings{
			Temperature: floatPtr(0.3), // Lower temperature for more consistent structured output
			MaxTokens:   intPtr(800),
		},
		OutputType: agent.NewStructuredOutputType(&ReservationStatus{}),
		ChatModel:  chatModel,
	})
	if err != nil {
		log.Fatalf("❌ Failed to create reservation agent: %v", err)
	}

	log.Printf("✅ Reservation agent '%s' created successfully", reservationAgent.Name())

	// Simulate a user interaction sequence
	userInputs := []string{
		"I want to make a restaurant reservation, I'm Mr. Lee",      // Initial incomplete request
		"My phone is 0912345678, I want tomorrow evening at 7pm",   // Partial info
		"4 people",                                                  // Final missing piece
	}

	sessionID := fmt.Sprintf("reservation-%d", time.Now().Unix())
	session := agent.NewSession(sessionID)
	log.Printf("🆔 Session ID: %s", sessionID)

	ctx := context.Background()
	maxTurns := 5
	
	fmt.Println("\n💬 Starting reservation collection process...")
	fmt.Println(strings.Repeat("=", 60))

	for turn := 0; turn < len(userInputs) && turn < maxTurns; turn++ {
		userInput := userInputs[turn]
		
		fmt.Printf("\n🔄 Turn %d/%d\n", turn+1, len(userInputs))
		fmt.Printf("👤 User: %s\n", userInput)

		// Log the request
		log.Printf("REQUEST[%d]: Processing user input", turn+1)
		log.Printf("REQUEST[%d]: Input: %s", turn+1, userInput)
		log.Printf("REQUEST[%d]: Turn %d of maximum %d", turn+1, turn+1, maxTurns)

		// Get agent response with structured output
		startTime := time.Now()
		response, structuredOutput, err := reservationAgent.Chat(ctx, session, userInput)
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("❌ ERROR[%d]: Failed to get response: %v", turn+1, err)
			continue
		}

		// Log response details
		log.Printf("RESPONSE[%d]: Duration: %v", turn+1, duration)
		log.Printf("RESPONSE[%d]: Content length: %d characters", turn+1, len(response.Content))

		// Display the text response
		fmt.Printf("🤖 Assistant: %s\n", response.Content)

		// Process structured output
		if structuredOutput != nil {
			if reservationStatus, ok := structuredOutput.(*ReservationStatus); ok {
				log.Printf("STRUCTURED[%d]: Parsed reservation status successfully", turn+1)
				log.Printf("STRUCTURED[%d]: Missing fields: %v", turn+1, reservationStatus.MissingFields)
				log.Printf("STRUCTURED[%d]: Collected info: %v", turn+1, reservationStatus.CollectedInfo)
				log.Printf("STRUCTURED[%d]: Completion flag: %t", turn+1, reservationStatus.CompletionFlag)

				// Display structured information
				fmt.Println("\n📋 Reservation Status:")
				fmt.Printf("   • Missing fields: %s\n", strings.Join(reservationStatus.MissingFields, ", "))
				fmt.Printf("   • Collected info: %d items\n", len(reservationStatus.CollectedInfo))
				for key, value := range reservationStatus.CollectedInfo {
					fmt.Printf("     - %s: %s\n", key, value)
				}
				fmt.Printf("   • Status: %s\n", getStatusText(reservationStatus.CompletionFlag))
				if reservationStatus.NextStep != "" {
					fmt.Printf("   • Next step: %s\n", reservationStatus.NextStep)
				}

				// Check if task is completed
				if reservationStatus.CompletionFlag {
					log.Printf("COMPLETION[%d]: Task completed successfully!", turn+1)
					fmt.Println("\n🎉 Reservation completed successfully!")
					fmt.Printf("📞 Final reservation details:\n")
					for key, value := range reservationStatus.CollectedInfo {
						fmt.Printf("   • %s: %s\n", key, value)
					}
					break
				}

				// Log missing fields for debugging
				if len(reservationStatus.MissingFields) > 0 {
					log.Printf("PROGRESS[%d]: Still missing: %s", turn+1, strings.Join(reservationStatus.MissingFields, ", "))
				}
			} else {
				log.Printf("WARNING[%d]: Structured output type mismatch: %T", turn+1, structuredOutput)
			}
		} else {
			log.Printf("WARNING[%d]: No structured output received", turn+1)
		}

		// Check session state
		log.Printf("SESSION[%d]: Total messages in session: %d", turn+1, len(session.Messages()))

		// Add delay between turns
		time.Sleep(2 * time.Second)
	}

	// Handle case where max turns reached without completion
	if turn := len(userInputs); turn >= maxTurns {
		log.Printf("TIMEOUT: Reached maximum turns (%d) without completion", maxTurns)
		fmt.Printf("\n⏰ Maximum number of turns (%d) reached\n", maxTurns)
	}

	// Display final session summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("📊 Final Session Summary:\n")
	fmt.Printf("   • Session ID: %s\n", session.ID())
	fmt.Printf("   • Total messages: %d\n", len(session.Messages()))
	fmt.Printf("   • Created at: %s\n", session.CreatedAt().Format("2006-01-02 15:04:05"))
	fmt.Printf("   • Updated at: %s\n", session.UpdatedAt().Format("2006-01-02 15:04:05"))

	log.Printf("SUMMARY: Session %s completed with %d messages", session.ID(), len(session.Messages()))
		
	// Log all messages for debugging
	for i, msg := range session.Messages() {
		log.Printf("MESSAGE[%d]: Role=%s, Content length=%d", i+1, msg.Role, len(msg.Content))
	}

	fmt.Println("🎯 Task completion example finished!")
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }

func getStatusText(completed bool) string {
	if completed {
		return "✅ Complete"
	}
	return "⏳ In Progress"
}