package agent

import (
	"context"

	"github.com/davidleitw/go-agent/pkg/conditions"
)

// Condition is an alias for conditions.Condition for backward compatibility
// and to avoid circular imports in flow.go
type Condition = conditions.Condition

// ConditionAdapter adapts between agent.Session and conditions.Session
type conditionSessionAdapter struct {
	session Session
}

func (a *conditionSessionAdapter) Messages() []conditions.Message {
	agentMessages := a.session.Messages()
	conditionMessages := make([]conditions.Message, len(agentMessages))
	
	for i, msg := range agentMessages {
		conditionMessages[i] = conditions.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
	}
	
	return conditionMessages
}

// EvaluateCondition evaluates a conditions.Condition with agent.Session
func EvaluateCondition(ctx context.Context, cond conditions.Condition, session Session, data map[string]interface{}) (bool, error) {
	adapter := &conditionSessionAdapter{session: session}
	return cond.Evaluate(ctx, adapter, data)
}

// Wrapper functions for backward compatibility with existing agent.go code
func WhenMissingFields(fields ...string) Condition {
	return conditions.Missing(fields...)
}

func WhenContains(text string) Condition {
	return conditions.Contains(text)
}

func WhenMessageCount(minCount int) Condition {
	return conditions.Count(minCount)
}

func WhenFunc(name string, fn func(Session) bool) Condition {
	// Adapt the function to work with conditions.Session
	adaptedFn := func(condSess conditions.Session) bool {
		// Convert back to agent.Session - we need to implement a reverse adapter
		// For now, we'll use a simple approach
		if adapter, ok := condSess.(*conditionSessionAdapter); ok {
			return fn(adapter.session)
		}
		// Fallback: this shouldn't happen in normal usage
		return false
	}
	return conditions.Func(name, adaptedFn)
}

func And(conds ...Condition) Condition {
	return conditions.And(conds...)
}

func Or(conds ...Condition) Condition {
	return conditions.Or(conds...)
}

func Not(condition Condition) Condition {
	return conditions.Not(condition)
}