package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/joho/godotenv"
)

// OnboardingAgent implements a custom Agent with advanced condition testing and flow rules
type OnboardingAgent struct {
	name         string
	description  string
	instructions string
	model        string
	settings     *agent.ModelSettings
	tools        []agent.Tool
	flowRules    []agent.FlowRule
	outputType   agent.OutputType
	chatModel    agent.ChatModel
	sessionStore agent.SessionStore
	maxTurns     int
	toolTimeout  time.Duration

	// Custom fields for condition testing
	conditionEvaluations map[string]int
	ruleActivations      map[string]int
	userData             map[string]any
}

// NewOnboardingAgent creates a new onboarding agent with advanced condition testing
func NewOnboardingAgent(config OnboardingAgentConfig) *OnboardingAgent {
	return &OnboardingAgent{
		name:                 config.Name,
		description:          config.Description,
		instructions:         config.Instructions,
		model:                config.Model,
		settings:             config.Settings,
		tools:                config.Tools,
		flowRules:            config.FlowRules,
		outputType:           config.OutputType,
		chatModel:            config.ChatModel,
		sessionStore:         config.SessionStore,
		maxTurns:             config.MaxTurns,
		toolTimeout:          config.ToolTimeout,
		conditionEvaluations: make(map[string]int),
		ruleActivations:      make(map[string]int),
		userData:             make(map[string]any),
	}
}

type OnboardingAgentConfig struct {
	Name         string
	Description  string
	Instructions string
	Model        string
	Settings     *agent.ModelSettings
	Tools        []agent.Tool
	FlowRules    []agent.FlowRule
	OutputType   agent.OutputType
	ChatModel    agent.ChatModel
	SessionStore agent.SessionStore
	MaxTurns     int
	ToolTimeout  time.Duration
}

// Agent interface implementation
func (a *OnboardingAgent) Name() string {
	return a.name
}

func (a *OnboardingAgent) Description() string {
	return a.description
}

func (a *OnboardingAgent) GetOutputType() agent.OutputType {
	return a.outputType
}

func (a *OnboardingAgent) GetTools() []agent.Tool {
	return a.tools
}

func (a *OnboardingAgent) GetFlowRules() []agent.FlowRule {
	return a.flowRules
}

// Chat implements sophisticated condition testing and flow rule management
func (a *OnboardingAgent) Chat(ctx context.Context, session agent.Session, userInput string) (*agent.Message, any, error) {
	// Add user message to session
	userMessage := agent.NewUserMessage(userInput)
	session.AddMessage(userMessage)

	// Prepare enhanced instructions with condition context
	enhancedInstructions := a.buildEnhancedInstructions()

	// Prepare messages for chat model
	messages := session.Messages()
	systemMessage := agent.NewSystemMessage(enhancedInstructions)
	allMessages := []agent.Message{systemMessage}
	allMessages = append(allMessages, messages...)

	// Execute conversation loop with advanced condition testing
	for turn := 0; turn < a.maxTurns; turn++ {
		log.Printf("🔄 TURN[%d]: Starting conversation turn with condition evaluation", turn+1)

		// Evaluate flow rules with detailed logging
		flowData := a.createFlowData(userInput, len(messages))
		a.evaluateFlowRules(ctx, session, flowData)

		// Update messages after flow rule application
		messages = session.Messages()
		allMessages = []agent.Message{systemMessage}
		allMessages = append(allMessages, messages...)

		// Call chat model with current context
		response, err := a.chatModel.GenerateChatCompletion(
			ctx,
			allMessages,
			a.model,
			a.settings,
			a.tools,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate chat completion: %w", err)
		}

		// Add response to session
		session.AddMessage(*response)

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			log.Printf("✅ TURN[%d]: Conversation completed without tool calls", turn+1)

			// Generate structured output
			var structuredOutput any
			if a.outputType != nil && response.Content != "" {
				instance := a.outputType.NewInstance()
				if err := json.Unmarshal([]byte(response.Content), instance); err == nil {
					if a.outputType.Validate(instance) == nil {
						structuredOutput = instance
					}
				}
			}

			// Save session
			if err := a.sessionStore.Save(ctx, session); err != nil {
				log.Printf("⚠️  Warning: failed to save session: %v", err)
			}

			a.logConditionStats()
			return response, structuredOutput, nil
		}

		// Execute tool calls with condition awareness
		if err := a.executeToolCalls(ctx, session, response.ToolCalls); err != nil {
			return nil, nil, fmt.Errorf("failed to execute tool calls: %w", err)
		}

		// Update messages for next iteration
		messages = session.Messages()
		allMessages = []agent.Message{systemMessage}
		allMessages = append(allMessages, messages...)

		log.Printf("📊 TURN[%d]: Condition evaluations: %v", turn+1, a.conditionEvaluations)
		log.Printf("📊 TURN[%d]: Rule activations: %v", turn+1, a.ruleActivations)
	}

	// If we've exhausted max turns, return the last response
	messages = session.Messages()
	if len(messages) > 0 {
		lastMessage := &messages[len(messages)-1]
		if lastMessage.Role == agent.RoleAssistant {
			a.logConditionStats()
			return lastMessage, nil, fmt.Errorf("reached maximum turns (%d) without completion", a.maxTurns)
		}
	}

	a.logConditionStats()
	return nil, nil, fmt.Errorf("conversation ended unexpectedly after %d turns", a.maxTurns)
}

// buildEnhancedInstructions creates context-aware instructions with condition info
func (a *OnboardingAgent) buildEnhancedInstructions() string {
	base := a.instructions

	// Add condition evaluation context
	if len(a.conditionEvaluations) > 0 {
		base += "\n\nCondition Evaluation Context:\n"
		for conditionName, count := range a.conditionEvaluations {
			base += fmt.Sprintf("- %s: evaluated %d times\n", conditionName, count)
		}
	}

	// Add rule activation context
	if len(a.ruleActivations) > 0 {
		base += "\nRule Activation Context:\n"
		for ruleName, count := range a.ruleActivations {
			base += fmt.Sprintf("- %s: activated %d times\n", ruleName, count)
		}
	}

	// Add user data context
	if len(a.userData) > 0 {
		base += "\nCurrent User Data:\n"
		for key, value := range a.userData {
			base += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	// Add advanced condition testing guidelines
	base += `
Advanced Condition Testing Guidelines:
1. Monitor condition evaluations for pattern analysis
2. Track rule activations for flow optimization
3. Maintain user data state across conversations
4. Provide detailed feedback on condition states
5. Adapt responses based on condition history
`

	return base
}

// createFlowData creates context data for flow rule evaluation
func (a *OnboardingAgent) createFlowData(userInput string, messageCount int) map[string]any {
	flowData := make(map[string]any)
	flowData["userInput"] = userInput
	flowData["messageCount"] = messageCount

	// Add current user data to flow context
	for key, value := range a.userData {
		flowData[key] = value
	}

	return flowData
}

// evaluateFlowRules evaluates all flow rules with detailed condition tracking
func (a *OnboardingAgent) evaluateFlowRules(ctx context.Context, session agent.Session, flowData map[string]any) {
	log.Printf("🎯 Evaluating %d flow rules", len(a.flowRules))

	for _, rule := range a.flowRules {
		// Track condition evaluations
		a.conditionEvaluations[rule.Condition.Name()]++

		log.Printf("🔍 CONDITION[%s]: Evaluating...", rule.Condition.Name())

		shouldTrigger, err := rule.Evaluate(ctx, session, flowData)
		if err != nil {
			log.Printf("❌ CONDITION[%s]: Evaluation failed: %v", rule.Condition.Name(), err)
			continue
		}

		if shouldTrigger {
			log.Printf("✅ CONDITION[%s]: Triggered! Applying rule '%s'", rule.Condition.Name(), rule.Name)

			// Track rule activations
			a.ruleActivations[rule.Name]++

			// Apply flow rule actions
			if err := rule.Action.Apply(ctx, session, flowData); err != nil {
				log.Printf("❌ RULE[%s]: Action failed: %v", rule.Name, err)
			} else {
				log.Printf("✅ RULE[%s]: Action applied successfully", rule.Name)
			}
		} else {
			log.Printf("⏸️  CONDITION[%s]: Not triggered", rule.Condition.Name())
		}
	}
}

// executeToolCalls handles tool execution with condition awareness
func (a *OnboardingAgent) executeToolCalls(ctx context.Context, session agent.Session, toolCalls []agent.ToolCall) error {
	log.Printf("🔧 Executing %d tool calls with condition awareness", len(toolCalls))

	for i, toolCall := range toolCalls {
		// Find matching tool
		var matchingTool agent.Tool
		for _, tool := range a.tools {
			if tool.Name() == toolCall.Function.Name {
				matchingTool = tool
				break
			}
		}

		if matchingTool == nil {
			log.Printf("❌ Tool '%s' not found", toolCall.Function.Name)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: tool '%s' not found", toolCall.Function.Name),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Parse tool arguments
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Printf("❌ Invalid arguments for tool '%s': %v", toolCall.Function.Name, err)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: invalid arguments - %v", err),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Execute tool with timeout
		toolCtx, cancel := context.WithTimeout(ctx, a.toolTimeout)

		log.Printf("🔧 TOOL[%d/%d]: Executing %s", i+1, len(toolCalls), toolCall.Function.Name)
		start := time.Now()

		result, err := matchingTool.Execute(toolCtx, args)
		cancel()

		duration := time.Since(start)

		if err != nil {
			log.Printf("❌ TOOL[%d/%d]: %s failed in %v: %v", i+1, len(toolCalls), toolCall.Function.Name, duration, err)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: %v", err),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Convert result to JSON string
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Printf("❌ TOOL[%d/%d]: Failed to serialize result for %s: %v", i+1, len(toolCalls), toolCall.Function.Name, err)
			errorMsg := agent.NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				fmt.Sprintf("Error: failed to serialize result - %v", err),
			)
			session.AddMessage(errorMsg)
			continue
		}

		// Add successful tool result
		toolMsg := agent.NewToolMessage(
			toolCall.ID,
			toolCall.Function.Name,
			string(resultJSON),
		)
		session.AddMessage(toolMsg)

		// Update user data based on tool results
		a.updateUserData(toolCall.Function.Name, result)

		log.Printf("✅ TOOL[%d/%d]: %s completed in %v", i+1, len(toolCalls), toolCall.Function.Name, duration)
	}

	return nil
}

// updateUserData updates internal user data based on tool results
func (a *OnboardingAgent) updateUserData(toolName string, result any) {
	if resultMap, ok := result.(map[string]any); ok {
		switch toolName {
		case "collect_user_info":
			if fieldType, ok := resultMap["field_type"].(string); ok {
				if value, ok := resultMap["value"].(string); ok {
					if isValid, ok := resultMap["is_valid"].(bool); ok && isValid {
						a.userData[fieldType] = value
						log.Printf("📝 USER_DATA: Updated %s = %s", fieldType, value)
					}
				}
			}
		case "validate_completion":
			if stage, ok := resultMap["completion_stage"].(string); ok {
				a.userData["completion_stage"] = stage
			}
			if missing, ok := resultMap["missing_fields"].([]any); ok {
				a.userData["missing_fields"] = missing
			}
		}
	}
}

// logConditionStats logs detailed condition and rule statistics
func (a *OnboardingAgent) logConditionStats() {
	log.Println("📊 CONDITION TESTING STATISTICS:")
	log.Printf("📊 Total conditions evaluated: %d", len(a.conditionEvaluations))
	for condition, count := range a.conditionEvaluations {
		log.Printf("📊   - %s: %d evaluations", condition, count)
	}

	log.Printf("📊 Total rules activated: %d", len(a.ruleActivations))
	for rule, count := range a.ruleActivations {
		log.Printf("📊   - %s: %d activations", rule, count)
	}

	log.Printf("📊 User data collected: %d fields", len(a.userData))
	for key, value := range a.userData {
		log.Printf("📊   - %s: %v", key, value)
	}
}

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

func (u *UserStatusOutput) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_id": map[string]any{
				"type":        "string",
				"description": "Unique user identifier",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "User's full name",
			},
			"email": map[string]any{
				"type":        "string",
				"description": "User's email address",
			},
			"phone": map[string]any{
				"type":        "string",
				"description": "User's phone number",
			},
			"preferences": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "User's preferences list",
			},
			"completion_stage": map[string]any{
				"type":        "string",
				"description": "Current stage of onboarding",
				"enum":        []string{"basic_info", "contact_details", "preferences", "completed"},
			},
			"missing_fields": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "List of missing required fields",
			},
			"completion_flag": map[string]any{
				"type":        "boolean",
				"description": "Whether onboarding is complete",
			},
			"message": map[string]any{
				"type":        "string",
				"description": "Message to display to user",
			},
		},
		"required": []string{"user_id", "completion_stage", "missing_fields", "completion_flag", "message"},
	}
}

func (u *UserStatusOutput) NewInstance() any {
	return &UserStatusOutput{}
}

func (u *UserStatusOutput) Validate(instance any) error {
	// Basic validation - ensure it's the right type
	if _, ok := instance.(*UserStatusOutput); !ok {
		return fmt.Errorf("instance is not a UserStatusOutput")
	}
	return nil
}

// CollectInfoTool simulates collecting user information
type CollectInfoTool struct{}

func (t *CollectInfoTool) Name() string {
	return "collect_user_info"
}

func (t *CollectInfoTool) Description() string {
	return "Collect and validate user information during onboarding process"
}

func (t *CollectInfoTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"field_type": map[string]any{
				"type":        "string",
				"description": "Type of information to collect",
				"enum":        []string{"name", "email", "phone", "preferences"},
			},
			"value": map[string]any{
				"type":        "string",
				"description": "The collected value",
			},
			"user_id": map[string]any{
				"type":        "string",
				"description": "User identifier",
			},
		},
		"required": []string{"field_type", "value", "user_id"},
	}
}

func (t *CollectInfoTool) Execute(ctx context.Context, args map[string]any) (any, error) {
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

	result := map[string]any{
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

func (t *ValidationTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_data": map[string]any{
				"type":        "object",
				"description": "Current user data to validate",
			},
		},
		"required": []string{"user_data"},
	}
}

func (t *ValidationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	userData, ok := args["user_data"].(map[string]any)
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

	result := map[string]any{
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
	name        string
	fields      []string
	description string
}

func NewMissingFieldsCondition(name string, fields []string) agent.Condition {
	return &MissingFieldsCondition{
		name:        name,
		fields:      fields,
		description: fmt.Sprintf("Checks if any of the fields %v are missing", fields),
	}
}

func (c *MissingFieldsCondition) Name() string {
	return c.name
}

func (c *MissingFieldsCondition) Description() string {
	return c.description
}

func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
	missingFields, exists := data["missing_fields"]
	if !exists {
		return false, nil
	}

	missingList, ok := missingFields.([]any)
	if !ok {
		return false, nil
	}

	// Convert to string slice
	missing := make([]string, 0, len(missingList))
	for _, v := range missingList {
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
	name        string
	stage       string
	description string
}

func NewCompletionStageCondition(name string, stage string) agent.Condition {
	return &CompletionStageCondition{
		name:        name,
		stage:       stage,
		description: fmt.Sprintf("Checks if the current completion stage is '%s'", stage),
	}
}

func (c *CompletionStageCondition) Name() string {
	return c.name
}

func (c *CompletionStageCondition) Description() string {
	return c.description
}

func (c *CompletionStageCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
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
	name        string
	minCount    int
	description string
}

func NewMessageCountCondition(name string, minCount int) agent.Condition {
	return &MessageCountCondition{
		name:        name,
		minCount:    minCount,
		description: fmt.Sprintf("Checks if the message count is at least %d", minCount),
	}
}

func (c *MessageCountCondition) Name() string {
	return c.name
}

func (c *MessageCountCondition) Description() string {
	return c.description
}

func (c *MessageCountCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
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
	chatModel, err := agent.NewOpenAIChatModel(apiKey)
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
	longConversationCondition := NewMessageCountCondition("long_conversation", 6)

	// Create flow rules
	flowRules := []agent.FlowRule{}

	// Rule 1: Missing contact information
	contactRule := agent.NewFlowRule("collect-contact-info", missingContactCondition).
		WithDescription("Prompt user for missing contact information").
		WithNewInstructions("The user is missing contact information. Please ask specifically for their email and phone number. Be encouraging and explain why this information is needed.").
		WithRecommendedTools("collect_user_info").
		WithSystemMessage("Contact information collection needed").
		Build()
	flowRules = append(flowRules, contactRule)

	// Rule 2: Missing preferences
	prefsRule := agent.NewFlowRule("collect-preferences", missingPrefsCondition).
		WithDescription("Collect user preferences").
		WithNewInstructions("The user needs to specify their preferences. Ask them about their interests, hobbies, or preferred topics (1-5 items, comma-separated).").
		WithRecommendedTools("collect_user_info", "validate_completion").
		WithSystemMessage("Preferences collection needed").
		Build()
	flowRules = append(flowRules, prefsRule)

	// Rule 3: At basic info stage
	basicInfoRule := agent.NewFlowRule("handle-basic-info", basicInfoStageCondition).
		WithDescription("Handle users at basic info stage").
		WithNewInstructions("The user is just starting. Collect their basic information first (name), then we'll move to contact details.").
		WithRecommendedTools("collect_user_info").
		WithSystemMessage("Starting basic information collection").
		Build()
	flowRules = append(flowRules, basicInfoRule)

	// Rule 4: Long conversation optimization
	longConvRule := agent.NewFlowRule("optimize-long-conversation", longConversationCondition).
		WithDescription("Optimize for long conversations").
		WithNewInstructions("This conversation has been going on for a while. Please provide a summary of what we've collected so far and clearly state what's still needed to complete the onboarding.").
		WithRecommendedTools("validate_completion").
		WithSystemMessage("Long conversation detected - providing summary").
		Build()
	flowRules = append(flowRules, longConvRule)

	// Create custom onboarding agent with advanced condition testing
	log.Println("🤖 Creating custom onboarding agent with advanced condition testing...")
	onboardingAgent := NewOnboardingAgent(OnboardingAgentConfig{
		Name:        "onboarding-agent",
		Description: "An intelligent onboarding agent with advanced condition testing and flow rule management",
		Instructions: `You are an advanced onboarding specialist that helps new users complete their registration process with sophisticated condition testing.

Your goal is to collect the following information:
1. Name (required)
2. Email (required) 
3. Phone number (required)
4. Preferences (required) - 1-5 interests/topics

Advanced Capabilities:
- Sophisticated condition evaluation and testing
- Dynamic flow rule management
- Comprehensive user data tracking
- Adaptive conversation flow based on conditions
- Real-time condition state monitoring

Use the available tools to:
- collect_user_info: Gather and validate user information
- validate_completion: Check if onboarding is complete

Always provide the structured output showing:
- Current user data
- What stage they're at
- What's missing
- Whether they're complete
- A helpful message

Be friendly, encouraging, and clear about what information is needed and why.`,
		Model: "gpt-4",
		Settings: &agent.ModelSettings{
			Temperature: floatPtr(0.7),
			MaxTokens:   intPtr(2000),
		},
		Tools:        tools,
		FlowRules:    flowRules,
		OutputType:   agent.NewStructuredOutputType(&UserStatusOutput{}),
		ChatModel:    chatModel,
		SessionStore: agent.NewInMemorySessionStore(),
		MaxTurns:     10,
		ToolTimeout:  30 * time.Second,
	})

	log.Printf("✅ Onboarding agent '%s' created successfully", onboardingAgent.Name())
	log.Printf("📋 Registered %d flow rules with various condition types", len(flowRules))

	// Test scenarios to validate different conditions
	testScenarios := []string{
		"Hi! I want to sign up for your service.",                       // Should trigger basic_info stage condition
		"My name is John Smith",                                         // Basic info collected
		"I don't want to give my contact details yet",                   // Should trigger missing contact condition
		"Okay fine, my email is john.smith@email.com",                   // Partial contact info
		"My phone is 0123456789",                                        // Contact info complete
		"I like reading, cooking, and traveling",                        // Should collect preferences
		"Actually, let me add that I also enjoy photography and hiking", // Additional conversation (testing message count condition)
		"Is that everything? Am I done now?",                            // Final validation
	}

	sessionID := fmt.Sprintf("onboarding-%d", time.Now().Unix())
	session := agent.NewSession(sessionID)
	log.Printf("🆔 Session ID: %s", sessionID)

	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🧪 Advanced Condition Testing & Flow Rules Demo")
	fmt.Println("Testing sophisticated condition evaluation and flow rule management...")
	fmt.Println(strings.Repeat("=", 80))

	for i, scenario := range testScenarios {
		fmt.Printf("\n🔄 Test %d/%d\n", i+1, len(testScenarios))
		fmt.Printf("👤 User: %s\n", scenario)

		log.Printf("REQUEST[%d]: Processing user input with condition testing", i+1)
		start := time.Now()

		response, structuredOutput, err := onboardingAgent.Chat(ctx, session, scenario)
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
	fmt.Println("✅ Advanced condition testing demo completed!")
	fmt.Println("🎯 The agent successfully demonstrated:")
	fmt.Println("   • Sophisticated condition evaluation")
	fmt.Println("   • Dynamic flow rule management")
	fmt.Println("   • Comprehensive user data tracking")
	fmt.Println("   • Adaptive conversation flow")
	fmt.Println("   • Real-time condition state monitoring")
	fmt.Println("   • Advanced pattern analysis")
	fmt.Println("   • Custom validation logic")
	fmt.Println("   • Tool recommendations")
	fmt.Println("   • Dynamic instruction updates")
	fmt.Println(strings.Repeat("=", 80))
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }
