# Session Management Example

This example demonstrates the new simplified Session interface in go-agent and shows how to effectively manage conversation sessions.

## New Session Interface (v0.0.2+)

The Session interface has been simplified to focus on core functionality:

```go
type Session interface {
    ID() string
    Messages() []Message
    AddMessage(role, content string) Message
    GetData(key string) interface{}
    SetData(key string, value interface{})
}
```

## Key Improvements

- **Simplified API**: Reduced from 12 methods to 5 core methods
- **Thread-safe**: All operations are safe for concurrent use
- **Immutable messages**: Messages returned are copies to prevent external modification
- **Flexible data storage**: Store any type of data with the session
- **Better performance**: Optimized for common use cases

## Running the Example

```bash
# Run the example
go run main.go

# Run the tests
go test -v
```

## What This Example Demonstrates

### 1. Basic Session Operations
- Creating sessions with unique IDs
- Adding messages with different roles (user, assistant, system, tool)
- Retrieving conversation history
- Message immutability and timestamps

### 2. Session Data Storage
- Storing various data types (strings, numbers, maps, timestamps)
- Retrieving and updating session data
- Handling non-existent keys gracefully

### 3. Session Cloning
- Creating independent copies of sessions
- Verifying data and message independence
- Use cases for session branching

### 4. Agent Integration
- Using sessions with agents for conversations
- Preserving session context across chat interactions
- Managing session state during agent operations

### 5. Concurrent Access
- Thread-safe operations across multiple goroutines
- Concurrent message addition and data updates
- Performance under concurrent load

## Session Data Storage Examples

```go
session := agent.NewSession("example")

// Store different types of data
session.SetData("user_id", "user_12345")
session.SetData("login_attempts", 3)
session.SetData("preferences", map[string]string{
    "language": "en",
    "theme":    "dark",
})
session.SetData("last_activity", time.Now())

// Retrieve data with type assertions
userID := session.GetData("user_id").(string)
attempts := session.GetData("login_attempts").(int)

if prefs, ok := session.GetData("preferences").(map[string]string); ok {
    fmt.Printf("User preferences: %+v\n", prefs)
}
```

## Message Management Examples

```go
session := agent.NewSession("conversation")

// Add different types of messages
userMsg := session.AddMessage(agent.RoleUser, "Hello!")
assistantMsg := session.AddMessage(agent.RoleAssistant, "Hi there! How can I help?")
systemMsg := session.AddMessage(agent.RoleSystem, "User authenticated")

// Access message properties
fmt.Printf("User said: %s at %s\n", userMsg.Content, userMsg.Timestamp)

// Get all messages
for i, msg := range session.Messages() {
    fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
}
```

## Advanced Features

### Session Cloning
```go
// Clone a session for branching conversations
if cloneable, ok := session.(interface{ Clone() Session }); ok {
    branchedSession := cloneable.Clone()
    // Now you have two independent sessions
}
```

### Agent Integration
```go
// Use sessions with agents
agent, _ := agent.NewBasicAgent(agent.BasicAgentConfig{
    Name:      "my-agent",
    ChatModel: chatModel,
})

session := agent.NewSession("chat-session")
response, output, err := agent.Chat(ctx, session, "Hello!")
```

## Migration from Previous Version

If you're migrating from the previous Session interface:

**Old API:**
```go
session.AddUserMessage("Hello")
session.AddAssistantMessage("Hi there")
session.GetLastMessage()
session.GetMessagesByRole(agent.RoleUser)
```

**New API:**
```go
session.AddMessage(agent.RoleUser, "Hello")
session.AddMessage(agent.RoleAssistant, "Hi there")

// For advanced operations, use helper functions or type assertions
messages := session.Messages()
lastMessage := messages[len(messages)-1] // Get last message

// Filter by role
var userMessages []agent.Message
for _, msg := range session.Messages() {
    if msg.Role == agent.RoleUser {
        userMessages = append(userMessages, msg)
    }
}
```

## Performance Notes

- Session operations are optimized for concurrent access
- Message retrieval returns copies to ensure immutability
- Data storage supports any serializable Go type
- Memory usage is optimized for typical conversation lengths

## Testing

The example includes comprehensive tests demonstrating:
- Basic functionality verification
- Concurrent access patterns
- Integration with mock agents
- Complex data structure handling
- Message type validation

Run tests with:
```bash
go test -v -race  # Include race detection
```