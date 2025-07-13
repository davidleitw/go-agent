# Advanced Conditions Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates the **sophisticated condition system** in go-agent, showcasing how to create intelligent, adaptive conversational agents with elegant condition-based flow control. Through a comprehensive user onboarding scenario, you'll learn to implement complex behavioral patterns with simple, readable code.

## 🎯 What This Example Demonstrates

This example implements a **smart onboarding agent** that collects user information (name, email, phone, preferences) while demonstrating:

- **🔄 Simple Field-Based Conditions** - Automatic prompting for missing information
- **🧠 Combination Logic** - AND/OR/NOT condition combinations  
- **⏱️ Conversation Management** - Message count and timing-based triggers
- **💭 Content Recognition** - Keyword and phrase detection
- **🕐 Time-Aware Behavior** - Business hours and weekend adaptations
- **😤 Emotional Intelligence** - Frustration and difficulty detection
- **📊 Structured Output** - Complete user profiles as JSON
- **🛠️ Function-Based Tools** - Information collection and validation

## 🏗️ Core Features Showcased

### 1. Simple Field-Based Conditions
```go
OnMissingInfo("name").Ask("Hello! I'd love to help you get started. What's your name?").Build().
OnMissingInfo("email").Ask("Great! Now I'll need your email address to set up your account.").Build().
```
Automatically detects missing required fields and prompts users appropriately.

### 2. Combination Conditions with Logical Operators
```go
When(agent.And(
    agent.WhenMissingFields("email"),
    agent.WhenMissingFields("phone"),
)).Ask("I'll need both your email address and phone number to proceed.").Build().

When(agent.Or(
    agent.WhenContains("frustrated"),
    agent.WhenContains("difficult"),
    agent.WhenMessageCount(8),
)).Ask("Let me help streamline this for you!").Build().
```
Combines multiple conditions with logical operators for sophisticated behavior.

### 3. Conversation Management
```go
OnMessageCount(6).Summarize().Build().
OnMessageCount(10).Ask("We've been chatting for a while! Let me help summarize...").Build().
```
Manages long conversations by providing summaries and guidance.

### 4. Content-Based Triggers
```go
When(agent.WhenContains("help")).Ask("Of course! How can I help?").Build().
When(agent.WhenContains("skip")).Ask("All information is required, but let's work together!").Build().
```
Responds intelligently to specific keywords and user sentiment.

### 5. Time-Aware Behavior
```go
When(agent.WhenFunc("business_hours", func(session agent.Session) bool {
    now := time.Now()
    hour := now.Hour()
    return hour >= 9 && hour <= 17
})).Ask("Since it's business hours, I can provide immediate assistance!").Build().
```
Adapts behavior based on time of day, weekends, or other temporal factors.

### 6. Behavioral Pattern Detection
```go
When(agent.WhenFunc("retry_attempt", func(session agent.Session) bool {
    messages := session.Messages()
    retryCount := 0
    for _, msg := range messages {
        if strings.Contains(strings.ToLower(msg.Content), "try again") {
            retryCount++
        }
    }
    return retryCount >= 2
})).Ask("I see you've had to retry a few times. Let me provide extra guidance!").Build().
```
Detects complex behavioral patterns and responds empathetically.

## 🛠️ Function-Based Tools

### Information Collection Tool
```go
collectTool := agent.NewTool("collect_info", 
    "Collect and validate user information",
    func(field, value string) map[string]any {
        // Automatic validation logic
        valid := len(value) > 0
        if field == "email" {
            valid = strings.Contains(value, "@")
        }
        if field == "phone" {
            valid = len(value) >= 10
        }
        
        return map[string]any{
            "field":     field,
            "value":     value,
            "is_valid":  valid,
            "timestamp": time.Now().Format(time.RFC3339),
        }
    })
```

### Profile Validation Tool
```go
validateTool := agent.NewTool("validate_profile",
    "Check if user profile is complete",
    func(userData map[string]any) map[string]any {
        required := []string{"name", "email", "phone", "preferences"}
        missing := []string{}
        
        for _, field := range required {
            if value, exists := userData[field]; !exists || value == "" {
                missing = append(missing, field)
            }
        }
        
        return map[string]any{
            "is_complete":   len(missing) == 0,
            "missing_fields": missing,
            "completion_rate": float64(len(required)-len(missing))/float64(len(required)),
        }
    })
```

## 📊 Structured Output

The agent produces a complete `UserProfile` struct as JSON output:

```go
type UserProfile struct {
    Name         string   `json:"name"`
    Email        string   `json:"email"`
    Phone        string   `json:"phone"`
    Preferences  []string `json:"preferences"`
    CompletedAt  string   `json:"completed_at,omitempty"`
    IsComplete   bool     `json:"is_complete"`
    MissingInfo  []string `json:"missing_info,omitempty"`
    StatusText   string   `json:"status_text"`
}
```

## 🚀 Running the Example

### Prerequisites
1. Go 1.22 or higher
2. OpenAI API key

### Setup
```bash
# 1. Configure your API key
cp .env.example .env
# Edit .env and add your OPENAI_API_KEY

# 2. Run the example
cd examples/advanced-conditions
go run main.go
```

## 📋 Test Scenarios

The example runs through comprehensive test scenarios:

1. **Initial Contact** - Triggers name collection condition
2. **Help Request** - Demonstrates content-based conditions  
3. **Information Gathering** - Shows progressive field collection
4. **Skip Attempts** - Handles user resistance elegantly
5. **Frustration Detection** - Responds to user difficulty
6. **Completion Tracking** - Produces structured output

### Sample Output
```
🎯 Advanced Conditions Example
Demonstrating elegant condition usage and flow control
============================================================

🚀 Testing Advanced Condition System
============================================================

🔄 Test 1/8: Initial contact - should trigger name missing condition
👤 User: Hi there! I want to sign up.
🤖 Assistant: Hello! I'd love to help you get started. What's your name?

🔄 Test 2/8: Provides name - should ask for email
👤 User: My name is Alex Chen
🤖 Assistant: Great! Now I'll need your email address to set up your account.

🔄 Test 3/8: Contains 'help' - should trigger help condition
👤 User: Can you help me understand what information you need?
🤖 Assistant: Of course! I'm here to help you complete your profile. What specific information do you need assistance with?

📊 Profile Status: Complete=false, Missing=[email, phone, preferences]
📈 Messages in session: 6

🎉 Profile completed successfully!
```

## 🎓 Condition Types Reference

| Condition Type | Syntax | Use Case | Example |
|----------------|--------|----------|---------|
| **Field-Based** | `OnMissingInfo("field")` | Required field collection | Missing email/phone |
| **Message Count** | `OnMessageCount(n)` | Conversation management | Long conversations |
| **Content Detection** | `WhenContains("text")` | Keyword responses | Help requests |
| **Custom Functions** | `WhenFunc("name", fn)` | Complex logic | Business hours |
| **Logical AND** | `And(cond1, cond2)` | Multiple requirements | Email AND phone |
| **Logical OR** | `Or(cond1, cond2)` | Alternative triggers | Frustrated OR difficult |
| **Logical NOT** | `Not(condition)` | Negation logic | Not during weekends |

## 🏆 Key Learning Objectives

After studying this example, you should understand:

1. **Progressive Condition Complexity** - Building from simple to sophisticated
2. **Elegant Syntax Design** - Readable, maintainable condition definitions
3. **Behavioral Intelligence** - Creating emotionally aware conversational agents
4. **Tool Integration** - Combining conditions with function-based tools
5. **Structured Output** - Managing complex data collection workflows
6. **Time-Aware Behavior** - Creating context-sensitive responses
7. **Pattern Recognition** - Detecting and responding to user behavioral patterns

## 🔄 Before vs After: Condition Systems

### Traditional Approach (Complex & Verbose)
```go
// Manual condition checking
if len(session.Messages()) > 5 {
    if containsKeyword(lastMessage, "help") {
        if isBusinessHours() && userSeemsFrustrated(session) {
            // Complex nested logic...
        }
    }
}
```

### go-agent Approach (Elegant & Readable)
```go
// Declarative condition system
When(agent.And(
    agent.WhenMessageCount(5),
    agent.WhenContains("help"),
    agent.WhenFunc("business_hours", isBusinessHours),
    agent.WhenFunc("frustrated", detectFrustration),
)).Ask("Let me provide immediate assistance!").Build()
```

## 🎯 Best Practices Demonstrated

1. **Start Simple** - Begin with basic field conditions, add complexity incrementally
2. **Combine Thoughtfully** - Use AND/OR to create meaningful condition combinations  
3. **Handle Edge Cases** - Include conditions for common user behaviors (help, skip, frustration)
4. **Maintain Readability** - Use descriptive names for custom conditions
5. **Test Thoroughly** - Verify conditions trigger in expected scenarios
6. **Progressive Enhancement** - Layer conditions from simple to sophisticated

## 🔄 Next Steps

After mastering this example:
1. **[Multi-Tool Agent](../multi-tool-agent/)** - Coordinate multiple capabilities
2. **[Task Completion](../task-completion/)** - Advanced structured output handling
3. **Custom Conditions** - Create your own condition types
4. **Integration Patterns** - Combine with external systems

## 🐛 Common Patterns & Solutions

### Pattern: Progressive Information Collection
```go
OnMissingInfo("name").Ask("What's your name?").Build().
OnMissingInfo("email").Ask("What's your email?").Build().
```

### Pattern: Frustration Detection & Response
```go
When(agent.Or(
    agent.WhenContains("frustrated"),
    agent.WhenMessageCount(8),
)).Ask("Let me help make this easier!").Build().
```

### Pattern: Time-Aware Behavior
```go
When(agent.WhenFunc("after_hours", func(session agent.Session) bool {
    return time.Now().Hour() > 17
})).Ask("I'm available 24/7 to help!").Build().
```

## 💡 Key APIs Demonstrated

### Condition Creation
- `OnMissingInfo(fields...)` - Field-based conditions
- `OnMessageCount(n)` - Message count thresholds
- `WhenContains(text)` - Content detection
- `WhenFunc(name, fn)` - Custom function conditions

### Logical Operators
- `agent.And(conditions...)` - All conditions must be true
- `agent.Or(conditions...)` - Any condition must be true  
- `agent.Not(condition)` - Negates a condition

### Flow Actions
- `.Ask(message)` - Prompt user with specific message
- `.Summarize()` - Request conversation summary
- `.UseTemplate(template)` - Apply dynamic instructions

This example showcases how go-agent's condition system enables the creation of **intelligent, adaptive, and empathetic** conversational agents with **elegant, maintainable code**.