# Task Completion Example

This example demonstrates structured output and data collection workflows using the go-agent framework. It shows how to build an agent that systematically collects information and provides structured responses.

## 🎯 Purpose

- Show how to implement structured output types
- Demonstrate systematic data collection workflows
- Illustrate JSON schema generation and validation
- Showcase task-oriented conversation patterns
- Provide examples of progress tracking and completion states

## 🚀 Running the Example

```bash
# From the project root directory
go run cmd/examples/task-completion/main.go
```

## 📋 Prerequisites

- OpenAI API key set in environment variable `OPENAI_API_KEY`
- Go 1.21 or later

## 🏗️ Code Structure & Implementation

### 1. Structured Output Definition

```go
// TaskStatusOutput represents the structured output for task completion tracking
type TaskStatusOutput struct {
    Name           *string  `json:"name"`
    Phone          *string  `json:"phone"`
    Date           *string  `json:"date"`
    Time           *string  `json:"time"`
    PartySize      *int     `json:"party_size"`
    CompletionFlag bool     `json:"completion_flag"`
    MissingFields  []string `json:"missing_fields"`
}
```

**Purpose**: Define structured output format for task completion tracking
**API Usage**:
- Pointer types (`*string`, `*int`) allow for null values in JSON
- `json` tags define field names in output
- `CompletionFlag` indicates whether task is complete
- `MissingFields` provides list of required fields still needed

### 2. Output Type Implementation

```go
// TaskOutputType implements the agent.OutputType interface
type TaskOutputType struct{}

func (t *TaskOutputType) Name() string {
    return "TaskStatusOutput"
}

func (t *TaskOutputType) Description() string {
    return "Structured output for tracking task completion status with missing fields"
}

func (t *TaskOutputType) Schema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "name": map[string]any{
                "type":        "string",
                "description": "Customer's full name",
            },
            "phone": map[string]any{
                "type":        "string",
                "description": "Customer's phone number",
            },
            "date": map[string]any{
                "type":        "string",
                "description": "Reservation date in YYYY-MM-DD format",
            },
            "time": map[string]any{
                "type":        "string",
                "description": "Reservation time in HH:MM format",
            },
            "party_size": map[string]any{
                "type":        "integer",
                "description": "Number of people in the party",
            },
            "completion_flag": map[string]any{
                "type":        "boolean",
                "description": "Whether all required information has been collected",
            },
            "missing_fields": map[string]any{
                "type":        "array",
                "items":       map[string]any{"type": "string"},
                "description": "List of fields that are still missing",
            },
        },
        "required": []string{"completion_flag", "missing_fields"},
    }
}

func (t *TaskOutputType) NewInstance() any {
    return &TaskStatusOutput{}
}

func (t *TaskOutputType) Validate(data any) error {
    _, ok := data.(*TaskStatusOutput)
    if !ok {
        return fmt.Errorf("expected *TaskStatusOutput, got %T", data)
    }
    return nil
}
```

**Purpose**: Implement the `agent.OutputType` interface for structured output
**API Usage**:
- `Name()` - Returns type identifier
- `Description()` - Provides human-readable description
- `Schema()` - Returns JSON schema defining the structure
- `NewInstance()` - Creates new instance of the output type
- `Validate(data)` - Validates that data matches expected type

### 3. Agent Configuration with Structured Output

```go
// Create structured output type
taskOutputType := &TaskOutputType{}

// Create reservation assistant with structured output
assistant, err := agent.NewBasicAgent(agent.BasicAgentConfig{
    Name:        "reservation-assistant",
    Description: "A helpful assistant for restaurant reservations",
    Instructions: `You are a restaurant reservation assistant. Your job is to collect all necessary information for a reservation.

Required information:
- Customer's full name
- Phone number
- Reservation date
- Reservation time
- Party size (number of people)

Always respond in JSON format with the current status of information collection.
Ask for missing information one field at a time.
Be polite and helpful throughout the process.`,
    Model: "gpt-4o-mini",
    ModelSettings: &agent.ModelSettings{
        Temperature: floatPtr(0.3), // Lower temperature for consistent data collection
        MaxTokens:   intPtr(800),
    },
    OutputType: taskOutputType, // Register structured output type
    ChatModel:  chatModel,
})
```

**Purpose**: Configure agent with structured output capability
**API Usage**:
- `OutputType: taskOutputType` - Register the structured output type
- Low temperature (0.3) for consistent data collection behavior
- Specific instructions for systematic information gathering
- JSON format requirement in instructions

### 4. Conversation Loop with Structured Output

```go
reservationSteps := []struct {
    input       string
    description string
}{
    {
        input:       "I want to make a restaurant reservation, I'm Mr. Lee",
        description: "initial request with name",
    },
    {
        input:       "My phone is 0912345678, I want tomorrow evening at 7pm",
        description: "phone and time information",
    },
    {
        input:       "4 people",
        description: "party size completion",
    },
}

for i, step := range reservationSteps {
    fmt.Printf("👤 User: %s\n", step.input)
    
    response, structuredOutput, err := assistant.Chat(ctx, session, step.input)
    if err != nil {
        log.Printf("❌ ERROR[%d]: %v", i+1, err)
        continue
    }
    
    fmt.Printf("🤖 Assistant: %s\n", response.Content)
    
    // Handle structured output
    if structuredOutput != nil {
        if taskStatus, ok := structuredOutput.(*TaskStatusOutput); ok {
            log.Printf("✅ STRUCTURED OUTPUT[%d]: Task completion status received", i+1)
            log.Printf("📊 COMPLETION: %t", taskStatus.CompletionFlag)
            log.Printf("📋 MISSING FIELDS: %v", taskStatus.MissingFields)
        }
    } else {
        log.Printf("WARNING[%d]: No structured output received", i+1)
    }
}
```

**Purpose**: Execute task completion workflow with structured output processing
**API Usage**:
- `assistant.Chat()` returns structured output as third parameter
- Type assertion to extract specific output type
- Structured output provides task completion status
- Missing fields list guides next steps

### 5. Progress Tracking and Validation

```go
// Example of processing structured output
if taskStatus, ok := structuredOutput.(*TaskStatusOutput); ok {
    // Check completion status
    if taskStatus.CompletionFlag {
        fmt.Println("✅ Task completed successfully!")
        fmt.Printf("📋 Final data: Name=%s, Phone=%s, Date=%s, Time=%s, Party=%d\n",
            safeString(taskStatus.Name),
            safeString(taskStatus.Phone),
            safeString(taskStatus.Date),
            safeString(taskStatus.Time),
            safeInt(taskStatus.PartySize))
    } else {
        fmt.Printf("⏳ Task in progress. Missing: %v\n", taskStatus.MissingFields)
    }
}

// Helper functions for safe dereferencing
func safeString(s *string) string {
    if s == nil {
        return "N/A"
    }
    return *s
}

func safeInt(i *int) int {
    if i == nil {
        return 0
    }
    return *i
}
```

**Purpose**: Track progress and validate completion state
**Usage**: Safely handle pointer types and provide progress feedback

## 🔧 Key APIs Demonstrated

### Output Type Interface
- `agent.OutputType` interface implementation
- `Name()` - Type identifier
- `Description()` - Type description
- `Schema()` - JSON schema definition
- `NewInstance()` - Instance creation
- `Validate(data)` - Data validation

### Agent Configuration
- `agent.BasicAgentConfig{}`
- `OutputType: outputType` - Structured output registration
- `ModelSettings{}` - Model configuration for structured output

### Structured Output Processing
- Third return value from `Chat()` method
- Type assertion for specific output types
- Progress tracking through structured data
- Validation and completion checking

## 📊 Example Output

```
🏪 Task Completion Example - Restaurant Reservation
============================================================
✅ OpenAI API key loaded
📝 Creating reservation agent with structured output...
✅ Reservation agent 'reservation-assistant' created successfully

💬 Starting reservation collection process...
============================================================

🔄 Turn 1/3
👤 User: I want to make a restaurant reservation, I'm Mr. Lee
🤖 Assistant: {"name":"Mr. Lee","phone":null,"date":null,"time":null,"party_size":null,"completion_flag":false,"missing_fields":["phone","date","time","party_size"]}

Could you please provide your phone number?
WARNING[1]: No structured output received

🔄 Turn 2/3
👤 User: My phone is 0912345678, I want tomorrow evening at 7pm
🤖 Assistant: {"name":"Mr. Lee","phone":"0912345678","date":"2023-10-04","time":"19:00","party_size":null,"completion_flag":false,"missing_fields":["party_size"]}

Thank you! Could you please let me know the number of people in your party?
WARNING[2]: No structured output received

🔄 Turn 3/3
👤 User: 4 people
🤖 Assistant: {"name":"Mr. Lee","phone":"0912345678","date":"2023-10-04","time":"19:00","party_size":4,"completion_flag":true,"missing_fields":[]}

Thank you for providing all the details! Your reservation is confirmed for Mr. Lee with 4 people on October 4th, 2023, at 7:00 PM.
WARNING[3]: No structured output received
```

## 🎓 Learning Objectives

After studying this example, you should understand:

1. **Structured Output Types**: How to define and implement output types
2. **JSON Schema Generation**: How to create schemas for structured data
3. **Data Collection Workflows**: How to systematically gather information
4. **Progress Tracking**: How to monitor task completion status
5. **Type Safety**: How to handle pointer types and null values
6. **Validation**: How to validate structured output data

## 🔄 Next Steps

- Try the [Multi-Tool Agent Example](../multi-tool-agent/) to learn about multiple tool coordination
- Explore the [Condition Testing Example](../condition-testing/) for advanced flow control
- Study the [Calculator Tool Example](../calculator-tool/) for tool implementation patterns

## 🐛 Common Issues

1. **Structured Output Parsing**: OpenAI API output format may not always match expected structure
2. **Null Pointer Handling**: Use pointer types and safe dereferencing for optional fields
3. **JSON Schema Validation**: Ensure schema accurately reflects the data structure
4. **Completion Logic**: Implement robust logic for determining task completion

## 💡 Customization Ideas

- Add validation for specific field formats (phone numbers, dates)
- Implement multi-step workflows with branching logic
- Add confirmation steps before completion
- Create different output types for different scenarios
- Add data persistence and retrieval capabilities

## 🏗️ Architecture Benefits

### Structured Output Advantages
- **Type Safety**: Compile-time checking of output structure
- **Validation**: Automatic validation of output format
- **Progress Tracking**: Built-in completion status monitoring
- **Flexibility**: Easy to extend with new fields or validation rules

### Use Cases
- **Data Collection**: Systematic gathering of user information
- **Form Processing**: Multi-step form completion
- **Survey Systems**: Structured questionnaire responses
- **Booking Systems**: Reservation and appointment scheduling
- **Configuration Wizards**: Step-by-step setup processes

## 🔍 Advanced Features

- **Pointer Types**: Handle optional fields with null values
- **JSON Schema**: Automatic schema generation for validation
- **Progress Tracking**: Real-time completion status monitoring
- **Field Validation**: Type checking and format validation
- **Error Handling**: Graceful handling of incomplete or invalid data