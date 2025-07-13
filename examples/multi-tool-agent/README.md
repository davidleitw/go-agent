# Multi-Tool Agent Example

This example demonstrates advanced agent implementation with multiple tool coordination, showcasing how to build sophisticated AI agents that can intelligently coordinate various capabilities.

## 🎯 Purpose

- Show how to implement custom agents with the `agent.Agent` interface
- Demonstrate coordination between multiple tools
- Illustrate advanced state management and tool usage tracking
- Showcase dynamic instruction enhancement based on tool usage
- Provide examples of complex conversation flows

## 🚀 Running the Example

```bash
# From the project root directory
go run cmd/examples/multi-tool-agent/main.go
```

## 📋 Prerequisites

- OpenAI API key set in environment variable `OPENAI_API_KEY`
- Go 1.21 or later

## 🏗️ Code Structure & Implementation

### 1. Custom Agent Implementation

```go
// MultiToolAgent implements the agent.Agent interface with advanced tool coordination
type MultiToolAgent struct {
    name         string
    description  string
    chatModel    agent.ChatModel
    tools        []agent.Tool
    
    // Advanced state management
    toolUsageStats map[string]int
    lastToolUsed   string
    instructions   string
}

// Name returns the agent's name
func (a *MultiToolAgent) Name() string {
    return a.name
}

// Description returns the agent's description
func (a *MultiToolAgent) Description() string {
    return a.description
}

// GetTools returns the tools available to this agent
func (a *MultiToolAgent) GetTools() []agent.Tool {
    return a.tools
}

// GetOutputType returns nil as this agent doesn't use structured output
func (a *MultiToolAgent) GetOutputType() agent.OutputType {
    return nil
}

// GetFlowRules returns nil as this agent doesn't use flow rules
func (a *MultiToolAgent) GetFlowRules() []agent.FlowRule {
    return nil
}
```

**Purpose**: Implement custom agent with advanced capabilities
**API Usage**:
- `agent.Agent` interface requires all methods to be implemented
- Custom state management through struct fields
- Tool usage statistics tracking
- Dynamic instruction enhancement

### 2. Multiple Tool Definitions

#### Weather Tool
```go
// WeatherTool provides weather information for locations
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
    return "get_weather"
}

func (t *WeatherTool) Description() string {
    return "Get current weather information for a specific location"
}

func (t *WeatherTool) Schema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "location": map[string]any{
                "type":        "string",
                "description": "The city and country/state for weather lookup",
            },
        },
        "required": []string{"location"},
    }
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    location := args["location"].(string)
    
    // Simulate weather API call
    weatherData := map[string]any{
        "location":    location,
        "temperature": 20 + rand.Intn(20), // Random temperature 20-40°C
        "condition":   []string{"Sunny", "Cloudy", "Rainy", "Partly Cloudy"}[rand.Intn(4)],
        "humidity":    40 + rand.Intn(40), // Random humidity 40-80%
        "windSpeed":   float64(5 + rand.Intn(20)), // Random wind speed 5-25 km/h
    }
    
    return weatherData, nil
}
```

#### Calculator Tool
```go
// CalculatorTool performs mathematical calculations
type CalculatorTool struct{}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    operation := args["operation"].(string)
    operand1 := args["operand1"].(float64)
    
    var result float64
    switch operation {
    case "add":
        operand2 := args["operand2"].(float64)
        result = operand1 + operand2
    case "subtract":
        operand2 := args["operand2"].(float64)
        result = operand1 - operand2
    case "multiply":
        operand2 := args["operand2"].(float64)
        result = operand1 * operand2
    case "divide":
        operand2 := args["operand2"].(float64)
        if operand2 == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        result = operand1 / operand2
    default:
        return nil, fmt.Errorf("unsupported operation: %s", operation)
    }
    
    return map[string]any{
        "operation": operation,
        "operand1":  operand1,
        "operand2":  args["operand2"],
        "result":    result,
    }, nil
}
```

#### Time Tool
```go
// TimeTool provides current time information for different timezones
type TimeTool struct{}

func (t *TimeTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    timezone := "UTC"
    if tz, exists := args["timezone"]; exists && tz != nil {
        timezone = tz.(string)
    }
    
    var loc *time.Location
    var err error
    
    if timezone == "UTC" {
        loc = time.UTC
    } else {
        loc, err = time.LoadLocation(timezone)
        if err != nil {
            return nil, fmt.Errorf("invalid timezone: %s", timezone)
        }
    }
    
    now := time.Now().In(loc)
    
    return map[string]any{
        "timezone":     timezone,
        "current_time": now.Format("Monday, January 2, 2006 at 3:04:05 PM"),
        "timestamp":    now.Unix(),
    }, nil
}
```

#### Notification Tool
```go
// NotificationTool schedules notifications
type NotificationTool struct{}

func (t *NotificationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    message := args["message"].(string)
    
    var scheduleTime time.Time
    var err error
    
    if timeStr, exists := args["time"]; exists && timeStr != nil {
        scheduleTime, err = time.Parse("15:04", timeStr.(string))
        if err != nil {
            return nil, fmt.Errorf("invalid time format: %s", timeStr)
        }
        
        // Set to today's date
        now := time.Now()
        scheduleTime = time.Date(now.Year(), now.Month(), now.Day(), 
                               scheduleTime.Hour(), scheduleTime.Minute(), 0, 0, now.Location())
        
        // If the time has passed today, schedule for tomorrow
        if scheduleTime.Before(now) {
            scheduleTime = scheduleTime.Add(24 * time.Hour)
        }
    } else {
        scheduleTime = time.Now().Add(5 * time.Minute) // Default: 5 minutes from now
    }
    
    minutesFromNow := int(time.Until(scheduleTime).Minutes())
    
    return map[string]any{
        "message":        message,
        "scheduled_time": scheduleTime.Format("15:04"),
        "minutes_from_now": minutesFromNow,
        "status":         "scheduled",
    }, nil
}
```

### 3. Advanced Chat Implementation

```go
// Chat implements the core conversation logic with tool coordination
func (a *MultiToolAgent) Chat(ctx context.Context, session agent.Session, userInput string) (*agent.Message, any, error) {
    // Add user message to session
    userMessage := agent.NewUserMessage(userInput)
    session.AddMessage(userMessage)
    
    // Build enhanced instructions based on tool usage
    enhancedInstructions := a.buildEnhancedInstructions()
    
    // Prepare messages with system instructions
    messages := []agent.Message{agent.NewSystemMessage(enhancedInstructions)}
    messages = append(messages, session.Messages()...)
    
    // Execute conversation loop with tool coordination
    maxTurns := 5
    for turn := 0; turn < maxTurns; turn++ {
        log.Printf("🔄 TURN[%d]: Starting conversation turn", turn+1)
        
        // Get model response
        response, err := a.chatModel.GenerateChatCompletion(
            ctx,
            messages,
            "gpt-4o-mini",
            &agent.ModelSettings{
                Temperature: floatPtr(0.7),
                MaxTokens:   intPtr(1000),
            },
            a.tools,
        )
        if err != nil {
            return nil, nil, fmt.Errorf("failed to generate response: %w", err)
        }
        
        // Add response to session
        session.AddMessage(*response)
        
        // Handle tool calls if present
        if len(response.ToolCalls) > 0 {
            log.Printf("🔧 Executing %d tool calls", len(response.ToolCalls))
            
            if err := a.executeToolCalls(ctx, session, response.ToolCalls); err != nil {
                return nil, nil, fmt.Errorf("tool execution failed: %w", err)
            }
            
            // Update messages for next turn
            messages = []agent.Message{agent.NewSystemMessage(enhancedInstructions)}
            messages = append(messages, session.Messages()...)
            
            // Log current statistics
            a.logToolUsageStats(turn + 1)
            continue
        }
        
        // No tool calls, conversation is complete
        log.Printf("✅ TURN[%d]: Conversation completed without tool calls", turn+1)
        return response, nil, nil
    }
    
    return nil, nil, fmt.Errorf("reached maximum turns (%d) without completion", maxTurns)
}
```

**Purpose**: Implement sophisticated conversation logic with tool coordination
**API Usage**:
- Custom implementation of `Chat()` method
- Tool execution loop with error handling
- Dynamic instruction enhancement
- Comprehensive logging and statistics

### 4. Tool Execution and Coordination

```go
// executeToolCalls handles multiple tool calls with coordination
func (a *MultiToolAgent) executeToolCalls(ctx context.Context, session agent.Session, toolCalls []agent.ToolCall) error {
    for i, toolCall := range toolCalls {
        log.Printf("🔧 TOOL[%d/%d]: Executing %s", i+1, len(toolCalls), toolCall.Function.Name)
        
        // Find matching tool
        var tool agent.Tool
        for _, t := range a.tools {
            if t.Name() == toolCall.Function.Name {
                tool = t
                break
            }
        }
        
        if tool == nil {
            return fmt.Errorf("tool not found: %s", toolCall.Function.Name)
        }
        
        // Parse arguments
        var args map[string]any
        if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
            return fmt.Errorf("failed to parse tool arguments: %w", err)
        }
        
        // Execute tool with timing
        startTime := time.Now()
        result, err := tool.Execute(ctx, args)
        duration := time.Since(startTime)
        
        if err != nil {
            log.Printf("❌ TOOL[%d/%d]: %s failed in %v: %v", i+1, len(toolCalls), tool.Name(), duration, err)
            
            // Add error message to session
            errorMsg := agent.NewToolMessage(toolCall.ID, tool.Name(), fmt.Sprintf("Error: %v", err))
            session.AddMessage(errorMsg)
            continue
        }
        
        log.Printf("✅ TOOL[%d/%d]: %s completed in %v", i+1, len(toolCalls), tool.Name(), duration)
        
        // Update usage statistics
        a.toolUsageStats[tool.Name()]++
        a.lastToolUsed = tool.Name()
        
        // Add successful result to session
        resultJSON, _ := json.Marshal(result)
        toolMsg := agent.NewToolMessage(toolCall.ID, tool.Name(), string(resultJSON))
        session.AddMessage(toolMsg)
    }
    
    return nil
}
```

**Purpose**: Coordinate multiple tool executions with comprehensive error handling
**Features**:
- Tool lookup and validation
- Argument parsing and validation
- Execution timing and logging
- Statistics tracking
- Error recovery

### 5. Dynamic Instruction Enhancement

```go
// buildEnhancedInstructions creates dynamic instructions based on tool usage
func (a *MultiToolAgent) buildEnhancedInstructions() string {
    baseInstructions := a.instructions
    
    // Add tool usage statistics
    if len(a.toolUsageStats) > 0 {
        baseInstructions += "\n\nTool Usage Statistics:\n"
        for toolName, count := range a.toolUsageStats {
            baseInstructions += fmt.Sprintf("- %s: used %d times\n", toolName, count)
        }
    }
    
    // Add last tool used context
    if a.lastToolUsed != "" {
        baseInstructions += fmt.Sprintf("\nLast tool used: %s\n", a.lastToolUsed)
        baseInstructions += "Consider whether the user might want to use related tools or follow up on the previous result.\n"
    }
    
    return baseInstructions
}
```

**Purpose**: Provide context-aware instructions based on conversation history
**Features**:
- Tool usage statistics integration
- Last tool used context
- Dynamic instruction enhancement

## 🔧 Key APIs Demonstrated

### Custom Agent Implementation
- `agent.Agent` interface implementation
- `Name()`, `Description()` - Agent identification
- `GetTools()` - Tool registration
- `GetOutputType()`, `GetFlowRules()` - Optional capabilities
- `Chat(ctx, session, input)` - Core conversation logic

### Tool Coordination
- Multiple tool registration and management
- Tool execution with error handling
- Usage statistics tracking
- Dynamic instruction enhancement

### Advanced State Management
- Tool usage statistics (`toolUsageStats`)
- Last tool used tracking (`lastToolUsed`)
- Dynamic instruction building
- Session state management

## 📊 Example Output

```
✅ OpenAI API key loaded (length: 164)
🔧 Creating OpenAI chat model...
🤖 Creating custom multi-tool AI assistant...
✅ Multi-tool assistant 'multi-tool-assistant' created successfully

🚀 Testing custom multi-tool agent with different scenarios...
============================================================

🔄 Test 1/6
👤 User: What's the weather like in Tokyo, Japan?
🔄 TURN[1]: Starting conversation turn
🔧 Executing 1 tool calls
🔧 TOOL[1/1]: Executing get_weather
🌤️  WEATHER[Tokyo, Japan]: 24°C, Partly Cloudy
✅ TOOL[1/1]: get_weather completed in 11.709µs
📊 TURN[1]: Tool usage stats: map[get_weather:1]
🔄 TURN[2]: Starting conversation turn
✅ TURN[2]: Conversation completed without tool calls
🤖 Assistant: The current weather in Tokyo, Japan is partly cloudy with a temperature of 24°C. The humidity is at 57% and the wind is blowing at a speed of 17.0 km/h.

🔄 Test 2/6
👤 User: Calculate the result of 25 * 4 + 10
🔄 TURN[1]: Starting conversation turn
🔧 Executing 1 tool calls
🔧 TOOL[1/1]: Executing calculate
🧮 CALC[multiply]: 25.00 × 4.00 = 100.00
✅ TOOL[1/1]: calculate completed in 7.427µs
📊 TURN[1]: Tool usage stats: map[calculate:1 get_weather:1]
🔄 TURN[2]: Starting conversation turn
🔧 Executing 1 tool calls
🔧 TOOL[1/1]: Executing calculate
🧮 CALC[add]: 100.00 + 10.00 = 110.00
✅ TOOL[1/1]: calculate completed in 7.004µs
📊 TURN[2]: Tool usage stats: map[calculate:2 get_weather:1]
🔄 TURN[3]: Starting conversation turn
✅ TURN[3]: Conversation completed without tool calls
🤖 Assistant: The result of the calculation 25 * 4 + 10 is 110.
```

## 🎓 Learning Objectives

After studying this example, you should understand:

1. **Custom Agent Implementation**: How to implement the `agent.Agent` interface
2. **Multi-Tool Coordination**: How to manage and coordinate multiple tools
3. **Advanced State Management**: How to track tool usage and conversation state
4. **Dynamic Instructions**: How to enhance instructions based on context
5. **Error Handling**: Comprehensive error handling in complex scenarios
6. **Performance Monitoring**: Tool execution timing and statistics

## 🔄 Next Steps

- Try the [Condition Testing Example](../condition-testing/) to learn about flow control
- Explore the [Basic Chat Example](../basic-chat/) for simpler patterns
- Study the [Calculator Tool Example](../calculator-tool/) for individual tool implementation

## 🐛 Common Issues

1. **Tool Coordination**: Ensure tools don't conflict with each other
2. **State Management**: Properly track and update agent state
3. **Error Propagation**: Handle tool failures gracefully
4. **Performance**: Monitor tool execution times and optimize as needed

## 💡 Customization Ideas

- Add more sophisticated tool selection logic
- Implement tool dependency management
- Add caching for frequently used tool results
- Create tool composition for complex operations
- Add user preference tracking for tool usage

## 🏗️ Architecture Benefits

### Custom Agent Advantages
- **Full Control**: Complete control over conversation flow
- **Advanced State**: Sophisticated state management capabilities
- **Tool Coordination**: Intelligent coordination between multiple tools
- **Performance Optimization**: Custom optimization strategies
- **Extensibility**: Easy to add new capabilities and features

### Use Cases
- **Personal Assistants**: Multi-capability AI assistants
- **Workflow Automation**: Complex multi-step processes
- **Data Integration**: Combining multiple data sources
- **Service Orchestration**: Coordinating multiple services
- **Decision Support**: Complex analysis with multiple tools

## 🔍 Advanced Features

- **Tool Usage Statistics**: Track and analyze tool usage patterns
- **Dynamic Instructions**: Context-aware instruction enhancement
- **Error Recovery**: Graceful handling of tool failures
- **Performance Monitoring**: Detailed execution timing and logging
- **State Persistence**: Maintain conversation state across turns