package agent

import (
	"context"
	"fmt"
)

// Condition defines a reusable logical condition for flow control.
// Conditions must be stateless and thread-safe.
type Condition interface {
	// Name returns the unique name of this condition
	Name() string
	
	// Description returns a human-readable description of what this condition checks
	Description() string
	
	// Evaluate assesses the condition based on session state and context data
	// The data parameter typically contains tool results or parsed user intent
	Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error)
}

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

// FlowRuleBuilder provides a fluent API for creating FlowRule instances
type FlowRuleBuilder struct {
	name        string
	description string
	condition   Condition
	action      FlowAction
	priority    int
	enabled     bool
	err         error
}

// NewFlowRule creates a new FlowRuleBuilder with the given name and condition
func NewFlowRule(name string, condition Condition) *FlowRuleBuilder {
	if name == "" {
		return &FlowRuleBuilder{err: fmt.Errorf("flow rule name cannot be empty")}
	}
	if condition == nil {
		return &FlowRuleBuilder{err: fmt.Errorf("condition cannot be nil")}
	}
	
	return &FlowRuleBuilder{
		name:      name,
		condition: condition,
		enabled:   true, // Default to enabled
	}
}

// WithDescription sets the rule's description
func (b *FlowRuleBuilder) WithDescription(description string) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.description = description
	return b
}

// WithPriority sets the rule's priority (higher = evaluated earlier)
func (b *FlowRuleBuilder) WithPriority(priority int) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.priority = priority
	return b
}

// WithAction sets the action to take when the condition is met
func (b *FlowRuleBuilder) WithAction(action FlowAction) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action = action
	return b
}

// WithNewInstructions sets the new instructions template
func (b *FlowRuleBuilder) WithNewInstructions(template string) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.NewInstructionsTemplate = template
	return b
}

// WithRecommendedTools sets the recommended tools for the next turn
func (b *FlowRuleBuilder) WithRecommendedTools(toolNames ...string) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.RecommendedToolNames = toolNames
	return b
}

// WithNotification enables notification and sets details
func (b *FlowRuleBuilder) WithNotification(details map[string]interface{}) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.TriggerNotification = true
	b.action.NotificationDetails = details
	return b
}

// WithNextAgent sets the next agent to transition to
func (b *FlowRuleBuilder) WithNextAgent(agentName string) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.NextAgentName = agentName
	return b
}

// WithStopExecution makes the rule stop further processing
func (b *FlowRuleBuilder) WithStopExecution() *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.StopExecution = true
	return b
}

// WithSystemMessage adds a system message when the rule triggers
func (b *FlowRuleBuilder) WithSystemMessage(message string) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.AddSystemMessage = message
	return b
}

// WithClearHistory makes the rule clear session history
func (b *FlowRuleBuilder) WithClearHistory() *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.action.ClearHistory = true
	return b
}

// WithModelSettings sets model settings for the next turn
func (b *FlowRuleBuilder) WithModelSettings(settings *ModelSettings) *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	if settings != nil {
		if err := settings.Validate(); err != nil {
			b.err = fmt.Errorf("invalid model settings: %w", err)
			return b
		}
	}
	b.action.SetModelSettings = settings
	return b
}

// Disable disables the rule
func (b *FlowRuleBuilder) Disable() *FlowRuleBuilder {
	if b.err != nil {
		return b
	}
	b.enabled = false
	return b
}

// Build creates the FlowRule instance
func (b *FlowRuleBuilder) Build() (FlowRule, error) {
	if b.err != nil {
		return FlowRule{}, b.err
	}
	
	return FlowRule{
		Name:        b.name,
		Description: b.description,
		Condition:   b.condition,
		Action:      b.action,
		Priority:    b.priority,
		Enabled:     b.enabled,
	}, nil
}

// Validate checks if the FlowRule is valid
func (fr *FlowRule) Validate() error {
	if fr.Name == "" {
		return fmt.Errorf("flow rule name cannot be empty")
	}
	if fr.Condition == nil {
		return fmt.Errorf("flow rule condition cannot be nil")
	}
	if fr.Action.SetModelSettings != nil {
		if err := fr.Action.SetModelSettings.Validate(); err != nil {
			return fmt.Errorf("invalid model settings in flow action: %w", err)
		}
	}
	return nil
}

// Evaluate evaluates the rule's condition
func (fr *FlowRule) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	if !fr.Enabled {
		return false, nil
	}
	return fr.Condition.Evaluate(ctx, session, data)
}

// Apply applies the rule's action based on provided context data
func (fr *FlowAction) Apply(ctx context.Context, session Session, data map[string]interface{}) error {
	// Add system message if specified
	if fr.AddSystemMessage != "" {
		message := fr.AddSystemMessage
		// Apply template substitution if needed
		if len(data) > 0 {
			message = applyTemplate(message, data)
		}
		session.AddSystemMessage(message)
	}
	
	// Clear history if requested
	if fr.ClearHistory {
		session.Clear()
	}
	
	return nil
}

// Common condition implementations

// AlwaysCondition always evaluates to true
type AlwaysCondition struct {
	name string
}

func NewAlwaysCondition(name string) *AlwaysCondition {
	return &AlwaysCondition{name: name}
}

func (c *AlwaysCondition) Name() string {
	return c.name
}

func (c *AlwaysCondition) Description() string {
	return "Always evaluates to true"
}

func (c *AlwaysCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	return true, nil
}

// NeverCondition always evaluates to false
type NeverCondition struct {
	name string
}

func NewNeverCondition(name string) *NeverCondition {
	return &NeverCondition{name: name}
}

func (c *NeverCondition) Name() string {
	return c.name
}

func (c *NeverCondition) Description() string {
	return "Always evaluates to false"
}

func (c *NeverCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	return false, nil
}

// MessageCountCondition checks if the session has a certain number of messages
type MessageCountCondition struct {
	name      string
	minCount  int
	maxCount  int
	operator  string // "eq", "gt", "lt", "gte", "lte", "between"
}

func NewMessageCountCondition(name string, operator string, count int) *MessageCountCondition {
	return &MessageCountCondition{
		name:     name,
		operator: operator,
		minCount: count,
	}
}

func NewMessageCountBetweenCondition(name string, minCount, maxCount int) *MessageCountCondition {
	return &MessageCountCondition{
		name:     name,
		operator: "between",
		minCount: minCount,
		maxCount: maxCount,
	}
}

func (c *MessageCountCondition) Name() string {
	return c.name
}

func (c *MessageCountCondition) Description() string {
	switch c.operator {
	case "eq":
		return fmt.Sprintf("Checks if message count equals %d", c.minCount)
	case "gt":
		return fmt.Sprintf("Checks if message count is greater than %d", c.minCount)
	case "lt":
		return fmt.Sprintf("Checks if message count is less than %d", c.minCount)
	case "gte":
		return fmt.Sprintf("Checks if message count is greater than or equal to %d", c.minCount)
	case "lte":
		return fmt.Sprintf("Checks if message count is less than or equal to %d", c.minCount)
	case "between":
		return fmt.Sprintf("Checks if message count is between %d and %d", c.minCount, c.maxCount)
	default:
		return "Checks message count"
	}
}

func (c *MessageCountCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	count := len(session.Messages())
	
	switch c.operator {
	case "eq":
		return count == c.minCount, nil
	case "gt":
		return count > c.minCount, nil
	case "lt":
		return count < c.minCount, nil
	case "gte":
		return count >= c.minCount, nil
	case "lte":
		return count <= c.minCount, nil
	case "between":
		return count >= c.minCount && count <= c.maxCount, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", c.operator)
	}
}

// DataKeyExistsCondition checks if a specific key exists in the context data
type DataKeyExistsCondition struct {
	name string
	key  string
}

func NewDataKeyExistsCondition(name, key string) *DataKeyExistsCondition {
	return &DataKeyExistsCondition{
		name: name,
		key:  key,
	}
}

func (c *DataKeyExistsCondition) Name() string {
	return c.name
}

func (c *DataKeyExistsCondition) Description() string {
	return fmt.Sprintf("Checks if key '%s' exists in context data", c.key)
}

func (c *DataKeyExistsCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	_, exists := data[c.key]
	return exists, nil
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