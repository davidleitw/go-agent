package conditions

import (
	"context"
	"fmt"
	"strings"
)

// And combines multiple conditions with AND logic
func And(conditions ...Condition) Condition {
	return &andCondition{
		conditions: conditions,
	}
}

// Or combines multiple conditions with OR logic
func Or(conditions ...Condition) Condition {
	return &orCondition{
		conditions: conditions,
	}
}

// Not negates a condition
func Not(condition Condition) Condition {
	return &notCondition{
		condition: condition,
	}
}

// Convenience methods with fluent interface
type ConditionBuilder struct {
	condition Condition
}

// NewBuilder creates a new condition builder from a condition
func NewBuilder(condition Condition) *ConditionBuilder {
	return &ConditionBuilder{condition: condition}
}

// And chains this condition with another using AND logic
func (cb *ConditionBuilder) And(other Condition) *ConditionBuilder {
	return &ConditionBuilder{
		condition: And(cb.condition, other),
	}
}

// Or chains this condition with another using OR logic
func (cb *ConditionBuilder) Or(other Condition) *ConditionBuilder {
	return &ConditionBuilder{
		condition: Or(cb.condition, other),
	}
}

// Build returns the final condition
func (cb *ConditionBuilder) Build() Condition {
	return cb.condition
}

// Add fluent interface methods to basic conditions
func (c *missingFieldsCondition) And(other Condition) *ConditionBuilder {
	return NewBuilder(c).And(other)
}

func (c *missingFieldsCondition) Or(other Condition) *ConditionBuilder {
	return NewBuilder(c).Or(other)
}

func (c *containsCondition) And(other Condition) *ConditionBuilder {
	return NewBuilder(c).And(other)
}

func (c *containsCondition) Or(other Condition) *ConditionBuilder {
	return NewBuilder(c).Or(other)
}

func (c *messageCountCondition) And(other Condition) *ConditionBuilder {
	return NewBuilder(c).And(other)
}

func (c *messageCountCondition) Or(other Condition) *ConditionBuilder {
	return NewBuilder(c).Or(other)
}

// =====================================================
// Logic Condition Implementations
// =====================================================

type andCondition struct {
	conditions []Condition
}

func (c *andCondition) Name() string {
	names := make([]string, len(c.conditions))
	for i, cond := range c.conditions {
		names[i] = cond.Name()
	}
	return fmt.Sprintf("and(%s)", strings.Join(names, ","))
}

func (c *andCondition) Description() string {
	descriptions := make([]string, len(c.conditions))
	for i, cond := range c.conditions {
		descriptions[i] = cond.Description()
	}
	return fmt.Sprintf("All of: [%s]", strings.Join(descriptions, ", "))
}

func (c *andCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	for _, cond := range c.conditions {
		result, err := cond.Evaluate(ctx, session, data)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}
	return true, nil
}

type orCondition struct {
	conditions []Condition
}

func (c *orCondition) Name() string {
	names := make([]string, len(c.conditions))
	for i, cond := range c.conditions {
		names[i] = cond.Name()
	}
	return fmt.Sprintf("or(%s)", strings.Join(names, ","))
}

func (c *orCondition) Description() string {
	descriptions := make([]string, len(c.conditions))
	for i, cond := range c.conditions {
		descriptions[i] = cond.Description()
	}
	return fmt.Sprintf("Any of: [%s]", strings.Join(descriptions, ", "))
}

func (c *orCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	for _, cond := range c.conditions {
		result, err := cond.Evaluate(ctx, session, data)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

type notCondition struct {
	condition Condition
}

func (c *notCondition) Name() string {
	return fmt.Sprintf("not(%s)", c.condition.Name())
}

func (c *notCondition) Description() string {
	return fmt.Sprintf("Not: %s", c.condition.Description())
}

func (c *notCondition) Evaluate(ctx context.Context, session Session, data map[string]interface{}) (bool, error) {
	result, err := c.condition.Evaluate(ctx, session, data)
	if err != nil {
		return false, err
	}
	return !result, nil
}