# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent Mascot" width="300"/>
  
  [![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)
</div>

A clean yet feature-complete AI agent framework for Go. We designed this framework to be easy to get started with while maintaining high extensibility, allowing you to quickly integrate AI agent capabilities into your Go projects.

## Why go-agent?

While there are several excellent agent frameworks available, we wanted to create something with a focus on simplicity and Go-idiomatic design. Our design philosophy is "Context is Everything" + **Easy to Start, Easy to Scale**:

**Easy to Start:**
- Get going with just one `Execute()` method
- Clear module responsibilities - no need to understand the entire framework
- Rich examples and documentation that you can follow immediately

**Highly Extensible:**
- Modular design - use only what you need
- Clear interface definitions make custom implementations easy
- Open Provider pattern allows integration with any data source

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/davidleitw/go-agent/agent"
    "github.com/davidleitw/go-agent/llm/openai"
)

func main() {
    // Create LLM model
    model := openai.New(llm.Config{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    
    // Create simple agent
    myAgent := agent.NewSimpleAgent(model)
    
    // Start conversation
    response, err := myAgent.Execute(context.Background(), agent.Request{
        Input: "Plan a 3-day trip to Tokyo for me",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Output)
    fmt.Printf("Used %d tokens\n", response.Usage.LLMTokens.TotalTokens)
}
```

## Framework Architecture

We break down complex AI agent functionality into several independent but well-coordinated modules:

### [Agent Module](./agent/) - Core Controller
This is the brain of the framework, coordinating all other modules. Provides a simple `Execute()` interface and flexible Builder pattern for easy configuration.

**Key Features:**
- Clean Agent interface - one method does everything
- Builder pattern makes configuration intuitive
- Automatic session management - no state worries
- Built-in convenience functions for common patterns

### [Session Module](./session/) - Memory Management
Handles conversation state and history. Supports TTL auto-expiration, concurrent safety, and complete JSON serialization.

**Key Features:**
- Key-Value state storage for any data type
- Unified history format supporting multiple conversation types
- Automatic TTL management with expired session cleanup
- Thread-safe for multi-goroutine usage

### [Context Module](./context/) - Information Aggregation
This module's job is to package information from various sources (conversation history, system prompts, external data, etc.) into a unified format that LLMs can understand.

**Key Features:**
- Unified Context data structure
- Extensible Provider system
- Automatic Session history to Context conversion
- Rich Metadata support

### [Tool Module](./tool/) - Tool Integration
Enables your AI agents to call external functions like database queries, API calls, calculations, etc.

**Key Features:**
- Simple Tool interface - easy to implement custom tools
- JSON Schema-based parameter definitions
- Thread-safe tool registry
- Complete error handling mechanisms

### [LLM Module](./llm/) - Language Model Interface
Provides unified language model interface. Currently supports OpenAI, with plans to expand to other providers.

**Key Features:**
- Clear Model interface
- Built-in tool calling support
- Complete token usage tracking
- Support for custom endpoints and proxies

## Current Development Status

**Ready to Use:**
- Complete module interface design and implementation
- Session management with TTL support
- Context provider system
- Tool registration and execution framework
- OpenAI integration
- Rich test coverage

**In Development:**
- Agent core execution logic (LLM calls, tool orchestration, iterative thinking, etc.)
- More LLM provider support
- Streaming response support
- More built-in tools and examples

**Future Plans:**
- Redis/Database Session storage
- Asynchronous tool execution
- Advanced Context management features
- MCP (Model Context Protocol) tool integration

## Design Philosophy

### "Context is Everything"
We believe the core of AI agents is context management. Whether it's conversation history, user preferences, external data, or tool execution results, everything needs to be provided to LLMs in a consistent way.

We're planning to organize talks and compile resources about Context Engineering to help the community better understand this approach.

## Contributing

This project is under active development, and we welcome all forms of participation:

**Interface Design Discussion (Most Important!):**
- Think some interface design isn't intuitive enough?
- Have better API design ideas?
- Feel some functionality abstraction levels are wrong?
- Want a module to provide different usage patterns?

We deeply believe good interface design is key to framework success - any friends with interface ideas are very welcome to discuss!

**Feature Suggestions:**
- What new features would you like to see?
- What usage difficulties have you encountered?
- What real-world scenarios haven't we considered?

**Code Contributions:**
- Implement new LLM providers
- Build more practical tools
- Improve performance and stability
- Add more tests and examples

**Documentation and Examples:**
- Write usage tutorials
- Create real-world application examples
- Translate documentation

Feel free to open Issues for discussion or submit PRs directly. We're happy to work together to make this framework better.

## Getting Started

1. **Check Module Documentation**: Each folder has detailed READMEs - suggest starting with [Agent Module](./agent/)
2. **Run Tests**: `go test ./...` to see if everything works
3. **Join Discussion**: Open Issues with questions or ideas

## License

MIT License - Use it however you want, but we're not responsible for any losses.

---

**Project Status: Under Active Development** | **Last Updated: 2024**

Looking forward to seeing what interesting things you build with this framework!