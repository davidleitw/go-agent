package conditions

// Any creates a condition that triggers when ANY of the specified fields are missing (OR logic)
func Any(fields ...string) Condition {
	if len(fields) == 1 {
		return Missing(fields[0])
	}
	
	conditions := make([]Condition, len(fields))
	for i, field := range fields {
		conditions[i] = Missing(field)
	}
	return Or(conditions...)
}

// All creates a condition that triggers when ALL of the specified fields are missing (AND logic)
func All(fields ...string) Condition {
	if len(fields) == 1 {
		return Missing(fields[0])
	}
	
	conditions := make([]Condition, len(fields))
	for i, field := range fields {
		conditions[i] = Missing(field)
	}
	return And(conditions...)
}

// Either creates a condition that triggers when either condition is true (alias for Or)
func Either(cond1, cond2 Condition) Condition {
	return Or(cond1, cond2)
}

// Both creates a condition that triggers when both conditions are true (alias for And)
func Both(cond1, cond2 Condition) Condition {
	return And(cond1, cond2)
}

// None creates a condition that triggers when none of the conditions are true
func None(conditions ...Condition) Condition {
	return Not(Or(conditions...))
}