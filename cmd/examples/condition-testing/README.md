# Condition Testing Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![Chinese](https://img.shields.io/badge/README-Chinese-red.svg)](README-zh.md)

This example demonstrates advanced condition validation and flow rule implementation using a user onboarding scenario. It tests various condition types and validates that the flow control system works as expected.

## Overview

The condition testing example showcases:
- **Multiple Condition Types**: Testing different kinds of conditions with various trigger scenarios
- **Flow Rule Orchestration**: Demonstrating how flow rules modify agent behavior dynamically
- **Custom Condition Implementation**: Creating domain-specific conditions beyond built-in types
- **Structured Output Integration**: Using conditions with structured data validation
- **Dynamic Instruction Updates**: Testing how conditions can change agent instructions on-the-fly

## Condition Types Tested

### 1. 🎯 Missing Fields Condition
Tests whether specific data fields are missing from the collected information.

**Implementation**:
```go
type MissingFieldsCondition struct {
    name   string
    fields []string  // Fields to check for absence
}
```

**Use Cases**:
- Missing contact information (email, phone)
- Missing user preferences
- Incomplete profile data

### 2. 📋 Completion Stage Condition
Checks the current stage of the onboarding process.

**Implementation**:
```go
type CompletionStageCondition struct {
    name  string
    stage string  // Target stage to match
}
```

**Stages**:
- `basic_info`: Initial information collection
- `contact_details`: Email and phone collection
- `preferences`: Interest and preference collection
- `completed`: Onboarding finished

### 3. 💬 Message Count Condition
Triggers based on the number of messages in the conversation.

**Implementation**:
```go
type MessageCountCondition struct {
    name     string
    minCount int  // Minimum message threshold
}
```

**Use Cases**:
- Long conversation optimization
- Progress summaries
- Escalation triggers

### 4. 🔍 Data Key Exists Condition (Built-in)
Uses the framework's built-in condition to check for data presence.

**Usage**:
```go
condition := agent.NewDataKeyExistsCondition("check_missing", "missing_fields")
```

## Flow Rules Configuration

### Rule 1: Contact Information Collection
**Condition**: Missing email or phone
**Actions**:
- Update instructions to specifically ask for contact details
- Recommend `collect_user_info` tool
- Provide encouraging messaging

### Rule 2: Preferences Collection
**Condition**: Missing preferences data
**Actions**:
- Ask for user interests and hobbies
- Recommend both collection and validation tools
- Guide preference format (1-5 items, comma-separated)

### Rule 3: Basic Information Handling
**Condition**: User at basic_info stage
**Actions**:
- Focus on name collection first
- Set expectations for next steps
- Provide structured onboarding flow

### Rule 4: Long Conversation Optimization
**Condition**: More than 6 messages exchanged
**Actions**:
- Provide progress summary
- Clearly state remaining requirements
- Recommend completion validation

## Tools Integration

### CollectInfoTool
Gathers and validates user information with field-specific validation rules.

**Validation Logic**:
- **Name**: Minimum 2 characters
- **Email**: Must contain @ and . symbols
- **Phone**: Minimum 10 digits, must start with 0
- **Preferences**: 1-5 comma-separated items

### ValidationTool
Assesses completion status and determines next steps.

**Capabilities**:
- Calculate completion percentage
- Identify missing fields
- Determine current stage
- Provide completion recommendations

## Structured Output

The agent returns a `UserStatusOutput` structure containing:

```json
{
  "user_id": "user-12345",
  "name": "John Smith",
  "email": "john.smith@email.com", 
  "phone": "0123456789",
  "preferences": ["reading", "cooking", "traveling"],
  "completion_stage": "completed",
  "missing_fields": [],
  "completion_flag": true,
  "message": "Onboarding complete! Welcome aboard."
}
```

## Test Scenarios

The example runs through a comprehensive test suite:

### Scenario 1: Initial Contact
**Input**: "Hi! I want to sign up for your service."
**Expected**: Basic info stage condition triggers
**Validates**: Stage-based flow control

### Scenario 2: Name Collection
**Input**: "My name is John Smith"
**Expected**: Progress to contact details stage
**Validates**: Information collection and stage progression

### Scenario 3: Resistance Handling
**Input**: "I don't want to give my contact details yet"
**Expected**: Missing contact condition triggers
**Validates**: Conditional instruction updates

### Scenario 4-5: Partial Contact Info
**Input**: Email first, then phone
**Expected**: Gradual completion tracking
**Validates**: Incremental progress conditions

### Scenario 6: Preferences Collection
**Input**: Interest and hobby information
**Expected**: Preferences condition handling
**Validates**: Custom field validation

### Scenario 7: Extended Conversation
**Input**: Additional conversation turns
**Expected**: Message count condition triggers
**Validates**: Long conversation optimization

### Scenario 8: Completion Validation
**Input**: "Is that everything? Am I done now?"
**Expected**: Final validation and completion
**Validates**: End-to-end condition flow

## Running the Example

### Prerequisites
1. Go 1.22 or later
2. OpenAI API key

### Setup
1. **Configure Environment**:
   ```bash
   cp .env.example .env
   # Edit .env and add your OpenAI API key
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the Example**:
   ```bash
   go run main.go
   ```

## Expected Output

```
🧪 Condition Testing & Flow Rules Demo
Testing various condition types and flow rule triggers...
============================================================

🔄 Test 1/8
👤 User: Hi! I want to sign up for your service.
🎯 CONDITION[at_basic_info_stage]: Stage 'basic_info' == 'basic_info' ? true
🤖 Agent: Welcome! I'll help you get signed up. Let's start with your name...
📊 Status: Stage=basic_info, Missing=[name email phone preferences], Complete=false

🔄 Test 3/8  
👤 User: I don't want to give my contact details yet
🎯 CONDITION[missing_contact_info]: Field 'email' is missing
🎯 CONDITION[missing_contact_info]: Field 'phone' is missing
🤖 Agent: I understand your concern about privacy. We need your contact information to...
📊 Status: Stage=contact_details, Missing=[email phone], Complete=false

🔄 Test 7/8
👤 User: Actually, let me add that I also enjoy photography and hiking
🎯 CONDITION[long_conversation]: Message count 7 >= 6 ? true
🤖 Agent: Great additional interests! Let me summarize what we've collected so far...
📊 Status: Stage=completed, Missing=[], Complete=true
```

## Learning Outcomes

This example demonstrates:

1. **Condition Diversity**: Different types of conditions for various use cases
2. **Flow Control**: How conditions modify agent behavior dynamically
3. **Custom Logic**: Implementing domain-specific conditional logic
4. **State Management**: Tracking complex state through structured output
5. **User Experience**: Creating natural, adaptive conversation flows

## Key Implementation Patterns

### Custom Condition Implementation
```go
func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]interface{}) (bool, error) {
    // Extract missing fields from structured output
    missingFields, exists := data["missing_fields"]
    if !exists {
        return false, nil
    }
    
    // Check if any target fields are missing
    for _, targetField := range c.fields {
        if contains(missing, targetField) {
            return true, nil  // Condition met
        }
    }
    
    return false, nil  // Condition not met
}
```

### Flow Rule Creation
```go
rule, err := agent.NewFlowRule("collect-contact-info", missingContactCondition).
    WithDescription("Prompt user for missing contact information").
    WithNewInstructions("Focus on collecting email and phone...").
    WithRecommendedTools("collect_user_info").
    WithSystemMessage("Contact information needed").
    Build()
```

### Structured Output Integration
```go
agent.New(
    // ... other options
    agent.WithStructuredOutput(&UserStatusOutput{}),
    agent.WithFlowRules(flowRules...),
)
```

## Architecture Benefits

1. **Modularity**: Conditions are independent and reusable
2. **Testability**: Each condition can be tested in isolation
3. **Flexibility**: Easy to add new condition types
4. **Maintainability**: Clear separation between logic and configuration
5. **Scalability**: Framework handles complex condition orchestration

## Common Use Cases

This pattern is ideal for:
- **User Onboarding**: Multi-step registration processes
- **Form Validation**: Dynamic form behavior based on user input
- **Workflow Management**: Conditional process flows
- **Customer Support**: Context-aware response handling
- **E-commerce**: Adaptive checkout and recommendation flows