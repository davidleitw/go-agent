package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/conditions"
	"github.com/joho/godotenv"
)

// UserProfile represents structured output for user onboarding
type UserProfile struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone"`
	Preferences  []string `json:"preferences"`
	CompletedAt  string   `json:"completed_at,omitempty"`
	IsComplete   bool     `json:"is_complete"`
	MissingInfo  []string `json:"missing_info,omitempty"`
	StatusText   string   `json:"status_text"`
}

func (u *UserProfile) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
				"description": "User's full name",
			},
			"email": map[string]any{
				"type": "string",
				"description": "User's email address",
			},
			"phone": map[string]any{
				"type": "string", 
				"description": "User's phone number",
			},
			"preferences": map[string]any{
				"type": "array",
				"items": map[string]any{"type": "string"},
				"description": "User's preferences/interests",
			},
			"completed_at": map[string]any{
				"type": "string",
				"description": "Completion timestamp",
			},
			"is_complete": map[string]any{
				"type": "boolean",
				"description": "Whether profile is complete",
			},
			"missing_info": map[string]any{
				"type": "array",
				"items": map[string]any{"type": "string"},
				"description": "Missing required information",
			},
			"status_text": map[string]any{
				"type": "string",
				"description": "Current status description",
			},
		},
		"required": []string{"is_complete", "status_text"},
	}
}

func (u *UserProfile) NewInstance() any {
	return &UserProfile{}
}

func (u *UserProfile) Validate(instance any) error {
	_, ok := instance.(*UserProfile)
	if !ok {
		return fmt.Errorf("instance is not a UserProfile")
	}
	return nil
}

func main() {
	fmt.Println("🎯 Advanced Conditions Example")
	fmt.Println("Demonstrating elegant condition usage and flow control")
	fmt.Println(strings.Repeat("=", 60))

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	// Create tools for information collection
	collectTool := agent.NewTool("collect_info", 
		"Collect and validate user information",
		func(field, value string) map[string]any {
			// Simulate validation
			valid := len(value) > 0
			if field == "email" {
				valid = strings.Contains(value, "@")
			}
			if field == "phone" {
				valid = len(value) >= 10
			}
			
			result := map[string]any{
				"field":     field,
				"value":     value,
				"is_valid":  valid,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			
			if valid {
				log.Printf("✅ COLLECTED: %s = %s", field, value)
			} else {
				log.Printf("❌ INVALID: %s = %s", field, value)
			}
			
			return result
		})

	validateTool := agent.NewTool("validate_profile",
		"Check if user profile is complete",
		func(userData map[string]any) map[string]any {
			required := []string{"name", "email", "phone", "preferences"}
			missing := []string{}
			
			for _, field := range required {
				if value, exists := userData[field]; !exists || value == "" {
					missing = append(missing, field)
				}
			}
			
			isComplete := len(missing) == 0
			status := "in_progress"
			if isComplete {
				status = "completed"
			}
			
			result := map[string]any{
				"is_complete":   isComplete,
				"missing_fields": missing,
				"status":        status,
				"completion_rate": float64(len(required)-len(missing))/float64(len(required)),
			}
			
			log.Printf("🔍 VALIDATION: Complete=%v, Missing=%v", isComplete, missing)
			return result
		})

	// Create agent with advanced condition-based flow rules
	onboardingAgent, err := agent.New("onboarding-specialist").
		WithOpenAI(apiKey).
		WithModel("gpt-4").
		WithDescription("An intelligent onboarding agent with sophisticated condition handling").
		WithInstructions(`You are an expert onboarding specialist. Your goal is to collect complete user information with personalized, adaptive communication.

Required Information:
1. Name (minimum 2 characters)
2. Email (valid format with @)
3. Phone (minimum 10 digits)
4. Preferences (1-5 interests/hobbies)

Use available tools to collect and validate information. Adapt your communication style based on conversation patterns and user responses.`).
		
		WithTools(collectTool, validateTool).
		WithOutputType(agent.NewStructuredOutputType(&UserProfile{})).
		
		// Simple field-based conditions
		OnMissingInfo("name").Ask("Hello! I'd love to help you get started. What's your name?").Build().
		OnMissingInfo("email").Ask("Great! Now I'll need your email address to set up your account.").Build().
		OnMissingInfo("phone").Ask("Perfect! Could you also provide your phone number for account verification?").Build().
		OnMissingInfo("preferences").Ask("Excellent! Finally, tell me about your interests or hobbies (1-5 topics).").Build().
		
		// Combination conditions using And/Or
		When(conditions.And(
			conditions.Missing("email"),
			conditions.Missing("phone"),
		)).Ask("I'll need both your email address and phone number to proceed. Could you provide both?").Build().
		
		// Message count conditions for conversation management
		OnMessageCount(6).Summarize().Build().
		OnMessageCount(10).Ask("We've been chatting for a while! Let me help summarize what we've collected and what's still needed.").Build().
		
		// Content-based conditions
		When(conditions.Contains("help")).Ask("Of course! I'm here to help you complete your profile. What specific information do you need assistance with?").Build().
		When(conditions.Contains("skip")).Ask("I understand you might want to skip some steps, but all the information is required for your account. Let's work together to complete it quickly!").Build().
		When(conditions.Contains("later")).Ask("I appreciate that you're busy! However, completing your profile now will give you immediate access to all features. Shall we finish the remaining steps?").Build().
		
		// Custom function-based conditions
		When(agent.WhenFunc("business_hours", func(session agent.Session) bool {
			now := time.Now()
			hour := now.Hour()
			return hour >= 9 && hour <= 17 // 9 AM to 5 PM
		})).Ask("Since it's during business hours, I can provide immediate assistance with any questions!").Build().
		
		When(agent.WhenFunc("weekend", func(session agent.Session) bool {
			now := time.Now()
			return now.Weekday() == time.Saturday || now.Weekday() == time.Sunday
		})).Ask("Hope you're having a great weekend! I'm here to help you get set up quickly.").Build().
		
		// Advanced combination: Multiple conditions with OR
		When(conditions.Or(
			conditions.Contains("frustrated"),
			conditions.Contains("difficult"),
			conditions.Count(8),
		)).Ask("I sense this process might be taking longer than expected. Let me help streamline this for you! What's the quickest way I can assist?").Build().
		
		// Data-driven conditions (these would work with session data)
		When(agent.WhenFunc("retry_attempt", func(session agent.Session) bool {
			messages := session.Messages()
			retryCount := 0
			for _, msg := range messages {
				if strings.Contains(strings.ToLower(msg.Content), "try again") ||
				   strings.Contains(strings.ToLower(msg.Content), "retry") {
					retryCount++
				}
			}
			return retryCount >= 2
		})).Ask("I see you've had to retry a few times. Let me provide extra guidance to make this easier!").Build().
		
		Build()

	if err != nil {
		log.Fatalf("❌ Failed to create onboarding agent: %v", err)
	}

	log.Printf("✅ Advanced onboarding agent created with sophisticated condition handling")

	// Test scenarios to demonstrate different condition types
	testScenarios := []struct {
		input       string
		description string
	}{
		{
			input:       "Hi there! I want to sign up.",
			description: "Initial contact - should trigger name missing condition",
		},
		{
			input:       "My name is Alex Chen",
			description: "Provides name - should ask for email",
		},
		{
			input:       "Can you help me understand what information you need?",
			description: "Contains 'help' - should trigger help condition",
		},
		{
			input:       "alex.chen@example.com",
			description: "Provides email - should ask for phone",
		},
		{
			input:       "Can I skip the phone number?",
			description: "Contains 'skip' - should trigger skip condition",
		},
		{
			input:       "Okay fine, my phone is +1-555-0123",
			description: "Provides phone - should ask for preferences",
		},
		{
			input:       "This is getting difficult...",
			description: "Contains 'difficult' - should trigger frustration condition",
		},
		{
			input:       "I like reading, hiking, and photography",
			description: "Provides preferences - should complete profile",
		},
	}

	sessionID := fmt.Sprintf("advanced-conditions-%d", time.Now().Unix())
	ctx := context.Background()

	fmt.Println("\n🚀 Testing Advanced Condition System")
	fmt.Println("Demonstrating various condition types and elegant flow control...")
	fmt.Println(strings.Repeat("=", 60))

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Test %d/%d: %s\n", i+1, len(testScenarios), scenario.description)
		fmt.Printf("👤 User: %s\n", scenario.input)

		log.Printf("REQUEST[%d]: %s", i+1, scenario.description)
		start := time.Now()

		response, err := onboardingAgent.Chat(ctx, scenario.input, 
			agent.WithSession(sessionID))
		if err != nil {
			log.Printf("❌ ERROR[%d]: %v", i+1, err)
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		duration := time.Since(start)
		log.Printf("RESPONSE[%d]: Duration: %.3fs", i+1, duration.Seconds())

		fmt.Printf("🤖 Assistant: %s\n", response.Message)

		// Display structured output if available
		if response.Data != nil {
			if profile, ok := response.Data.(*UserProfile); ok {
				fmt.Printf("📊 Profile Status: Complete=%v, Missing=%v\n", 
					profile.IsComplete, profile.MissingInfo)
				log.Printf("PROFILE[%d]: Complete=%v, Missing=%v", 
					i+1, profile.IsComplete, profile.MissingInfo)
				
				if profile.IsComplete {
					fmt.Println("🎉 Profile completed successfully!")
					break
				}
			}
		}

		// Show session metadata
		fmt.Printf("📈 Messages in session: %d\n", len(response.Session.Messages()))

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ Advanced Conditions Demo Completed!")
	fmt.Println("🎯 Demonstrated condition types:")
	fmt.Println("   • Simple field-based conditions (OnMissingInfo)")
	fmt.Println("   • Combination conditions (And/Or)")
	fmt.Println("   • Message count conditions (OnMessageCount)")
	fmt.Println("   • Content-based conditions (WhenContains)")
	fmt.Println("   • Custom function conditions (WhenFunc)")
	fmt.Println("   • Time-based conditions (business hours, weekends)")
	fmt.Println("   • Complex behavioral conditions (retry attempts)")
	fmt.Println("   • Dynamic flow adaptation")
	fmt.Println("   • Elegant error handling and user guidance")
	fmt.Println(strings.Repeat("=", 80))
}