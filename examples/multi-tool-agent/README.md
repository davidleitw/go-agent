# Multi-Tool Agent Example

This example demonstrates how to create an AI agent that can use multiple tools to perform different tasks. The agent can access weather information, perform calculations, and get time data.

## What You'll Learn

- How to create multiple tools with different purposes
- How to give an agent access to multiple tools
- How agents intelligently select the right tool for each task
- How tools can work together in a single conversation

## Tools Included

### 1. Weather Tool
- **Purpose**: Get weather information for any location
- **Usage**: "What's the weather in Tokyo?"
- **Returns**: Temperature, condition, humidity, wind speed

### 2. Calculator Tool  
- **Purpose**: Perform mathematical calculations
- **Operations**: Addition, subtraction, multiplication, division, square root
- **Usage**: "Calculate 15 + 7" or "What's the square root of 144?"

### 3. Time Tool
- **Purpose**: Get current time and date information
- **Features**: Different timezones, day of week, unix timestamp
- **Usage**: "What time is it in New York?"

## Running the Example

Make sure you have your OpenAI API key set up:

```bash
# Copy the environment file
cp ../../.env.example .env

# Edit .env and add your OpenAI API key
```

Run the example:

```bash
go run main.go
```

## Example Output

```
🤖 Multi-Tool Agent Example
=========================
This example demonstrates an agent that can use multiple tools:
- Weather information
- Mathematical calculations
- Current time and date

👤 User: What's the weather like in Tokyo?
🤖 Assistant: The current weather in Tokyo is sunny with a temperature of 22°C. The humidity is at 65% and there's a light wind at 5 km/h.

👤 User: Calculate the square root of 144
🤖 Assistant: The square root of 144 is 12.

👤 User: What time is it in New York?
🤖 Assistant: The current time in New York (EST timezone) is 2024-01-15 14:30:25. Today is Monday.

👤 User: Can you tell me the weather in Paris and what time it is there?
🤖 Assistant: In Paris, the weather is currently sunny with a temperature of 22°C, 65% humidity, and light winds at 5 km/h. The current time in Paris would be around 20:30:25 (8:30 PM) today, Monday.

✅ Multi-tool demonstration completed!
💬 Total messages in session: 10
```

## Key Features Demonstrated

### 1. Tool Definition
Each tool implements the `agent.Tool` interface:

```go
type WeatherTool struct{}

func (w *WeatherTool) Name() string {
    return "get_weather"
}

func (w *WeatherTool) Description() string {
    return "Get current weather information for a specified location"
}

func (w *WeatherTool) Schema() map[string]any {
    // JSON Schema definition for parameters
}

func (w *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // Tool implementation
}
```

### 2. Agent with Multiple Tools
```go
agent, err := agent.NewBasicAgent(agent.BasicAgentConfig{
    Name:        "multi-tool-assistant",
    Description: "An AI assistant with access to weather, calculator, and time tools",
    Instructions: `You are a helpful assistant with access to multiple tools...`,
    Model:       "gpt-4o-mini",
    Tools:       []agent.Tool{weatherTool, calculatorTool, timeTool},
    ChatModel:   chatModel,
})
```

### 3. Intelligent Tool Selection
The agent automatically:
- Analyzes user requests
- Selects the appropriate tool(s)
- Uses multiple tools when needed
- Provides comprehensive responses

## Benefits

- **Modularity**: Each tool has a single, clear purpose
- **Reusability**: Tools can be used across different agents
- **Extensibility**: Easy to add new tools
- **Intelligence**: Agent selects tools automatically based on context

## Next Steps

Try modifying the example:
1. Add a new tool (e.g., currency converter, unit converter)
2. Modify existing tools to add more features
3. Create an agent with different tool combinations
4. Experiment with different instructions to change behavior

This example shows how simple it is to create powerful, multi-functional AI agents using the go-agent framework!