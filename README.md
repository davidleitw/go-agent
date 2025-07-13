# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent" width="200" height="200">
</div>

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

A lightweight Go AI agent framework for building intelligent conversations and automated workflows with efficiency.

## Features

- 🚀 **Lightweight & Fast**: Minimal abstractions focused on core functionality
- ⚡ **Functional Options**: Clean, intuitive APIs using Go's functional options pattern
- 🔌 **Pluggable Architecture**: Support for multiple LLM providers and storage backends
- 🛠️ **Tool Integration**: Easy integration of custom tools and function calling
- 🔄 **Flow Control**: Dynamic conversation flow with conditional rules
- 📝 **Structured Output**: Built-in support for validated JSON output
- 💾 **Session Management**: Persistent conversation history for backend scenarios
- 🧪 **Testing Support**: Comprehensive mocking and testing utilities

## Quick Start

### Installation

```bash
go get github.com/davidleitw/go-agent
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/davidleitw/go-agent/pkg/agent"
    "github.com/davidleitw/go-agent/pkg/openai"
)

func main() {
    // Create OpenAI chat model
    chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
    if err != nil {
        log.Fatal(err)
    }

    // Create an agent with functional options
    assistant, err := agent.New(
        agent.WithName("helpful-assistant"),
        agent.WithDescription("A helpful AI assistant"),
        agent.WithInstructions("You are a helpful assistant. Be concise and friendly."),
        agent.WithChatModel(chatModel),
        agent.WithModel("gpt-4"),
        agent.WithModelSettings(&agent.ModelSettings{
            Temperature: floatPtr(0.7),
            MaxTokens:   intPtr(1000),
        }),
        agent.WithSessionStore(agent.NewInMemorySessionStore()),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Have a conversation - much simpler!
    ctx := context.Background()
    response, _, err := assistant.Chat(ctx, "session-1", "Hello! How are you?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Assistant:", response.Content)
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }
```

### With Tools

```go
// Define a custom tool
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
    return "get_weather"
}

func (t *WeatherTool) Description() string {
    return "Get current weather for a location"
}

func (t *WeatherTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "The city and state/country",
            },
        },
        "required": []string{"location"},
    }
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    location := args["location"].(string)
    // Simulate weather API call
    return map[string]interface{}{
        "location":    location,
        "temperature": "22°C",
        "condition":   "Sunny",
    }, nil
}

// Create OpenAI chat model
chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
if err != nil {
    log.Fatal(err)
}

// Create agent with tool - much cleaner!
weatherAgent, err := agent.New(
    agent.WithName("weather-assistant"),
    agent.WithInstructions("You can help users get weather information."),
    agent.WithChatModel(chatModel),
    agent.WithModel("gpt-4"),
    agent.WithTools(&WeatherTool{}),
    agent.WithSessionStore(agent.NewInMemorySessionStore()),
)
```

### Structured Output

```go
// Define output structure
type TaskResult struct {
    Title    string   `json:"title" validate:"required"`
    Priority string   `json:"priority" validate:"required,oneof=low medium high"`
    Tags     []string `json:"tags"`
}

// Create OpenAI chat model
chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
if err != nil {
    log.Fatal(err)
}

// Create agent with structured output - much simpler!
taskAgent, err := agent.New(
    agent.WithName("task-creator"),
    agent.WithInstructions("Create tasks based on user input. Return structured JSON."),
    agent.WithChatModel(chatModel),
    agent.WithModel("gpt-4"),
    agent.WithStructuredOutput(&TaskResult{}), // Automatically generates schema
    agent.WithSessionStore(agent.NewInMemorySessionStore()),
)

// The agent will automatically validate and parse the output
response, structuredOutput, err := taskAgent.Chat(ctx, "session-1", "Create a high priority task for code review")
if taskResult, ok := structuredOutput.(*TaskResult); ok {
    fmt.Printf("Created task: %s (Priority: %s)\n", taskResult.Title, taskResult.Priority)
}
```

### Flow Rules

```go
// Create conditional flow rules
missingInfoCondition := agent.NewDataKeyExistsCondition("missing_info_check", "missing_fields")

flowRule, err := agent.NewFlowRule("collect-missing-info", missingInfoCondition).
    WithDescription("Prompt user for missing information").
    WithNewInstructions("Please ask the user for the following missing information: {{missing_fields}}").
    WithRecommendedTools("collect_info").
    WithSystemMessage("The user needs to provide additional information.").
    Build()

// Create OpenAI chat model
chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
if err != nil {
    log.Fatal(err)
}

// Create agent with flow rules
smartAgent, err := agent.New(
    agent.WithName("smart-assistant"),
    agent.WithInstructions("You are a smart assistant that adapts based on context."),
    agent.WithChatModel(chatModel),
    agent.WithModel("gpt-4"),
    agent.WithFlowRules(flowRule),
    agent.WithSessionStore(agent.NewInMemorySessionStore()),
)
```

## Architecture

The framework is designed with clean separation of concerns:

- **`pkg/agent/`**: Core interfaces, implementations, and public APIs
- **`pkg/openai/`**: OpenAI ChatModel implementation

### Core Components

1. **Agent**: Complete AI agent with configuration and execution capabilities
2. **Session**: Conversation history and state management for backend scenarios
3. **Tools**: External capabilities that agents can use
4. **Flow Rules**: Dynamic behavior control based on conditions
5. **Chat Models**: Abstraction for different LLM providers
6. **Storage**: Pluggable session persistence backends

## Supported LLM Providers

- ✅ **OpenAI** (GPT-4, GPT-3.5-turbo, etc.)
- 🔜 **Anthropic** (Claude 3.5 Sonnet, etc.)
- 🔜 **Google** (Gemini)
- 🔜 **Local models** (via Ollama)

## Storage Backends

- ✅ **In-Memory**: For development and testing
- 🔜 **Redis**: For production distributed systems
- 🔜 **PostgreSQL**: For advanced querying and analytics

## Examples

See the [`cmd/examples/`](./cmd/examples/) directory for complete working examples. Each example is a standalone Go program that demonstrates specific features of the go-agent framework.

### 🚀 Quick Setup

1. **Configure your OpenAI API key**:
   ```bash
   # Copy the example environment file
   cp .env.example .env
   
   # Edit .env and add your OpenAI API key
   # OPENAI_API_KEY=your_openai_api_key_here
   ```

2. **Install dependencies** (for examples):
   ```bash
   go mod download
   ```

### 📋 Available Examples

#### 1. **Basic Chat** (`cmd/examples/basic-chat/`)
Simple conversational AI demonstrating core framework usage.

**Features**:
- Environment variable configuration (.env support)
- Basic agent creation with functional options
- Simple conversation flow
- Detailed logging for troubleshooting

**Run the example**:
```bash
cd cmd/examples/basic-chat
go run main.go
```

**What it demonstrates**:
- Agent creation with `agent.New()`
- OpenAI integration
- Session management
- Basic conversation handling

---

#### 2. **Task Completion** (`cmd/examples/task-completion/`)
Advanced example showing condition validation and iterative information collection.

**Features**:
- **Condition-based flow**: Demonstrates missing information detection
- **Structured output**: Uses JSON schema for status tracking
- **Iterative collection**: Simulates a restaurant reservation system
- **Completion detection**: LLM sets completion flag when all conditions are met
- **Safety limits**: Maximum 5 iterations to prevent token overuse

**Run the example**:
```bash
cd cmd/examples/task-completion
go run main.go
```

**What it demonstrates**:
- Structured output with custom types (`ReservationStatus`)
- Condition validation logic
- Multi-turn conversation management
- LLM-driven completion flag detection
- Detailed process logging

**Simulated Flow**:
1. User: "我想要預訂餐廳，我是李先生" → Missing: phone, date, time, party_size
2. User: "我的電話是0912345678，想要明天晚上7點" → Missing: party_size
3. User: "4個人" → All conditions met, completion_flag = true

---

#### 3. **Calculator Tool** (`cmd/examples/calculator-tool/`)
Demonstrates tool integration and OpenAI function calling.

**Features**:
- **Custom tool implementation**: Mathematical calculator
- **Function calling**: OpenAI tool integration
- **Multiple operations**: Add, subtract, multiply, divide, power, square root
- **Structured results**: Tool returns detailed calculation steps
- **Error handling**: Division by zero, invalid operations, etc.

**Run the example**:
```bash
cd cmd/examples/calculator-tool
go run main.go
```

**What it demonstrates**:
- Custom tool implementation (`agent.Tool` interface)
- OpenAI function calling mechanism
- Tool parameter validation
- Structured tool responses
- Tool execution logging

**Supported Operations**:
- `add`: Addition (15 + 27)
- `subtract`: Subtraction (125 - 47)
- `multiply`: Multiplication (13 × 7)
- `divide`: Division (144 ÷ 12)
- `power`: Exponentiation (2^8)
- `sqrt`: Square root (√64)

### 🔧 Troubleshooting

All examples include detailed logging to help you understand the execution flow:

- **REQUEST**: User input and request parameters
- **AGENT**: Agent processing and decision making
- **TOOL**: Tool execution details and results
- **RESPONSE**: LLM responses and parsing results
- **SESSION**: Session state changes
- **STRUCTURED**: Structured output parsing
- **ERROR**: Error details and recovery

**Common Issues**:

1. **Missing API Key**: Make sure `OPENAI_API_KEY` is set in your `.env` file
2. **Import Errors**: Ensure you're running from the example directory
3. **Module Issues**: Run `go mod tidy` in the example directory

**Example Logs**:
```
✅ OpenAI API key loaded (length: 51)
📝 Creating AI agent...
✅ Agent 'helpful-assistant' created successfully
REQUEST[1]: Sending user input to agent
RESPONSE[1]: Duration: 1.234s
SESSION[1]: Total messages: 2
```

## Development

### Prerequisites

- Go 1.21 or later
- (Optional) golangci-lint for linting

### Building

```bash
make build
```

### Testing

```bash
# Run all tests
make test

# Run only unit tests
make unit-test

# Run with coverage
make coverage
```

### Linting

```bash
make lint
```

## API Documentation

For detailed API documentation, see:

- [Getting Started Guide](./docs/getting-started.md)
- [API Reference](./docs/api-reference.md)
- [Architecture Overview](./docs/architecture.md)
- [Examples](./docs/examples.md)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Additional LLM providers (Anthropic, Google, etc.)
- [ ] Advanced storage backends (Redis, PostgreSQL)
- [ ] Streaming response support
- [ ] Multi-agent orchestration
- [ ] Observability and metrics
- [ ] Web UI for agent management
- [ ] Plugin system for custom extensions

## Support

- 📖 [Documentation](./docs/)
- 🐛 [Issue Tracker](https://github.com/davidleitw/go-agent/issues)
- 💬 [Discussions](https://github.com/davidleitw/go-agent/discussions)
