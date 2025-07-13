# Simple Schema Collection Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates the fundamental concepts of schema-based information collection in go-agent. Learn how to automatically gather structured information from users through natural conversation.

## Overview

Traditional chatbots require complex state management to collect user information. With go-agent's schema system, you simply define what information you need, and the framework automatically:

- **Extracts** information from user messages using LLM semantic understanding
- **Identifies** missing required fields
- **Asks** for missing information using natural prompts
- **Remembers** collected information across conversation turns

## Quick Start

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# Run the example
go run examples/simple-schema/main.go
```

## Code Walkthrough

### 1. Field Definition

```go
import "github.com/davidleitw/go-agent/pkg/schema"

// Define required fields
emailField := schema.Define("email", "Please provide your email address")
issueField := schema.Define("issue", "Please describe your issue")

// Define optional fields
phoneField := schema.Define("phone", "Contact number for urgent matters").Optional()
```

**Key Points:**
- `schema.Define(name, prompt)` creates a required field by default
- `.Optional()` makes a field optional
- The prompt is what users see when the field is missing

### 2. Agent Creation

```go
assistant, err := agent.New("simple-assistant").
    WithOpenAI(os.Getenv("OPENAI_API_KEY")).
    WithModel("gpt-4o-mini").
    WithInstructions("You are a helpful assistant that collects user information.").
    Build()
```

**Key Points:**
- Standard agent creation with OpenAI integration
- Instructions help the agent understand its role
- No special configuration needed for schema support

### 3. Schema Application

```go
response, err := assistant.Chat(ctx, "I need some help",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("issue", "Please describe your issue"),
    ),
)
```

**Key Points:**
- `agent.WithSchema()` applies schema to a specific conversation
- Multiple fields can be defined in one call
- Schema only applies to this chat interaction

### 4. Information Extraction

```go
userInput := "My email is user@example.com and I have a billing question"

response, err := assistant.Chat(ctx, userInput,
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("question", "Please describe your billing question in detail"),
    ),
)

// The framework automatically extracts:
// - email: "user@example.com" ✓
// - question: Missing, will ask for it
```

**Key Points:**
- LLM understands context and meaning, not just exact matches
- Partial information is extracted and remembered
- Missing information is identified automatically

### 5. Multi-Turn Conversations

```go
// Turn 1: User provides partial information
response1, _ := assistant.Chat(ctx, "Hi, my name is John",
    agent.WithSchema(
        schema.Define("name", "Please provide your name"),
        schema.Define("email", "Please provide your email address"),
        schema.Define("topic", "What would you like to discuss?"),
    ),
)
// Agent asks for missing email and topic

// Turn 2: User provides more information
response2, _ := assistant.Chat(ctx, "My email is john@example.com and I want to talk about pricing",
    agent.WithSchema(/* same schema */),
)
// Agent acknowledges completion
```

**Key Points:**
- Session automatically maintains conversation context
- Previously collected information is remembered
- Schema continues until all required fields are collected

### 6. Checking Collection Status

```go
// Check if collection is ongoing
if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
    fmt.Println("Still collecting information...")
    if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
        fmt.Printf("Missing: %v\n", missingFields)
    }
} else {
    fmt.Println("All required information collected!")
}
```

**Key Points:**
- `response.Metadata["schema_collection"]` indicates if collection is active
- `response.Metadata["missing_fields"]` shows what's still needed
- Use this to implement custom collection logic

## Example Scenarios

### Scenario 1: Basic Email Collection

```go
// User: "I need some help"
// Schema: email (required)
// Result: Agent asks for email address

response, err := assistant.Chat(ctx, "I need some help",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
    ),
)
// Response: "I'd be happy to help! Could you please provide your email address?"
```

### Scenario 2: Multiple Required Fields

```go
// User: "I want to report an issue"
// Schema: email, issue (both required)
// Result: Agent asks for missing information

response, err := assistant.Chat(ctx, "I want to report an issue",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("issue", "Please describe your issue"),
    ),
)
// Response: Agent asks for email (issue type already understood from context)
```

### Scenario 3: Mixed Required and Optional

```go
// User: "I need assistance with my account"
// Schema: email (required), issue (required), phone (optional)
// Result: Agent focuses on required fields first

response, err := assistant.Chat(ctx, "I need assistance with my account",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("issue", "Please describe the issue you're experiencing"),
        schema.Define("phone", "Contact number for urgent matters").Optional(),
    ),
)
// Response: Agent asks for email and issue description
```

### Scenario 4: Partial Information Provided

```go
// User: "My email is user@example.com and I have a billing question"
// Schema: email, question, account_type (email extracted, others missing)
// Result: Agent asks for remaining required information

response, err := assistant.Chat(ctx, 
    "My email is user@example.com and I have a billing question",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("question", "Please describe your billing question in detail"),
        schema.Define("account_type", "What type of account do you have?").Optional(),
    ),
)
// Response: Agent acknowledges email, asks for detailed question
```

## Best Practices

### 1. Field Design

**Good field prompts:**
```go
schema.Define("email", "Please provide your email address")
schema.Define("issue", "Please describe your issue in detail")
schema.Define("urgency", "How urgent is this? (low/medium/high)")
```

**Avoid:**
```go
schema.Define("email", "Email")  // Too brief
schema.Define("data", "Enter data")  // Too vague
schema.Define("field1", "Input required")  // Not descriptive
```

### 2. Required vs Optional

**Use Required for:**
- Essential contact information (email)
- Core business data (issue description)
- Information needed to proceed

**Use Optional for:**
- Nice-to-have details (phone number)
- Preference settings (contact method)
- Additional context (previous attempts)

### 3. Field Naming

```go
// Good: Descriptive and clear
schema.Define("customer_email", "Your email address")
schema.Define("issue_description", "Describe the problem you're experiencing")
schema.Define("preferred_contact_method", "How would you like us to contact you?")

// Avoid: Generic or confusing names
schema.Define("data1", "Email")
schema.Define("info", "Details")
schema.Define("field", "Information")
```

### 4. Error Handling

```go
response, err := assistant.Chat(ctx, userInput, agent.WithSchema(schema...))
if err != nil {
    log.Printf("Chat failed: %v", err)
    return
}

// Check if collection is complete
if schemaActive := response.Metadata["schema_collection"]; schemaActive == true {
    // Continue collection
    fmt.Println("Collecting more information...")
} else {
    // Process collected data
    fmt.Println("Information collection complete!")
    processCollectedData(response.Session)
}
```

## Testing

Run the example tests:

```bash
go test ./examples/simple-schema/
```

The tests verify:
- Field definition and properties
- Schema application to conversations
- Information extraction accuracy
- Missing field identification

## Integration Examples

### With Web Forms

```go
// Convert web form to schema
func webFormToSchema(formFields []FormField) []*schema.Field {
    var schemaFields []*schema.Field
    for _, field := range formFields {
        schemaField := schema.Define(field.Name, field.Label)
        if !field.Required {
            schemaField = schemaField.Optional()
        }
        schemaFields = append(schemaFields, schemaField)
    }
    return schemaFields
}
```

### With Validation

```go
// Add custom validation
func validateEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

// Use in collection flow
if email := extractFieldFromSession("email", session); email != "" {
    if !validateEmail(email) {
        // Ask for email again with specific guidance
        return agent.WithSchema(
            schema.Define("email", "Please provide a valid email address (e.g., user@example.com)"),
        )
    }
}
```

## Related Examples

- **[Customer Support](../customer-support/)**: Real-world support bot using specialized schemas
- **[Dynamic Schema](../dynamic-schema/)**: Advanced intent-based schema selection
- **[Basic Chat](../basic-chat/)**: Foundation concepts without schema
- **[Task Completion](../task-completion/)**: Structured output with validation

## Next Steps

1. **Try the Example**: Run the code and experiment with different inputs
2. **Modify Fields**: Add your own fields and prompts
3. **Test Edge Cases**: Try partial information, typos, different phrasing
4. **Explore Advanced**: Move on to customer-support or dynamic-schema examples

## Troubleshooting

**Issue**: Information not being extracted
**Solution**: Check that field names are descriptive and prompts are clear

**Issue**: Agent asks for information already provided
**Solution**: Verify the field name matches the semantic meaning in user input

**Issue**: Collection never completes
**Solution**: Ensure all required fields are reasonable and achievable

For more help, see the [main examples documentation](../README.md) or [schema collection guide](../../docs/schema-collection.md).