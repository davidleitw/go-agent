# Multi-Tool Agent Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![Chinese](https://img.shields.io/badge/README-Chinese-red.svg)](README-zh.md)

This example demonstrates an advanced AI assistant that can intelligently select and use multiple tools based on user context and requests.

## Overview

The multi-tool agent example showcases:
- **Context-Aware Tool Selection**: Agent automatically chooses appropriate tools based on user input
- **Multiple Tool Integration**: Weather, calculator, time, and notification tools working together
- **Sequential Tool Usage**: Agent can use multiple tools in sequence for complex requests
- **Real-world Scenarios**: Practical examples of multi-tool interactions

## Available Tools

### 1. 🌤️ Weather Tool (`get_weather`)
Retrieves weather information for any location.

**Capabilities**:
- Current weather conditions
- Temperature (Celsius/Fahrenheit)
- Humidity and wind speed
- Multiple location support

**Parameters**:
- `location` (required): City and country
- `unit` (optional): "celsius" or "fahrenheit"

### 2. 🧮 Calculator Tool (`calculate`)
Performs mathematical calculations.

**Operations**:
- Basic arithmetic: add, subtract, multiply, divide
- Advanced: power, square root
- Error handling: division by zero, negative square roots

**Parameters**:
- `operation` (required): Type of calculation
- `a` (required): First number
- `b` (optional): Second number (not needed for sqrt)

### 3. ⏰ Time Tool (`get_time`)
Provides current time information across timezones.

**Features**:
- Multiple timezone support
- Various time formats (ISO, human-readable, timestamp)
- Additional info: day of week, day of year

**Parameters**:
- `timezone` (optional): Timezone identifier (default: "UTC")
- `format` (optional): "iso", "human", or "timestamp"

### 4. 📢 Notification Tool (`send_notification`)
Simulates sending notifications and reminders.

**Features**:
- Immediate or scheduled notifications
- Priority levels
- Custom titles and messages

**Parameters**:
- `message` (required): Notification content
- `title` (optional): Notification title
- `priority` (optional): "low", "normal", "high", "urgent"
- `delay_minutes` (optional): Delay before sending

## Test Scenarios

The example runs through various scenarios to test tool combinations:

### Single Tool Usage
1. **Weather Query**: "What's the weather like in Tokyo, Japan?"
2. **Math Calculation**: "Calculate the square root of 144"
3. **Time Query**: "What time is it in New York?"
4. **Notification**: "Send me a reminder about my dentist appointment"

### Multi-Tool Combinations
5. **Weather + Time**: "What's the weather in London and what time is it there?"
6. **Calculator + Weather**: "Calculate 25 * 4 and tell me the weather in San Francisco"
7. **Time + Notification**: "Get time in Tokyo and remind me about meeting in 30 minutes"

## Expected Agent Behavior

The agent should:
1. **Parse User Intent**: Understand what tools are needed
2. **Select Appropriate Tools**: Choose the right tool(s) for each request
3. **Execute in Sequence**: Use multiple tools when necessary
4. **Provide Comprehensive Responses**: Combine results meaningfully

## Running the Example

### Prerequisites
1. Go 1.22 or later
2. OpenAI API key

### Setup
1. **Configure Environment**:
   ```bash
   cp .env.example .env
   # Edit .env and add your OpenAI API key
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the Example**:
   ```bash
   go run main.go
   ```

## Example Output

```
🚀 Multi-Tool AI Assistant Demo
Testing various tool combinations...
============================================================

🔄 Test 1/7
👤 User: What's the weather like in Tokyo, Japan?
🌤️  WEATHER[Tokyo, Japan]: 22°C, Partly Cloudy
🤖 Assistant: The current weather in Tokyo, Japan is partly cloudy with a temperature of 22°C. The humidity is 62% and wind speed is 12.0 km/h.

🔄 Test 2/7
👤 User: Calculate the square root of 144
🧮 CALC[sqrt]: √144.00 = 12.00
🤖 Assistant: The square root of 144 is 12.

🔄 Test 5/7
👤 User: What's the weather in London and what time is it there?
🌤️  WEATHER[London, UK]: 18°C, Cloudy
⏰ TIME[Europe/London]: Monday, January 15, 2024 at 2:30:45 PM
🤖 Assistant: In London, UK, the weather is currently cloudy with a temperature of 18°C. The local time is Monday, January 15, 2024 at 2:30:45 PM.
```

## Learning Outcomes

This example demonstrates:

1. **Tool Selection Logic**: How the agent chooses tools based on context
2. **Multi-Tool Coordination**: Combining results from multiple tools
3. **Error Handling**: Graceful handling of tool failures
4. **Response Synthesis**: Creating coherent responses from tool outputs
5. **Context Understanding**: Parsing complex user requests

## Key Implementation Details

### Tool Registration
```go
tools := []agent.Tool{
    &WeatherTool{},
    &CalculatorTool{},
    &TimeTool{},
    &NotificationTool{},
}
```

### Agent Configuration
```go
agent.New(
    agent.WithName("multi-tool-assistant"),
    agent.WithInstructions(`You are a helpful assistant with access to multiple tools...`),
    agent.WithTools(tools...),
    // ... other options
)
```

### Tool Interface Implementation
Each tool implements the `agent.Tool` interface:
- `Name()`: Tool identifier
- `Description()`: Tool purpose
- `Schema()`: Parameter definition
- `Execute()`: Tool logic

## Architecture Benefits

1. **Modularity**: Each tool is independent and reusable
2. **Extensibility**: Easy to add new tools
3. **Testability**: Tools can be tested individually
4. **Scalability**: Framework handles tool orchestration

## Common Use Cases

This pattern is ideal for:
- **Virtual Assistants**: Multi-capability AI helpers
- **Workflow Automation**: Chaining different operations
- **API Orchestration**: Combining multiple service calls
- **Business Process Automation**: Complex multi-step workflows