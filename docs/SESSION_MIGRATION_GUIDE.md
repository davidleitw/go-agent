# Session API Migration Guide

This guide helps you migrate from the old Session interface to the new simplified Session interface introduced in v0.0.2.

## Overview

The Session interface has been significantly simplified to improve performance, usability, and maintainability. The new interface reduces method count from 12 to 5 core methods while maintaining all essential functionality.

## API Changes Summary

### Old Interface (v0.0.1)
```go
type Session interface {
    ID() string
    Messages() []Message
    AddMessage(msg Message)
    AddUserMessage(content string) Message
    AddAssistantMessage(content string) Message
    AddSystemMessage(content string) Message
    AddToolMessage(toolCallID, toolName, content string) Message
    GetLastMessage() *Message
    GetMessagesByRole(role string) []Message
    MessageCount() int
    Clear()
    Clone() Session
    // ... plus data storage methods
}
```

### New Interface (v0.0.2+)
```go
type Session interface {
    ID() string
    Messages() []Message
    AddMessage(role, content string) Message
    GetData(key string) interface{}
    SetData(key string, value interface{})
}
```

## Migration Steps

### 1. Update Message Addition

**Old API:**
```go
// Role-specific methods
session.AddUserMessage("Hello!")
session.AddAssistantMessage("Hi there!")
session.AddSystemMessage("User authenticated")
session.AddToolMessage("call_123", "calculator", `{"result": 42}`)

// Generic method with Message object
userMsg := agent.NewUserMessage("Hello!")
session.AddMessage(userMsg)
```

**New API:**
```go
// Single method with role parameter
session.AddMessage(agent.RoleUser, "Hello!")
session.AddMessage(agent.RoleAssistant, "Hi there!")
session.AddMessage(agent.RoleSystem, "User authenticated")
session.AddMessage(agent.RoleTool, `{"result": 42}`)
```

### 2. Update Message Retrieval

**Old API:**
```go
// Get last message
lastMsg := session.GetLastMessage()

// Get messages by role
userMessages := session.GetMessagesByRole(agent.RoleUser)

// Get message count
count := session.MessageCount()
```

**New API:**
```go
// Get all messages and process
messages := session.Messages()

// Get last message
var lastMsg *agent.Message
if len(messages) > 0 {
    lastMsg = &messages[len(messages)-1]
}

// Filter messages by role
var userMessages []agent.Message
for _, msg := range messages {
    if msg.Role == agent.RoleUser {
        userMessages = append(userMessages, msg)
    }
}

// Get message count
count := len(session.Messages())
```

### 3. Update Session Operations

**Old API:**
```go
// Clear session
session.Clear()

// Clone session
clonedSession := session.Clone()
```

**New API:**
```go
// Clear session (use type assertion for advanced features)
if clearable, ok := session.(interface{ ClearMessages() }); ok {
    clearable.ClearMessages()
}

// Clone session (use type assertion for advanced features)
if cloneable, ok := session.(interface{ Clone() Session }); ok {
    clonedSession := cloneable.Clone()
}
```

### 4. Update Data Storage

**Old API:**
```go
// Data storage was often custom or part of larger session objects
// No standardized way to store session data
```

**New API:**
```go
// Built-in data storage
session.SetData("user_id", "user_12345")
session.SetData("context", map[string]string{"theme": "dark"})

// Retrieve data
userID := session.GetData("user_id").(string)
context := session.GetData("context").(map[string]string)
```

## Migration Examples

### Example 1: Basic Chat Application

**Old Code:**
```go
func handleUserInput(session agent.Session, input string) error {
    // Add user message
    session.AddUserMessage(input)
    
    // Process with AI
    response := generateAIResponse(session.Messages())
    
    // Add AI response
    session.AddAssistantMessage(response)
    
    // Check message count
    if session.MessageCount() > 10 {
        session.Clear()
    }
    
    return nil
}
```

**New Code:**
```go
func handleUserInput(session agent.Session, input string) error {
    // Add user message
    session.AddMessage(agent.RoleUser, input)
    
    // Process with AI
    response := generateAIResponse(session.Messages())
    
    // Add AI response
    session.AddMessage(agent.RoleAssistant, response)
    
    // Check message count
    if len(session.Messages()) > 10 {
        if clearable, ok := session.(interface{ ClearMessages() }); ok {
            clearable.ClearMessages()
        }
    }
    
    return nil
}
```

### Example 2: Tool Integration

**Old Code:**
```go
func executeTool(session agent.Session, toolCall agent.ToolCall) error {
    result, err := callExternalAPI(toolCall.Function.Arguments)
    if err != nil {
        session.AddToolMessage(toolCall.ID, toolCall.Function.Name, 
            fmt.Sprintf("Error: %v", err))
        return err
    }
    
    resultJSON, _ := json.Marshal(result)
    session.AddToolMessage(toolCall.ID, toolCall.Function.Name, string(resultJSON))
    return nil
}
```

**New Code:**
```go
func executeTool(session agent.Session, toolCall agent.ToolCall) error {
    result, err := callExternalAPI(toolCall.Function.Arguments)
    if err != nil {
        session.AddMessage(agent.RoleTool, fmt.Sprintf("Error: %v", err))
        return err
    }
    
    resultJSON, _ := json.Marshal(result)
    session.AddMessage(agent.RoleTool, string(resultJSON))
    return nil
}
```

### Example 3: Session Analysis

**Old Code:**
```go
func analyzeConversation(session agent.Session) {
    userMsgCount := len(session.GetMessagesByRole(agent.RoleUser))
    assistantMsgCount := len(session.GetMessagesByRole(agent.RoleAssistant))
    
    lastMsg := session.GetLastMessage()
    if lastMsg != nil {
        fmt.Printf("Last message from: %s\n", lastMsg.Role)
    }
    
    fmt.Printf("User: %d, Assistant: %d messages\n", userMsgCount, assistantMsgCount)
}
```

**New Code:**
```go
func analyzeConversation(session agent.Session) {
    messages := session.Messages()
    userMsgCount := 0
    assistantMsgCount := 0
    
    for _, msg := range messages {
        switch msg.Role {
        case agent.RoleUser:
            userMsgCount++
        case agent.RoleAssistant:
            assistantMsgCount++
        }
    }
    
    if len(messages) > 0 {
        lastMsg := messages[len(messages)-1]
        fmt.Printf("Last message from: %s\n", lastMsg.Role)
    }
    
    fmt.Printf("User: %d, Assistant: %d messages\n", userMsgCount, assistantMsgCount)
}
```

## Helper Functions

To ease migration, you can create helper functions that provide the old API semantics:

```go
// Helper functions for backward compatibility
func AddUserMessage(session agent.Session, content string) agent.Message {
    return session.AddMessage(agent.RoleUser, content)
}

func AddAssistantMessage(session agent.Session, content string) agent.Message {
    return session.AddMessage(agent.RoleAssistant, content)
}

func AddSystemMessage(session agent.Session, content string) agent.Message {
    return session.AddMessage(agent.RoleSystem, content)
}

func GetLastMessage(session agent.Session) *agent.Message {
    messages := session.Messages()
    if len(messages) == 0 {
        return nil
    }
    return &messages[len(messages)-1]
}

func GetMessagesByRole(session agent.Session, role string) []agent.Message {
    var filtered []agent.Message
    for _, msg := range session.Messages() {
        if msg.Role == role {
            filtered = append(filtered, msg)
        }
    }
    return filtered
}

func MessageCount(session agent.Session) int {
    return len(session.Messages())
}
```

## Benefits of the New API

### 1. Simplified Interface
- Reduced from 12 methods to 5 core methods
- Easier to understand and implement
- Less cognitive overhead

### 2. Better Performance
- Optimized message storage and retrieval
- Reduced memory allocations
- Thread-safe operations

### 3. Improved Testability
- Cleaner mock implementations
- Easier to create test scenarios
- Better integration with test frameworks

### 4. Enhanced Flexibility
- Built-in data storage for session context
- Support for session cloning
- Extensible through type assertions

## Breaking Changes

### Removed Methods
- `AddUserMessage()` → Use `AddMessage(agent.RoleUser, content)`
- `AddAssistantMessage()` → Use `AddMessage(agent.RoleAssistant, content)`
- `AddSystemMessage()` → Use `AddMessage(agent.RoleSystem, content)`
- `AddToolMessage()` → Use `AddMessage(agent.RoleTool, content)`
- `GetLastMessage()` → Use `Messages()[len(Messages())-1]`
- `GetMessagesByRole()` → Filter `Messages()` manually
- `MessageCount()` → Use `len(Messages())`
- `Clear()` → Use type assertion for `ClearMessages()`
- `Clone()` → Use type assertion for `Clone()`

### Changed Behavior
- Message addition now requires explicit role specification
- Advanced features require type assertions
- Data storage is now built-in with `SetData()`/`GetData()`

## Testing Migration

### Old Test Patterns
```go
func TestOldSession(t *testing.T) {
    session := agent.NewSession("test")
    
    session.AddUserMessage("Hello")
    if session.MessageCount() != 1 {
        t.Error("Expected 1 message")
    }
    
    lastMsg := session.GetLastMessage()
    if lastMsg.Content != "Hello" {
        t.Error("Wrong message content")
    }
}
```

### New Test Patterns
```go
func TestNewSession(t *testing.T) {
    session := agent.NewSession("test")
    
    session.AddMessage(agent.RoleUser, "Hello")
    messages := session.Messages()
    if len(messages) != 1 {
        t.Error("Expected 1 message")
    }
    
    if messages[0].Content != "Hello" {
        t.Error("Wrong message content")
    }
}
```

## Migration Checklist

- [ ] Update all `AddUserMessage()` calls to `AddMessage(agent.RoleUser, content)`
- [ ] Update all `AddAssistantMessage()` calls to `AddMessage(agent.RoleAssistant, content)`
- [ ] Update all `AddSystemMessage()` calls to `AddMessage(agent.RoleSystem, content)`
- [ ] Update all `AddToolMessage()` calls to `AddMessage(agent.RoleTool, content)`
- [ ] Replace `GetLastMessage()` with manual indexing of `Messages()`
- [ ] Replace `GetMessagesByRole()` with manual filtering of `Messages()`
- [ ] Replace `MessageCount()` with `len(Messages())`
- [ ] Update `Clear()` calls to use type assertion
- [ ] Update `Clone()` calls to use type assertion
- [ ] Migrate session data storage to `SetData()`/`GetData()`
- [ ] Update test cases to use new API
- [ ] Run comprehensive tests to ensure compatibility

## Need Help?

If you encounter issues during migration:

1. Check the [Session Management Example](../examples/session-management/) for complete usage patterns
2. Review the [API Documentation](../README.md#session-management) for detailed interface information
3. Open an issue on [GitHub](https://github.com/davidleitw/go-agent/issues) if you need assistance

The new Session API provides a cleaner, more performant foundation for building conversational AI applications while maintaining all essential functionality.