# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent" width="200" height="200">
</div>

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

A lightweight Go AI agent framework focused on building intelligent conversations and automated workflows.

## Why choose go-agent

Honestly, most AI frameworks out there are overly complex. What we really want is simple: give it an API key, create an agent, and start chatting. That's it.

go-agent's design philosophy is simple: make common things super easy, and make complex things possible. You shouldn't need to write 60 lines of code to create a basic chatbot. You only need 5.

## Quick Start

First, install go-agent:

```bash
go get github.com/davidleitw/go-agent
```

Then write your first AI agent. Really, it's this simple:

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

See that? No need to manually create OpenAI clients, manage sessions, or deal with complex configuration structs. The framework handles all of this automatically.

## Adding Tool Capabilities

When your agent needs to perform actual operations, like checking weather or doing calculations, you need tools. Previously, defining a tool required writing a bunch of interface implementations. Now you just write a function:

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

The framework automatically generates JSON Schema from your functions, handles parameter validation, and manages the tool calling flow. You don't need to manually handle OpenAI's function calling format.

## Structured Output

Sometimes you want AI to respond with specific data formats, like JSON. The traditional approach is to pray in your prompt that AI returns the correct format, then manually parse it. Now you just define a struct and the framework handles everything else:

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

The framework automatically generates JSON Schema, validates AI output, and parses it into your Go struct. No more manual JSON parsing errors.

## Intelligent Flow Control

This is one of go-agent's most powerful features. You can make agents automatically adjust their behavior based on conversation state. For example, when user information is incomplete, automatically guide them to complete it:

```go
// Create an agent that automatically collects missing information
onboardingAgent, err := agent.New("onboarding-specialist").
    WithOpenAI(apiKey).
    WithInstructions("You are an onboarding expert who needs to collect basic user information.").
    
    // When name is missing, automatically ask
    OnMissingInfo("name").Ask("What's your name?").Build().
    
    // When email is missing, automatically ask
    OnMissingInfo("email").Ask("Please provide your email address.").Build().
    
    // When conversation gets too long, automatically summarize
    OnMessageCount(6).Summarize().Build().
    
    // When user says "help", provide assistance
    When(agent.WhenContains("help")).Ask("How can I help you?").Build().
    
    // Complex condition combinations: when multiple information is missing
    When(agent.And(
        agent.WhenMissingFields("email"),
        agent.WhenMissingFields("phone"),
    )).Ask("I need both your email and phone number to proceed.").Build().
    
    Build()
```

These conditions are automatically checked during each conversation, making your agent smarter and more human-like. No need to write a bunch of if-else statements in your code.

## Core Design Philosophy

We had several core principles when designing go-agent:

**Make simple things super simple**: Creating a basic chatbot shouldn't require reading documentation. The API should be intuitive enough that you know how to use it at first glance.

**Make complex things possible**: When you need advanced features like multi-tool coordination, conditional flows, structured output, the framework should provide powerful abstractions instead of making you reinvent the wheel.

**Automated default behavior**: Infrastructure like session management, tool calling loops, error handling should work correctly by default without manual management.

### Architecture Components

The framework consists of several main parts:

**Agent**: The brain of your AI assistant, responsible for handling conversation logic. We provide `agent.New()` for quick creation while preserving full interfaces for customization.

**Session**: Automatically manages conversation history. You don't need to manually track messages; the framework handles it.

**Tools**: Capabilities that allow agents to perform actual operations. Use `agent.NewTool()` to quickly turn any function into a tool.

**Conditions**: The core of intelligent flow control. Define complex conversation logic with natural language style APIs.

**Chat Models**: Abstraction for different LLM providers. Currently supports OpenAI, with more coming soon.

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

Each example has detailed README instructions on how to run and key learning points. We recommend starting with basic-chat, then trying advanced-conditions.

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