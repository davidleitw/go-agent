# Basic Chat Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates the fundamental usage of the go-agent framework by creating a simple conversational AI agent.

## Overview

The basic chat example showcases:
- **Environment Configuration**: Loading OpenAI API key from `.env` file
- **Agent Creation**: Using functional options pattern to configure an agent
- **Simple Conversations**: Executing multiple conversation turns
- **Session Management**: Tracking conversation history
- **Detailed Logging**: Comprehensive logging for debugging and monitoring

## Code Structure

### Key Components

1. **Environment Setup**
   ```go
   if err := godotenv.Load("../../../.env"); err != nil {
       log.Printf("Warning: Could not load .env file: %v", err)
   }
   
   apiKey := os.Getenv("OPENAI_API_KEY")
   if apiKey == "" {
       log.Fatal("❌ OPENAI_API_KEY environment variable is required")
   }
   ```
   - Loads environment variables from `.env` file
   - Validates that OpenAI API key is present
   - Provides clear error messages if configuration is missing

2. **Agent Creation**
   ```go
   assistant, err := agent.New(
       agent.WithName("helpful-assistant"),
       agent.WithDescription("A helpful AI assistant for general conversations"),
       agent.WithInstructions("You are a helpful, friendly AI assistant..."),
       agent.WithOpenAI(apiKey),
       agent.WithModel("gpt-4"),
       agent.WithModelSettings(&agent.ModelSettings{
           Temperature: floatPtr(0.7),
           MaxTokens:   intPtr(1000),
       }),
       agent.WithSessionStore(agent.NewInMemorySessionStore()),
       agent.WithDebugLogging(),
   )
   ```
   - Uses functional options pattern for clean configuration
   - Configures OpenAI with GPT-4 model
   - Sets temperature to 0.7 for balanced creativity/consistency
   - Enables debug logging for detailed execution traces

3. **Conversation Flow**
   ```go
   conversations := []struct {
       user     string
       expected string
   }{
       {
           user:     "Hello! How are you doing today?",
           expected: "greeting response",
       },
       // ... more examples
   }
   ```
   - Predefined conversation examples for consistent testing
   - Each turn demonstrates different types of interactions

4. **Response Processing**
   ```go
   response, structuredOutput, err := assistant.Chat(ctx, sessionID, conv.user)
   if err != nil {
       log.Printf("❌ ERROR[%d]: Failed to get response: %v", i+1, err)
       continue
   }
   
   fmt.Printf("🤖 Assistant: %s\n", response.Content)
   ```
   - Handles errors gracefully
   - Logs response details for debugging
   - Displays formatted output to user

## Logging System

The example includes comprehensive logging at multiple levels:

- **REQUEST**: User input and request parameters
- **RESPONSE**: LLM response details including duration and content length
- **SESSION**: Session state tracking and message counts
- **ERROR**: Detailed error information

### Log Output Example
```
✅ OpenAI API key loaded (length: 51)
📝 Creating AI agent...
✅ Agent 'helpful-assistant' created successfully
🆔 Session ID: basic-chat-1704067200
REQUEST[1]: Sending user input to agent
REQUEST[1]: Input: Hello! How are you doing today?
RESPONSE[1]: Duration: 1.234s
RESPONSE[1]: Content length: 87 characters
SESSION[1]: Total messages: 2
```

## Running the Example

### Prerequisites
1. Go 1.22 or later
2. OpenAI API key

### Setup
1. **Configure API Key**:
   ```bash
   # From the root directory
   cp .env.example .env
   # Edit .env and add your OPENAI_API_KEY
   ```

2. **Install Dependencies**:
   ```bash
   cd cmd/examples/basic-chat
   go mod tidy
   ```

3. **Run the Example**:
   ```bash
   go run main.go
   ```

### Expected Output
```
🤖 Basic Chat Agent Example
==================================================
✅ OpenAI API key loaded (length: 51)
📝 Creating AI agent...
✅ Agent 'helpful-assistant' created successfully

💬 Starting conversation...
==================================================

🔄 Turn 1/3
👤 User: Hello! How are you doing today?
🤖 Assistant: Hello! I'm doing great, thank you for asking...

🔄 Turn 2/3
👤 User: What's the weather like?
🤖 Assistant: I don't have access to current weather data...

🔄 Turn 3/3
👤 User: Can you help me write a simple Python function to add two numbers?
🤖 Assistant: Certainly! Here's a simple Python function...

==================================================
✅ Conversation completed successfully!
📊 Session Summary:
   • Session ID: basic-chat-1704067200
   • Total messages: 6
   • Created at: 2024-01-01 12:00:00
   • Updated at: 2024-01-01 12:00:15
```

## Key Learning Points

1. **Functional Options Pattern**: Clean and extensible configuration
2. **Error Handling**: Robust error checking and graceful degradation
3. **Session Management**: Automatic conversation history tracking
4. **Logging Strategy**: Multi-level logging for debugging and monitoring
5. **Environment Configuration**: Secure API key management

## Troubleshooting

### Common Issues

1. **Missing API Key**
   ```
   ❌ OPENAI_API_KEY environment variable is required
   ```
   - Solution: Ensure your `.env` file contains a valid OpenAI API key

2. **Import Errors**
   ```
   package basic-chat is not in GOROOT or GOPATH
   ```
   - Solution: Run `go mod tidy` from the example directory

3. **Network Issues**
   ```
   Failed to get response: connection timeout
   ```
   - Solution: Check internet connection and OpenAI API status

### Debug Tips

1. **Enable Debug Logging**: The example already includes `agent.WithDebugLogging()`
2. **Check Session State**: Examine the session summary at the end
3. **Monitor Response Times**: Look for unusually slow responses in logs
4. **Validate API Key**: Ensure the key has sufficient credits and permissions

## Next Steps

After running this example successfully:
1. Try the **Task Completion** example for advanced condition handling
2. Explore the **Calculator Tool** example for function calling
3. Modify the conversation examples to test different scenarios
4. Experiment with different model settings (temperature, max tokens)