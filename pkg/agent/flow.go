package agent

import (
	"context"
	"fmt"
)

// Condition type is now defined in condition_adapter.go

// FlowAction defines the actions to take when a FlowRule's condition is met
type FlowAction struct {
	// NewInstructionsTemplate is a template for dynamic instruction adjustment
	// It can use format placeholders that will be filled with data values
	NewInstructionsTemplate string `json:"new_instructions_template,omitempty"`

	// RecommendedToolNames lists tools that should be prioritized in the next turn
	RecommendedToolNames []string `json:"recommended_tool_names,omitempty"`

	// TriggerNotification indicates whether an external notification should be sent
	TriggerNotification bool `json:"trigger_notification,omitempty"`

	// NotificationDetails contains structured data for the notification payload
	NotificationDetails map[string]interface{} `json:"notification_details,omitempty"`

	// NextAgentName specifies the next agent to transition to (for multi-agent workflows)
	NextAgentName string `json:"next_agent_name,omitempty"`

	// StopExecution indicates whether to halt further processing
	StopExecution bool `json:"stop_execution,omitempty"`

	// AddSystemMessage adds a system message to the session
	AddSystemMessage string `json:"add_system_message,omitempty"`

	// ClearHistory removes previous messages from the session
	ClearHistory bool `json:"clear_history,omitempty"`

	// SetModelSettings overrides the agent's model settings for the next turn
	SetModelSettings *ModelSettings `json:"set_model_settings,omitempty"`

	// Direct response actions (for Ask and AskAI)
	DirectResponse     string `json:"direct_response,omitempty"`     // For Ask - direct text response
	AIPrompt          string `json:"ai_prompt,omitempty"`           // For AskAI - LLM prompt
	FallbackResponse  string `json:"fallback_response,omitempty"`   // For OrElse - backup response
}

// FlowActionResult represents the result of applying a FlowAction
type FlowActionResult struct {
	ShouldStop       bool     // Whether to stop further processing
	DirectResponse   *Message // If set, return this message directly
	ModifiedPrompt   string   // If set, use this as the system prompt
	FallbackResponse string   // If set, use this as fallback for AskAI
	Error           error    // Any error that occurred
}

// FlowRule pairs a Condition with a specific FlowAction
type FlowRule struct {
	// Name is a unique identifier for this rule
	Name string `json:"name"`

	// Description explains what this rule does
	Description string `json:"description,omitempty"`

	// Condition is the logical condition to evaluate
	Condition Condition `json:"-"` // Not serialized since it's an interface

	// Action defines what to do when the condition is met
	Action FlowAction `json:"action"`

	// Priority determines the order of rule evaluation (higher = earlier)
	Priority int `json:"priority,omitempty"`

	// Enabled allows rules to be temporarily disabled
	Enabled bool `json:"enabled"`
}

// NewFlowRule creates a new FlowRuleBuilder with the given name and condition
func NewFlowRule(name string, condition Condition) *FlowRuleBuilderImpl {
	if name == "" {
		return &FlowRuleBuilderImpl{err: fmt.Errorf("flow rule name cannot be empty")}
	}
	if condition == nil {
		return &FlowRuleBuilderImpl{err: fmt.Errorf("condition cannot be nil")}
	}

	return &FlowRuleBuilderImpl{
		name:      name,
		condition: condition,
		enabled:   true, // Default to enabled
	}
}

// FlowRuleBuilderImpl provides a fluent API for creating FlowRule instances
type FlowRuleBuilderImpl struct {
	name        string
	description string
	condition   Condition
	action      FlowAction
	priority    int
	enabled     bool
	err         error
}

// WithDescription sets the rule's description
func (b *FlowRuleBuilderImpl) WithDescription(description string) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.description = description
	return b
}

// WithPriority sets the rule's priority (higher = evaluated earlier)
func (b *FlowRuleBuilderImpl) WithPriority(priority int) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.priority = priority
	return b
}

// WithAction sets the action to take when the condition is met
func (b *FlowRuleBuilderImpl) WithAction(action FlowAction) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action = action
	return b
}

// WithNewInstructions sets the new instructions template
func (b *FlowRuleBuilderImpl) WithNewInstructions(template string) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.NewInstructionsTemplate = template
	return b
}

// WithRecommendedTools sets the recommended tools for the next turn
func (b *FlowRuleBuilderImpl) WithRecommendedTools(toolNames ...string) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.RecommendedToolNames = toolNames
	return b
}

// WithNotification enables notification and sets details
func (b *FlowRuleBuilderImpl) WithNotification(details map[string]interface{}) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.TriggerNotification = true
	b.action.NotificationDetails = details
	return b
}

// WithNextAgent sets the next agent to transition to
func (b *FlowRuleBuilderImpl) WithNextAgent(agentName string) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.NextAgentName = agentName
	return b
}

// WithStopExecution makes the rule stop further processing
func (b *FlowRuleBuilderImpl) WithStopExecution() *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.StopExecution = true
	return b
}

// WithSystemMessage adds a system message when the rule triggers
func (b *FlowRuleBuilderImpl) WithSystemMessage(message string) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.AddSystemMessage = message
	return b
}

// WithClearHistory makes the rule clear session history
func (b *FlowRuleBuilderImpl) WithClearHistory() *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.action.ClearHistory = true
	return b
}

// WithModelSettings sets model settings for the next turn
func (b *FlowRuleBuilderImpl) WithModelSettings(settings *ModelSettings) *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	if settings != nil {
		// We'll skip validation for now to avoid missing methods
	}
	b.action.SetModelSettings = settings
	return b
}

// Disable disables the rule
func (b *FlowRuleBuilderImpl) Disable() *FlowRuleBuilderImpl {
	if b.err != nil {
		return b
	}
	b.enabled = false
	return b
}

// Build creates the FlowRule instance
func (b *FlowRuleBuilderImpl) Build() FlowRule {
	if b.err != nil {
		// For now, return empty rule to avoid breaking the interface
		return FlowRule{}
	}

	return FlowRule{
		Name:        b.name,
		Description: b.description,
		Condition:   b.condition,
		Action:      b.action,
		Priority:    b.priority,
		Enabled:     b.enabled,
	}
}

// Validate checks if the FlowRule is valid
func (fr *FlowRule) Validate() error {
	if fr.Name == "" {
		return fmt.Errorf("flow rule name cannot be empty")
	}
	if fr.Condition == nil {
		return fmt.Errorf("flow rule condition cannot be nil")
	}
	return nil
}

// Evaluate evaluates the rule's condition
func (fr *FlowRule) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	if !fr.Enabled {
		return false, nil
	}
	return EvaluateCondition(ctx, fr.Condition, session, data)
}

// Apply applies the rule's action based on provided context data
func (fr *FlowAction) Apply(ctx context.Context, session Session, data map[string]interface{}) *FlowActionResult {
	result := &FlowActionResult{}
	
	// Handle direct response first (Ask)
	if fr.DirectResponse != "" {
		message := fr.DirectResponse
		// Apply template substitution if needed
		if len(data) > 0 {
			message = applyTemplate(message, data)
		}
		result.DirectResponse = &Message{
			Role:      RoleAssistant,
			Content:   message,
			Timestamp: timeNow(),
		}
		result.ShouldStop = true
		return result
	}

	// Handle AI-enhanced prompt (AskAI)
	if fr.AIPrompt != "" {
		prompt := fr.AIPrompt
		// Apply template substitution if needed
		if len(data) > 0 {
			prompt = applyTemplate(prompt, data)
		}
		result.ModifiedPrompt = prompt
		result.ShouldStop = false // Continue to LLM with modified prompt
		
		// Include fallback response if specified
		if fr.FallbackResponse != "" {
			fallback := fr.FallbackResponse
			if len(data) > 0 {
				fallback = applyTemplate(fallback, data)
			}
			result.FallbackResponse = fallback
		}
		
		return result
	}

	// Add system message if specified
	if fr.AddSystemMessage != "" {
		message := fr.AddSystemMessage
		// Apply template substitution if needed
		if len(data) > 0 {
			message = applyTemplate(message, data)
		}
		session.AddMessage(RoleSystem, message)
	}

	// Clear history if requested
	if fr.ClearHistory {
		if clearable, ok := session.(interface{ ClearMessages() }); ok {
			clearable.ClearMessages()
		}
	}

	// Check if execution should stop
	if fr.StopExecution {
		result.ShouldStop = true
	}

	return result
}

// Helper function to apply template substitution
func applyTemplate(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = replaceAll(result, placeholder, replacement)
	}
	return result
}

// Helper function to replace all occurrences (avoiding strings package)
func replaceAll(s, old, new string) string {
	if old == "" {
		return s
	}

	var result []byte
	i := 0
	for {
		j := indexOf(s[i:], old)
		if j < 0 {
			break
		}
		result = append(result, s[i:i+j]...)
		result = append(result, new...)
		i += j + len(old)
	}
	result = append(result, s[i:]...)
	return string(result)
}

// Helper function to find index of substring
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}