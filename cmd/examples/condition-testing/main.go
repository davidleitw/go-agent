package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/openai"
	"github.com/joho/godotenv"
)

// UserStatusOutput represents the current status of user onboarding
type UserStatusOutput struct {
	UserID          string   `json:"user_id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Phone           string   `json:"phone"`
	Preferences     []string `json:"preferences"`
	CompletionStage string   `json:"completion_stage"`
	MissingFields   []string `json:"missing_fields"`
	CompletionFlag  bool     `json:"completion_flag"`
	Message         string   `json:"message"`
}

func (u *UserStatusOutput) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "Unique user identifier",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "User's full name",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "User's email address",
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "User's phone number",
			},
			"preferences": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "User's preferences list",
			},
			"completion_stage": map[string]interface{}{
				"type":        "string",
				"description": "Current stage of onboarding",
				"enum":        []string{"basic_info", "contact_details", "preferences", "completed"},
			},
			"missing_fields": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of missing required fields",
			},
			"completion_flag": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether onboarding is complete",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Message to display to user",
			},
		},
		"required": []string{"user_id", "completion_stage", "missing_fields", "completion_flag", "message"},
	}
}

func (u *UserStatusOutput) NewInstance() interface{} {
	return &UserStatusOutput{}
}

// CollectInfoTool simulates collecting user information
type CollectInfoTool struct{}

func (t *CollectInfoTool) Name() string {
	return "collect_user_info"
}

func (t *CollectInfoTool) Description() string {
	return "Collect and validate user information during onboarding process"
}

func (t *CollectInfoTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of information to collect",
				"enum":        []string{"name", "email", "phone", "preferences"},
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "The collected value",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "User identifier",
			},
		},
		"required": []string{"field_type", "value", "user_id"},
	}
}

func (t *CollectInfoTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	fieldType, ok := args["field_type"].(string)
	if !ok {
		return nil, fmt.Errorf("field_type must be a string")
	}

	value, ok := args["value"].(string)
	if !ok {
		return nil, fmt.Errorf("value must be a string")
	}

	userID, ok := args["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id must be a string")
	}

	// Simulate validation logic
	var isValid bool
	var errorMsg string

	switch fieldType {
	case "name":
		isValid = len(strings.TrimSpace(value)) >= 2
		if !isValid {
			errorMsg = "Name must be at least 2 characters long"
		}
	case "email":
		isValid = strings.Contains(value, "@") && strings.Contains(value, ".")
		if !isValid {
			errorMsg = "Email must be a valid email address"
		}
	case "phone":
		cleaned := strings.ReplaceAll(value, " ", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		isValid = len(cleaned) >= 10 && strings.HasPrefix(cleaned, "0")
		if !isValid {
			errorMsg = "Phone number must be at least 10 digits and start with 0"
		}
	case "preferences":
		prefs := strings.Split(value, ",")
		isValid = len(prefs) >= 1 && len(prefs) <= 5
		if !isValid {
			errorMsg = "Please provide 1-5 preferences separated by commas"
		}
	default:
		return nil, fmt.Errorf("unsupported field_type: %s", fieldType)
	}

	result := map[string]interface{}{
		"user_id":    userID,
		"field_type": fieldType,
		"value":      value,
		"is_valid":   isValid,
		"error_msg":  errorMsg,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if isValid {
		log.Printf("✅ COLLECT[%s]: %s = %s", userID, fieldType, value)
	} else {
		log.Printf("❌ COLLECT[%s]: %s validation failed - %s", userID, fieldType, errorMsg)
	}

	return result, nil
}

// ValidationTool provides additional validation capabilities
type ValidationTool struct{}

func (t *ValidationTool) Name() string {
	return "validate_completion"
}

func (t *ValidationTool) Description() string {
	return "Validate if user onboarding is complete and determine next steps"
}

func (t *ValidationTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user_data": map[string]interface{}{
				"type":        "object",
				"description": "Current user data to validate",
			},
		},
		"required": []string{"user_data"},
	}
}

func (t *ValidationTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	userData, ok := args["user_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("user_data must be an object")
	}

	// Check required fields
	requiredFields := []string{"name", "email", "phone", "preferences"}
	missingFields := []string{}

	for _, field := range requiredFields {
		if value, exists := userData[field]; !exists || value == "" {
			missingFields = append(missingFields, field)
		}
	}

	isComplete := len(missingFields) == 0
	var stage string
	if isComplete {
		stage = "completed"
	} else if len(missingFields) == len(requiredFields) {
		stage = "basic_info"
	} else if contains(missingFields, "email") || contains(missingFields, "phone") {
		stage = "contact_details"
	} else {
		stage = "preferences"
	}

	result := map[string]interface{}{
		"is_complete":      isComplete,
		"missing_fields":   missingFields,
		"completion_stage": stage,
		"total_fields":     len(requiredFields),
		"completed_fields": len(requiredFields) - len(missingFields),
		"completion_rate":  float64(len(requiredFields)-len(missingFields)) / float64(len(requiredFields)),
	}

	log.Printf("🔍 VALIDATE: Stage=%s, Missing=%v, Complete=%v", stage, missingFields, isComplete)
	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Custom Conditions

// MissingFieldsCondition checks if specific fields are missing
type MissingFieldsCondition struct {
	name   string
	fields []string
}

func NewMissingFieldsCondition(name string, fields []string) agent.Condition {
	return &MissingFieldsCondition{
		name:   name,
		fields: fields,
	}
}

func (c *MissingFieldsCondition) Name() string {
	return c.name
}

func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]interface{}) (bool, error) {
	missingFields, exists := data["missing_fields"]
	if !exists {
		return false, nil
	}

	missingList, ok := missingFields.([]interface{})
	if !ok {
		return false, nil
	}

	// Convert to string slice
	missing := make([]string, len(missingList))
	for i, v := range missingList {
		if s, ok := v.(string); ok {
			missing = append(missing, s)
		}
	}

	// Check if any of our target fields are in the missing list
	for _, targetField := range c.fields {
		for _, missingField := range missing {
			if missingField == targetField {
				log.Printf("🎯 CONDITION[%s]: Field '%s' is missing", c.name, targetField)
				return true, nil
			}
		}
	}

	log.Printf("✅ CONDITION[%s]: All target fields present", c.name)
	return false, nil
}

// CompletionStageCondition checks the current completion stage
type CompletionStageCondition struct {
	name  string
	stage string
}

func NewCompletionStageCondition(name string, stage string) agent.Condition {
	return &CompletionStageCondition{
		name:  name,
		stage: stage,
	}
}

func (c *CompletionStageCondition) Name() string {
	return c.name
}

func (c *CompletionStageCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]interface{}) (bool, error) {
	currentStage, exists := data["completion_stage"]
	if !exists {
		return false, nil
	}

	stage, ok := currentStage.(string)
	if !ok {
		return false, nil
	}

	result := stage == c.stage
	log.Printf("🎯 CONDITION[%s]: Stage '%s' == '%s' ? %v", c.name, stage, c.stage, result)
	return result, nil
}

// MessageCountCondition checks the number of messages in session
type MessageCountCondition struct {
	name     string
	minCount int
}

func NewMessageCountCondition(name string, minCount int) agent.Condition {
	return &MessageCountCondition{
		name:     name,
		minCount: minCount,
	}
}

func (c *MessageCountCondition) Name() string {
	return c.name
}

func (c *MessageCountCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]interface{}) (bool, error) {
	messageCount := len(session.Messages())
	result := messageCount >= c.minCount
	log.Printf("🎯 CONDITION[%s]: Message count %d >= %d ? %v", c.name, messageCount, c.minCount, result)
	return result, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Verify OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}
	log.Printf("✅ OpenAI API key loaded (length: %d)", len(apiKey))

	// Create OpenAI chat model
	log.Println("🔧 Creating OpenAI chat model...")
	chatModel, err := openai.NewChatModel(apiKey, nil)
	if err != nil {
		log.Fatalf("❌ Failed to create OpenAI chat model: %v", err)
	}

	// Create tools
	tools := []agent.Tool{
		&CollectInfoTool{},
		&ValidationTool{},
	}

	// Create custom conditions
	missingContactCondition := NewMissingFieldsCondition("missing_contact_info", []string{"email", "phone"})
	missingPrefsCondition := NewMissingFieldsCondition("missing_preferences", []string{"preferences"})
	basicInfoStageCondition := NewCompletionStageCondition("at_basic_info_stage", "basic_info")
	contactStageCondition := NewCompletionStageCondition("at_contact_stage", "contact_details")
	longConversationCondition := NewMessageCountCondition("long_conversation", 6)

	// Create flow rules
	flowRules := []agent.FlowRule{}

	// Rule 1: Missing contact information
	contactRule, err := agent.NewFlowRule("collect-contact-info", missingContactCondition).
		WithDescription("Prompt user for missing contact information").
		WithNewInstructions("The user is missing contact information. Please ask specifically for their email and phone number. Be encouraging and explain why this information is needed.").
		WithRecommendedTools("collect_user_info").
		WithSystemMessage("Contact information collection needed").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create contact rule: %v", err)
	}
	flowRules = append(flowRules, contactRule)

	// Rule 2: Missing preferences
	prefsRule, err := agent.NewFlowRule("collect-preferences", missingPrefsCondition).
		WithDescription("Collect user preferences").
		WithNewInstructions("The user needs to specify their preferences. Ask them about their interests, hobbies, or preferred topics (1-5 items, comma-separated).").
		WithRecommendedTools("collect_user_info", "validate_completion").
		WithSystemMessage("Preferences collection needed").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create preferences rule: %v", err)
	}
	flowRules = append(flowRules, prefsRule)

	// Rule 3: At basic info stage
	basicInfoRule, err := agent.NewFlowRule("handle-basic-info", basicInfoStageCondition).
		WithDescription("Handle users at basic info stage").
		WithNewInstructions("The user is just starting. Collect their basic information first (name), then we'll move to contact details.").
		WithRecommendedTools("collect_user_info").
		WithSystemMessage("Starting basic information collection").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create basic info rule: %v", err)
	}
	flowRules = append(flowRules, basicInfoRule)

	// Rule 4: Long conversation optimization
	longConvRule, err := agent.NewFlowRule("optimize-long-conversation", longConversationCondition).
		WithDescription("Optimize for long conversations").
		WithNewInstructions("This conversation has been going on for a while. Please provide a summary of what we've collected so far and clearly state what's still needed to complete the onboarding.").
		WithRecommendedTools("validate_completion").
		WithSystemMessage("Long conversation detected - providing summary").
		Build()
	if err != nil {
		log.Fatalf("❌ Failed to create long conversation rule: %v", err)
	}
	flowRules = append(flowRules, longConvRule)

	// Create onboarding agent with conditions and flow rules
	log.Println("📝 Creating onboarding agent with condition testing...")
	onboardingAgent, err := agent.New(
		agent.WithName("onboarding-agent"),
		agent.WithDescription("An intelligent onboarding agent that uses conditions and flow rules to guide user registration"),
		agent.WithInstructions(`You are an onboarding specialist that helps new users complete their registration process.

Your goal is to collect the following information:
1. Name (required)
2. Email (required) 
3. Phone number (required)
4. Preferences (required) - 1-5 interests/topics

Use the available tools to:
- collect_user_info: Gather and validate user information
- validate_completion: Check if onboarding is complete

Always provide the structured output showing:
- Current user data
- What stage they're at
- What's missing
- Whether they're complete
- A helpful message

Be friendly, encouraging, and clear about what information is needed and why.`),
		agent.WithChatModel(chatModel),
		agent.WithModel("gpt-4"),
		agent.WithModelSettings(&agent.ModelSettings{
			Temperature: floatPtr(0.7),
			MaxTokens:   intPtr(2000),
		}),
		agent.WithTools(tools...),
		agent.WithFlowRules(flowRules...),
		agent.WithStructuredOutput(&UserStatusOutput{}),
		agent.WithSessionStore(agent.NewInMemorySessionStore()),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create onboarding agent: %v", err)
	}

	log.Printf("✅ Onboarding agent '%s' created successfully", onboardingAgent.Name())
	log.Printf("📋 Registered %d flow rules with various condition types", len(flowRules))

	// Test scenarios to validate different conditions
	testScenarios := []string{
		"Hi! I want to sign up for your service.",                                    // Should trigger basic_info stage condition
		"My name is John Smith",                                                      // Basic info collected
		"I don't want to give my contact details yet",                               // Should trigger missing contact condition
		"Okay fine, my email is john.smith@email.com",                               // Partial contact info
		"My phone is 0123456789",                                                     // Contact info complete
		"I like reading, cooking, and traveling",                                     // Should collect preferences
		"Actually, let me add that I also enjoy photography and hiking",             // Additional conversation (testing message count condition)
		"Is that everything? Am I done now?",                                         // Final validation
	}

	sessionID := fmt.Sprintf("onboarding-%d", time.Now().Unix())
	log.Printf("🆔 Session ID: %s", sessionID)

	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🧪 Condition Testing & Flow Rules Demo")
	fmt.Println("Testing various condition types and flow rule triggers...")
	fmt.Println(strings.Repeat("=", 80))

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Test %d/%d\n", i+1, len(testScenarios))
		fmt.Printf("👤 User: %s\n", scenario)

		log.Printf("REQUEST[%d]: Processing user input", i+1)
		start := time.Now()

		response, structuredOutput, err := onboardingAgent.Chat(ctx, sessionID, scenario)
		if err != nil {
			log.Printf("❌ ERROR[%d]: %v", i+1, err)
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		duration := time.Since(start)
		log.Printf("RESPONSE[%d]: Duration: %.3fs", i+1, duration.Seconds())

		fmt.Printf("🤖 Agent: %s\n", response.Content)

		// Display structured output if available
		if structuredOutput != nil {
			if userStatus, ok := structuredOutput.(*UserStatusOutput); ok {
				fmt.Printf("📊 Status: Stage=%s, Missing=%v, Complete=%v\n",
					userStatus.CompletionStage, userStatus.MissingFields, userStatus.CompletionFlag)
				log.Printf("STRUCTURED[%d]: Stage=%s, Missing=%v, Complete=%v",
					i+1, userStatus.CompletionStage, userStatus.MissingFields, userStatus.CompletionFlag)
			}
		}

		// Add a small delay between requests
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ Condition testing demo completed!")
	fmt.Println("📋 Flow rules and conditions were tested across different scenarios:")
	fmt.Println("   • Missing fields conditions")
	fmt.Println("   • Completion stage conditions") 
	fmt.Println("   • Message count conditions")
	fmt.Println("   • Custom validation logic")
	fmt.Println("   • Dynamic instruction updates")
	fmt.Println("   • Tool recommendations")
	fmt.Println(strings.Repeat("=", 80))
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }