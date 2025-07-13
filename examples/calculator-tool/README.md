# Calculator Tool Example

This example demonstrates how to integrate custom tools with an AI agent, showcasing the powerful tool calling capabilities of the go-agent framework.

## 🎯 Purpose

- Show how to implement the `agent.Tool` interface
- Demonstrate tool schema definition and validation
- Illustrate tool execution with error handling
- Showcase structured mathematical operations
- Provide comprehensive logging for tool interactions

## 🚀 Running the Example

```bash
# From the project root directory
go run cmd/examples/calculator-tool/main.go
```

## 📋 Prerequisites

- OpenAI API key set in environment variable `OPENAI_API_KEY`
- Go 1.21 or later

## 🏗️ Code Structure & Implementation

### 1. Tool Data Structure

```go
// CalculationResult represents the structured output for mathematical calculations
type CalculationResult struct {
    Expression    string    `json:"expression"`
    Result        float64   `json:"result"`
    Steps         []string  `json:"steps"`
    OperationType string    `json:"operation_type"`
    Timestamp     time.Time `json:"timestamp"`
}
```

**Purpose**: Define structured output format for calculation results
**Usage**: This structure provides detailed information about each calculation performed

### 2. Tool Implementation

```go
// CalculatorTool implements the agent.Tool interface for mathematical operations
type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
    return "calculator"
}

func (t *CalculatorTool) Description() string {
    return "Perform mathematical calculations including basic arithmetic, powers, and square roots"
}
```

**Purpose**: Implement basic tool identification methods
**API Usage**:
- `Name()` - Returns unique tool identifier used by the agent
- `Description()` - Provides human-readable tool description for the AI model

### 3. Tool Schema Definition

```go
func (t *CalculatorTool) Schema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "operation": map[string]any{
                "type":        "string",
                "description": "The mathematical operation to perform",
                "enum":        []string{"add", "subtract", "multiply", "divide", "power", "sqrt"},
            },
            "operand1": map[string]any{
                "type":        "number",
                "description": "The first number",
            },
            "operand2": map[string]any{
                "type":        "number",
                "description": "The second number (not required for sqrt)",
            },
        },
        "required": []string{"operation", "operand1"},
    }
}
```

**Purpose**: Define JSON schema for tool parameters
**API Usage**:
- `Schema()` - Returns JSON schema that describes expected parameters
- Schema includes parameter types, descriptions, and validation rules
- `enum` constrains operation to specific values
- `required` array specifies mandatory parameters

### 4. Tool Execution Logic

```go
func (t *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    log.Printf("TOOL: Calculator tool execution started")
    log.Printf("TOOL: Input arguments: %v", args)

    // Parameter validation
    operation, ok := args["operation"].(string)
    if !ok {
        return nil, fmt.Errorf("operation must be a string")
    }

    operand1, ok := args["operand1"].(float64)
    if !ok {
        if val, err := convertToFloat64(args["operand1"]); err == nil {
            operand1 = val
        } else {
            return nil, fmt.Errorf("operand1 must be a number")
        }
    }

    // Operation execution
    switch operation {
    case "add":
        if !hasOperand2 {
            return nil, fmt.Errorf("add operation requires two operands")
        }
        result = operand1 + operand2
        expression = fmt.Sprintf("%.2f + %.2f", operand1, operand2)
        steps = []string{fmt.Sprintf("Addition: %.2f + %.2f", operand1, operand2), fmt.Sprintf("Result: %.2f", result)}
    // ... more operations
    }

    return CalculationResult{
        Expression:    expression,
        Result:        result,
        Steps:         steps,
        OperationType: operation,
        Timestamp:     time.Now(),
    }, nil
}
```

**Purpose**: Core tool execution with comprehensive error handling
**API Usage**:
- `Execute(ctx, args)` - Main tool execution method
- `ctx` - Context for cancellation and timeouts
- `args` - Map of parameters passed from the AI model
- Returns structured result or error
- Includes parameter validation and type conversion
- Provides detailed logging for debugging

### 5. Agent Configuration with Tools

```go
// Create calculator tool
calcTool := &CalculatorTool{}

// Create math assistant agent with calculator tool
assistant, err := agent.NewBasicAgent(agent.BasicAgentConfig{
    Name:         "math-assistant",
    Description:  "A helpful AI assistant specialized in mathematical calculations",
    Instructions: `You are a math assistant that can perform calculations using the calculator tool.
                   Always use the calculator tool for mathematical operations.
                   Explain your calculations step by step.
                   Present results in a clear, formatted manner.`,
    Model:        "gpt-4o-mini",
    ModelSettings: &agent.ModelSettings{
        Temperature: floatPtr(0.1), // Low temperature for consistent math
        MaxTokens:   intPtr(1000),
    },
    Tools:     []agent.Tool{calcTool}, // Register the calculator tool
    ChatModel: chatModel,
})
```

**Purpose**: Configure agent with tool integration
**API Usage**:
- `Tools: []agent.Tool{calcTool}` - Register tools with the agent
- Low temperature (0.1) for consistent mathematical responses
- Specific instructions for tool usage
- `BasicAgent` automatically handles tool calling workflow

### 6. Conversation Loop with Tool Calls

```go
calculations := []struct {
    input       string
    description string
}{
    {
        input:       "Calculate 15 + 27",
        description: "basic addition",
    },
    {
        input:       "What is 144 divided by 12?",
        description: "division operation",
    },
    // ... more examples
}

for i, calc := range calculations {
    fmt.Printf("👤 User: %s\n", calc.input)
    
    response, structuredOutput, err := assistant.Chat(ctx, session, calc.input)
    if err != nil {
        log.Printf("❌ ERROR[%d]: %v", i+1, err)
        continue
    }
    
    fmt.Printf("🤖 Assistant: %s\n", response.Content)
}
```

**Purpose**: Execute calculations through natural language
**API Usage**:
- Agent automatically determines when to use tools
- Tool calls are transparent to the user
- Agent provides natural language explanations of results

### 7. Type Conversion Helper

```go
func convertToFloat64(val any) (float64, error) {
    switch v := val.(type) {
    case float64:
        return v, nil
    case int:
        return float64(v), nil
    case string:
        return strconv.ParseFloat(v, 64)
    default:
        return 0, fmt.Errorf("cannot convert %T to float64", val)
    }
}
```

**Purpose**: Handle different numeric types from JSON
**Usage**: Converts various numeric representations to float64 for calculations

## 🔧 Key APIs Demonstrated

### Tool Interface
- `agent.Tool` interface implementation
- `Name()` - Tool identifier
- `Description()` - Tool description
- `Schema()` - Parameter schema definition
- `Execute(ctx, args)` - Tool execution logic

### Agent Configuration
- `agent.BasicAgentConfig{}`
- `Tools: []agent.Tool{}` - Tool registration
- `ModelSettings{}` - Model configuration for tool usage

### Tool Execution Flow
1. User provides natural language input
2. Agent determines tool usage necessity
3. Agent calls tool with extracted parameters
4. Tool executes and returns structured result
5. Agent incorporates result into natural language response

## 📊 Example Output

```
🧮 Calculator Tool Example
==================================================
✅ OpenAI API key loaded
🛠️  Creating calculator tool...
📝 Creating math assistant agent...
✅ Math assistant 'math-assistant' created with calculator tool

🧮 Starting calculator demonstrations...
==================================================

🔄 Calculation 1/6
👤 User: Calculate 15 + 27
TOOL: Calculator tool execution started
TOOL: Input arguments: map[operand1:15 operand2:27 operation:add]
TOOL: Operation: add, Operand1: 15.000000, Operand2: 27.000000 (has: true)
TOOL: Calculation completed successfully
TOOL: Expression: 15.00 + 27.00
TOOL: Result: 42.00
🤖 Assistant: To calculate 15 + 27:

1. **Addition**: We add the two numbers together:
   15 + 27 = 42

The result of 15 + 27 is **42**.

🔄 Calculation 2/6
👤 User: What is 144 divided by 12?
TOOL: Calculator tool execution started
TOOL: Input arguments: map[operand1:144 operand2:12 operation:divide]
TOOL: Operation: divide, Operand1: 144.000000, Operand2: 12.000000 (has: true)
TOOL: Calculation completed successfully
TOOL: Expression: 144.00 ÷ 12.00
TOOL: Result: 12.00
🤖 Assistant: To calculate 144 ÷ 12:

1. **Division**: We divide 144 by 12:
   144 ÷ 12 = 12

The result of 144 ÷ 12 is **12**.
```

## 🎓 Learning Objectives

After studying this example, you should understand:

1. **Tool Interface Implementation**: How to create custom tools
2. **Schema Definition**: How to define tool parameters and validation
3. **Tool Registration**: How to register tools with agents
4. **Error Handling**: Robust error handling in tool execution
5. **Type Conversion**: Handling different data types from JSON
6. **Logging Strategy**: Comprehensive logging for tool debugging

## 🔄 Next Steps

- Try the [Task Completion Example](../task-completion/) to learn about structured output
- Explore the [Multi-Tool Agent Example](../multi-tool-agent/) for multiple tool coordination
- Study the [Condition Testing Example](../condition-testing/) for advanced flow control

## 🐛 Common Issues

1. **Schema Validation**: Ensure schema matches actual parameter usage
2. **Type Conversion**: Handle different numeric types properly
3. **Error Messages**: Provide clear error messages for debugging
4. **Tool Timeout**: Consider timeout handling for long-running operations

## 💡 Customization Ideas

- Add more mathematical operations (trigonometry, logarithms)
- Implement calculation history and memory
- Add input validation for edge cases (division by zero)
- Create composite operations (solve equations)
- Add unit conversion capabilities

## 🧮 Supported Operations

| Operation | Description | Parameters | Example |
|-----------|-------------|------------|---------|
| `add` | Addition | operand1, operand2 | 15 + 27 = 42 |
| `subtract` | Subtraction | operand1, operand2 | 100 - 25 = 75 |
| `multiply` | Multiplication | operand1, operand2 | 8 × 7 = 56 |
| `divide` | Division | operand1, operand2 | 144 ÷ 12 = 12 |
| `power` | Exponentiation | operand1, operand2 | 2^8 = 256 |
| `sqrt` | Square Root | operand1 | √64 = 8 |

## 🔍 Advanced Features

- **Structured Output**: Each calculation returns detailed result information
- **Step-by-Step Explanation**: Tools provide calculation steps
- **Error Handling**: Comprehensive validation and error reporting
- **Type Flexibility**: Automatic conversion between numeric types
- **Logging**: Detailed execution logging for debugging