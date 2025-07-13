# Go-Agent Examples

This directory contains comprehensive examples demonstrating the capabilities of the go-agent framework. Each example showcases different aspects of building AI agents with various levels of complexity.

## 🚀 Quick Start

### Prerequisites

1. **Go 1.21 or later** installed on your system
2. **OpenAI API Key** - Get one from [OpenAI Platform](https://platform.openai.com/)

### Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd go-agent
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up environment variables**:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key-here"
   ```

### Running Examples

Each example can be run directly from the project root:

```bash
# Run basic chat example
go run examples/basic-chat/main.go

# Run calculator tool example  
go run examples/calculator-tool/main.go

# Run task completion example
go run examples/task-completion/main.go

# Run multi-tool agent example
go run examples/multi-tool-agent/main.go

# Run condition testing example
go run examples/condition-testing/main.go

# Run simple schema example
go run examples/simple-schema/main.go

# Run customer support example
go run examples/customer-support/main.go

# Run dynamic schema example
go run examples/dynamic-schema/main.go
```

## 📚 Examples Overview

### 1. Basic Chat (`basic-chat/`)
**Purpose**: Demonstrates the simplest possible AI agent implementation

**What it shows**:
- Basic conversation flow
- Using `BasicAgent` for simple scenarios
- Session management
- Message handling

**Key APIs**:
- `agent.NewBasicAgent()`
- `agent.NewSession()`
- `agent.Chat()`

**Use Case**: Perfect starting point for understanding the framework fundamentals.

**Detailed Documentation**: [README](basic-chat/README.md) | [繁體中文](basic-chat/README-zh.md)

---

### 2. Calculator Tool (`calculator-tool/`)
**Purpose**: Shows how to integrate external tools with an AI agent

**What it shows**:
- Tool implementation and registration
- Tool execution and error handling
- Mathematical operations integration
- Structured tool responses

**Key APIs**:
- `agent.Tool` interface
- Tool schema definition
- Tool execution context
- Error handling in tools

**Use Case**: Learn how to extend agent capabilities with custom tools.

**Detailed Documentation**: [README](calculator-tool/README.md) | [繁體中文](calculator-tool/README-zh.md)

---

### 3. Task Completion (`task-completion/`)
**Purpose**: Demonstrates structured output and data collection workflows

**What it shows**:
- Structured output types
- Data validation
- Task-oriented conversations
- Progress tracking

**Key APIs**:
- `agent.OutputType` interface
- `agent.NewStructuredOutputType()`
- JSON schema generation
- Output validation

**Use Case**: Building agents that collect and structure user data systematically.

**Detailed Documentation**: [README](task-completion/README.md) | [繁體中文](task-completion/README-zh.md)

---

### 4. Multi-Tool Agent (`multi-tool-agent/`)
**Purpose**: Advanced example showing custom agent implementation with multiple tools

**What it shows**:
- Custom agent implementation
- Multiple tool coordination
- Tool usage statistics
- Dynamic instruction enhancement
- Advanced state management

**Key APIs**:
- `agent.Agent` interface implementation
- Custom chat logic
- Tool orchestration
- State tracking

**Use Case**: Building sophisticated agents that coordinate multiple capabilities intelligently.

**Detailed Documentation**: [README](multi-tool-agent/README.md) | [繁體中文](multi-tool-agent/README-zh.md)

---

### 5. Condition Testing (`condition-testing/`)
**Purpose**: Advanced flow control and conditional logic in conversations

**What it shows**:
- Flow rules and conditions
- Dynamic conversation flow
- User onboarding processes
- Conditional tool execution
- Advanced state management

**Key APIs**:
- `agent.FlowRule` interface
- `agent.Condition` interface
- Dynamic flow control
- Conditional actions

**Use Case**: Creating agents with complex, adaptive conversation flows.

**Detailed Documentation**: [README](condition-testing/README.md) | [繁體中文](condition-testing/README-zh.md)

---

### 6. Simple Schema (`simple-schema/`)
**Purpose**: Basic schema-based information collection

**What it shows**:
- Field definition with `schema.Define()`
- Required vs optional fields
- Automatic information extraction
- Natural conversation flow
- Session-based information persistence

**Key APIs**:
- `schema.Define()` - Field definition
- `schema.Field.Optional()` - Optional field marking
- `agent.WithSchema()` - Schema application
- Schema collection metadata

**Use Case**: Learn the fundamentals of intelligent information collection.

**Detailed Documentation**: [README](simple-schema/README.md) | [繁體中文](simple-schema/README-zh.md)

---

### 7. Customer Support (`customer-support/`)
**Purpose**: Real-world customer support bot with intelligent information collection

**What it shows**:
- Professional support workflows
- Specialized schemas for different support types
- Multi-turn conversation handling
- Contextual information extraction
- Support ticket information gathering

**Key APIs**:
- Dynamic schema selection
- Support-specific field definitions
- Multi-schema workflows
- Professional prompt design

**Use Case**: Building production-ready customer support systems.

**Detailed Documentation**: [README](customer-support/README.md) | [繁體中文](customer-support/README-zh.md)

---

### 8. Dynamic Schema (`dynamic-schema/`)
**Purpose**: Advanced schema selection and multi-step workflows

**What it shows**:
- Intent classification systems
- Dynamic schema selection based on context
- Multi-step information collection workflows
- Complex conversation management
- Real-time schema adaptation

**Key APIs**:
- Intent-based schema selection
- Multi-step workflow orchestration
- Advanced conversation analytics
- Complex business logic integration

**Use Case**: Building sophisticated conversation systems with adaptive data collection.

**Detailed Documentation**: [README](dynamic-schema/README.md) | [繁體中文](dynamic-schema/README-zh.md)

## 🏗️ Architecture Patterns

### BasicAgent vs Custom Agent

**Use BasicAgent when**:
- Simple, straightforward conversations
- Standard tool usage patterns
- Minimal state management needs
- Quick prototyping

**Use Custom Agent when**:
- Complex state management required
- Advanced tool coordination needed
- Custom conversation flow logic
- Sophisticated error handling

### Tool Integration Patterns

**Simple Tool**: Single-purpose, stateless operations
```go
type CalculatorTool struct{}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // Implementation
}
```

**Complex Tool**: Multi-operation, stateful tools
```go
type WeatherTool struct {
    apiKey string
    cache  map[string]WeatherData
}
```

### Flow Control Patterns

**Condition-Based**: React to conversation state
```go
type MissingFieldsCondition struct {
    requiredFields []string
}

func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
    // Check if fields are missing
}
```

**Rule-Based**: Apply actions when conditions are met
```go
type FlowRule struct {
    Name      string
    Condition agent.Condition
    Action    agent.Action
}
```

## 🔧 Development Guidelines

### Adding New Examples

1. Create a new directory under `cmd/examples/`
2. Implement `main.go` with comprehensive logging
3. Add detailed README with code explanations
4. Include `.env.example` if needed
5. Test thoroughly with various scenarios

### Code Style

- Use descriptive variable names
- Add comprehensive logging for debugging
- Handle errors gracefully
- Include performance metrics where relevant
- Document complex logic with comments

### Testing

- Test with various input scenarios
- Verify error handling
- Check resource cleanup
- Validate output formats
- Test edge cases

## 🐛 Troubleshooting

### Common Issues

1. **OpenAI API Key Issues**:
   - Ensure the key is set correctly
   - Check for rate limits
   - Verify key permissions

2. **Tool Execution Errors**:
   - Check tool argument validation
   - Verify tool schema matches usage
   - Review timeout settings

3. **Flow Rule Problems**:
   - Debug condition evaluation
   - Check action implementations
   - Verify rule ordering

### Debug Tips

- Enable detailed logging
- Use small test cases
- Check API responses
- Validate input/output formats
- Monitor resource usage

## 📖 Further Reading

- [Go-Agent Documentation](../../README.md)
- [API Reference](../../docs/api.md)
- [Architecture Guide](../../docs/architecture.md)
- [Best Practices](../../docs/best-practices.md)

## 🤝 Contributing

We welcome contributions! Please:

1. Follow the existing code style
2. Add comprehensive tests
3. Update documentation
4. Include example usage
5. Test with multiple scenarios

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details. 