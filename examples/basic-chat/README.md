# Basic Chat Example

This example demonstrates the **elegant and simplified** usage of the go-agent framework for creating conversational AI assistants with minimal code.

## 🎯 What This Example Shows

- **One-line agent creation** with fluent builder pattern
- **Automatic OpenAI integration** - just provide your API key
- **Built-in session management** - no manual session handling required
- **Clean conversation flow** with automatic message tracking
- **Zero boilerplate code** - focus on your agent's behavior, not infrastructure

## 🚀 Running the Example

```bash
cd cmd/examples/basic-chat
go run main.go
```

Make sure you have your `OPENAI_API_KEY` set in your environment or `.env` file.

## 🏗️ Key Features

### Elegant Agent Creation
```go
assistant, err := goagent.New("helpful-assistant").
    WithOpenAI(apiKey).
    WithModel("gpt-4o-mini").
    WithDescription("A helpful AI assistant for general conversations").
    WithInstructions("You are a helpful, friendly AI assistant...").
    WithTemperature(0.7).
    WithMaxTokens(1000).
    Build()
```

### Simplified Conversation
```go
// Sessions are automatically managed
response, err := assistant.Chat(ctx, userInput, 
    goagent.WithSession(sessionID))
```

### Clean Response Handling
```go
// Access everything through the response object
fmt.Println(response.Message)           // Agent's text response
data := response.Data                   // Structured output (if any)
session := response.Session             // Session state
metadata := response.Metadata           // Additional info
```

## 🔧 Code Highlights

1. **Fluent Builder Pattern**: Chain configuration methods for readable, self-documenting code.

2. **Automatic Provider Setup**: Just specify `WithOpenAI(apiKey)` and everything else is handled.

3. **Zero Boilerplate**: No manual chat model creation, session stores, or runners needed.

4. **Type-Safe Configuration**: Full compile-time checking with IntelliSense support.

5. **Progressive Complexity**: Start simple, add features as needed without rewriting.

## 📊 Expected Output

The example will demonstrate:
1. **Elegant agent creation** with fluent API
2. **Automatic conversation management** 
3. **Clean response handling**
4. **Built-in logging and metadata**
5. **Error handling with helpful messages**

Each conversation turn shows:
- User input
- Agent response with timing
- Automatic session management
- Zero configuration overhead

## 🔄 Comparison: Before vs After

### Before (Old API - ~60 lines)
```go
// Manual chat model creation
chatModel, err := openai.NewChatModel(apiKey, nil)
// Complex configuration struct
agent, err := agent.NewBasicAgent(agent.BasicAgentConfig{
    Name: "assistant",
    Description: "...",
    Instructions: "...",
    Model: "gpt-4o-mini",
    ModelSettings: &agent.ModelSettings{...},
    ChatModel: chatModel,
})
// Manual session management
session := agent.NewSession(sessionID)
// Complex response handling
response, output, err := agent.Chat(ctx, session, input)
```

### After (New API - ~15 lines)
```go
// One-line creation with fluent API
agent, err := goagent.New("assistant").
    WithOpenAI(apiKey).
    WithModel("gpt-4o-mini").
    WithInstructions("...").
    Build()

// Simple conversation with automatic session management
response, err := agent.Chat(ctx, input, goagent.WithSession(sessionID))
fmt.Println(response.Message)
```

## 🎓 Learning Objectives

After studying this example, you should understand:

1. **Simplified Agent Creation**: How the new fluent API eliminates boilerplate
2. **Automatic Session Management**: How sessions are handled transparently
3. **Clean Response Processing**: How to access all response data elegantly
4. **Progressive Complexity**: How to start simple and add features incrementally
5. **Type Safety**: How the builder pattern provides compile-time validation

## 🔄 Next Steps

After running this basic example, explore:
- **[Calculator Tool Example](../calculator-tool/)** - Adding function-based tools
- **[Advanced Conditions Example](../advanced-conditions/)** - Sophisticated flow control
- **[Multi-Tool Agent Example](../multi-tool-agent/)** - Coordinating multiple capabilities
- **[Task Completion Example](../task-completion/)** - Structured output handling

## 💡 Key APIs Demonstrated

### Agent Creation APIs
- `goagent.New(name)` - Start building an agent
- `.WithOpenAI(apiKey)` - Configure OpenAI provider
- `.WithModel(model)` - Set the language model
- `.WithInstructions(text)` - Define agent behavior
- `.WithTemperature(float)` - Set creativity level
- `.Build()` - Create the final agent

### Conversation APIs
- `agent.Chat(ctx, input, opts...)` - Have a conversation
- `goagent.WithSession(id)` - Specify session ID
- `response.Message` - Get agent's response text
- `response.Session` - Access session state

## 🐛 Common Issues

1. **API Key Problems**: Ensure `OPENAI_API_KEY` is set correctly
2. **Import Errors**: Make sure to import `pkg/goagent` not `pkg/agent`
3. **Build Errors**: Remember to call `.Build()` at the end of the chain
4. **Context Timeout**: Use appropriate context timeouts for longer conversations

This example showcases how go-agent makes AI agent development **elegant, simple, and productive** with minimal code and maximum clarity.