# Smart Information Collection with Schema

go-agent's schema system provides intelligent, automatic information collection from users through natural conversation. Instead of manually managing state and explicitly prompting for missing information, you simply define what you need and let the framework handle the rest.

## Quick Start

```go
import "github.com/davidleitw/go-agent/pkg/schema"

// Define the information you need
response, err := agent.Chat(ctx, "I need support",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("issue", "Describe your issue in detail"),
        schema.Define("priority", "How urgent is this?").Optional(),
    ),
)
```

That's it. The agent will automatically:
- Extract any information already provided in the user's message
- Identify what's still missing
- Ask for missing information using natural, contextual prompts
- Remember collected information across conversation turns

## Core Concepts

### Field Definition

```go
// Required field (default)
email := schema.Define("email", "Please provide your email address")

// Optional field
phone := schema.Define("phone", "Phone number for urgent contact").Optional()
```

Every field has:
- **Name**: Internal identifier for the data
- **Prompt**: Human-readable request shown to users when the field is missing
- **Required**: Whether the field must be collected (default: true)

### Intelligent Extraction

The framework uses LLM semantic understanding to extract information:

```go
// User says: "My email is john@example.com and I have a billing issue"
// Framework automatically extracts:
// - email: "john@example.com" 
// - issue_type: "billing" (if you defined this field)
```

No rigid pattern matching - the LLM understands context and meaning.

### Natural Collection

When information is missing, the framework generates contextual prompts:

```go
// Instead of: "Please provide field X"
// You get: "Thanks for contacting support! To help you with your billing issue, 
//           I'll need your email address for follow-up."
```

Prompts are contextual, friendly, and maintain conversation flow.

## Real-World Examples

### Customer Support Bot

```go
supportSchema := []*schema.Field{
    schema.Define("email", "Please provide your email for follow-up"),
    schema.Define("issue_category", "What type of issue are you experiencing?"),
    schema.Define("description", "Please describe the issue in detail"),
    schema.Define("order_id", "Order number if this relates to a purchase").Optional(),
    schema.Define("urgency", "How urgent is this issue?").Optional(),
}

// Handles the complete support intake process automatically
response, err := supportBot.Chat(ctx, userMessage, agent.WithSchema(supportSchema...))
```

### Lead Qualification

```go
salesSchema := []*schema.Field{
    schema.Define("email", "Please provide your business email"),
    schema.Define("company", "What company do you represent?"),
    schema.Define("team_size", "How many people are on your team?"),
    schema.Define("use_case", "How would you use our product?"),
    schema.Define("timeline", "When are you looking to get started?").Optional(),
    schema.Define("budget", "Do you have a budget range in mind?").Optional(),
}
```

### User Onboarding

```go
onboardingSchema := []*schema.Field{
    schema.Define("name", "What should we call you?"),
    schema.Define("role", "What's your role at your company?"),
    schema.Define("goals", "What are you hoping to achieve?"),
    schema.Define("experience", "How familiar are you with tools like ours?").Optional(),
}
```

## Advanced Usage

### Dynamic Schema Selection

Adapt collection based on user intent:

```go
func getSchemaForIntent(intent string) []*schema.Field {
    switch intent {
    case "technical_support":
        return []*schema.Field{
            schema.Define("email", "Email for technical follow-up"),
            schema.Define("error_message", "What error are you seeing?"),
            schema.Define("steps_taken", "What have you tried so far?"),
        }
    case "billing":
        return []*schema.Field{
            schema.Define("email", "Account email address"),
            schema.Define("account_id", "Your account number"),
            schema.Define("billing_question", "What's your billing question?"),
        }
    default:
        return generalSupportSchema()
    }
}

// Use in conversation
intent := classifyUserIntent(userInput)
schema := getSchemaForIntent(intent)
response, err := agent.Chat(ctx, userInput, agent.WithSchema(schema...))
```

### Multi-Step Workflows

Break complex collection into manageable steps:

```go
func getTechnicalSupportWorkflow() [][]*schema.Field {
    return [][]*schema.Field{
        { // Step 1: Basic info
            schema.Define("email", "Your email address"),
            schema.Define("issue_summary", "Brief description of the issue"),
        },
        { // Step 2: Technical details
            schema.Define("error_message", "Exact error message"),
            schema.Define("browser", "What browser are you using?"),
            schema.Define("steps_to_reproduce", "How can we reproduce this?"),
        },
        { // Step 3: Impact assessment
            schema.Define("affected_users", "How many users are affected?"),
            schema.Define("business_impact", "How does this impact your business?"),
            schema.Define("workaround", "Any temporary workarounds?").Optional(),
        },
    }
}

// Execute workflow step by step
workflow := getTechnicalSupportWorkflow()
for i, stepSchema := range workflow {
    response, err := agent.Chat(ctx, userInput, 
        agent.WithSession(sessionID),
        agent.WithSchema(stepSchema...),
    )
    // Check if step is complete and proceed
}
```

### Integration with Flow Control

Combine with existing flow control features:

```go
agent := agent.New("smart-support").
    WithOpenAI(apiKey).
    WithInstructions("Professional customer support assistant").
    
    // Use flow control for routing
    When(agent.Contains("billing")).
        Ask("I'll help you with billing. Let me gather some information.").
        
    // Use conditions for escalation
    When(agent.Contains("urgent")).
        UseTemplate("This is marked as urgent. I'll prioritize your request.").
        
    Build()

// Schema collection happens automatically during chat
response, err := agent.Chat(ctx, userInput,
    agent.WithSchema(getBillingSchema()...),
)
```

## Best Practices

### Field Design

**Good field prompts are:**
- **Clear**: "Please provide your email address"
- **Contextual**: "Email for order updates"  
- **Natural**: "What's your phone number for urgent contact?"

**Avoid:**
- Technical jargon: "Enter primary identifier"
- Robotic language: "Input field EMAIL required"
- Ambiguous requests: "Please provide information"

### Schema Organization

**Group related fields:**
```go
// Contact information
contactSchema := []*schema.Field{
    schema.Define("name", "Your full name"),
    schema.Define("email", "Email address"),
    schema.Define("phone", "Phone number").Optional(),
}

// Issue details  
issueSchema := []*schema.Field{
    schema.Define("category", "Type of issue"),
    schema.Define("description", "Detailed description"),
    schema.Define("urgency", "How urgent is this?"),
}
```

**Use optional fields strategically:**
- Required: Information you absolutely need
- Optional: Nice-to-have information that shouldn't block progress

### Error Handling

```go
response, err := agent.Chat(ctx, userInput, agent.WithSchema(schema...))
if err != nil {
    // Handle chat errors
    return err
}

// Check if collection is complete
if isCollectionComplete(response) {
    // Process collected information
    processUserData(response.Session)
} else {
    // Continue collection in next turn
    handleMissingInformation(response)
}

func isCollectionComplete(response *agent.Response) bool {
    // Schema collection is complete when metadata doesn't indicate missing fields
    if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
        return false // Still collecting
    }
    return true // Collection complete
}
```

### Performance Considerations

**Efficient schema usage:**
- Keep schemas focused (5-7 fields max per step)
- Use multi-step workflows for complex collection
- Cache schema definitions for reuse
- Consider user experience - don't overwhelm with too many questions

**Memory management:**
```go
// Reuse schema definitions
var (
    contactSchema = []*schema.Field{
        schema.Define("name", "Your name"),
        schema.Define("email", "Email address"),
    }
    
    supportSchema = []*schema.Field{
        schema.Define("issue", "Describe your issue"),
        schema.Define("priority", "How urgent?"),
    }
)

// Combine as needed
response, err := agent.Chat(ctx, userInput,
    agent.WithSchema(append(contactSchema, supportSchema...)...),
)
```

## Integration Examples

### With Web Forms

```go
// Convert web form to schema
func webFormToSchema(formFields []FormField) []*schema.Field {
    var schema []*schema.Field
    for _, field := range formFields {
        schemaField := schema.Define(field.Name, field.Label)
        if !field.Required {
            schemaField = schemaField.Optional()
        }
        schema = append(schema, schemaField)
    }
    return schema
}
```

### With Validation

```go
// Add custom validation to collected data
func validateCollectedData(session agent.Session) error {
    // Extract collected information
    messages := session.Messages()
    
    // Validate email format, phone numbers, etc.
    for _, msg := range messages {
        if msg.Role == agent.RoleUser {
            if email := extractEmail(msg.Content); email != "" {
                if !isValidEmail(email) {
                    return fmt.Errorf("invalid email format: %s", email)
                }
            }
        }
    }
    
    return nil
}
```

### With External Systems

```go
// Save collected information to CRM
func saveToExternalSystem(session agent.Session) error {
    collectedData := extractCollectedData(session)
    
    // Map to external system format
    leadData := mapToLeadData(collectedData)
    
    // Save to CRM/database
    return crmClient.CreateLead(leadData)
}
```

## Troubleshooting

**Common Issues:**

1. **Information not being extracted:**
   - Check that field names are descriptive
   - Verify prompts are clear
   - Ensure user input contains the information

2. **Endless collection loops:**
   - Validate schema definitions
   - Check for typos in field names  
   - Ensure required fields are reasonable

3. **Poor user experience:**
   - Review prompt language for naturalness
   - Consider breaking complex schemas into steps
   - Test with real user conversations

**Debug tips:**
```go
// Check collection metadata
if metadata, ok := response.Metadata["missing_fields"].([]string); ok {
    fmt.Printf("Still missing: %v\n", metadata)
}

// Inspect session state
fmt.Printf("Messages in session: %d\n", len(response.Session.Messages()))
for _, msg := range response.Session.Messages() {
    fmt.Printf("%s: %s\n", msg.Role, msg.Content)
}
```

## Next Steps

- Explore the [examples](../examples/) directory for complete working implementations
- Try the [simple-schema](../examples/simple-schema/) example for basic usage
- Build a [customer-support](../examples/customer-support/) bot for real-world scenarios
- Experiment with [dynamic-schema](../examples/dynamic-schema/) selection for advanced use cases

The schema system is designed to handle 90% of information collection scenarios automatically while remaining flexible for custom requirements. Start simple and gradually add complexity as needed.