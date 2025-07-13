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

// CustomerRequest represents the structured output for customer service
type CustomerRequest struct {
	RequestType    string   `json:"request_type"`
	Priority       string   `json:"priority"`
	Summary        string   `json:"summary"`
	RequiredAction string   `json:"required_action"`
	Tags           []string `json:"tags"`
	IsComplete     bool     `json:"is_complete"`
}

func (c *CustomerRequest) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"request_type": map[string]any{
				"type":        "string",
				"description": "Type of customer request (question, complaint, feature_request, etc.)",
			},
			"priority": map[string]any{
				"type":        "string",
				"description": "Priority level (low, medium, high, urgent)",
			},
			"summary": map[string]any{
				"type":        "string",
				"description": "Brief summary of the request",
			},
			"required_action": map[string]any{
				"type":        "string",
				"description": "What action should be taken to resolve this",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Relevant tags for categorization",
			},
			"is_complete": map[string]any{
				"type":        "boolean",
				"description": "Whether the request has been fully processed",
			},
		},
		"required": []string{"request_type", "summary", "is_complete"},
	}
}

func (c *CustomerRequest) NewInstance() any {
	return &CustomerRequest{}
}

func (c *CustomerRequest) Validate(instance any) error {
	_, ok := instance.(*CustomerRequest)
	if !ok {
		return fmt.Errorf("instance is not a CustomerRequest")
	}
	return nil
}

func main() {
	fmt.Println("🌟 LLM Enhanced Responses Example")
	fmt.Println("Demonstrating intelligent contextual responses...")
	fmt.Println("=" + strings.Repeat("=", 60))

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

	// Create a customer service agent with smart LLM-enhanced responses
	customerAgent, err := agent.New("smart-customer-service").
		WithOpenAI(apiKey).
		WithModel("gpt-4o-mini").
		WithDescription("An intelligent customer service agent with contextual response capabilities").
		WithInstructions(`You are an expert customer service representative with advanced conversation skills.

Core Capabilities:
- Understand customer emotions and adapt your tone accordingly
- Provide contextual responses based on conversation history
- Escalate complex issues appropriately
- Maintain professional yet friendly communication

When using AskAI enhanced responses, consider:
- The customer's emotional state from previous messages
- The complexity and type of their request
- Whether they seem frustrated or satisfied
- The conversation length and flow`).

		WithOutputType(agent.NewStructuredOutputType(&CustomerRequest{})).

		// Greeting and initial response
		When(conditions.Count(1)).
			Ask("Hello! I'm here to help you today. What can I assist you with?").
			Build().

		// Detect frustration and respond with empathy (LLM-enhanced)
		When(conditions.Or(
			conditions.Contains("frustrated"),
			conditions.Contains("angry"),
			conditions.Contains("terrible"),
			conditions.Contains("awful"),
		)).
			AskAI("The customer seems frustrated or upset. Based on their message and our conversation, provide an empathetic response that acknowledges their feelings and offers concrete help. Be genuine and professional.").
			OrElse("I understand your frustration, and I sincerely apologize for any inconvenience. Let me help you resolve this issue right away.").
			Build().

		// Handle compliments with contextual appreciation
		When(conditions.Or(
			conditions.Contains("great"),
			conditions.Contains("excellent"),
			conditions.Contains("thank you"),
			conditions.Contains("helpful"),
		)).
			AskAI("The customer is expressing satisfaction or gratitude. Based on our conversation context, provide a warm, professional response that reinforces their positive experience and encourages continued engagement.").
			OrElse("Thank you so much for your kind words! I'm delighted I could help. Is there anything else I can assist you with?").
			Build().

		// Complex technical questions get detailed, contextual responses
		When(conditions.Or(
			conditions.Contains("technical"),
			conditions.Contains("API"),
			conditions.Contains("integration"),
			conditions.Contains("developer"),
		)).
			AskAI("The customer has a technical question. Based on the conversation context and their specific needs, provide a detailed yet accessible explanation. Include relevant examples or next steps if appropriate.").
			OrElse("I'd be happy to help with your technical question. Let me provide you with detailed information and guide you through the solution.").
			Build().

		// Long conversations get summarization offers
		When(conditions.Count(8)).
			AskAI("We've been discussing several topics. Based on our conversation, provide a helpful summary of what we've covered and suggest next steps or ask if there are other areas where you can help.").
			OrElse("We've covered quite a bit today! Let me summarize what we've discussed and see if there are any other ways I can help you.").
			Build().

		// Handle urgent requests with priority acknowledgment
		When(conditions.Or(
			conditions.Contains("urgent"),
			conditions.Contains("emergency"),
			conditions.Contains("ASAP"),
			conditions.Contains("critical"),
		)).
			AskAI("The customer has marked this as urgent or critical. Based on their request and context, provide an immediate acknowledgment of the urgency and outline your action plan to address their needs quickly.").
			OrElse("I understand this is urgent for you. Let me prioritize your request and work on a solution immediately.").
			Build().

		// Feature requests get contextual product guidance
		When(conditions.Or(
			conditions.Contains("feature"),
			conditions.Contains("enhancement"),
			conditions.Contains("suggestion"),
			conditions.Contains("would be nice"),
		)).
			AskAI("The customer is suggesting a feature or enhancement. Based on their specific suggestion and context, provide an encouraging response that explains how their feedback is valuable and what the typical process is for feature consideration.").
			OrElse("Thank you for that great suggestion! Feature requests like yours help us improve our product. I'll make sure your feedback reaches our product team.").
			Build().

		Build()

	if err != nil {
		log.Fatalf("❌ Failed to create customer service agent: %v", err)
	}

	log.Printf("✅ Smart customer service agent created with LLM-enhanced responses")

	// Test scenarios showcasing different response types
	testScenarios := []struct {
		input       string
		description string
		expected    string
	}{
		{
			input:       "Hi, I'm having some issues with your service",
			description: "Initial greeting",
			expected:    "Direct response",
		},
		{
			input:       "This is really frustrating! Nothing is working properly and I'm getting angry about it",
			description: "Frustrated customer",
			expected:    "AskAI empathetic response",
		},
		{
			input:       "Actually, your support has been excellent and very helpful",
			description: "Positive feedback",
			expected:    "AskAI appreciation response",
		},
		{
			input:       "I need help with API integration for my developer team",
			description: "Technical question",
			expected:    "AskAI technical response",
		},
		{
			input:       "This is really urgent - we have a critical system down",
			description: "Urgent request",
			expected:    "AskAI urgency acknowledgment",
		},
		{
			input:       "It would be nice if you had a feature for automated backups",
			description: "Feature suggestion",
			expected:    "AskAI feature response",
		},
		{
			input:       "Let me know if there's anything else",
			description: "Continuing conversation",
			expected:    "Context-aware response",
		},
		{
			input:       "Perfect, thanks for all your help today!",
			description: "Final appreciation",
			expected:    "Message count + appreciation response",
		},
	}

	sessionID := fmt.Sprintf("llm-enhanced-%d", time.Now().Unix())
	ctx := context.Background()

	fmt.Println("\n🚀 Testing LLM-Enhanced Customer Service")
	fmt.Println("Observing how responses adapt to context and customer sentiment...")
	fmt.Println("=" + strings.Repeat("=", 70))

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Scenario %d/%d: %s\n", i+1, len(testScenarios), scenario.description)
		fmt.Printf("👤 Customer: %s\n", scenario.input)
		fmt.Printf("🎯 Expected: %s\n", scenario.expected)

		start := time.Now()

		response, err := customerAgent.Chat(ctx, scenario.input, 
			agent.WithSession(sessionID))
		if err != nil {
			log.Printf("❌ ERROR[%d]: %v", i+1, err)
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		duration := time.Since(start)
		log.Printf("RESPONSE[%d]: Duration: %.3fs", i+1, duration.Seconds())

		fmt.Printf("🤖 Agent: %s\n", response.Message)

		// Show structured output if available
		if response.Data != nil {
			if request, ok := response.Data.(*CustomerRequest); ok {
				fmt.Printf("📊 Analysis: Type=%s, Priority=%s, Complete=%v\n", 
					request.RequestType, request.Priority, request.IsComplete)
			}
		}

		fmt.Printf("📈 Messages: %d | Duration: %.2fs\n", len(response.Session.Messages()), duration.Seconds())

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ LLM-Enhanced Responses Demo Completed!")
	fmt.Println("\n🎯 Demonstrated Features:")
	fmt.Println("   🧠 Contextual AI responses based on conversation history")
	fmt.Println("   💝 Emotion-aware adaptations (empathy for frustration)")
	fmt.Println("   🎉 Dynamic appreciation for positive feedback") 
	fmt.Println("   🔧 Technical depth adjustment based on query type")
	fmt.Println("   ⚡ Urgency recognition and priority acknowledgment")
	fmt.Println("   💡 Feature request handling with product guidance")
	fmt.Println("   📋 Conversation summarization for long interactions")
	fmt.Println("   🔄 Fallback responses ensure reliability")
	fmt.Println("   📊 Structured output for request analysis")
	fmt.Println("\n💡 Key Benefits:")
	fmt.Println("   • More natural, context-aware conversations")
	fmt.Println("   • Improved customer satisfaction through empathy")
	fmt.Println("   • Consistent quality with intelligent fallbacks")
	fmt.Println("   • Scalable LLM-enhanced customer service")
	fmt.Println("=" + strings.Repeat("=", 80))
}