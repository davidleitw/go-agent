# Task Completion Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates advanced condition validation and iterative information collection using structured output and LLM-driven completion detection.

## Overview

The task completion example showcases:
- **Structured Output**: Using JSON schema for tracking missing information
- **Condition Validation**: LLM-driven detection of missing required fields
- **Iterative Collection**: Multi-turn conversation to gather complete information
- **Completion Detection**: Automatic flag setting when all conditions are met
- **Safety Controls**: Maximum iteration limits to prevent token overuse

## Scenario: Restaurant Reservation System

This example simulates a restaurant reservation assistant that must collect:
1. **Customer name** (`name`)
2. **Phone number** (`phone`)
3. **Date** (`date`)
4. **Time** (`time`)
5. **Party size** (`party_size`)

The agent uses structured output to track progress and determine when all required information has been collected.

## Code Structure

### Key Components

1. **Structured Output Definition**
   ```go
   type ReservationStatus struct {
       MissingFields  []string          `json:"missing_fields"`
       CollectedInfo  map[string]string `json:"collected_info"`
       CompletionFlag bool              `json:"completion_flag"`
       Message        string            `json:"message"`
       NextStep       string            `json:"next_step,omitempty"`
   }
   ```
   - `MissingFields`: Array of still-needed information
   - `CollectedInfo`: Key-value pairs of gathered data
   - `CompletionFlag`: Boolean indicating task completion
   - `Message`: User-friendly status message
   - `NextStep`: Optional guidance for next action

2. **Agent Configuration**
   ```go
   reservationAgent, err := agent.New(
       agent.WithName("reservation-assistant"),
       agent.WithInstructions(`You are a restaurant reservation assistant...`),
       agent.WithOpenAI(apiKey),
       agent.WithModel("gpt-4"),
       agent.WithModelSettings(&agent.ModelSettings{
           Temperature: floatPtr(0.3), // Lower for consistent structured output
           MaxTokens:   intPtr(800),
       }),
       agent.WithStructuredOutput(&ReservationStatus{}),
       agent.WithSessionStore(agent.NewInMemorySessionStore()),
       agent.WithDebugLogging(),
   )
   ```
   - Lower temperature (0.3) for more consistent structured output
   - Structured output automatically generates JSON schema
   - Specific instructions for reservation collection

3. **Simulated User Flow**
   ```go
   userInputs := []string{
       "我想要預訂餐廳，我是李先生",                    // Initial incomplete request
       "我的電話是0912345678，想要明天晚上7點",        // Partial info
       "4個人",                               // Final missing piece
   }
   ```
   - Gradually provides information across multiple turns
   - Tests the agent's ability to track partial progress

4. **Iteration Control**
   ```go
   maxTurns := 5
   for turn := 0; turn < len(userInputs) && turn < maxTurns; turn++ {
       // Process each turn
       response, structuredOutput, err := reservationAgent.Chat(ctx, sessionID, userInput)
       
       if reservationStatus, ok := structuredOutput.(*ReservationStatus); ok {
           if reservationStatus.CompletionFlag {
               log.Printf("COMPLETION[%d]: Task completed successfully!", turn+1)
               break
           }
       }
   }
   ```
   - Safety limit of 5 turns maximum
   - Automatic completion detection via `CompletionFlag`
   - Graceful exit when task is complete

## Structured Output Processing

### Response Analysis
```go
if structuredOutput != nil {
    if reservationStatus, ok := structuredOutput.(*ReservationStatus); ok {
        // Display status
        fmt.Printf("   • Missing fields: %s\n", strings.Join(reservationStatus.MissingFields, ", "))
        fmt.Printf("   • Collected info: %d items\n", len(reservationStatus.CollectedInfo))
        
        // Check completion
        if reservationStatus.CompletionFlag {
            fmt.Println("\n🎉 Reservation completed successfully!")
            break
        }
    }
}
```

### Expected Flow Progression

**Turn 1**: `"我想要預訂餐廳，我是李先生"`
```json
{
  "missing_fields": ["phone", "date", "time", "party_size"],
  "collected_info": {"name": "李先生"},
  "completion_flag": false,
  "message": "我已經記錄您的姓名。還需要您的電話號碼、用餐日期、時間和人數。"
}
```

**Turn 2**: `"我的電話是0912345678，想要明天晚上7點"`
```json
{
  "missing_fields": ["party_size"],
  "collected_info": {
    "name": "李先生",
    "phone": "0912345678", 
    "date": "明天",
    "time": "晚上7點"
  },
  "completion_flag": false,
  "message": "很好！最後請告訴我有幾個人用餐？"
}
```

**Turn 3**: `"4個人"`
```json
{
  "missing_fields": [],
  "collected_info": {
    "name": "李先生",
    "phone": "0912345678",
    "date": "明天", 
    "time": "晚上7點",
    "party_size": "4個人"
  },
  "completion_flag": true,
  "message": "完美！預訂已完成。"
}
```

## Logging System

The example provides detailed logging at multiple levels:

### Log Categories
- **REQUEST**: Input processing and turn tracking
- **RESPONSE**: LLM response details and timing
- **STRUCTURED**: Structured output parsing and validation
- **PROGRESS**: Missing field tracking and completion status
- **COMPLETION**: Task completion detection
- **SESSION**: Conversation state management

### Sample Log Output
```
🏪 Task Completion Example - Restaurant Reservation
============================================================
✅ OpenAI API key loaded
📝 Creating reservation agent with structured output...
✅ Reservation agent 'reservation-assistant' created successfully

💬 Starting reservation collection process...
============================================================

🔄 Turn 1/3
👤 User: 我想要預訂餐廳，我是李先生
REQUEST[1]: Processing user input
RESPONSE[1]: Duration: 2.1s
STRUCTURED[1]: Parsed reservation status successfully
STRUCTURED[1]: Missing fields: [phone date time party_size]
STRUCTURED[1]: Completion flag: false
PROGRESS[1]: Still missing: phone, date, time, party_size

🔄 Turn 2/3
👤 User: 我的電話是0912345678，想要明天晚上7點
STRUCTURED[2]: Missing fields: [party_size]
PROGRESS[2]: Still missing: party_size

🔄 Turn 3/3
👤 User: 4個人
COMPLETION[3]: Task completed successfully!
🎉 Reservation completed successfully!
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
   cd cmd/examples/task-completion
   go mod tidy
   ```

3. **Run the Example**:
   ```bash
   go run main.go
   ```

## Key Learning Points

### 1. Structured Output Design
- **Clear Schema**: Well-defined JSON structure for tracking state
- **Progress Tracking**: Array of missing fields for transparency
- **Completion Detection**: Boolean flag for automatic termination

### 2. LLM-Driven Logic
- **Condition Evaluation**: LLM determines what information is missing
- **Dynamic Instructions**: Agent adapts prompts based on current state
- **Natural Language Processing**: Extracts relevant data from conversational input

### 3. Safety Mechanisms
- **Iteration Limits**: Prevents infinite loops and excessive token usage
- **Error Handling**: Graceful degradation when parsing fails
- **State Validation**: Ensures structured output matches expected format

### 4. Backend-Friendly Design
- **Session Persistence**: Conversation state maintained across turns
- **Structured Data**: Easy integration with databases and APIs
- **Audit Trail**: Complete logging for debugging and analysis

## Troubleshooting

### Common Issues

1. **Inconsistent Structured Output**
   - **Cause**: Temperature too high or unclear instructions
   - **Solution**: Lower temperature (0.1-0.3) and refine prompts

2. **Completion Flag Never Set**
   - **Cause**: LLM not recognizing completion criteria
   - **Solution**: Add explicit completion examples in instructions

3. **Missing Field Detection Issues**
   - **Cause**: Ambiguous field names or requirements
   - **Solution**: Use clear, specific field names and validation rules

### Debug Tips

1. **Monitor Structured Output**: Check if JSON parsing succeeds
2. **Track Field Changes**: Watch how `missing_fields` array evolves
3. **Validate Instructions**: Ensure LLM understands completion criteria
4. **Test Edge Cases**: Try incomplete or ambiguous user inputs

## Customization

### Adapting to Different Scenarios

1. **Change Required Fields**:
   ```go
   // For hotel booking
   type BookingStatus struct {
       MissingFields  []string `json:"missing_fields"`
       CollectedInfo  map[string]string `json:"collected_info"`
       CompletionFlag bool `json:"completion_flag"`
       // Add hotel-specific fields
       RoomType      string `json:"room_type,omitempty"`
       CheckIn       string `json:"check_in,omitempty"`
       CheckOut      string `json:"check_out,omitempty"`
   }
   ```

2. **Modify Instructions**:
   ```go
   agent.WithInstructions(`You are a hotel booking assistant. Collect:
   1. Guest name, 2. Phone number, 3. Check-in date, 
   4. Check-out date, 5. Room preferences...`)
   ```

3. **Adjust Iteration Limits**:
   ```go
   maxTurns := 10 // For more complex scenarios
   ```

## Next Steps

After understanding this example:
1. Implement your own structured output types
2. Experiment with different completion criteria
3. Add validation logic for collected information
4. Integrate with external APIs for real bookings
5. Explore the **Calculator Tool** example for function calling