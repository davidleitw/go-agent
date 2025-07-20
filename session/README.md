# Session Module

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

The Session module provides conversation state and history management for the Go Agent Framework. It's designed to be simple yet complete, supporting in-memory storage with extensible interfaces for future implementations.

## Features

- **State Management**: Key-Value storage supporting any data type
- **History Tracking**: Unified Entry structure supporting multiple conversation types
- **Lifecycle Management**: TTL support with automatic expiration cleanup
- **Thread Safety**: All operations are concurrent-safe
- **JSON Serialization**: Full JSON support for all data structures with proper tags
- **Extensible**: Reserved interfaces for Redis, database, and other persistent implementations

## Quick Start

```go
import (
    "github.com/davidleitw/go-agent/session"
    "github.com/davidleitw/go-agent/session/memory"
)

// Create store
store := memory.NewStore()
defer store.Close() // Clean up resources

// Create session
sess := store.Create(
    session.WithTTL(24 * time.Hour),
    session.WithID("custom-id"), // Optional
)

// State management
sess.Set("current_task", "booking_flight")
sess.Set("user_preference", map[string]string{
    "language": "zh-TW",
    "currency": "TWD",
})

// Add conversation history
sess.AddEntry(session.NewMessageEntry("user", "Book a flight to Tokyo"))
sess.AddEntry(session.NewToolCallEntry("search_flights", map[string]any{
    "destination": "Tokyo",
}))

// Save (no-op for memory implementation, but keeps interface consistent)
err := store.Save(sess)

// Retrieve session
retrieved, err := store.Get(sess.ID())
history := retrieved.GetHistory(10) // Get recent 10 entries

// JSON serialization support
entryJSON, _ := json.Marshal(history[0])
fmt.Println(string(entryJSON))
```

## API Reference

### Session Interface

```go
type Session interface {
    // Basic information
    ID() string
    CreatedAt() time.Time
    UpdatedAt() time.Time
    
    // State management
    Get(key string) (any, bool)
    Set(key string, value any)
    Delete(key string)
    
    // History management
    AddEntry(entry Entry) error
    GetHistory(limit int) []Entry
}
```

### SessionStore Interface

```go
type SessionStore interface {
    Create(opts ...CreateOption) Session
    Get(id string) (Session, error)
    Save(session Session) error
    Delete(id string) error
    DeleteExpired() error
    Close() error
}
```

### Entry Types

Four types of conversation records are supported:

1. **Message**: User/assistant/system messages
2. **ToolCall**: Tool invocation records
3. **ToolResult**: Tool execution results
4. **Thinking**: Internal reasoning process (reserved)

Each type has corresponding creation functions and type-safe extraction functions:

```go
// Creation
entry := session.NewMessageEntry("user", "Hello")
entry := session.NewToolCallEntry("search", params)
entry := session.NewToolResultEntry("search", result, err)

// Extraction
if content, ok := session.GetMessageContent(entry); ok {
    fmt.Printf("%s: %s\n", content.Role, content.Text)
}
```

## Design Decisions

### Why Keep the Save() Method?

Although `Save()` is a no-op in the memory implementation, we keep it to:
- Support future persistent implementations (Redis, databases)
- Allow batch update optimizations
- Maintain interface consistency

### Why Remove Touch() and ExpiresAt()?

- `Touch()`: UpdatedAt is automatically updated on every modification
- `ExpiresAt()`: Internal implementation detail, no need to expose to users

### Background Cleanup Mechanism

- Automatic cleanup of expired sessions every 5 minutes
- Provides `Close()` method for graceful shutdown
- Manual cleanup available via `DeleteExpired()`

## Extension Guide

### Implementing Custom SessionStore

```go
type MyStore struct {
    // Your implementation
}

func (s *MyStore) Create(opts ...session.CreateOption) session.Session {
    // Implement creation logic
}

func (s *MyStore) Get(id string) (session.Session, error) {
    // Implement retrieval logic
}

func (s *MyStore) Save(sess session.Session) error {
    // Implement save logic
}

// ... other methods
```

### Planned Extensions

1. **Redis Store**
   - Distributed session support
   - Automatic expiration (using Redis TTL)
   - Batch operation optimization

2. **Database Store**
   - SQL/NoSQL support
   - Transactional guarantees
   - Query and analytics capabilities

3. **Hybrid Storage**
   - Memory as L1 cache
   - Redis/DB as persistent layer
   - Automatic synchronization

## Integration with Agent System

*To be implemented: This section will be filled in after the Agent core functionality is complete*

```go
// Expected integration approach
agent := agent.New(
    agent.WithSessionStore(store),
    // Other configurations...
)

// Agent automatically manages sessions internally
response := agent.Process(ctx, request)
```

## Performance Considerations

- Memory implementation suitable for single-machine applications
- Recommend limiting history size (e.g., max 1000 entries)
- Regular cleanup of expired sessions to prevent memory leaks
- Large-scale applications should use Redis/DB implementations

## JSON Serialization

All session data structures include proper JSON tags for serialization:

```go
// Serialize a complete session history
history := session.GetHistory(0)
jsonData, err := json.Marshal(history)

// Serialize individual entries
entry := session.NewMessageEntry("user", "Hello")
entryJSON, err := json.Marshal(entry)

// Deserialize entries
var deserializedEntry session.Entry
err = json.Unmarshal(entryJSON, &deserializedEntry)
```

### JSON Structure Examples

**Message Entry:**
```json
{
  "id": "uuid-string",
  "type": "message", 
  "timestamp": "2024-01-01T12:00:00Z",
  "content": {
    "role": "user",
    "text": "Hello world"
  },
  "metadata": {}
}
```

**Tool Call Entry:**
```json
{
  "id": "uuid-string",
  "type": "tool_call",
  "timestamp": "2024-01-01T12:00:00Z", 
  "content": {
    "tool": "search",
    "parameters": {"query": "flights"}
  },
  "metadata": {}
}
```

**Tool Result Entry:**
```json
{
  "id": "uuid-string",
  "type": "tool_result",
  "timestamp": "2024-01-01T12:00:00Z",
  "content": {
    "tool": "search", 
    "success": true,
    "result": ["result1", "result2"],
    "error": ""
  },
  "metadata": {}
}
```

## Best Practices

1. **State Management**
   - Use structured key naming (e.g., `task.current`, `user.preference`)
   - Avoid storing overly large objects
   - Encrypt sensitive information before storage

2. **History Records**
   - Use limit parameter appropriately to avoid loading too much history
   - Regularly archive old records
   - Use Metadata to add additional information

3. **Lifecycle**
   - Set reasonable TTL based on use case
   - Remember to call `Close()` to clean up resources
   - Monitor session count to prevent leaks

## Error Handling

```go
sess, err := store.Get(sessionID)
if errors.Is(err, session.ErrSessionNotFound) {
    // Create new session
    sess = store.Create()
}
```

## Example Projects

Check the `/examples/session/` directory for complete examples.

## Contributing

Welcome to submit Issues and Pull Requests! Please ensure:
- Add appropriate tests
- Update relevant documentation
- Follow Go coding conventions

## License

MIT License - See LICENSE file in the project root directory.