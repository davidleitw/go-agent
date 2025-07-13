# Customer Support Bot Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates a real-world customer support bot that uses intelligent schema-based information collection to handle different types of support requests professionally and efficiently.

## Overview

Customer support requires collecting different information based on the type of request. This example shows how to:

- **Classify** support requests into different categories
- **Apply** specialized schemas for each support type
- **Collect** information naturally across multiple conversation turns
- **Handle** partial information and context switching
- **Maintain** professional support agent behavior

## Quick Start

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# Run the example
go run examples/customer-support/main.go
```

## Code Walkthrough

### 1. Support Category Definition

```go
// Define different types of support requests
const (
    GeneralSupport    = "general"
    BillingSupport    = "billing"
    TechnicalSupport  = "technical"
)
```

**Key Points:**
- Different support types require different information
- Each category has its own specialized schema
- Professional categorization improves user experience

### 2. Schema Definition for Each Support Type

#### General Support Schema
```go
func getGeneralSupportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "Please provide your email for follow-up"),
        schema.Define("issue_category", "What type of issue are you experiencing?"),
        schema.Define("description", "Please describe the issue in detail"),
        schema.Define("order_id", "Order number if this relates to a purchase").Optional(),
        schema.Define("urgency", "How urgent is this issue?").Optional(),
        schema.Define("previous_contact", "Have you contacted us about this before?").Optional(),
    }
}
```

#### Billing Support Schema
```go
func getBillingSupportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "Please provide the email associated with your account"),
        schema.Define("account_number", "What is your account ID or number?"),
        schema.Define("billing_question", "Please describe your billing question or concern"),
        schema.Define("amount_disputed", "If disputing a charge, what amount?").Optional(),
        schema.Define("payment_method", "What payment method was used?").Optional(),
        schema.Define("billing_period", "Which billing period does this concern?").Optional(),
    }
}
```

#### Technical Support Schema
```go
func getTechnicalSupportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "Please provide your email for technical follow-up"),
        schema.Define("error_message", "What error message are you seeing?"),
        schema.Define("steps_taken", "What troubleshooting steps have you already tried?"),
        schema.Define("browser", "What browser are you using?").Optional(),
        schema.Define("device_type", "What device are you using? (desktop/mobile/tablet)").Optional(),
        schema.Define("operating_system", "What operating system?").Optional(),
    }
}
```

**Key Points:**
- Each schema tailored to collect relevant information
- Required fields focus on essential support data
- Optional fields gather helpful context
- Field prompts are professional and specific

### 3. Support Bot Creation

```go
supportBot, err := agent.New("customer-support-bot").
    WithOpenAI(apiKey).
    WithModel("gpt-4o-mini").
    WithInstructions(`You are a professional customer support assistant. 
    Your goal is to help customers efficiently by collecting the necessary 
    information to resolve their issues. Be empathetic, clear, and helpful.`).
    Build()
```

**Key Points:**
- Professional instructions set the right tone
- Emphasizes efficiency and empathy
- Model choice balances cost and capability

### 4. Support Type Detection

```go
func detectSupportType(userInput string) string {
    input := strings.ToLower(userInput)
    
    // Check for billing keywords
    billingKeywords := []string{"billing", "payment", "charge", "invoice", "refund", "subscription"}
    for _, keyword := range billingKeywords {
        if strings.Contains(input, keyword) {
            return BillingSupport
        }
    }
    
    // Check for technical keywords
    technicalKeywords := []string{"error", "bug", "login", "password", "technical", "broken", "not working"}
    for _, keyword := range technicalKeywords {
        if strings.Contains(input, keyword) {
            return TechnicalSupport
        }
    }
    
    return GeneralSupport
}
```

**Key Points:**
- Simple keyword-based classification
- Fallback to general support for unclear cases
- Can be enhanced with ML classification

### 5. Dynamic Schema Application

```go
func handleSupportRequest(ctx context.Context, bot agent.Agent, userInput string) {
    // Detect the type of support needed
    supportType := detectSupportType(userInput)
    
    // Get appropriate schema
    var supportSchema []*schema.Field
    switch supportType {
    case BillingSupport:
        supportSchema = getBillingSupportSchema()
    case TechnicalSupport:
        supportSchema = getTechnicalSupportSchema()
    default:
        supportSchema = getGeneralSupportSchema()
    }
    
    // Apply schema to conversation
    response, err := bot.Chat(ctx, userInput,
        agent.WithSchema(supportSchema...),
    )
}
```

**Key Points:**
- Dynamic schema selection based on detected intent
- Seamless switching between different support types
- Maintains conversation context throughout

### 6. Multi-Turn Conversation Handling

```go
// Conversation flow example
func runSupportConversation(ctx context.Context, bot agent.Agent) {
    sessionID := "support-conversation"
    
    // First interaction
    response1, _ := bot.Chat(ctx, "I need help with something",
        agent.WithSession(sessionID),
        agent.WithSchema(getGeneralSupportSchema()...),
    )
    
    // Customer provides more information
    response2, _ := bot.Chat(ctx, "My email is john.doe@example.com and it's a technical issue",
        agent.WithSession(sessionID),
        agent.WithSchema(getTechnicalSupportSchema()...),  // Switch to technical schema
    )
    
    // Continue until all information collected
    response3, _ := bot.Chat(ctx, "I can't access my dashboard. It shows 'Connection timeout' error",
        agent.WithSession(sessionID),
        agent.WithSchema(getTechnicalSupportSchema()...),
    )
}
```

**Key Points:**
- Session maintains conversation context
- Schema can change based on new information
- Information accumulates across turns

## Example Scenarios

### Scenario 1: General Support Request

```go
// User: "I'm having trouble with my account"
// Detection: General support (no specific keywords)
// Schema: General support schema applied
// Result: Collects email, issue category, description

userInput := "I'm having trouble with my account"
supportType := detectSupportType(userInput)  // Returns "general"
schema := getGeneralSupportSchema()

response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(schema...),
)
// Bot asks for email and more details about the account issue
```

### Scenario 2: Billing Inquiry

```go
// User: "I have a question about my last invoice"
// Detection: Billing support (contains "invoice")
// Schema: Billing-specific schema applied
// Result: Collects email, account number, billing question

userInput := "I have a question about my last invoice"
supportType := detectSupportType(userInput)  // Returns "billing"
schema := getBillingSupportSchema()

response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(schema...),
)
// Bot asks for email, account number, and specific billing question
```

### Scenario 3: Technical Support

```go
// User: "I'm getting an error when trying to log in"
// Detection: Technical support (contains "error" and "login")
// Schema: Technical support schema applied
// Result: Collects email, error message, troubleshooting steps

userInput := "I'm getting an error when trying to log in"
supportType := detectSupportType(userInput)  // Returns "technical"
schema := getTechnicalSupportSchema()

response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(schema...),
)
// Bot asks for email, specific error message, and steps already tried
```

### Scenario 4: Customer Provides Partial Information

```go
// User: "Hi, my email is customer@example.com and I have a billing issue with order #12345"
// Detection: Billing support (contains "billing")
// Schema: Billing schema, but email and order info already extracted
// Result: Asks for remaining required information

userInput := "Hi, my email is customer@example.com and I have a billing issue with order #12345"
response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(getBillingSupportSchema()...),
)
// Bot acknowledges email and order, asks for account number and billing question details
```

## Advanced Features

### 1. Support Priority Detection

```go
func detectPriority(userInput string) string {
    input := strings.ToLower(userInput)
    
    urgentKeywords := []string{"urgent", "critical", "emergency", "asap", "immediately"}
    for _, keyword := range urgentKeywords {
        if strings.Contains(input, keyword) {
            return "high"
        }
    }
    
    return "normal"
}

// Apply priority to schema
priority := detectPriority(userInput)
if priority == "high" {
    // Add urgency field as required for high-priority requests
    schema = append(schema, schema.Define("urgency_reason", "Why is this urgent?"))
}
```

### 2. Escalation Logic

```go
func shouldEscalate(session agent.Session, supportType string) bool {
    messages := session.Messages()
    
    // Escalate if conversation is getting long
    if len(messages) > 10 {
        return true
    }
    
    // Escalate technical issues with multiple failed attempts
    if supportType == TechnicalSupport {
        for _, msg := range messages {
            if strings.Contains(strings.ToLower(msg.Content), "still not working") {
                return true
            }
        }
    }
    
    return false
}
```

### 3. Information Validation

```go
func validateSupportInfo(extractedData map[string]interface{}) []string {
    var errors []string
    
    // Validate email format
    if email, ok := extractedData["email"].(string); ok {
        if !isValidEmail(email) {
            errors = append(errors, "Invalid email format")
        }
    }
    
    // Validate account number format
    if accountNum, ok := extractedData["account_number"].(string); ok {
        if len(accountNum) < 6 {
            errors = append(errors, "Account number too short")
        }
    }
    
    return errors
}
```

### 4. Support Ticket Creation

```go
type SupportTicket struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`
    Priority     string                 `json:"priority"`
    CustomerInfo map[string]interface{} `json:"customer_info"`
    CreatedAt    time.Time              `json:"created_at"`
    Status       string                 `json:"status"`
}

func createSupportTicket(session agent.Session, supportType string) (*SupportTicket, error) {
    // Extract information from session
    extractedInfo := extractAllInformation(session)
    
    ticket := &SupportTicket{
        ID:           generateTicketID(),
        Type:         supportType,
        Priority:     determinePriority(extractedInfo),
        CustomerInfo: extractedInfo,
        CreatedAt:    time.Now(),
        Status:       "open",
    }
    
    return ticket, nil
}
```

## Testing

Run the example tests:

```bash
go test ./examples/customer-support/
```

The tests verify:
- Support type detection accuracy
- Schema selection for different support types
- Information collection across conversation turns
- Professional response generation

### Test Cases

```go
func TestSupportTypeDetection(t *testing.T) {
    testCases := []struct {
        input    string
        expected string
    }{
        {"I have a billing question", "billing"},
        {"Getting a login error", "technical"},
        {"Need help with my account", "general"},
        {"Refund my payment", "billing"},
        {"App not working", "technical"},
    }
    
    for _, tc := range testCases {
        result := detectSupportType(tc.input)
        assert.Equal(t, tc.expected, result)
    }
}
```

## Integration Examples

### With CRM Systems

```go
func saveToCRM(ticket *SupportTicket) error {
    crmData := map[string]interface{}{
        "customer_email": ticket.CustomerInfo["email"],
        "issue_type":     ticket.Type,
        "description":    ticket.CustomerInfo["description"],
        "priority":       ticket.Priority,
        "created_at":     ticket.CreatedAt,
    }
    
    return crmClient.CreateTicket(crmData)
}
```

### With Help Desk Software

```go
func createHelpDeskTicket(ticket *SupportTicket) error {
    helpdeskTicket := helpdesk.Ticket{
        Subject:     generateSubject(ticket),
        Description: formatDescription(ticket),
        Priority:    mapPriority(ticket.Priority),
        Customer:    ticket.CustomerInfo["email"].(string),
        Category:    ticket.Type,
    }
    
    return helpdeskClient.Create(helpdeskTicket)
}
```

### With Knowledge Base

```go
func suggestKnowledgeBaseArticles(supportType string, description string) []KBArticle {
    switch supportType {
    case TechnicalSupport:
        return searchTechnicalArticles(description)
    case BillingSupport:
        return searchBillingArticles(description)
    default:
        return searchGeneralArticles(description)
    }
}
```

## Best Practices

### 1. Professional Communication

```go
// Good: Professional and empathetic
"I'm sorry to hear you're experiencing this issue. To help you better, could you please provide..."

// Avoid: Too casual or robotic
"Give me your email" or "Please input the required data fields"
```

### 2. Information Prioritization

```go
// Prioritize essential information first
requiredFields := []*schema.Field{
    schema.Define("email", "Contact email for follow-up"),
    schema.Define("issue_description", "Description of the problem"),
}

// Collect nice-to-have information later
optionalFields := []*schema.Field{
    schema.Define("phone", "Phone number for urgent contact").Optional(),
    schema.Define("preferred_contact_time", "Best time to contact you").Optional(),
}
```

### 3. Context Preservation

```go
// Always use session to maintain context
response, err := supportBot.Chat(ctx, userInput,
    agent.WithSession(sessionID),
    agent.WithSchema(schema...),
)

// Don't lose previous conversation context
// Each turn builds on the previous information
```

### 4. Error Recovery

```go
if err != nil {
    // Graceful error handling in support context
    fallbackResponse := "I apologize, but I'm having trouble processing your request. " +
                        "Could you please try again or contact our support team directly?"
    return &agent.Response{Message: fallbackResponse}
}
```

## Performance Considerations

### 1. Schema Caching

```go
var (
    generalSchema   []*schema.Field
    billingSchema   []*schema.Field
    technicalSchema []*schema.Field
)

func init() {
    // Pre-create schemas to avoid repeated allocation
    generalSchema = getGeneralSupportSchema()
    billingSchema = getBillingSupportSchema()
    technicalSchema = getTechnicalSupportSchema()
}
```

### 2. Conversation Limits

```go
const (
    MaxConversationTurns = 20
    MaxResponseTime      = 30 * time.Second
)

func enforceConversationLimits(session agent.Session) error {
    if len(session.Messages()) > MaxConversationTurns {
        return errors.New("conversation too long, escalating to human agent")
    }
    return nil
}
```

## Related Examples

- **[Simple Schema](../simple-schema/)**: Foundation concepts for schema-based collection
- **[Dynamic Schema](../dynamic-schema/)**: Advanced intent classification and workflow management
- **[Basic Chat](../basic-chat/)**: Core conversation concepts
- **[Multi-Tool Agent](../multi-tool-agent/)**: Adding external capabilities to support bots

## Next Steps

1. **Try Different Support Types**: Test with various support scenarios
2. **Customize Schemas**: Modify fields for your specific business needs
3. **Add Integrations**: Connect to your CRM or help desk system
4. **Enhance Classification**: Improve support type detection with ML models
5. **Test Edge Cases**: Handle unusual requests and error scenarios

## Troubleshooting

**Issue**: Wrong support type detected
**Solution**: Improve keyword lists or implement ML-based classification

**Issue**: Information not being extracted properly
**Solution**: Review field names and prompts for clarity

**Issue**: Conversations getting too long
**Solution**: Implement escalation logic and conversation limits

**Issue**: Schema switching not working
**Solution**: Ensure session continuity and proper schema application

For more help, see the [main examples documentation](../README.md) or [schema collection guide](../../docs/schema-collection.md).