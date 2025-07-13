// Package main demonstrates a customer support bot with comprehensive information collection.
//
// This example shows how to use schema-based collection for a real-world customer support
// scenario, including required and optional fields, contextual prompts, and conversation flow.
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

// supportSchema defines the complete information required for customer support tickets.
// This includes both required information for proper support and optional details
// that can help provide better assistance.
func supportSchema() []*schema.Field {
	return []*schema.Field{
		// Required information for all support requests
		schema.Define("email", "Please provide your email address for follow-up"),
		schema.Define("issue_category", "What type of issue are you experiencing? (technical/billing/account/other)"),
		schema.Define("description", "Please describe the issue in detail"),

		// Optional information that can help with support
		schema.Define("order_id", "If this relates to an order, please provide the order number").Optional(),
		schema.Define("urgency", "How urgent is this issue? (low/medium/high)").Optional(),
		schema.Define("previous_contact", "Have you contacted support about this before?").Optional(),
	}
}

// billingSchema defines specific fields for billing-related inquiries
func billingSchema() []*schema.Field {
	return []*schema.Field{
		// Required for billing issues
		schema.Define("email", "Please provide the email associated with your account"),
		schema.Define("account_number", "What is your account number?"),
		schema.Define("billing_question", "Please describe your billing question in detail"),

		// Optional billing-specific information
		schema.Define("amount_disputed", "If disputing a charge, what amount?").Optional(),
		schema.Define("payment_method", "What payment method are you using?").Optional(),
		schema.Define("billing_period", "Which billing period does this concern?").Optional(),
	}
}

// technicalSchema defines fields for technical support issues
func technicalSchema() []*schema.Field {
	return []*schema.Field{
		// Required for technical support
		schema.Define("email", "Please provide your email for technical follow-up"),
		schema.Define("error_message", "What error message are you seeing?"),
		schema.Define("steps_taken", "What troubleshooting steps have you already tried?"),

		// Optional technical information
		schema.Define("browser", "What browser are you using?").Optional(),
		schema.Define("device_type", "What device are you using? (desktop/mobile/tablet)").Optional(),
		schema.Define("operating_system", "What operating system are you using?").Optional(),
	}
}

func main() {
	fmt.Println("🎧 Customer Support Bot Example")
	fmt.Println("Advanced schema-based information collection for customer support")
	fmt.Println("===============================================================")

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

	// Create a customer support agent with comprehensive information collection
	supportBot, err := agent.New("support-bot").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("Professional customer support assistant").
		WithInstructions(`You are a professional customer support assistant. Your goal is to efficiently 
			collect the necessary information to help resolve customer issues. Be empathetic, 
			professional, and ensure you gather all required details before proceeding with assistance.`).
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create support bot: %v", err)
	}

	log.Printf("✅ Customer support bot created successfully")

	ctx := context.Background()

	// Customer support scenarios demonstrating different schema usage
	scenarios := []struct {
		title       string
		description string
		userInput   string
		schema      []*schema.Field
	}{
		{
			title:       "General Support Request",
			description: "Customer with a general issue - uses comprehensive support schema",
			userInput:   "I'm having trouble with my account",
			schema:      supportSchema(),
		},
		{
			title:       "Billing Inquiry",
			description: "Customer with billing question - uses specialized billing schema",
			userInput:   "I have a question about my last invoice",
			schema:      billingSchema(),
		},
		{
			title:       "Technical Support",
			description: "Customer with technical issue - uses technical support schema",
			userInput:   "I'm getting an error when trying to log in",
			schema:      technicalSchema(),
		},
		{
			title:       "Partial Information Provided",
			description: "Customer provides some information upfront",
			userInput:   "Hi, my email is customer@example.com and I have a billing issue with order #12345",
			schema:      billingSchema(),
		},
	}

	// Run through each customer support scenario
	for i, scenario := range scenarios {
		fmt.Printf("\n📞 Scenario %d: %s\n", i+1, scenario.title)
		fmt.Printf("📝 %s\n", scenario.description)
		fmt.Printf("👤 Customer: %s\n", scenario.userInput)

		// Show what information we're collecting
		fmt.Printf("📋 Information to collect:\n")
		requiredFields := make([]string, 0)
		optionalFields := make([]string, 0)
		for _, field := range scenario.schema {
			if field.Required() {
				requiredFields = append(requiredFields, field.Name())
			} else {
				optionalFields = append(optionalFields, field.Name())
			}
		}
		fmt.Printf("   Required: %v\n", requiredFields)
		fmt.Printf("   Optional: %v\n", optionalFields)

		// Use unique session for each scenario
		sessionID := fmt.Sprintf("support-scenario-%d", i+1)

		start := time.Now()

		// Execute support interaction with schema
		response, err := supportBot.Chat(ctx, scenario.userInput,
			agent.WithSession(sessionID),
			agent.WithSchema(scenario.schema...),
		)
		if err != nil {
			log.Printf("❌ Error in scenario %d: %v", i+1, err)
			continue
		}

		duration := time.Since(start)

		fmt.Printf("🤖 Support Agent: %s\n", response.Message)

		// Show collection metadata
		if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
			if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
				fmt.Printf("📊 Missing required fields: %v\n", missingFields)
			}
			fmt.Printf("⏱️  Response time: %.3fs\n", duration.Seconds())
		} else {
			fmt.Printf("✅ All required information collected!\n")
			fmt.Printf("⏱️  Response time: %.3fs\n", duration.Seconds())
		}

		fmt.Println("---------------------------------------------------------------")
	}

	// Demonstrate multi-turn conversation with information collection
	fmt.Printf("\n🔄 Multi-Turn Support Conversation\n")
	fmt.Printf("Demonstrating information collection across multiple interactions\n")
	fmt.Printf("================================================================\n")

	conversationSession := "support-conversation"

	// Turn 1: Initial contact
	fmt.Printf("👤 Customer: I need help with something\n")
	response1, err := supportBot.Chat(ctx, "I need help with something",
		agent.WithSession(conversationSession),
		agent.WithSchema(supportSchema()...),
	)
	if err != nil {
		log.Printf("❌ Error in conversation turn 1: %v", err)
	} else {
		fmt.Printf("🤖 Support Agent: %s\n", response1.Message)
	}

	// Turn 2: Partial information
	fmt.Printf("\n👤 Customer: My email is john.doe@example.com and it's a technical issue\n")
	response2, err := supportBot.Chat(ctx, "My email is john.doe@example.com and it's a technical issue",
		agent.WithSession(conversationSession),
		agent.WithSchema(supportSchema()...),
	)
	if err != nil {
		log.Printf("❌ Error in conversation turn 2: %v", err)
	} else {
		fmt.Printf("🤖 Support Agent: %s\n", response2.Message)
	}

	// Turn 3: Complete information
	fmt.Printf("\n👤 Customer: I can't access my dashboard. It shows 'Connection timeout' error\n")
	response3, err := supportBot.Chat(ctx, "I can't access my dashboard. It shows 'Connection timeout' error",
		agent.WithSession(conversationSession),
		agent.WithSchema(supportSchema()...),
	)
	if err != nil {
		log.Printf("❌ Error in conversation turn 3: %v", err)
	} else {
		fmt.Printf("🤖 Support Agent: %s\n", response3.Message)
		fmt.Printf("📨 Total conversation turns: %d\n", len(response3.Session.Messages()))
	}

	// Show support ticket summary
	fmt.Printf("\n📋 Support Ticket Summary\n")
	fmt.Printf("Session ID: %s\n", conversationSession)
	fmt.Printf("Messages exchanged: %d\n", len(response3.Session.Messages()))
	if response3.Metadata["schema_collection"] == nil {
		fmt.Printf("Status: ✅ Ready for escalation (all info collected)\n")
	} else {
		fmt.Printf("Status: ⏳ Still collecting information\n")
	}

	fmt.Printf("\n===============================================================\n")
	fmt.Printf("✅ Customer Support Bot Example Complete!\n")
	fmt.Printf("🎯 Features Demonstrated:\n")
	fmt.Printf("   • Specialized schemas for different support types\n")
	fmt.Printf("   • Required vs optional field collection\n")
	fmt.Printf("   • Multi-turn conversation handling\n")
	fmt.Printf("   • Contextual information extraction\n")
	fmt.Printf("   • Professional support agent behavior\n")
	fmt.Printf("   • Comprehensive ticket information gathering\n")
	fmt.Printf("===============================================================\n")
}