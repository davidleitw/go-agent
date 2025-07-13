// Package main demonstrates dynamic schema selection based on conversation context.
//
// This example shows how schemas can be chosen contextually during conversation,
// how to adapt collection strategies based on user intent, and how to handle
// complex multi-step information gathering workflows.
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

// IntentClassifier represents a simple intent classification system
type IntentClassifier struct {
	patterns map[string][]string
}

// NewIntentClassifier creates a new intent classifier with predefined patterns
func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{
		patterns: map[string][]string{
			"technical_support": {
				"error", "bug", "crash", "login", "broken", "not working",
				"technical", "system", "website", "app", "software",
			},
			"billing_inquiry": {
				"billing", "charge", "payment", "invoice", "refund", "price",
				"cost", "subscription", "upgrade", "cancel", "money",
			},
			"account_management": {
				"account", "profile", "password", "settings", "preferences",
				"update", "change", "modify", "personal",
			},
			"product_inquiry": {
				"product", "feature", "how to", "tutorial", "guide",
				"documentation", "learn", "use", "setup",
			},
			"sales_inquiry": {
				"buy", "purchase", "demo", "trial", "pricing", "plan",
				"enterprise", "business", "quote", "sales",
			},
		},
	}
}

// ClassifyIntent determines the user's intent based on their input
func (ic *IntentClassifier) ClassifyIntent(input string) string {
	inputLower := strings.ToLower(input)
	scores := make(map[string]int)

	for intent, keywords := range ic.patterns {
		for _, keyword := range keywords {
			if strings.Contains(inputLower, keyword) {
				scores[intent]++
			}
		}
	}

	// Find the intent with the highest score
	maxScore := 0
	bestIntent := "general_inquiry"
	for intent, score := range scores {
		if score > maxScore {
			maxScore = score
			bestIntent = intent
		}
	}

	return bestIntent
}

// getSchemaForIntent dynamically selects appropriate schema based on user intent.
// This demonstrates how schemas can be chosen contextually during conversation.
func getSchemaForIntent(intent string) []*schema.Field {
	switch intent {
	case "technical_support":
		return []*schema.Field{
			schema.Define("email", "Please provide your email for technical follow-up"),
			schema.Define("error_description", "Please describe the error or issue you're experiencing"),
			schema.Define("steps_taken", "What troubleshooting steps have you already tried?"),
			schema.Define("environment", "What browser/device are you using?").Optional(),
			schema.Define("urgency", "How critical is this issue for your work? (low/medium/high)").Optional(),
		}

	case "billing_inquiry":
		return []*schema.Field{
			schema.Define("email", "Please provide the email associated with your account"),
			schema.Define("account_id", "What is your account ID or number?"),
			schema.Define("billing_question", "Please describe your billing question or concern"),
			schema.Define("amount", "If this involves a specific amount, please specify").Optional(),
			schema.Define("transaction_date", "When did the transaction occur? (if applicable)").Optional(),
		}

	case "account_management":
		return []*schema.Field{
			schema.Define("email", "Please provide your current account email"),
			schema.Define("request_type", "What would you like to change? (password/email/profile/settings)"),
			schema.Define("reason", "Please explain why you need to make this change"),
			schema.Define("verification_code", "If you have a verification code, please provide it").Optional(),
		}

	case "product_inquiry":
		return []*schema.Field{
			schema.Define("email", "Please provide your email for follow-up resources"),
			schema.Define("product_area", "Which product or feature are you interested in?"),
			schema.Define("use_case", "What are you trying to accomplish?"),
			schema.Define("experience_level", "How familiar are you with our products? (beginner/intermediate/advanced)").Optional(),
		}

	case "sales_inquiry":
		return []*schema.Field{
			schema.Define("email", "Please provide your business email"),
			schema.Define("company", "What company do you represent?"),
			schema.Define("team_size", "How many people are on your team?"),
			schema.Define("use_case", "How do you plan to use our product?"),
			schema.Define("timeline", "When are you looking to get started?").Optional(),
			schema.Define("budget", "Do you have a budget range in mind?").Optional(),
		}

	default: // general_inquiry
		return []*schema.Field{
			schema.Define("email", "Please provide your email address"),
			schema.Define("inquiry_topic", "What would you like to know about?"),
			schema.Define("preferred_contact", "How would you prefer to be contacted? (email/phone)").Optional(),
		}
	}
}

// getWorkflowForIntent returns a multi-step workflow for complex intents
func getWorkflowForIntent(intent string) [][]*schema.Field {
	switch intent {
	case "technical_support":
		// Multi-step workflow for technical issues
		return [][]*schema.Field{
			{ // Step 1: Basic information
				schema.Define("email", "Please provide your email for follow-up"),
				schema.Define("issue_summary", "Can you briefly describe the issue?"),
			},
			{ // Step 2: Detailed technical information
				schema.Define("error_message", "What exact error message do you see?"),
				schema.Define("steps_to_reproduce", "What steps lead to this error?"),
				schema.Define("browser_version", "What browser and version are you using?").Optional(),
			},
			{ // Step 3: Context and troubleshooting
				schema.Define("when_started", "When did this issue first occur?"),
				schema.Define("frequency", "How often does this happen?"),
				schema.Define("workaround", "Have you found any temporary workaround?").Optional(),
			},
		}

	case "sales_inquiry":
		// Multi-step workflow for sales qualification
		return [][]*schema.Field{
			{ // Step 1: Contact information
				schema.Define("email", "Please provide your business email"),
				schema.Define("company", "What company do you represent?"),
				schema.Define("role", "What's your role at the company?"),
			},
			{ // Step 2: Requirements gathering
				schema.Define("team_size", "How many people would use our product?"),
				schema.Define("current_solution", "What solution are you currently using?"),
				schema.Define("pain_points", "What challenges are you trying to solve?"),
			},
			{ // Step 3: Qualification and next steps
				schema.Define("timeline", "When are you looking to make a decision?"),
				schema.Define("budget_authority", "Are you involved in the buying decision?"),
				schema.Define("next_step", "What would be the ideal next step for you?").Optional(),
			},
		}

	default:
		// Single step for simple intents
		return [][]*schema.Field{getSchemaForIntent(intent)}
	}
}

func main() {
	fmt.Println("🎯 Dynamic Schema Selection Example")
	fmt.Println("Intelligent schema adaptation based on user intent and conversation flow")
	fmt.Println("=======================================================================")

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

	// Create adaptive assistant with dynamic schema capabilities
	adaptiveBot, err := agent.New("adaptive-assistant").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Intelligent assistant with adaptive information collection").
		WithInstructions(`You are an intelligent assistant that adapts to user needs. 
			Analyze user intent and collect information accordingly. Be professional, 
			helpful, and efficient in gathering the right information for each situation.`).
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create adaptive assistant: %v", err)
	}

	log.Printf("✅ Adaptive assistant created successfully")

	ctx := context.Background()
	classifier := NewIntentClassifier()

	// Test scenarios for dynamic schema selection
	scenarios := []struct {
		title     string
		userInput string
	}{
		{
			title:     "Technical Support Request",
			userInput: "I'm getting a login error when trying to access the system",
		},
		{
			title:     "Billing Question",
			userInput: "I have a question about charges on my latest invoice",
		},
		{
			title:     "Account Management",
			userInput: "I need to update my account settings and change my password",
		},
		{
			title:     "Product Inquiry",
			userInput: "Can you help me understand how to use the analytics feature?",
		},
		{
			title:     "Sales Request",
			userInput: "I'm interested in purchasing your enterprise plan for my company",
		},
		{
			title:     "General Question",
			userInput: "Hello, I have some general questions about your service",
		},
	}

	// Test intent classification and dynamic schema selection
	fmt.Printf("\n🧠 Intent Classification and Schema Selection\n")
	fmt.Printf("============================================\n")

	for i, scenario := range scenarios {
		fmt.Printf("\n📝 Scenario %d: %s\n", i+1, scenario.title)
		fmt.Printf("👤 User: %s\n", scenario.userInput)

		// Classify intent
		intent := classifier.ClassifyIntent(scenario.userInput)
		fmt.Printf("🎯 Detected Intent: %s\n", intent)

		// Get appropriate schema
		selectedSchema := getSchemaForIntent(intent)
		fmt.Printf("📋 Selected Schema (%d fields):\n", len(selectedSchema))
		for _, field := range selectedSchema {
			required := "required"
			if !field.Required() {
				required = "optional"
			}
			fmt.Printf("   - %s (%s): %s\n", field.Name(), required, field.Prompt())
		}

		// Execute with dynamically selected schema
		sessionID := fmt.Sprintf("dynamic-%d", i+1)
		start := time.Now()

		response, err := adaptiveBot.Chat(ctx, scenario.userInput,
			agent.WithSession(sessionID),
			agent.WithSchema(selectedSchema...),
		)
		if err != nil {
			log.Printf("❌ Error in scenario %d: %v", i+1, err)
			continue
		}

		duration := time.Since(start)
		fmt.Printf("🤖 Assistant: %s\n", response.Message)
		fmt.Printf("⏱️  Response time: %.3fs\n", duration.Seconds())

		fmt.Println("-----------------------------------------------------------------------")
	}

	// Demonstrate multi-step workflow
	fmt.Printf("\n🔄 Multi-Step Workflow Example\n")
	fmt.Printf("==============================\n")

	workflowInput := "I'm having a critical technical issue that's blocking my work"
	intent := classifier.ClassifyIntent(workflowInput)
	workflow := getWorkflowForIntent(intent)

	fmt.Printf("👤 User: %s\n", workflowInput)
	fmt.Printf("🎯 Intent: %s\n", intent)
	fmt.Printf("📊 Workflow Steps: %d\n", len(workflow))

	workflowSession := "workflow-demo"

	for step, stepSchema := range workflow {
		fmt.Printf("\n📋 Step %d/%d - Collecting:\n", step+1, len(workflow))
		for _, field := range stepSchema {
			required := "required"
			if !field.Required() {
				required = "optional"
			}
			fmt.Printf("   - %s (%s)\n", field.Name(), required)
		}

		// Simulate user input for demonstration
		var simulatedInput string
		switch step {
		case 0:
			simulatedInput = workflowInput
		case 1:
			simulatedInput = "Error: 'Connection timeout' appears when I click login. It happens every time I try to log in using Chrome version 120."
		case 2:
			simulatedInput = "This started yesterday morning. It happens every single time. I can access other websites fine."
		default:
			simulatedInput = "That's all the information I have."
		}

		fmt.Printf("👤 User: %s\n", simulatedInput)

		response, err := adaptiveBot.Chat(ctx, simulatedInput,
			agent.WithSession(workflowSession),
			agent.WithSchema(stepSchema...),
		)
		if err != nil {
			log.Printf("❌ Error in workflow step %d: %v", step+1, err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", response.Message)

		// Show progress
		if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
			if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
				fmt.Printf("📊 Still need: %v\n", missingFields)
			}
		} else {
			fmt.Printf("✅ Step %d completed!\n", step+1)
		}
	}

	// Show conversation analytics
	fmt.Printf("\n📈 Conversation Analytics\n")
	fmt.Printf("========================\n")

	// Get final session state
	finalResponse, err := adaptiveBot.Chat(ctx, "Thank you for your help",
		agent.WithSession(workflowSession),
	)
	if err == nil {
		fmt.Printf("📨 Total messages: %d\n", len(finalResponse.Session.Messages()))
		fmt.Printf("🎯 Workflow completion: Successfully collected technical support information\n")
		fmt.Printf("📋 Information gathered across multiple steps\n")
	}

	fmt.Printf("\n=======================================================================\n")
	fmt.Printf("✅ Dynamic Schema Selection Example Complete!\n")
	fmt.Printf("🎯 Advanced Features Demonstrated:\n")
	fmt.Printf("   • Intelligent intent classification\n")
	fmt.Printf("   • Dynamic schema selection based on context\n")
	fmt.Printf("   • Multi-step information collection workflows\n")
	fmt.Printf("   • Adaptive conversation strategies\n")
	fmt.Printf("   • Complex business logic integration\n")
	fmt.Printf("   • Real-time schema adaptation\n")
	fmt.Printf("=======================================================================\n")
}