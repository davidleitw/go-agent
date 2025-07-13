package conditions

import (
	"context"
	"fmt"
	"strings"
)

// Session interface for condition evaluation (minimal interface to avoid circular imports)
type Session interface {
	Messages() []Message
}

// Message represents a message in the conversation
type Message struct {
	Role      string
	Content   string
	Timestamp interface{}
}

// Condition defines a reusable logical condition for flow control
type Condition interface {
	Name() string
	Description() string
	Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error)
}

// Missing creates a condition that triggers when specified fields are missing
func Missing(fields ...string) Condition {
	return &missingFieldsCondition{
		fields: fields,
	}
}

// Contains creates a condition that triggers when user input contains text
func Contains(text string) Condition {
	return &containsCondition{
		text: text,
	}
}

// Count creates a condition that triggers when message count reaches threshold
func Count(minCount int) Condition {
	return &messageCountCondition{
		minCount: minCount,
	}
}

// DataEquals creates a condition that triggers when session data equals value
func DataEquals(key string, value interface{}) Condition {
	return &dataEqualsCondition{
		key:   key,
		value: value,
	}
}

// Func creates a custom condition from a function
func Func(name string, fn func(Session) bool) Condition {
	return &funcCondition{
		name: name,
		fn:   fn,
	}
}

// =====================================================
// Condition Implementations
// =====================================================

type missingFieldsCondition struct {
	fields []string
}

func (c *missingFieldsCondition) Name() string {
	return fmt.Sprintf("missing_fields_%s", strings.Join(c.fields, "_"))
}

func (c *missingFieldsCondition) Description() string {
	return fmt.Sprintf("Checks if any of the fields %v are missing", c.fields)
}

func (c *missingFieldsCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	// Check in session data
	messages := session.Messages()
	collectedData := make(map[string]bool)
	
	// Scan messages for collected information
	for _, msg := range messages {
		content := strings.ToLower(msg.Content)
		for _, field := range c.fields {
			if strings.Contains(content, field) {
				collectedData[field] = true
			}
		}
	}
	
	// Check if any field is missing
	for _, field := range c.fields {
		if !collectedData[field] {
			return true, nil
		}
	}
	
	return false, nil
}

type messageCountCondition struct {
	minCount int
}

func (c *messageCountCondition) Name() string {
	return fmt.Sprintf("message_count_%d", c.minCount)
}

func (c *messageCountCondition) Description() string {
	return fmt.Sprintf("Checks if message count is at least %d", c.minCount)
}

func (c *messageCountCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	return len(session.Messages()) >= c.minCount, nil
}

type containsCondition struct {
	text string
}

func (c *containsCondition) Name() string {
	return fmt.Sprintf("contains_%s", strings.ReplaceAll(c.text, " ", "_"))
}

func (c *containsCondition) Description() string {
	return fmt.Sprintf("Checks if user input contains '%s'", c.text)
}

func (c *containsCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	if userInput, ok := data["userInput"].(string); ok {
		return strings.Contains(strings.ToLower(userInput), strings.ToLower(c.text)), nil
	}
	return false, nil
}

type dataEqualsCondition struct {
	key   string
	value interface{}
}

func (c *dataEqualsCondition) Name() string {
	return fmt.Sprintf("data_%s_equals_%v", c.key, c.value)
}

func (c *dataEqualsCondition) Description() string {
	return fmt.Sprintf("Checks if data key '%s' equals '%v'", c.key, c.value)
}

func (c *dataEqualsCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	if val, ok := data[c.key]; ok {
		return val == c.value, nil
	}
	return false, nil
}

type funcCondition struct {
	name string
	fn   func(Session) bool
}

func (c *funcCondition) Name() string {
	return c.name
}

func (c *funcCondition) Description() string {
	return fmt.Sprintf("Custom function condition: %s", c.name)
}

func (c *funcCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	return c.fn(session), nil
}