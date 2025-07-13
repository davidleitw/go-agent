# Dynamic Schema Selection Example

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

This example demonstrates the most advanced schema-based information collection capabilities in go-agent, including intelligent intent classification, dynamic schema selection, and multi-step workflow orchestration.

## Overview

Real-world applications often need to handle multiple different conversation flows, each requiring different information. This example shows how to:

- **Classify** user intent from natural language input
- **Select** appropriate schemas dynamically based on context
- **Orchestrate** multi-step information collection workflows
- **Adapt** conversation strategies in real-time
- **Integrate** complex business logic with schema collection

## Quick Start

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# Run the example
go run examples/dynamic-schema/main.go
```

## Code Walkthrough

### 1. Intent Classification System

```go
type IntentClassifier struct {
    intentKeywords map[string][]string
}

func NewIntentClassifier() *IntentClassifier {
    return &IntentClassifier{
        intentKeywords: map[string][]string{
            "technical_support": {"error", "bug", "broken", "not working", "technical", "login", "password", "issue"},
            "billing_inquiry":   {"billing", "payment", "charge", "invoice", "refund", "subscription", "cost", "price"},
            "account_management": {"account", "profile", "settings", "change", "update", "delete", "privacy"},
            "product_inquiry":   {"feature", "how", "help", "tutorial", "guide", "documentation", "usage"},
            "sales_inquiry":     {"buy", "purchase", "pricing", "plan", "enterprise", "demo", "quote", "contact sales"},
            "general_inquiry":   {"hello", "hi", "support", "help", "question", "information"},
        },
    }
}
```

**Key Points:**
- Intent classification based on keyword analysis
- Extensible to ML-based classification
- Fallback to general inquiry for unknown intents
- Handles multiple languages and synonyms

### 2. Intent Classification Logic

```go
func (ic *IntentClassifier) ClassifyIntent(userInput string) string {
    input := strings.ToLower(userInput)
    intentScores := make(map[string]int)
    
    // Score each intent based on keyword matches
    for intent, keywords := range ic.intentKeywords {
        score := 0
        for _, keyword := range keywords {
            if strings.Contains(input, keyword) {
                score++
            }
        }
        if score > 0 {
            intentScores[intent] = score
        }
    }
    
    // Return intent with highest score
    if len(intentScores) == 0 {
        return "general_inquiry"
    }
    
    maxScore := 0
    bestIntent := "general_inquiry"
    for intent, score := range intentScores {
        if score > maxScore {
            maxScore = score
            bestIntent = intent
        }
    }
    
    return bestIntent
}
```

**Key Points:**
- Scoring system for accurate classification
- Higher scores for multiple keyword matches
- Robust fallback handling
- Simple yet effective for most use cases

### 3. Dynamic Schema Selection

```go
func getSchemaForIntent(intent string) []*schema.Field {
    switch intent {
    case "technical_support":
        return []*schema.Field{
            schema.Define("email", "Please provide your email for technical follow-up"),
            schema.Define("error_description", "Please describe the error or issue you're experiencing"),
            schema.Define("steps_taken", "What troubleshooting steps have you already tried?"),
            schema.Define("environment", "What browser/device are you using?").Optional(),
            schema.Define("urgency", "How critical is this issue for your work? (low/medium/high)").Optional(),
        }
    
    case "billing_inquiry":
        return []*schema.Field{
            schema.Define("email", "Please provide the email associated with your account"),
            schema.Define("account_id", "What is your account ID or number?"),
            schema.Define("billing_question", "Please describe your billing question or concern"),
            schema.Define("amount", "If this involves a specific amount, please specify").Optional(),
            schema.Define("transaction_date", "When did the transaction occur? (if applicable)").Optional(),
        }
    
    case "sales_inquiry":
        return []*schema.Field{
            schema.Define("email", "Please provide your business email"),
            schema.Define("company", "What company do you represent?"),
            schema.Define("team_size", "How many people are on your team?"),
            schema.Define("use_case", "How do you plan to use our product?"),
            schema.Define("timeline", "When are you looking to get started?").Optional(),
            schema.Define("budget", "Do you have a budget range in mind?").Optional(),
        }
    
    // ... more intent schemas
    
    default:
        return getGeneralInquirySchema()
    }
}
```

**Key Points:**
- Each intent has a specialized schema
- Required vs optional fields optimized per intent
- Professional field prompts for business context
- Fallback to general schema for unknown intents

### 4. Multi-Step Workflow System

```go
func getWorkflowForIntent(intent string) [][]*schema.Field {
    switch intent {
    case "technical_support":
        return [][]*schema.Field{
            { // Step 1: Basic contact and issue summary
                schema.Define("email", "Your email address"),
                schema.Define("issue_summary", "Brief description of the issue"),
            },
            { // Step 2: Technical details
                schema.Define("error_message", "Exact error message you're seeing"),
                schema.Define("steps_to_reproduce", "How can we reproduce this issue?"),
                schema.Define("browser_version", "What browser and version?").Optional(),
            },
            { // Step 3: Impact assessment
                schema.Define("when_started", "When did this issue first occur?"),
                schema.Define("frequency", "How often does this happen?"),
                schema.Define("workaround", "Any temporary workarounds you've found?").Optional(),
            },
        }
    
    case "sales_inquiry":
        return [][]*schema.Field{
            { // Step 1: Basic qualification
                schema.Define("email", "Business email address"),
                schema.Define("company", "Company name"),
                schema.Define("role", "Your role at the company"),
            },
            { // Step 2: Requirements gathering
                schema.Define("team_size", "Size of your team"),
                schema.Define("use_case", "Primary use case for our product"),
                schema.Define("current_solution", "What solution are you using now?").Optional(),
            },
            { // Step 3: Timeline and budget
                schema.Define("timeline", "When do you need to have this implemented?"),
                schema.Define("decision_process", "Who else is involved in the decision?"),
                schema.Define("budget_range", "Budget range you're considering").Optional(),
            },
        }
    
    default:
        // Single-step workflow for simple intents
        return [][]*schema.Field{getSchemaForIntent(intent)}
    }
}
```

**Key Points:**
- Complex intents broken into manageable steps
- Progressive information collection
- Context builds across workflow steps
- Flexible single-step fallback

### 5. Adaptive Conversation Strategy

```go
func runAdaptiveConversation(ctx context.Context, bot agent.Agent, scenarios []ConversationScenario) {
    classifier := NewIntentClassifier()
    
    for i, scenario := range scenarios {
        fmt.Printf("📝 Scenario %d: %s\n", i+1, scenario.Description)
        fmt.Printf("👤 User: %s\n", scenario.UserInput)
        
        // Classify intent
        intent := classifier.ClassifyIntent(scenario.UserInput)
        fmt.Printf("🎯 Detected Intent: %s\n", intent)
        
        // Get appropriate schema
        schema := getSchemaForIntent(intent)
        fmt.Printf("📋 Selected Schema (%d fields):\n", len(schema))
        for _, field := range schema {
            requiredText := "required"
            if !field.Required() {
                requiredText = "optional"
            }
            fmt.Printf("   - %s (%s): %s\n", field.Name(), requiredText, field.Prompt())
        }
        
        // Execute conversation with selected schema
        response, err := bot.Chat(ctx, scenario.UserInput,
            agent.WithSchema(schema...),
        )
        
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        fmt.Printf("🤖 Assistant: %s\n", response.Message)
        fmt.Printf("⏱️  Response time: %.3fs\n", time.Since(startTime).Seconds())
    }
}
```

**Key Points:**
- Real-time intent classification
- Schema selection based on detected intent
- Performance metrics and debugging info
- Error handling and graceful degradation

### 6. Multi-Step Workflow Execution

```go
func runMultiStepWorkflow(ctx context.Context, bot agent.Agent, intent string, userInput string) {
    workflow := getWorkflowForIntent(intent)
    sessionID := fmt.Sprintf("workflow-%s", intent)
    
    fmt.Printf("🔄 Multi-Step Workflow Example\n")
    fmt.Printf("👤 User: %s\n", userInput)
    fmt.Printf("🎯 Intent: %s\n", intent)
    fmt.Printf("📊 Workflow Steps: %d\n\n", len(workflow))
    
    for stepIndex, stepSchema := range workflow {
        fmt.Printf("📋 Step %d/%d - Collecting:\n", stepIndex+1, len(workflow))
        for _, field := range stepSchema {
            requiredText := "required"
            if !field.Required() {
                requiredText = "optional"
            }
            fmt.Printf("   - %s (%s)\n", field.Name(), requiredText)
        }
        
        // Execute this step
        response, err := bot.Chat(ctx, userInput,
            agent.WithSession(sessionID),
            agent.WithSchema(stepSchema...),
        )
        
        if err != nil {
            fmt.Printf("❌ Step %d failed: %v\n", stepIndex+1, err)
            break
        }
        
        fmt.Printf("👤 User: %s\n", userInput)
        fmt.Printf("🤖 Assistant: %s\n", response.Message)
        
        // Check if this step is complete
        if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
            if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
                fmt.Printf("📊 Still need: %v\n", missingFields)
            }
        } else {
            fmt.Printf("✅ Step %d completed!\n", stepIndex+1)
        }
        
        // Simulate user providing information for next step
        userInput = generateSimulatedUserResponse(stepIndex, intent)
        fmt.Printf("\n")
    }
}
```

**Key Points:**
- Step-by-step execution with progress tracking
- Session continuity across workflow steps
- Missing field identification and handling
- Simulated user responses for demonstration

## Example Scenarios

### Scenario 1: Technical Support Classification

```go
// User input: "I'm getting a login error when trying to access the system"
// Expected: Technical support intent detected
// Schema: Technical support fields (email, error_description, steps_taken, etc.)
// Workflow: 3-step technical troubleshooting process

userInput := "I'm getting a login error when trying to access the system"
intent := classifier.ClassifyIntent(userInput)
// Returns: "technical_support"

schema := getSchemaForIntent(intent)
// Returns: Technical support schema with 5 fields (3 required, 2 optional)

workflow := getWorkflowForIntent(intent)
// Returns: 3-step workflow for comprehensive technical support
```

### Scenario 2: Sales Inquiry Processing

```go
// User input: "I'm interested in purchasing your enterprise plan for my company"
// Expected: Sales inquiry intent detected
// Schema: Sales qualification fields (email, company, team_size, use_case, etc.)
// Workflow: 3-step sales qualification process

userInput := "I'm interested in purchasing your enterprise plan for my company"
intent := classifier.ClassifyIntent(userInput)
// Returns: "sales_inquiry"

schema := getSchemaForIntent(intent)
// Returns: Sales schema with 6 fields (4 required, 2 optional)

workflow := getWorkflowForIntent(intent)
// Returns: 3-step workflow for sales qualification
```

### Scenario 3: Billing Issue Handling

```go
// User input: "I have a question about charges on my latest invoice"
// Expected: Billing inquiry intent detected
// Schema: Billing-specific fields (email, account_id, billing_question, etc.)
// Workflow: Single-step billing information collection

userInput := "I have a question about charges on my latest invoice"
intent := classifier.ClassifyIntent(userInput)
// Returns: "billing_inquiry"

schema := getSchemaForIntent(intent)
// Returns: Billing schema with 5 fields (3 required, 2 optional)
```

## Advanced Features

### 1. Intent Confidence Scoring

```go
func (ic *IntentClassifier) ClassifyWithConfidence(userInput string) (string, float64) {
    input := strings.ToLower(userInput)
    intentScores := make(map[string]int)
    totalKeywords := 0
    
    for intent, keywords := range ic.intentKeywords {
        score := 0
        for _, keyword := range keywords {
            if strings.Contains(input, keyword) {
                score++
            }
        }
        intentScores[intent] = score
        totalKeywords += len(keywords)
    }
    
    maxScore := 0
    bestIntent := "general_inquiry"
    for intent, score := range intentScores {
        if score > maxScore {
            maxScore = score
            bestIntent = intent
        }
    }
    
    confidence := float64(maxScore) / float64(len(strings.Fields(input)))
    return bestIntent, confidence
}
```

### 2. Context-Aware Schema Adaptation

```go
func adaptSchemaToContext(baseSchema []*schema.Field, conversationHistory []agent.Message) []*schema.Field {
    adaptedSchema := make([]*schema.Field, 0, len(baseSchema))
    
    // Analyze conversation history for context
    hasUrgencyIndicators := false
    hasCompanyMention := false
    
    for _, msg := range conversationHistory {
        content := strings.ToLower(msg.Content)
        if strings.Contains(content, "urgent") || strings.Contains(content, "critical") {
            hasUrgencyIndicators = true
        }
        if strings.Contains(content, "company") || strings.Contains(content, "business") {
            hasCompanyMention = true
        }
    }
    
    // Adapt schema based on context
    for _, field := range baseSchema {
        adaptedField := field
        
        // Make urgency required if urgency indicators present
        if field.Name() == "urgency" && hasUrgencyIndicators {
            adaptedField = schema.Define(field.Name(), field.Prompt())
        }
        
        // Add company-specific fields if business context detected
        if field.Name() == "email" && hasCompanyMention {
            adaptedSchema = append(adaptedSchema, 
                schema.Define("company_name", "What company do you represent?"))
        }
        
        adaptedSchema = append(adaptedSchema, adaptedField)
    }
    
    return adaptedSchema
}
```

### 3. Workflow State Management

```go
type WorkflowState struct {
    Intent         string                 `json:"intent"`
    CurrentStep    int                    `json:"current_step"`
    TotalSteps     int                    `json:"total_steps"`
    CollectedData  map[string]interface{} `json:"collected_data"`
    CompletedSteps []int                  `json:"completed_steps"`
    StartTime      time.Time              `json:"start_time"`
}

func (ws *WorkflowState) IsComplete() bool {
    return ws.CurrentStep >= ws.TotalSteps
}

func (ws *WorkflowState) Progress() float64 {
    return float64(len(ws.CompletedSteps)) / float64(ws.TotalSteps)
}

func (ws *WorkflowState) NextStep() int {
    if ws.CurrentStep < ws.TotalSteps-1 {
        return ws.CurrentStep + 1
    }
    return ws.CurrentStep
}
```

### 4. Analytics and Reporting

```go
type ConversationAnalytics struct {
    TotalMessages     int                    `json:"total_messages"`
    IntentChanges     int                    `json:"intent_changes"`
    WorkflowComplete  bool                   `json:"workflow_complete"`
    CollectionRate    float64                `json:"collection_rate"`
    AverageResponse   time.Duration          `json:"average_response_time"`
    FieldsCollected   map[string]interface{} `json:"fields_collected"`
}

func generateAnalytics(session agent.Session, workflow [][]*schema.Field) *ConversationAnalytics {
    messages := session.Messages()
    
    // Calculate collection rate
    totalFields := 0
    collectedFields := 0
    
    for _, step := range workflow {
        for _, field := range step {
            totalFields++
            if field.Required() {
                // Check if field was collected
                if isFieldCollected(field.Name(), messages) {
                    collectedFields++
                }
            }
        }
    }
    
    collectionRate := float64(collectedFields) / float64(totalFields)
    
    return &ConversationAnalytics{
        TotalMessages:    len(messages),
        WorkflowComplete: collectionRate >= 0.8, // 80% collection rate threshold
        CollectionRate:   collectionRate,
        FieldsCollected:  extractCollectedFields(messages),
    }
}
```

## Testing

Run the comprehensive test suite:

```bash
go test ./examples/dynamic-schema/
```

### Test Coverage

#### Intent Classification Tests
```go
func TestIntentClassifier(t *testing.T) {
    classifier := NewIntentClassifier()
    
    testCases := []struct {
        input    string
        expected string
    }{
        {"I'm getting a login error", "technical_support"},
        {"I have a billing question about my invoice", "billing_inquiry"},
        {"I need to change my password", "account_management"},
        {"How do I use the analytics feature?", "product_inquiry"},
        {"I want to purchase your enterprise plan", "sales_inquiry"},
        {"Hello, I have some questions", "general_inquiry"},
    }
    
    for _, tc := range testCases {
        result := classifier.ClassifyIntent(tc.input)
        assert.Equal(t, tc.expected, result)
    }
}
```

#### Schema Selection Tests
```go
func TestGetSchemaForIntent(t *testing.T) {
    testCases := []struct {
        intent           string
        expectedRequired []string
        expectedOptional []string
    }{
        {
            intent:           "technical_support",
            expectedRequired: []string{"email", "error_description", "steps_taken"},
            expectedOptional: []string{"environment", "urgency"},
        },
        // ... more test cases
    }
    
    for _, tc := range testCases {
        schema := getSchemaForIntent(tc.intent)
        validateSchemaFields(t, schema, tc.expectedRequired, tc.expectedOptional)
    }
}
```

#### Workflow Tests
```go
func TestWorkflowExecution(t *testing.T) {
    mockModel := NewMockChatModel(
        `{"email": null, "error_description": null, "steps_taken": null}`,
        "I understand you're having a technical issue. Please provide your email and describe the error.",
    )
    
    bot, err := agent.New("test-bot").
        WithChatModel(mockModel).
        WithInstructions("You are a test assistant.").
        Build()
    
    require.NoError(t, err)
    
    workflow := getWorkflowForIntent("technical_support")
    require.Equal(t, 3, len(workflow))
    
    // Test each workflow step
    for i, step := range workflow {
        response, err := bot.Chat(context.Background(), 
            fmt.Sprintf("Test input for step %d", i+1),
            agent.WithSchema(step...),
        )
        require.NoError(t, err)
        require.NotEmpty(t, response.Message)
    }
}
```

## Performance Optimization

### 1. Schema Caching

```go
var (
    schemaCache     = make(map[string][]*schema.Field)
    workflowCache   = make(map[string][][]*schema.Field)
    cacheMutex      sync.RWMutex
)

func getCachedSchema(intent string) []*schema.Field {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    
    if schema, exists := schemaCache[intent]; exists {
        return schema
    }
    
    cacheMutex.RUnlock()
    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    
    // Double-check after acquiring write lock
    if schema, exists := schemaCache[intent]; exists {
        return schema
    }
    
    schema := getSchemaForIntent(intent)
    schemaCache[intent] = schema
    return schema
}
```

### 2. Intent Classification Optimization

```go
type OptimizedIntentClassifier struct {
    keywordTrie  *Trie
    intentScorer *IntentScorer
}

func (oic *OptimizedIntentClassifier) FastClassify(input string) string {
    // Use trie for O(1) keyword lookup
    keywords := oic.keywordTrie.FindKeywords(input)
    
    // Use optimized scoring algorithm
    return oic.intentScorer.CalculateBestIntent(keywords)
}
```

### 3. Workflow State Persistence

```go
type WorkflowStateManager struct {
    states map[string]*WorkflowState
    mutex  sync.RWMutex
}

func (wsm *WorkflowStateManager) SaveState(sessionID string, state *WorkflowState) {
    wsm.mutex.Lock()
    defer wsm.mutex.Unlock()
    wsm.states[sessionID] = state
}

func (wsm *WorkflowStateManager) LoadState(sessionID string) (*WorkflowState, bool) {
    wsm.mutex.RLock()
    defer wsm.mutex.RUnlock()
    state, exists := wsm.states[sessionID]
    return state, exists
}
```

## Integration Examples

### With Machine Learning Classification

```go
type MLIntentClassifier struct {
    modelEndpoint string
    fallbackClassifier *IntentClassifier
}

func (ml *MLIntentClassifier) ClassifyIntent(input string) string {
    // Try ML classification first
    result, confidence, err := ml.callMLModel(input)
    if err != nil || confidence < 0.8 {
        // Fallback to keyword-based classification
        return ml.fallbackClassifier.ClassifyIntent(input)
    }
    return result
}
```

### With External CRM Systems

```go
func syncWithCRM(collectedData map[string]interface{}, intent string) error {
    switch intent {
    case "sales_inquiry":
        return createCRMLead(collectedData)
    case "technical_support":
        return createSupportTicket(collectedData)
    case "billing_inquiry":
        return createBillingInquiry(collectedData)
    default:
        return createGeneralInquiry(collectedData)
    }
}
```

### With Analytics Platforms

```go
func trackConversationMetrics(analytics *ConversationAnalytics) {
    metrics := map[string]interface{}{
        "intent":              analytics.Intent,
        "workflow_completed":  analytics.WorkflowComplete,
        "collection_rate":     analytics.CollectionRate,
        "response_time":       analytics.AverageResponse.Milliseconds(),
        "total_messages":      analytics.TotalMessages,
    }
    
    analyticsClient.Track("conversation_completed", metrics)
}
```

## Best Practices

### 1. Intent Design

**Good intent categories:**
- Specific enough to guide schema selection
- General enough to handle variations
- Non-overlapping to avoid classification conflicts
- Business-relevant for your domain

**Example structure:**
```go
intents := map[string][]string{
    "technical_support": {"error", "bug", "broken", "not working"},
    "account_help":     {"account", "profile", "settings", "login"},
    "billing_support":  {"billing", "payment", "invoice", "charge"},
}
```

### 2. Schema Design

**Progressive disclosure:**
- Start with essential information
- Add details in subsequent steps
- Use optional fields for nice-to-have data

**Context awareness:**
- Adapt fields based on conversation history
- Consider user persona and intent
- Optimize for user experience

### 3. Workflow Design

**Step organization:**
- Logical information progression
- Reasonable step size (3-5 fields max)
- Clear completion criteria
- Escape routes for complex cases

### 4. Error Handling

```go
func handleWorkflowError(err error, step int, intent string) *agent.Response {
    switch {
    case errors.Is(err, ErrTooManyRetries):
        return &agent.Response{
            Message: "I'm having trouble collecting this information. Let me connect you with a human agent.",
        }
    case errors.Is(err, ErrInvalidInput):
        return &agent.Response{
            Message: "I didn't quite understand that. Could you please rephrase?",
        }
    default:
        return &agent.Response{
            Message: "Something went wrong. Let's start over or contact our support team.",
        }
    }
}
```

## Related Examples

- **[Simple Schema](../simple-schema/)**: Foundation concepts for schema-based collection
- **[Customer Support](../customer-support/)**: Real-world application with specialized schemas
- **[Basic Chat](../basic-chat/)**: Core conversation concepts
- **[Multi-Tool Agent](../multi-tool-agent/)**: Adding external capabilities

## Next Steps

1. **Experiment with Intent Classification**: Try different keyword sets and scoring algorithms
2. **Design Custom Workflows**: Create multi-step flows for your specific use cases
3. **Implement ML Classification**: Upgrade from keyword-based to ML-based intent detection
4. **Add Analytics**: Track workflow performance and user behavior
5. **Scale for Production**: Implement caching, persistence, and monitoring

## Troubleshooting

**Issue**: Poor intent classification accuracy
**Solution**: Review and expand keyword lists, consider ML-based classification

**Issue**: Workflows getting stuck
**Solution**: Implement timeout handling and escalation paths

**Issue**: Schema selection not matching user needs
**Solution**: Add context awareness and conversation history analysis

**Issue**: Performance issues with complex workflows
**Solution**: Implement caching and optimize schema/workflow creation

For comprehensive guidance, see the [schema collection documentation](../../docs/schema-collection.md) and [examples overview](../README.md).