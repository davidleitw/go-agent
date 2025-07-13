# Condition Testing Example

This example demonstrates advanced flow control and conditional logic in conversations using the go-agent framework. It shows how to build sophisticated agents with dynamic conversation flows based on conditions and rules.

## 🎯 Purpose

- Show how to implement flow rules and conditions
- Demonstrate dynamic conversation flow management
- Illustrate custom agent implementation for complex scenarios
- Showcase user onboarding and data collection workflows
- Provide examples of condition evaluation and rule activation

## 🚀 Running the Example

```bash
# From the project root directory
go run cmd/examples/condition-testing/main.go
```

## 📋 Prerequisites

- OpenAI API key set in environment variable `OPENAI_API_KEY`
- Go 1.21 or later

## 🏗️ Code Structure & Implementation

### 1. Custom Onboarding Agent

```go
// OnboardingAgent implements the agent.Agent interface with condition testing
type OnboardingAgent struct {
    name        string
    description string
    chatModel   agent.ChatModel
    tools       []agent.Tool
    flowRules   []agent.FlowRule
    
    // Advanced state management for condition testing
    conditionEvaluations map[string]int
    ruleActivations      map[string]int
    userData             map[string]string
    instructions         string
}

// Implementation of agent.Agent interface
func (a *OnboardingAgent) Name() string { return a.name }
func (a *OnboardingAgent) Description() string { return a.description }
func (a *OnboardingAgent) GetTools() []agent.Tool { return a.tools }
func (a *OnboardingAgent) GetOutputType() agent.OutputType { return nil }
func (a *OnboardingAgent) GetFlowRules() []agent.FlowRule { return a.flowRules }
```

**Purpose**: Implement custom agent with advanced condition testing capabilities
**API Usage**:
- `agent.Agent` interface implementation
- Custom state management for condition evaluation
- Flow rule integration
- User data tracking

### 2. Condition Implementations

#### Missing Fields Condition
```go
// MissingFieldsCondition checks if required fields are missing
type MissingFieldsCondition struct {
    requiredFields []string
    fieldName      string
}

func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
    log.Printf("🔍 CONDITION[%s]: Evaluating...", c.fieldName)
    
    // Check if user data exists in flow data
    userData, exists := data["userData"]
    if !exists {
        log.Printf("⏸️  CONDITION[%s]: Not triggered - no user data", c.fieldName)
        return false, nil
    }
    
    userDataMap, ok := userData.(map[string]string)
    if !ok {
        log.Printf("⏸️  CONDITION[%s]: Not triggered - invalid user data type", c.fieldName)
        return false, nil
    }
    
    // Check if any required fields are missing
    for _, field := range c.requiredFields {
        if value, exists := userDataMap[field]; !exists || value == "" {
            log.Printf("🎯 CONDITION[%s]: Triggered! Missing field: %s", c.fieldName, field)
            return true, nil
        }
    }
    
    log.Printf("⏸️  CONDITION[%s]: Not triggered - all fields present", c.fieldName)
    return false, nil
}
```

#### Completion Stage Condition
```go
// CompletionStageCondition checks if user has reached a specific completion stage
type CompletionStageCondition struct {
    targetStage string
    fieldName   string
}

func (c *CompletionStageCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
    log.Printf("🔍 CONDITION[%s]: Evaluating...", c.fieldName)
    
    userData, exists := data["userData"]
    if !exists {
        log.Printf("⏸️  CONDITION[%s]: Not triggered - no user data", c.fieldName)
        return false, nil
    }
    
    userDataMap, ok := userData.(map[string]string)
    if !ok {
        log.Printf("⏸️  CONDITION[%s]: Not triggered - invalid user data type", c.fieldName)
        return false, nil
    }
    
    // Check if user has basic info (name, email, phone)
    hasBasicInfo := userDataMap["name"] != "" && userDataMap["email"] != "" && userDataMap["phone"] != ""
    
    if c.targetStage == "basic_info" && hasBasicInfo && userDataMap["preferences"] == "" {
        log.Printf("🎯 CONDITION[%s]: Triggered! User at basic info stage", c.fieldName)
        return true, nil
    }
    
    log.Printf("⏸️  CONDITION[%s]: Not triggered", c.fieldName)
    return false, nil
}
```

#### Message Count Condition
```go
// MessageCountCondition checks if conversation has reached a certain length
type MessageCountCondition struct {
    minMessages int
    fieldName   string
}

func (c *MessageCountCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
    log.Printf("🔍 CONDITION[%s]: Evaluating...", c.fieldName)
    
    messageCount := len(session.Messages())
    log.Printf("🎯 CONDITION[%s]: Message count %d >= %d ? %t", c.fieldName, messageCount, c.minMessages, messageCount >= c.minMessages)
    
    if messageCount >= c.minMessages {
        log.Printf("🎯 CONDITION[%s]: Triggered! Long conversation detected", c.fieldName)
        return true, nil
    }
    
    log.Printf("⏸️  CONDITION[%s]: Not triggered", c.fieldName)
    return false, nil
}
```

**Purpose**: Implement various condition types for different scenarios
**API Usage**:
- `agent.Condition` interface implementation
- `Evaluate(ctx, session, data)` - Core condition evaluation method
- Context-aware condition checking
- Comprehensive logging for debugging

### 3. Flow Rule Actions

#### System Message Action
```go
// SystemMessageAction adds a system message to guide the conversation
type SystemMessageAction struct {
    message string
}

func (a *SystemMessageAction) Apply(ctx context.Context, session agent.Session, data map[string]any) error {
    log.Printf("✅ ACTION: Adding system message")
    session.AddSystemMessage(a.message)
    return nil
}
```

#### Data Collection Action
```go
// DataCollectionAction prompts for missing information
type DataCollectionAction struct {
    prompt string
}

func (a *DataCollectionAction) Apply(ctx context.Context, session agent.Session, data map[string]any) error {
    log.Printf("✅ ACTION: Prompting for data collection")
    session.AddSystemMessage(a.prompt)
    return nil
}
```

**Purpose**: Implement actions that modify conversation flow
**API Usage**:
- `agent.Action` interface implementation
- `Apply(ctx, session, data)` - Core action execution method
- Session manipulation for flow control

### 4. Tool Implementations

#### User Info Collection Tool
```go
// UserInfoCollectionTool collects and stores user information
type UserInfoCollectionTool struct {
    agent *OnboardingAgent
}

func (t *UserInfoCollectionTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    log.Printf("🔧 TOOL: Collecting user information")
    
    // Extract information from arguments
    for key, value := range args {
        if strValue, ok := value.(string); ok && strValue != "" {
            t.agent.userData[key] = strValue
            log.Printf("📝 USER_DATA: Updated %s = %s", key, strValue)
        }
    }
    
    return map[string]any{
        "status":      "collected",
        "fields":      len(t.agent.userData),
        "user_data":   t.agent.userData,
    }, nil
}
```

#### Completion Validation Tool
```go
// CompletionValidationTool validates if onboarding is complete
type CompletionValidationTool struct {
    agent *OnboardingAgent
}

func (t *CompletionValidationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    log.Printf("🔧 TOOL: Validating completion status")
    
    requiredFields := []string{"name", "email", "phone", "preferences"}
    missingFields := []string{}
    
    for _, field := range requiredFields {
        if value, exists := t.agent.userData[field]; !exists || value == "" {
            missingFields = append(missingFields, field)
        }
    }
    
    isComplete := len(missingFields) == 0
    
    return map[string]any{
        "complete":        isComplete,
        "missing_fields":  missingFields,
        "collected_data":  t.agent.userData,
        "completion_rate": float64(len(t.agent.userData)) / float64(len(requiredFields)),
    }, nil
}
```

**Purpose**: Provide tools for data collection and validation
**Features**:
- User data collection and storage
- Completion status validation
- Progress tracking

### 5. Advanced Chat Implementation with Condition Testing

```go
// Chat implements conversation with advanced condition testing
func (a *OnboardingAgent) Chat(ctx context.Context, session agent.Session, userInput string) (*agent.Message, any, error) {
    // Add user message
    userMessage := agent.NewUserMessage(userInput)
    session.AddMessage(userMessage)
    
    // Execute conversation loop with condition evaluation
    maxTurns := 5
    for turn := 0; turn < maxTurns; turn++ {
        log.Printf("🔄 TURN[%d]: Starting conversation turn with condition evaluation", turn+1)
        
        // Evaluate flow rules
        flowData := map[string]any{
            "userInput":    userInput,
            "messageCount": len(session.Messages()),
            "userData":     a.userData,
        }
        
        log.Printf("🎯 Evaluating %d flow rules", len(a.flowRules))
        for _, rule := range a.flowRules {
            // Track condition evaluations
            a.conditionEvaluations[rule.Name]++
            
            shouldTrigger, err := rule.Condition.Evaluate(ctx, session, flowData)
            if err != nil {
                log.Printf("❌ CONDITION[%s]: Evaluation failed: %v", rule.Name, err)
                continue
            }
            
            if shouldTrigger {
                log.Printf("✅ CONDITION[%s]: Triggered! Applying rule '%s'", rule.Name, rule.Name)
                
                // Track rule activations
                a.ruleActivations[rule.Name]++
                
                if err := rule.Action.Apply(ctx, session, flowData); err != nil {
                    log.Printf("❌ RULE[%s]: Action failed: %v", rule.Name, err)
                } else {
                    log.Printf("✅ RULE[%s]: Action applied successfully", rule.Name)
                }
            }
        }
        
        // Generate response with enhanced instructions
        enhancedInstructions := a.buildEnhancedInstructions()
        messages := []agent.Message{agent.NewSystemMessage(enhancedInstructions)}
        messages = append(messages, session.Messages()...)
        
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
        
        // Handle tool calls
        if len(response.ToolCalls) > 0 {
            log.Printf("🔧 Executing %d tool calls with condition awareness", len(response.ToolCalls))
            
            if err := a.executeToolCalls(ctx, session, response.ToolCalls); err != nil {
                log.Printf("❌ Tool execution failed: %v", err)
                continue
            }
            
            // Log statistics after tool execution
            a.logConditionStatistics(turn + 1)
            continue
        }
        
        // No tool calls, conversation complete
        log.Printf("✅ TURN[%d]: Conversation completed without tool calls", turn+1)
        return response, nil, nil
    }
    
    return nil, nil, fmt.Errorf("reached maximum turns without completion")
}
```

**Purpose**: Implement sophisticated conversation flow with condition testing
**Features**:
- Flow rule evaluation at each turn
- Condition statistics tracking
- Dynamic instruction enhancement
- Tool coordination with condition awareness

### 6. Flow Rule Configuration

```go
// Create flow rules with different condition types
flowRules := []agent.FlowRule{
    {
        Name: "collect-missing-contact-info",
        Condition: &MissingFieldsCondition{
            requiredFields: []string{"email", "phone"},
            fieldName:      "missing_contact_info",
        },
        Action: &DataCollectionAction{
            prompt: "Focus on collecting missing contact information (email and phone).",
        },
    },
    {
        Name: "collect-missing-preferences",
        Condition: &MissingFieldsCondition{
            requiredFields: []string{"preferences"},
            fieldName:      "missing_preferences",
        },
        Action: &DataCollectionAction{
            prompt: "Focus on collecting user preferences and interests.",
        },
    },
    {
        Name: "basic-info-completion",
        Condition: &CompletionStageCondition{
            targetStage: "basic_info",
            fieldName:   "at_basic_info_stage",
        },
        Action: &SystemMessageAction{
            message: "User has completed basic information. Now focus on preferences.",
        },
    },
    {
        Name: "optimize-long-conversation",
        Condition: &MessageCountCondition{
            minMessages: 6,
            fieldName:   "long_conversation",
        },
        Action: &SystemMessageAction{
            message: "This conversation is getting long. Provide a summary of current progress and focus on completion.",
        },
    },
}
```

**Purpose**: Configure flow rules for different scenarios
**Features**:
- Multiple condition types
- Flexible action configuration
- Rule naming and organization

## 🔧 Key APIs Demonstrated

### Flow Control Interface
- `agent.FlowRule` structure
- `agent.Condition` interface implementation
- `agent.Action` interface implementation
- `Evaluate(ctx, session, data)` - Condition evaluation
- `Apply(ctx, session, data)` - Action execution

### Custom Agent Implementation
- `agent.Agent` interface implementation
- `GetFlowRules()` - Flow rule registration
- Advanced state management
- Condition evaluation statistics

### Dynamic Flow Management
- Rule evaluation at each conversation turn
- Context-aware condition checking
- Action execution based on conditions
- Statistics tracking and logging

## 📊 Example Output

```
✅ OpenAI API key loaded (length: 164)
🔧 Creating OpenAI chat model...
🤖 Creating custom onboarding agent with advanced condition testing...
✅ Onboarding agent 'onboarding-agent' created successfully
📋 Registered 4 flow rules with various condition types

================================================================================
🧪 Advanced Condition Testing & Flow Rules Demo
Testing sophisticated condition evaluation and flow rule management...
================================================================================

🔄 Test 1/8
👤 User: Hi! I want to sign up for your service.
🔄 TURN[1]: Starting conversation turn with condition evaluation
🎯 Evaluating 4 flow rules
🔍 CONDITION[missing_contact_info]: Evaluating...
⏸️  CONDITION[missing_contact_info]: Not triggered
🔍 CONDITION[missing_preferences]: Evaluating...
⏸️  CONDITION[missing_preferences]: Not triggered
🔍 CONDITION[at_basic_info_stage]: Evaluating...
⏸️  CONDITION[at_basic_info_stage]: Not triggered
🔍 CONDITION[long_conversation]: Evaluating...
🎯 CONDITION[long_conversation]: Message count 1 >= 6 ? false
⏸️  CONDITION[long_conversation]: Not triggered
✅ TURN[1]: Conversation completed without tool calls
🤖 Agent: Great! We're excited to have you on board. Let's start the registration process.

First off, could you please provide your full name?
```

## 🎓 Learning Objectives

After studying this example, you should understand:

1. **Flow Rules**: How to implement and configure flow rules
2. **Condition Evaluation**: How to create custom conditions
3. **Action Execution**: How to implement actions that modify conversation flow
4. **Custom Agents**: How to build sophisticated agents with advanced capabilities
5. **State Management**: How to track and manage complex conversation state
6. **Dynamic Flow**: How to create adaptive conversation flows

## 🔄 Next Steps

- Try the [Multi-Tool Agent Example](../multi-tool-agent/) to learn about tool coordination
- Explore the [Task Completion Example](../task-completion/) for structured output
- Study the [Basic Chat Example](../basic-chat/) for simpler patterns

## 🐛 Common Issues

1. **Condition Logic**: Ensure conditions are properly implemented and tested
2. **Action Side Effects**: Be careful with actions that modify session state
3. **Rule Ordering**: Consider the order of rule evaluation
4. **State Consistency**: Maintain consistent state across turns

## 💡 Customization Ideas

- Add more sophisticated condition types
- Implement rule priorities and dependencies
- Add condition caching for performance
- Create composite conditions with logical operators
- Add rule activation history and analytics

## 🏗️ Architecture Benefits

### Flow Control Advantages
- **Adaptive Behavior**: Conversations adapt based on context
- **Reusable Logic**: Conditions and actions can be reused
- **Testable Components**: Each condition and action can be tested independently
- **Flexible Configuration**: Easy to modify flow rules without code changes
- **Comprehensive Logging**: Detailed logging for debugging and analysis

### Use Cases
- **User Onboarding**: Step-by-step user registration processes
- **Dynamic Surveys**: Adaptive questionnaires based on responses
- **Troubleshooting**: Context-aware support workflows
- **Sales Processes**: Adaptive sales conversation flows
- **Educational Systems**: Personalized learning paths

## 🔍 Advanced Features

- **Condition Statistics**: Track condition evaluation patterns
- **Rule Activation Tracking**: Monitor which rules are triggered
- **Dynamic Instructions**: Context-aware instruction enhancement
- **State Persistence**: Maintain conversation state across sessions
- **Performance Monitoring**: Detailed execution timing and metrics