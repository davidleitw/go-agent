# Calculator Tool Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates custom tool implementation and OpenAI function calling integration using a mathematical calculator tool.

## Overview

The calculator tool example showcases:
- **Custom Tool Implementation**: Creating tools that implement the `agent.Tool` interface
- **OpenAI Function Calling**: Integration with OpenAI's native function calling mechanism
- **Parameter Validation**: Robust input validation and type conversion
- **Structured Results**: Detailed calculation steps and results
- **Error Handling**: Graceful handling of mathematical errors (division by zero, etc.)

## Scenario: Mathematical Assistant

This example creates a math assistant that can perform various calculations using a custom calculator tool:
- **Basic Arithmetic**: Addition, subtraction, multiplication, division
- **Advanced Operations**: Power (exponentiation), square root
- **Detailed Steps**: Each calculation includes step-by-step breakdown
- **Error Prevention**: Validation for invalid operations

## Code Structure

### Key Components

1. **Tool Result Structure**
   ```go
   type CalculationResult struct {
       Expression    string    `json:"expression"`
       Result        float64   `json:"result"`
       Steps         []string  `json:"steps"`
       OperationType string    `json:"operation_type"`
       Timestamp     time.Time `json:"timestamp"`
   }
   ```
   - `Expression`: Human-readable mathematical expression
   - `Result`: Numerical result of the calculation
   - `Steps`: Array of calculation steps for transparency
   - `OperationType`: Type of mathematical operation performed
   - `Timestamp`: When the calculation was performed

2. **Custom Tool Implementation**
   ```go
   type CalculatorTool struct{}
   
   func (t *CalculatorTool) Name() string {
       return "calculator"
   }
   
   func (t *CalculatorTool) Description() string {
       return "Perform mathematical calculations including basic arithmetic, powers, and square roots"
   }
   
   func (t *CalculatorTool) Schema() map[string]interface{} {
       return map[string]interface{}{
           "type": "object",
           "properties": map[string]interface{}{
               "operation": map[string]interface{}{
                   "type": "string",
                   "enum": []string{"add", "subtract", "multiply", "divide", "power", "sqrt"},
               },
               "operand1": map[string]interface{}{
                   "type": "number",
                   "description": "The first number",
               },
               "operand2": map[string]interface{}{
                   "type": "number", 
                   "description": "The second number (not required for sqrt)",
               },
           },
           "required": []string{"operation", "operand1"},
       }
   }
   ```
   - Implements all four required methods of `agent.Tool` interface
   - Provides JSON schema for OpenAI function calling
   - Defines supported operations and parameters

3. **Tool Execution Logic**
   ```go
   func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
       // Input validation and type conversion
       operation := args["operation"].(string)
       operand1, _ := convertToFloat64(args["operand1"])
       
       // Operation-specific logic
       switch operation {
       case "add":
           result = operand1 + operand2
           expression = fmt.Sprintf("%.2f + %.2f", operand1, operand2)
           steps = []string{
               fmt.Sprintf("Addition: %.2f + %.2f", operand1, operand2),
               fmt.Sprintf("Result: %.2f", result),
           }
       // ... other operations
       }
       
       return CalculationResult{...}, nil
   }
   ```
   - Robust type conversion and validation
   - Operation-specific calculation logic
   - Detailed step-by-step breakdown
   - Error handling for edge cases

4. **Agent Configuration**
   ```go
   mathAssistant, err := agent.New(
       agent.WithName("math-assistant"),
       agent.WithInstructions(`You are a helpful math assistant...`),
       agent.WithOpenAI(apiKey),
       agent.WithModel("gpt-4"),
       agent.WithModelSettings(&agent.ModelSettings{
           Temperature: floatPtr(0.1), // Low temperature for precise calculations
           MaxTokens:   intPtr(1000),
       }),
       agent.WithTools(calculatorTool),
       agent.WithSessionStore(agent.NewInMemorySessionStore()),
       agent.WithDebugLogging(),
   )
   ```
   - Very low temperature (0.1) for mathematical precision
   - Specific instructions for mathematical assistance
   - Calculator tool registration

## Supported Operations

### 1. Addition (`add`)
```go
// Input: 15 + 27
{
  "operation": "add",
  "operand1": 15,
  "operand2": 27
}

// Output:
{
  "expression": "15.00 + 27.00",
  "result": 42,
  "steps": [
    "Addition: 15.00 + 27.00",
    "Result: 42.00"
  ],
  "operation_type": "add"
}
```

### 2. Subtraction (`subtract`)
```go
// Input: 125 - 47
{
  "operation": "subtract", 
  "operand1": 125,
  "operand2": 47
}
```

### 3. Multiplication (`multiply`)
```go
// Input: 13 × 7
{
  "operation": "multiply",
  "operand1": 13,
  "operand2": 7
}
```

### 4. Division (`divide`)
```go
// Input: 144 ÷ 12
{
  "operation": "divide",
  "operand1": 144,
  "operand2": 12
}
// Note: Includes division by zero protection
```

### 5. Power (`power`)
```go
// Input: 2^8
{
  "operation": "power",
  "operand1": 2,
  "operand2": 8
}
```

### 6. Square Root (`sqrt`)
```go
// Input: √64
{
  "operation": "sqrt",
  "operand1": 64
}
// Note: Only operand1 required, validates non-negative input
```

## Function Calling Flow

### 1. User Input Processing
```
User: "Calculate 15 + 27"
↓
Agent interprets natural language
↓ 
Agent decides to use calculator tool
↓
Function call generated
```

### 2. Tool Invocation
```go
// OpenAI generates this function call:
{
  "name": "calculator",
  "arguments": {
    "operation": "add",
    "operand1": 15,
    "operand2": 27
  }
}
```

### 3. Execution and Response
```go
// Tool executes and returns:
CalculationResult{
  Expression: "15.00 + 27.00",
  Result: 42.0,
  Steps: ["Addition: 15.00 + 27.00", "Result: 42.00"],
  OperationType: "add",
}

// Agent formats response:
"The result of 15 + 27 is 42. Here's how I calculated it:
1. Addition: 15.00 + 27.00  
2. Result: 42.00"
```

## Logging System

The example provides comprehensive logging for tool execution:

### Log Categories
- **TOOL**: Tool execution details and timing
- **TOOLCALL**: Function call parameters and results
- **REQUEST**: User input and calculation requests
- **RESPONSE**: Agent responses and tool integration
- **SESSION**: Conversation and tool usage tracking

### Sample Log Output
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
REQUEST[1]: Processing calculation request
TOOL: Calculator tool execution started
TOOL: Input arguments: map[operand1:15 operand2:27 operation:add]
TOOL: Operation: add, Operand1: 15.000000, Operand2: 27.000000 (has: true)
TOOL: Calculation completed successfully
TOOL: Expression: 15.00 + 27.00
TOOL: Result: 42.00
RESPONSE[1]: Duration: 2.3s
RESPONSE[1]: Tool calls: 1
🤖 Assistant: I'll calculate 15 + 27 for you using the calculator.

The result is **42**.

Here's the calculation breakdown:
- Addition: 15.00 + 27.00
- Result: 42.00

🔧 Tool Calls:
   • Tool: calculator
   • Arguments: {"operation":"add","operand1":15,"operand2":27}
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
   cd cmd/examples/calculator-tool
   go mod tidy
   ```

3. **Run the Example**:
   ```bash
   go run main.go
   ```

### Expected Output
The example will run through 6 predefined calculations:
1. Calculate 15 + 27
2. What is 144 divided by 12?
3. Find the square root of 64
4. Calculate 2 to the power of 8
5. What is 125 - 47?
6. Multiply 13 by 7

Each calculation will show:
- User request
- Agent response with explanation
- Tool call details
- Execution timing and session statistics

## Key Learning Points

### 1. Tool Interface Implementation
- **Method Requirements**: All four interface methods must be implemented
- **Schema Design**: JSON schema defines parameters for OpenAI
- **Type Safety**: Robust type conversion and validation

### 2. OpenAI Function Calling
- **Automatic Integration**: Framework handles OpenAI function calling protocol
- **Parameter Mapping**: Arguments automatically mapped from JSON to Go types
- **Response Formatting**: Tool results integrated into conversation flow

### 3. Error Handling Strategies
- **Input Validation**: Check types and required parameters
- **Mathematical Errors**: Handle division by zero, negative square roots
- **Graceful Degradation**: Continue operation when individual calculations fail

### 4. Structured Results
- **Detailed Output**: Include steps, expressions, and metadata
- **Audit Trail**: Timestamp and operation tracking
- **User Experience**: Clear explanations of calculations

## Troubleshooting

### Common Issues

1. **Tool Not Called**
   - **Cause**: Unclear instructions or schema issues
   - **Solution**: Refine tool description and ensure schema is valid

2. **Type Conversion Errors**
   ```
   TOOL: ERROR - operand1 conversion failed
   ```
   - **Cause**: OpenAI sending unexpected data types
   - **Solution**: Implement robust `convertToFloat64` function

3. **Mathematical Errors**
   ```
   division by zero is not allowed
   ```
   - **Cause**: Invalid mathematical operations
   - **Solution**: Add validation logic in tool execution

### Debug Tips

1. **Monitor Tool Calls**: Check if function calls are generated
2. **Validate Arguments**: Ensure OpenAI sends correct parameters
3. **Test Edge Cases**: Try invalid inputs and edge cases
4. **Schema Validation**: Verify JSON schema matches expectations

## Customization

### Adding New Operations

1. **Extend Schema**:
   ```go
   "enum": []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "sin", "cos", "log"}
   ```

2. **Implement Operation**:
   ```go
   case "sin":
       result = math.Sin(operand1)
       expression = fmt.Sprintf("sin(%.2f)", operand1)
       steps = []string{
           fmt.Sprintf("Sine of %.2f radians", operand1),
           fmt.Sprintf("Result: %.6f", result),
       }
   ```

### Creating Different Tools

```go
type WeatherTool struct{}

func (t *WeatherTool) Name() string { return "get_weather" }
func (t *WeatherTool) Description() string { return "Get current weather for a location" }
func (t *WeatherTool) Schema() map[string]interface{} { /* ... */ }
func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Weather API integration
}
```

## Next Steps

After understanding this example:
1. Create your own custom tools
2. Experiment with different parameter types
3. Integrate with external APIs
4. Add more sophisticated error handling
5. Explore multi-tool scenarios with complex workflows