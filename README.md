# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent" width="200" height="200">
</div>

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

A lightweight Go AI agent framework focused on building intelligent conversations and automated workflows.

## Why choose go-agent

go-agent provides intuitive interfaces for building AI applications. The framework focuses on minimal configuration: provide an API key, create an agent, and start conversing.

The design prioritizes simplicity for common use cases while maintaining flexibility for complex scenarios. Creating a basic chatbot requires minimal code.

## Quick Start

First, install go-agent:

```bash
go get github.com/davidleitw/go-agent
```

Create your first AI agent:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/davidleitw/go-agent/pkg/agent"
)

func main() {
    // Create an AI agent in one line
    assistant, err := agent.New("helpful-assistant").
        WithOpenAI(os.Getenv("OPENAI_API_KEY")).
        WithModel("gpt-4o-mini").
        WithInstructions("You are a helpful assistant. Be concise and friendly.").
        Build()
    if err != nil {
        panic(err)
    }

    // Start chatting
    response, err := assistant.Chat(context.Background(), "Hello! How are you today?")
    if err != nil {
        panic(err)
    }

    fmt.Println("Assistant:", response.Message)
}
```

The framework automatically handles OpenAI client creation, session management, and configuration setup.

## Core Features

### Tool Integration

**When to use**: When agents need to perform external operations like API calls, calculations, or data processing.

Tools enable agents to interact with external systems. Define tools using simple function syntax:

```go
// Create a weather query tool using function definition
weatherTool := agent.NewTool("get_weather", 
    "Get weather information for a specified location",
    func(location string) map[string]any {
        // Simulate weather API call
        return map[string]any{
            "location":    location,
            "temperature": "22°C",
            "condition":   "Sunny",
        }
    })

// Create an agent with tool capabilities
weatherAgent, err := agent.New("weather-assistant").
    WithOpenAI(apiKey).
    WithInstructions("You can help users get weather information.").
    WithTools(weatherTool).
    Build()
```

The framework automatically generates JSON Schema, handles parameter validation, and manages tool execution flow.

**Complete example**: [Calculator Tool Example](./examples/calculator-tool/)

### Structured Output

**When to use**: When you need agents to return data in specific formats for downstream processing.

Define structured output using Go structs:

```go
// Define your desired output format
type TaskResult struct {
    Title    string   `json:"title"`
    Priority string   `json:"priority"`
    Tags     []string `json:"tags"`
}

// Create an agent that returns structured data
taskAgent, err := agent.New("task-creator").
    WithOpenAI(apiKey).
    WithInstructions("Create tasks based on user input, return structured JSON data.").
    WithOutputType(&TaskResult{}).
    Build()

// Conversations automatically return parsed structures
response, err := taskAgent.Chat(ctx, "Create a high priority code review task")
if taskResult, ok := response.Data.(*TaskResult); ok {
    fmt.Printf("Created task: %s (Priority: %s)\n", taskResult.Title, taskResult.Priority)
}
```

The framework automatically generates JSON Schema, validates AI output, and parses responses into Go structs.

**Complete example**: [Task Completion Example](./examples/task-completion/)

### Schema-Based Information Collection

**When to use**: When you need to collect structured data from users across conversation turns, such as form filling, user onboarding, or support ticket creation.

The schema system automatically extracts information from user messages and manages collection state. This eliminates manual state management and provides natural conversation flow.

#### Basic Schema Definition

```go
import "github.com/davidleitw/go-agent/pkg/schema"

// Required fields (default)
emailField := schema.Define("email", "Please provide your email address")
issueField := schema.Define("issue", "Please describe your issue")

// Optional fields
phoneField := schema.Define("phone", "Contact number for urgent matters").Optional()
```

#### Applying Schema to Conversations

```go
supportBot, err := agent.New("support-agent").
    WithOpenAI(apiKey).
    WithInstructions("You are a customer support assistant.").
    Build()

response, err := supportBot.Chat(ctx, "I need help with my account",
    agent.WithSchema(
        schema.Define("email", "Please provide your email address"),
        schema.Define("issue", "Please describe your issue in detail"),
        schema.Define("urgency", "How urgent is this?").Optional(),
    ),
)
```

The framework intelligently:
- **Extracts** information from user messages using LLM semantic understanding
- **Identifies** missing required fields automatically  
- **Asks** for missing information using natural, contextual prompts
- **Remembers** collected information across conversation turns
- **Adapts** to different conversation styles and user input patterns

#### Dynamic Schema Selection

**When to use**: When different conversation types require different information (e.g., support requests vs. sales inquiries).

```go
func getSchemaForIntent(intent string) []*schema.Field {
    switch intent {
    case "technical_support":
        return []*schema.Field{
            schema.Define("email", "Email for technical follow-up"),
            schema.Define("error_message", "What error are you seeing?"),
            schema.Define("steps_taken", "What have you tried?"),
        }
    case "billing_inquiry":
        return []*schema.Field{
            schema.Define("email", "Account email address"),
            schema.Define("account_id", "Your account number"),
            schema.Define("billing_question", "Billing question details"),
        }
    }
}

// Apply schema based on detected intent
intent := detectIntent(userInput)
schema := getSchemaForIntent(intent)
response, err := agent.Chat(ctx, userInput, agent.WithSchema(schema...))
```

#### Multi-Step Workflows

**When to use**: For complex forms or processes that should be broken into logical steps.

```go
func getTechnicalSupportWorkflow() [][]*schema.Field {
    return [][]*schema.Field{
        { // Step 1: Contact info
            schema.Define("email", "Your email address"),
            schema.Define("issue_summary", "Brief issue description"),
        },
        { // Step 2: Technical details
            schema.Define("error_message", "Exact error message"),
            schema.Define("browser", "Browser and version"),
        },
        { // Step 3: Impact assessment
            schema.Define("urgency", "How critical is this?"),
            schema.Define("affected_users", "How many users affected?"),
        },
    }
}
```

**Complete examples**: 
- [Simple Schema Example](./examples/simple-schema/) - Basic usage
- [Customer Support Example](./examples/customer-support/) - Real-world scenarios  
- [Dynamic Schema Example](./examples/dynamic-schema/) - Advanced workflows

### Conditional Flow Control

**When to use**: When you need agents to respond differently based on conversation context, user state, or external conditions.

Flow control enables dynamic agent behavior through conditions and rules. This is essential for creating intelligent, context-aware conversations.

#### Built-in Conditions

Common conditions for typical conversation scenarios:

```go
// Text-based conditions
agent.Contains("help")        // User message contains "help"
agent.StartsWith("hello")     // User message starts with "hello"  
agent.Exactly("yes")          // User message is exactly "yes"
agent.MatchesRegex("\\d+")    // User message matches regex pattern

// Conversation state conditions
agent.MessageCount(5)         // Conversation has 5+ messages
agent.TurnsSince(3)          // 3+ turns since last condition match
agent.FirstMessage()         // This is the first message in session

// Data conditions
agent.MissingFields("email", "name")  // Required fields are missing
agent.HasField("phone")               // Specific field is present
agent.FieldEquals("status", "urgent") // Field has specific value
```

#### Custom Conditions

Implement the `Condition` interface for complex logic:

```go
type BusinessHoursCondition struct{}

func (c *BusinessHoursCondition) Name() string {
    return "business_hours"
}

func (c *BusinessHoursCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]interface{}) (bool, error) {
    now := time.Now()
    hour := now.Hour()
    return hour >= 9 && hour <= 17, nil
}

// Use custom condition
businessRule := agent.FlowRule{
    Name:      "office_hours_response",
    Condition: &BusinessHoursCondition{},
    Action: agent.FlowAction{
        NewInstructionsTemplate: "You can provide full support during business hours.",
    },
}
```

#### Combining Conditions

```go
// Logical operators
agent.And(agent.Contains("urgent"), agent.MissingFields("phone"))
agent.Or(agent.Contains("help"), agent.Contains("support"))
agent.Not(agent.HasField("email"))

// Complex condition combinations
complexCondition := agent.And(
    agent.Or(agent.Contains("billing"), agent.Contains("payment")),
    agent.MissingFields("account_id"),
    agent.MessageCount(2),
)
```

**Complete examples**:
- [Condition Testing Example](./examples/condition-testing/) - Basic flow control
- [Advanced Conditions Example](./examples/advanced-conditions/) - Complex scenarios

## Core Design Philosophy

The framework design follows these principles:

**Simplicity for common cases**: Basic functionality requires minimal configuration. Essential operations like creating agents and managing conversations use straightforward APIs.

**Flexibility for complex scenarios**: Advanced features including multi-tool coordination, conditional flows, and structured output are available through composable interfaces.

**Automatic infrastructure management**: Session management, tool execution, and error handling operate without manual intervention.

### Architecture Components

The framework consists of several main parts:

**Agent**: Core interface for conversation handling. Create using `agent.New()` or implement custom logic through the `Agent` interface.

**Session**: Manages conversation history and state. Automatic persistence and retrieval across conversation turns.

**Tools**: Enable external operations through the `Tool` interface. Convert functions to tools using `agent.NewTool()`.

**Conditions**: Flow control through the `Condition` interface. Built-in conditions available for common scenarios.

**Schema**: Information collection through the `schema` package. Automatic extraction and validation of structured data.

**Chat Models**: LLM provider abstraction. Supports OpenAI with additional providers in development.

## Supported LLM Providers

Currently mainly supports OpenAI models, including GPT-4, GPT-4o, GPT-3.5-turbo, etc. We're actively developing support for other providers:

**Supported**: OpenAI (full support, including function calling and structured output)

**In Development**: Anthropic Claude, Google Gemini, local models (via Ollama)

## Session Storage

The framework comes with in-memory session storage, suitable for development and testing. For production environments, we're developing Redis and PostgreSQL backend support.

Honestly though, for most applications, in-memory storage is sufficient. You can always implement your own storage backend.

## Examples

We've prepared complete examples in the [`examples/`](./examples/) directory, each is a directly executable Go program.

### Quick Setup

First, set up your OpenAI API key:

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env and add your OpenAI API key
```

### Main Examples

**Basic Chat (basic-chat)**: The simplest starting point, showing how to create a chatbot with just a few lines of code.

**Calculator Tool (calculator-tool)**: Shows how to let agents use tools, this example creates an assistant that can do math.

**Advanced Conditions (advanced-conditions)**: Shows intelligent flow control where agents automatically adjust behavior based on conversation state. This is our most recommended example, showcasing the framework's powerful features.

**Multi-Tool Agent (multi-tool-agent)**: Shows how to let one agent use multiple tools simultaneously, intelligently selecting appropriate tools to complete tasks.

**Task Completion (task-completion)**: Shows structured output and condition validation, simulating a restaurant reservation system.

**Simple Schema (simple-schema)**: Demonstrates basic schema-based information collection, showing how to define required and optional fields for automatic data gathering.

**Customer Support (customer-support)**: Real-world example showing how to build a professional customer support bot with intelligent information collection across different support scenarios.

**Dynamic Schema (dynamic-schema)**: Advanced example demonstrating dynamic schema selection based on user intent, multi-step workflows, and complex conversation management.

Each example has detailed README instructions on how to run and key learning points. We recommend starting with basic-chat, then trying simple-schema to understand information collection, followed by advanced-conditions for flow control.

## Common Issues

If you encounter problems, check these first:

**API Key Configuration Error**: Make sure your `.env` file has the correct `OPENAI_API_KEY`

**Import Errors**: Make sure you're running in the correct directory and using `github.com/davidleitw/go-agent/pkg/agent`

**Module Issues**: Run `go mod tidy` in the example directory

All examples have detailed log output to help you track execution flow and errors.

## Development

If you want to participate in development or customize the framework:

```bash
# Run tests
make test

# Code linting
make lint

# Build project
make build
```

Requires Go 1.22 or newer.

## Future Plans

We're currently developing these features:

More LLM provider support (Anthropic, Google, etc.), production-grade storage backends (Redis, PostgreSQL), streaming responses, multi-agent coordination, monitoring and observability features.

If you have specific needs or ideas, feel free to discuss them in [GitHub Issues](https://github.com/davidleitw/go-agent/issues).

## Summary

go-agent's goal is to enable Go developers to quickly build AI applications without needing to deeply understand the details of various LLM APIs. We believe good frameworks should make common tasks simple and complex tasks possible.

If you're considering adding AI features to your Go project, give go-agent a try. Start with a simple chatbot, and when you need more features, the framework will grow with your needs.