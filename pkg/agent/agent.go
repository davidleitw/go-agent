package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/davidleitw/go-agent/pkg/schema"
)

// Agent represents an AI agent that can have conversations.
// This is the core interface that users implement to create custom agents.
type Agent interface {
	// Configuration methods for agent metadata
	Name() string
	Description() string

	// Chat executes a conversation turn with the given session and user input.
	// Returns the agent's response message and optional structured output.
	Chat(ctx context.Context, session Session, userInput string) (*Message, any, error)

	// Optional capabilities that agents can implement

	// GetOutputType returns the expected structured output type, if any
	GetOutputType() OutputType

	// GetTools returns the tools available to this agent
	GetTools() []Tool

	// GetFlowRules returns the flow rules for dynamic behavior
	GetFlowRules() []FlowRule
}

// Tool defines an external capability that an agent can use.
type Tool interface {
	Name() string
	Description() string
	Schema() map[string]any
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// BasicAgentConfig holds configuration for creating a basic agent
type BasicAgentConfig struct {
	Name          string
	Description   string
	Instructions  string
	Model         string
	ModelSettings *ModelSettings
	Tools         []Tool
	OutputType    OutputType
	FlowRules     []FlowRule

	// Runtime dependencies
	ChatModel    ChatModel
	SessionStore SessionStore
	MaxTurns     int
	ToolTimeout  time.Duration
}

// NewBasicAgent creates a basic, ready-to-use agent with sensible defaults.
// This is the recommended way to create an agent for most use cases.
func NewBasicAgent(config BasicAgentConfig) (Agent, error) {
	// Set defaults
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.MaxTurns == 0 {
		config.MaxTurns = 5
	}
	if config.ToolTimeout == 0 {
		config.ToolTimeout = 30 * time.Second
	}

	// Validate required fields
	if config.Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if config.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if config.SessionStore == nil {
		config.SessionStore = NewInMemorySessionStore()
	}

	return &basicAgent{
		config: config,
	}, nil
}

// basicAgent implements the Agent interface
type basicAgent struct {
	config BasicAgentConfig
}

// Name returns the agent's name
func (a *basicAgent) Name() string {
	return a.config.Name
}

// Description returns the agent's description
func (a *basicAgent) Description() string {
	return a.config.Description
}

// GetOutputType returns the expected structured output type
func (a *basicAgent) GetOutputType() OutputType {
	return a.config.OutputType
}

// GetTools returns the tools available to this agent
func (a *basicAgent) GetTools() []Tool {
	return a.config.Tools
}

// GetFlowRules returns the flow rules for dynamic behavior
func (a *basicAgent) GetFlowRules() []FlowRule {
	return a.config.FlowRules
}

// Chat implements the Agent interface
func (a *basicAgent) Chat(ctx context.Context, session Session, userInput string) (*Message, any, error) {
	// Add user message to session
	session.AddMessage(RoleUser, userInput)

	// Get tools and flow rules from the agent interface
	tools := a.GetTools()
	flowRules := a.GetFlowRules()
	outputType := a.GetOutputType()

	// Apply flow rules if any (before building final message list)
	var modifiedPrompt string
	var fallbackResponse string
	if len(flowRules) > 0 {
		// Create flow data context using session messages (without system instruction yet)
		sessionMessages := session.Messages()
		flowData := make(map[string]any)
		flowData["userInput"] = userInput
		flowData["messageCount"] = len(sessionMessages)

		// Evaluate and apply all triggered flow rules
		for _, rule := range flowRules {
			shouldTrigger, err := rule.Evaluate(ctx, session, flowData)
			if err != nil {
				// Log but don't fail
				fmt.Printf("Warning: flow rule '%s' evaluation failed: %v\n", rule.Name, err)
				continue
			}

			if shouldTrigger {
				// Apply flow rule actions (these may modify the session)
				result := rule.Action.Apply(ctx, session, flowData)
				if result.Error != nil {
					fmt.Printf("Warning: flow rule '%s' action failed: %v\n", rule.Name, result.Error)
					continue
				}

				// Handle direct response (Ask action)
				if result.ShouldStop && result.DirectResponse != nil {
					// Save session and return the direct response immediately
					if err := a.config.SessionStore.Save(ctx, session); err != nil {
						fmt.Printf("Warning: failed to save session: %v\n", err)
					}
					return result.DirectResponse, nil, nil
				}

				// Handle modified prompt (AskAI action)
				if result.ModifiedPrompt != "" {
					modifiedPrompt = result.ModifiedPrompt
				}
				
				// Handle fallback response for AskAI
				if result.FallbackResponse != "" {
					fallbackResponse = result.FallbackResponse
				}

				// Handle stop execution
				if result.ShouldStop {
					break
				}
			}
		}
	}

	// Build final message list with system instruction (only once, at the end)
	messages := session.Messages()
	
	// Use modified prompt if available, otherwise use default instructions
	instructions := a.config.Instructions
	if modifiedPrompt != "" {
		// For AskAI, create a context-aware prompt
		contextualPrompt := fmt.Sprintf("%s\n\nCurrent situation: %s\nTask: %s", 
			a.config.Instructions, 
			buildContextualPrompt(session, userInput),
			modifiedPrompt)
		instructions = contextualPrompt
	}
	
	if instructions != "" {
		systemMessage := NewSystemMessage(instructions)
		// Insert system message at the beginning
		finalMessages := []Message{systemMessage}
		finalMessages = append(finalMessages, messages...)
		messages = finalMessages
	}

	// Execute conversation loop with tool calls
	maxTurns := a.config.MaxTurns
	for range maxTurns {
		// Call chat model
		response, err := a.config.ChatModel.GenerateChatCompletion(
			ctx,
			messages,
			a.config.Model,
			a.config.ModelSettings,
			tools,
		)
		if err != nil {
			// Use fallback response if available for AskAI failures
			if fallbackResponse != "" {
				session.AddMessage(RoleAssistant, fallbackResponse)
				
				// Save session
				if err := a.config.SessionStore.Save(ctx, session); err != nil {
					fmt.Printf("Warning: failed to save session: %v\n", err)
				}
				
				// Create fallback message to return
				fallbackMsg := &Message{
					Role:      RoleAssistant,
					Content:   fallbackResponse,
					Timestamp: timeNow(),
				}
				return fallbackMsg, nil, nil
			}
			return nil, nil, fmt.Errorf("failed to generate chat completion: %w", err)
		}

		// Add response to session (may contain tool calls)
		if advSession, ok := session.(SessionAdvanced); ok {
			advSession.AddComplexMessage(*response)
		} else {
			session.AddMessage(response.Role, response.Content)
		}

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			// Handle structured output if needed
			var structuredOutput any
			if outputType != nil && response.Content != "" {
				instance := outputType.NewInstance()
				if err := json.Unmarshal([]byte(response.Content), instance); err == nil {
					if outputType.Validate(instance) == nil {
						structuredOutput = instance
					}
				}
			}

			// Save session
			if err := a.config.SessionStore.Save(ctx, session); err != nil {
				// Log but don't fail
				fmt.Printf("Warning: failed to save session: %v\n", err)
			}

			return response, structuredOutput, nil
		}

		// Execute tool calls
		toolCallsHandled := 0
		for _, toolCall := range response.ToolCalls {
			// Find matching tool
			var matchingTool Tool
			for _, tool := range tools {
				if tool.Name() == toolCall.Function.Name {
					matchingTool = tool
					break
				}
			}

			if matchingTool == nil {
				// Tool not found
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: tool '%s' not found", toolCall.Function.Name),
				)
				if advSession, ok := session.(SessionAdvanced); ok {
					advSession.AddComplexMessage(errorMsg)
				} else {
					session.AddMessage(errorMsg.Role, errorMsg.Content)
				}
				toolCallsHandled++
				continue
			}

			// Parse tool arguments
			var args map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: invalid arguments - %v", err),
				)
				if advSession, ok := session.(SessionAdvanced); ok {
					advSession.AddComplexMessage(errorMsg)
				} else {
					session.AddMessage(errorMsg.Role, errorMsg.Content)
				}
				toolCallsHandled++
				continue
			}

			// Create context with timeout for tool execution
			toolCtx, cancel := context.WithTimeout(ctx, a.config.ToolTimeout)

			// Execute the tool
			result, err := matchingTool.Execute(toolCtx, args)
			cancel()

			if err != nil {
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: %v", err),
				)
				if advSession, ok := session.(SessionAdvanced); ok {
					advSession.AddComplexMessage(errorMsg)
				} else {
					session.AddMessage(errorMsg.Role, errorMsg.Content)
				}
				toolCallsHandled++
				continue
			}

			// Convert result to JSON string
			resultJSON, err := json.Marshal(result)
			if err != nil {
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: failed to serialize result - %v", err),
				)
				if advSession, ok := session.(SessionAdvanced); ok {
					advSession.AddComplexMessage(errorMsg)
				} else {
					session.AddMessage(errorMsg.Role, errorMsg.Content)
				}
				toolCallsHandled++
				continue
			}

			// Add successful tool result
			toolMsg := NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				string(resultJSON),
			)
			if advSession, ok := session.(SessionAdvanced); ok {
				advSession.AddComplexMessage(toolMsg)
			} else {
				session.AddMessage(toolMsg.Role, toolMsg.Content)
			}
			toolCallsHandled++
		}

		// If all tool calls were handled, continue to next turn
		if toolCallsHandled == len(response.ToolCalls) {
			// Update messages for next iteration
			messages = session.Messages()
			// Insert system message again if it exists
			if a.config.Instructions != "" {
				systemMessage := NewSystemMessage(a.config.Instructions)
				messages = append([]Message{systemMessage}, messages...)
			}
			continue
		}

		// If we couldn't handle some tool calls, return error
		return nil, nil, fmt.Errorf("failed to handle %d tool calls", len(response.ToolCalls)-toolCallsHandled)
	}

	// If we've exhausted max turns, return the last response
	messages = session.Messages()
	if len(messages) > 0 {
		lastMessage := &messages[len(messages)-1]
		if lastMessage.Role == RoleAssistant {
			return lastMessage, nil, fmt.Errorf("reached maximum turns (%d) without completion", maxTurns)
		}
	}

	return nil, nil, fmt.Errorf("conversation ended unexpectedly after %d turns", maxTurns)
}

// NewInMemorySessionStore creates a new in-memory SessionStore implementation
func NewInMemorySessionStore() SessionStore {
	return &inMemoryStore{
		sessions: make(map[string]Session),
	}
}

// inMemoryStore implements SessionStore using in-memory storage
type inMemoryStore struct {
	sessions map[string]Session
	mutex    sync.RWMutex
}

// Save persists a session in memory
func (s *inMemoryStore) Save(ctx context.Context, session Session) error {
	if session == nil {
		return ErrInvalidSession
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clone the session to prevent external modification if possible
	if cloneable, ok := session.(interface{ Clone() Session }); ok {
		s.sessions[session.ID()] = cloneable.Clone()
	} else {
		s.sessions[session.ID()] = session
	}
	return nil
}

// Load retrieves a session from memory by ID
func (s *inMemoryStore) Load(ctx context.Context, sessionID string) (Session, error) {
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	// Return a clone to prevent external modification if possible
	if cloneable, ok := session.(interface{ Clone() Session }); ok {
		return cloneable.Clone(), nil
	} else {
		return session, nil
	}
}

// Delete removes a session from memory
func (s *inMemoryStore) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSessionID
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return ErrSessionNotFound
	}

	delete(s.sessions, sessionID)
	return nil
}

// List returns all session IDs, optionally filtered
func (s *inMemoryStore) List(ctx context.Context, filter SessionFilter) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var sessionIDs []string
	for id := range s.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	return sessionIDs, nil
}

// Exists checks if a session exists
func (s *inMemoryStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, ErrInvalidSessionID
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.sessions[sessionID]
	return exists, nil
}

// Session creation is now handled in session.go

// NewStructuredOutputType creates an OutputType from a struct example
func NewStructuredOutputType(example any) OutputType {
	return &simpleOutputType{example: example}
}


// simpleOutputType implements OutputType for structured output
type simpleOutputType struct {
	example any
}

func (s *simpleOutputType) Name() string {
	t := reflect.TypeOf(s.example)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (s *simpleOutputType) Description() string {
	return fmt.Sprintf("Structured output type for %s", s.Name())
}

func (s *simpleOutputType) Schema() map[string]any {
	// Generate basic JSON schema from struct
	t := reflect.TypeOf(s.example)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := map[string]any{
		"type":       "object",
		"properties": make(map[string]any),
	}

	properties := schema["properties"].(map[string]any)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse json tag
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]

		// Determine field type
		fieldType := "string" // Default
		switch field.Type.Kind() {
		case reflect.String:
			fieldType = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldType = "integer"
		case reflect.Float32, reflect.Float64:
			fieldType = "number"
		case reflect.Bool:
			fieldType = "boolean"
		case reflect.Slice, reflect.Array:
			fieldType = "array"
		case reflect.Map, reflect.Struct:
			fieldType = "object"
		}

		properties[fieldName] = map[string]any{
			"type": fieldType,
		}
	}

	return schema
}

func (s *simpleOutputType) NewInstance() any {
	t := reflect.TypeOf(s.example)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t).Interface()
}

func (s *simpleOutputType) Validate(data any) error {
	// Basic validation - check if types match
	expectedType := reflect.TypeOf(s.example)
	actualType := reflect.TypeOf(data)

	if expectedType != actualType {
		return fmt.Errorf("type mismatch: expected %v, got %v", expectedType, actualType)
	}

	return nil
}

// =====================================================
// Simplified Agent API (formerly in pkg/goagent)
// =====================================================

// SimpleAgent represents a simplified AI agent with elegant API
type SimpleAgent interface {
	// Chat performs a conversation turn with the agent
	Chat(ctx context.Context, input string, opts ...ChatOption) (*Response, error)
}

// Response contains the agent's response and metadata
type Response struct {
	Message  string                 // The text response from the agent
	Data     any                    // Structured output if configured
	Session  Session                // Conversation session
	Metadata map[string]interface{} // Additional metadata
}

// ChatOption allows customizing individual chat requests
type ChatOption func(*chatConfig)

type chatConfig struct {
	sessionID     string
	variables     map[string]interface{}
	schemaFields  []*schema.Field // Fields to collect from user input
}

// WithSession specifies a session ID for the conversation
func WithSession(sessionID string) ChatOption {
	return func(c *chatConfig) {
		c.sessionID = sessionID
	}
}

// WithVariables provides variables for template substitution
func WithVariables(vars map[string]interface{}) ChatOption {
	return func(c *chatConfig) {
		c.variables = vars
	}
}

// WithSchema configures the agent to collect specific structured information from user input.
// The agent will intelligently extract and collect the specified fields through natural conversation.
//
// When fields are missing from user input, the agent will use the provided prompts to
// ask for the information in a contextual and natural way.
//
// Example:
//   agent.WithSchema(
//       schema.Define("email", "Please provide your email address"),
//       schema.Define("issue", "Please describe your issue"),
//       schema.Define("phone", "Contact number (optional)").Optional(),
//   )
func WithSchema(fields ...*schema.Field) ChatOption {
	return func(c *chatConfig) {
		c.schemaFields = fields
	}
}

// simpleAgentImpl is the internal implementation of the simplified Agent
type simpleAgentImpl struct {
	name         string
	coreAgent    Agent
	sessionStore SessionStore
	options      *agentOptions
}

// agentOptions holds all configuration for an agent
type agentOptions struct {
	// Basic configuration
	description  string
	instructions string
	model        string
	modelSettings *ModelSettings
	
	// Provider configuration
	provider     string
	apiKey       string
	customModel  ChatModel
	
	// Tools and capabilities
	tools      []Tool
	outputType OutputType
	
	// Flow rules and conditions
	flowRules []FlowRule
	
	// Runtime configuration
	sessionStore SessionStore
	maxTurns     int
	toolTimeout  time.Duration
}

// Chat implements the simplified chat interface
func (a *simpleAgentImpl) Chat(ctx context.Context, input string, opts ...ChatOption) (*Response, error) {
	// Apply chat options
	cfg := &chatConfig{
		sessionID: fmt.Sprintf("session-%d", timeNow().Unix()),
		variables: make(map[string]interface{}),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Load or create session
	session, err := a.sessionStore.Load(ctx, cfg.sessionID)
	if err != nil {
		// Create new session if not found
		session = NewSession(cfg.sessionID)
	}

	// Handle schema-based information collection if configured
	if len(cfg.schemaFields) > 0 {
		// Check if we need to collect any information
		collectionResponse, shouldReturn, err := a.handleSchemaCollection(ctx, session, input, cfg.schemaFields)
		if err != nil {
			return nil, fmt.Errorf("schema collection failed: %w", err)
		}
		if shouldReturn {
			// Save session after schema collection
			if err := a.sessionStore.Save(ctx, session); err != nil {
				// Log but don't fail on save error
			}
			return collectionResponse, nil
		}
	}

	// Execute the agent directly
	message, data, err := a.coreAgent.Chat(ctx, session, input)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Save session after chat
	if err := a.sessionStore.Save(ctx, session); err != nil {
		// Log but don't fail on save error
		// Could add proper logging here
	}

	// Build response
	response := &Response{
		Message:  message.Content,
		Data:     data,
		Session:  session,
		Metadata: make(map[string]interface{}),
	}

	// Add execution metadata
	response.Metadata["model"] = a.options.model
	response.Metadata["tools_used"] = len(message.ToolCalls)
	response.Metadata["timestamp"] = timeNow()

	return response, nil
}

// timeNow is a variable to allow mocking in tests
var timeNow = time.Now

// buildContextualPrompt creates a context summary for AskAI prompts
func buildContextualPrompt(session Session, currentInput string) string {
	messages := session.Messages()
	messageCount := len(messages)
	
	// Basic context
	context := fmt.Sprintf("User has sent %d messages in this conversation", messageCount)
	
	// Add recent message context (last 2 messages for brevity)
	if messageCount > 0 {
		recentMessages := messages
		if messageCount > 2 {
			recentMessages = messages[messageCount-2:]
		}
		
		context += ". Recent messages: "
		for _, msg := range recentMessages {
			if msg.Role == RoleUser || msg.Role == RoleAssistant {
				context += fmt.Sprintf("[%s: %s] ", msg.Role, msg.Content)
			}
		}
	}
	
	// Add current input
	context += fmt.Sprintf("Current user input: '%s'", currentInput)
	
	return context
}

// =====================================================
// Builder API for Simplified Agent Creation
// =====================================================

// Builder provides a fluent interface for creating agents
type Builder struct {
	name    string
	options *agentOptions
	err     error
}

// New creates a new agent builder with the given name
func New(name string) *Builder {
	return &Builder{
		name: name,
		options: &agentOptions{
			model:        "gpt-4o-mini",
			maxTurns:     10,
			toolTimeout:  30 * time.Second,
			provider:     "openai",
			modelSettings: &ModelSettings{
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(1000),
			},
		},
	}
}

// WithDescription sets the agent's description
func (b *Builder) WithDescription(description string) *Builder {
	if b.err != nil {
		return b
	}
	b.options.description = description
	return b
}

// WithInstructions sets the agent's system instructions
func (b *Builder) WithInstructions(instructions string) *Builder {
	if b.err != nil {
		return b
	}
	b.options.instructions = instructions
	return b
}

// WithModel sets the model to use (e.g., "gpt-4", "gpt-3.5-turbo")
func (b *Builder) WithModel(model string) *Builder {
	if b.err != nil {
		return b
	}
	b.options.model = model
	return b
}

// WithTemperature sets the model temperature
func (b *Builder) WithTemperature(temp float64) *Builder {
	if b.err != nil {
		return b
	}
	b.options.modelSettings.Temperature = &temp
	return b
}

// WithMaxTokens sets the maximum tokens for responses
func (b *Builder) WithMaxTokens(tokens int) *Builder {
	if b.err != nil {
		return b
	}
	b.options.modelSettings.MaxTokens = &tokens
	return b
}

// WithOpenAI configures the agent to use OpenAI
func (b *Builder) WithOpenAI(apiKey string) *Builder {
	if b.err != nil {
		return b
	}
	b.options.provider = "openai"
	b.options.apiKey = apiKey
	return b
}

// WithChatModel sets a custom chat model implementation
func (b *Builder) WithChatModel(model ChatModel) *Builder {
	if b.err != nil {
		return b
	}
	b.options.customModel = model
	b.options.provider = "custom"
	return b
}

// WithTool adds a tool to the agent
func (b *Builder) WithTool(tool Tool) *Builder {
	if b.err != nil {
		return b
	}
	b.options.tools = append(b.options.tools, tool)
	return b
}

// WithTools adds multiple tools to the agent
func (b *Builder) WithTools(tools ...Tool) *Builder {
	if b.err != nil {
		return b
	}
	b.options.tools = append(b.options.tools, tools...)
	return b
}

// WithOutputType sets the structured output type
func (b *Builder) WithOutputType(outputType OutputType) *Builder {
	if b.err != nil {
		return b
	}
	b.options.outputType = outputType
	return b
}

// WithSessionStore sets a custom session store
func (b *Builder) WithSessionStore(store SessionStore) *Builder {
	if b.err != nil {
		return b
	}
	b.options.sessionStore = store
	return b
}

// WithMaxTurns sets the maximum conversation turns
func (b *Builder) WithMaxTurns(turns int) *Builder {
	if b.err != nil {
		return b
	}
	b.options.maxTurns = turns
	return b
}

// WithToolTimeout sets the timeout for tool execution
func (b *Builder) WithToolTimeout(timeout time.Duration) *Builder {
	if b.err != nil {
		return b
	}
	b.options.toolTimeout = timeout
	return b
}

// When adds a condition-based flow rule
func (b *Builder) When(condition interface{}) *AgentFlowRuleBuilder {
	return &AgentFlowRuleBuilder{
		agentBuilder: b,
		condition:    condition,
	}
}

// OnMissingInfo is a shorthand for common missing field conditions
func (b *Builder) OnMissingInfo(fields ...string) *AgentFlowRuleBuilder {
	return b.When(WhenMissingFields(fields...))
}

// OnMessageCount is a shorthand for message count conditions
func (b *Builder) OnMessageCount(count int) *AgentFlowRuleBuilder {
	return b.When(WhenMessageCount(count))
}

// Build creates the agent with all configurations
func (b *Builder) Build() (SimpleAgent, error) {
	if b.err != nil {
		return nil, b.err
	}

	// Validate required fields
	if b.name == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	// Set default instructions if not provided
	if b.options.instructions == "" {
		b.options.instructions = fmt.Sprintf("You are %s, a helpful AI assistant.", b.name)
		if b.options.description != "" {
			b.options.instructions += " " + b.options.description
		}
	}

	// Create chat model based on provider
	var chatModel ChatModel
	var err error

	switch b.options.provider {
	case "openai":
		if b.options.apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required")
		}
		chatModel, err = b.createOpenAIChatModel()
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
		}
	case "custom":
		if b.options.customModel == nil {
			return nil, fmt.Errorf("custom chat model is required")
		}
		chatModel = b.options.customModel
	default:
		return nil, fmt.Errorf("unsupported provider: %s", b.options.provider)
	}

	// Create session store if not provided
	if b.options.sessionStore == nil {
		b.options.sessionStore = NewInMemorySessionStore()
	}

	// Create the core agent directly
	coreAgent, err := NewBasicAgent(BasicAgentConfig{
		Name:          b.name,
		Description:   b.options.description,
		Instructions:  b.options.instructions,
		Model:         b.options.model,
		ModelSettings: b.options.modelSettings,
		Tools:         b.options.tools,
		OutputType:    b.options.outputType,
		FlowRules:     b.options.flowRules,
		ChatModel:     chatModel,
		SessionStore:  b.options.sessionStore,
		MaxTurns:      b.options.maxTurns,
		ToolTimeout:   b.options.toolTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create core agent: %w", err)
	}

	// Return the simplified agent
	return &simpleAgentImpl{
		name:         b.name,
		coreAgent:    coreAgent,
		sessionStore: b.options.sessionStore,
		options:      b.options,
	}, nil
}

// AgentFlowRuleBuilder provides fluent interface for creating flow rules for agents
type AgentFlowRuleBuilder struct {
	agentBuilder *Builder
	condition    interface{}
	actions      []flowAction
}

type flowAction struct {
	actionType string
	data       interface{}
}

// Ask sets the agent to ask for specific information
func (f *AgentFlowRuleBuilder) Ask(message string) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "ask",
		data:       message,
	})
	return f
}

// AskAI sets the agent to use LLM to generate a contextual response
func (f *AgentFlowRuleBuilder) AskAI(instruction string) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "ask_ai",
		data:       instruction,
	})
	return f
}

// OrElse sets a fallback message if AskAI fails
func (f *AgentFlowRuleBuilder) OrElse(fallbackMessage string) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "fallback",
		data:       fallbackMessage,
	})
	return f
}

// UseTemplate applies a new instruction template
func (f *AgentFlowRuleBuilder) UseTemplate(template string) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "template",
		data:       template,
	})
	return f
}

// EnableTools recommends specific tools for the next turn
func (f *AgentFlowRuleBuilder) EnableTools(toolNames ...string) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "tools",
		data:       toolNames,
	})
	return f
}

// Summarize requests a summary of the conversation
func (f *AgentFlowRuleBuilder) Summarize() *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "summarize",
		data:       "Please provide a summary of our conversation so far.",
	})
	return f
}

// Escalate marks the conversation for escalation
func (f *AgentFlowRuleBuilder) Escalate(reason string) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "escalate",
		data:       reason,
	})
	return f
}

// OutputJSON sets the expected structured output type
func (f *AgentFlowRuleBuilder) OutputJSON(outputType OutputType) *AgentFlowRuleBuilder {
	f.actions = append(f.actions, flowAction{
		actionType: "output",
		data:       outputType,
	})
	return f
}

// And allows chaining with additional conditions
func (f *AgentFlowRuleBuilder) And(condition interface{}) *AgentFlowRuleBuilder {
	// Convert existing condition to Condition if needed
	existingCond := f.convertCondition(f.condition)
	newCond := f.convertCondition(condition)
	
	f.condition = And(existingCond, newCond)
	return f
}

// Or allows chaining with alternative conditions
func (f *AgentFlowRuleBuilder) Or(condition interface{}) *AgentFlowRuleBuilder {
	// Convert existing condition to Condition if needed
	existingCond := f.convertCondition(f.condition)
	newCond := f.convertCondition(condition)
	
	f.condition = Or(existingCond, newCond)
	return f
}

// Build adds the flow rule to the agent and returns the builder
func (f *AgentFlowRuleBuilder) Build() *Builder {
	if f.agentBuilder.err != nil {
		return f.agentBuilder
	}

	// Convert condition to Condition
	condition := f.convertCondition(f.condition)
	if condition == nil {
		f.agentBuilder.err = fmt.Errorf("invalid condition provided")
		return f.agentBuilder
	}

	// Create flow action from accumulated actions
	var flowAction FlowAction
	
	// Apply actions in order
	for _, action := range f.actions {
		switch action.actionType {
		case "ask":
			if msg, ok := action.data.(string); ok {
				flowAction.DirectResponse = msg
			}
		case "ask_ai":
			if instruction, ok := action.data.(string); ok {
				flowAction.AIPrompt = instruction
			}
		case "fallback":
			if fallback, ok := action.data.(string); ok {
				flowAction.FallbackResponse = fallback
			}
		case "template":
			if template, ok := action.data.(string); ok {
				flowAction.NewInstructionsTemplate = template
			}
		case "tools":
			if tools, ok := action.data.([]string); ok {
				flowAction.RecommendedToolNames = tools
			}
		case "summarize":
			flowAction.NewInstructionsTemplate = action.data.(string)
		case "escalate":
			flowAction.TriggerNotification = true
			flowAction.NotificationDetails = map[string]interface{}{
				"reason": action.data,
				"type":   "escalation",
			}
		case "output":
			// This would need to be handled at agent level
			f.agentBuilder.options.outputType = action.data.(OutputType)
		}
	}

	// Create the flow rule
	ruleName := fmt.Sprintf("rule_%s", condition.Name())
	rule := NewFlowRule(ruleName, condition).
		WithNewInstructions(flowAction.NewInstructionsTemplate).
		WithRecommendedTools(flowAction.RecommendedToolNames...).
		Build()

	// Add to agent's flow rules
	f.agentBuilder.options.flowRules = append(f.agentBuilder.options.flowRules, rule)
	
	return f.agentBuilder
}

// convertCondition converts various condition formats to Condition
func (f *AgentFlowRuleBuilder) convertCondition(condition interface{}) Condition {
	switch c := condition.(type) {
	case Condition:
		return c
	case string:
		// Parse string conditions like "email missing", "phone missing", etc.
		return f.parseStringCondition(c)
	case func(Session) bool:
		return WhenFunc("custom", c)
	default:
		return nil
	}
}

// parseStringCondition parses natural language condition strings
func (f *AgentFlowRuleBuilder) parseStringCondition(condition string) Condition {
	// Simple parsing for common patterns
	switch {
	case strings.Contains(condition, "missing"):
		// Extract field name from "field missing" pattern
		field := extractFieldName(condition)
		if field != "" {
			return WhenMissingFields(field)
		}
	case strings.Contains(condition, "contains"):
		// Extract text from "contains text" pattern
		text := extractContainsText(condition)
		if text != "" {
			return WhenContains(text)
		}
	}
	
	// Fallback: create a simple contains condition
	return WhenContains(condition)
}

// Helper functions for string parsing
func extractFieldName(condition string) string {
	// Simple extraction: get word before "missing"
	parts := strings.Fields(condition)
	for i, part := range parts {
		if part == "missing" && i > 0 {
			return parts[i-1]
		}
	}
	return ""
}

func extractContainsText(condition string) string {
	// Simple extraction: get text after "contains"
	parts := strings.Fields(condition)
	for i, part := range parts {
		if part == "contains" && i < len(parts)-1 {
			return parts[i+1]
		}
	}
	return ""
}

// handleSchemaCollection processes schema-based information collection.
// Returns (response, shouldReturn, error) where shouldReturn indicates
// whether the caller should return immediately with the response.
func (a *simpleAgentImpl) handleSchemaCollection(ctx context.Context, session Session, input string, fields []*schema.Field) (*Response, bool, error) {
	// Add user input to session for context
	session.AddMessage(RoleUser, input)
	
	// Analyze conversation to extract available information
	extractedData, err := a.extractInformationFromConversation(ctx, session, fields)
	if err != nil {
		return nil, false, fmt.Errorf("failed to extract information: %w", err)
	}
	
	// Find missing required fields
	missingFields := a.findMissingRequiredFields(extractedData, fields)
	
	if len(missingFields) == 0 {
		// All required information collected, continue with normal flow
		return nil, false, nil
	}
	
	// Generate response asking for missing information
	response, err := a.generateCollectionResponse(ctx, session, missingFields)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate collection response: %w", err)
	}
	
	// Add assistant response to session
	session.AddMessage(RoleAssistant, response.Message)
	
	return response, true, nil
}

// extractInformationFromConversation uses LLM to extract structured information from conversation
func (a *simpleAgentImpl) extractInformationFromConversation(ctx context.Context, session Session, fields []*schema.Field) (map[string]interface{}, error) {
	// Build prompt for information extraction
	prompt := a.buildExtractionPrompt(session, fields)
	
	// Get the chat model from the core agent
	var chatModel ChatModel
	if basicAgent, ok := a.coreAgent.(*basicAgent); ok {
		chatModel = basicAgent.config.ChatModel
	} else {
		return nil, fmt.Errorf("unable to get chat model from core agent")
	}

	// Create a temporary basic agent for extraction (without schema to avoid recursion)
	extractorAgent, err := NewBasicAgent(BasicAgentConfig{
		Name:         "extractor",
		Instructions: prompt,
		Model:        a.options.model,
		ChatModel:    chatModel,
		SessionStore: NewInMemorySessionStore(),
		MaxTurns:     1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor agent: %w", err)
	}
	
	// Create temporary session for extraction
	extractSession := NewSession("extract-temp")
	
	// Execute extraction
	response, _, err := extractorAgent.Chat(ctx, extractSession, "Extract information from the conversation.")
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}
	
	// Parse the response as JSON
	var extractedData map[string]interface{}
	if err := json.Unmarshal([]byte(response.Content), &extractedData); err != nil {
		// If JSON parsing fails, return empty map (no information extracted)
		return make(map[string]interface{}), nil
	}
	
	return extractedData, nil
}

// buildExtractionPrompt creates a prompt for information extraction
func (a *simpleAgentImpl) buildExtractionPrompt(session Session, fields []*schema.Field) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an information extraction assistant. Analyze the conversation below and extract any available information that matches the expected fields.\n\n")
	
	prompt.WriteString("Expected fields:\n")
	for _, field := range fields {
		required := "required"
		if !field.Required() {
			required = "optional"
		}
		prompt.WriteString(fmt.Sprintf("- %s: %s (%s)\n", field.Name(), field.Prompt(), required))
	}
	
	prompt.WriteString("\nConversation history:\n")
	messages := session.Messages()
	for _, msg := range messages {
		if msg.Role == RoleUser || msg.Role == RoleAssistant {
			prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}
	
	prompt.WriteString("\nInstructions:\n")
	prompt.WriteString("1. Extract information from the conversation that matches the expected fields\n")
	prompt.WriteString("2. Return ONLY a JSON object with the extracted information\n")
	prompt.WriteString("3. Use null for fields where no information was found\n")
	prompt.WriteString("4. Be conservative - only extract information that is clearly stated\n\n")
	
	prompt.WriteString("Example response format:\n")
	prompt.WriteString("{\n")
	for i, field := range fields {
		comma := ","
		if i == len(fields)-1 {
			comma = ""
		}
		prompt.WriteString(fmt.Sprintf("  \"%s\": \"extracted_value_or_null\"%s\n", field.Name(), comma))
	}
	prompt.WriteString("}\n")
	
	return prompt.String()
}

// findMissingRequiredFields identifies which required fields are still missing
func (a *simpleAgentImpl) findMissingRequiredFields(extractedData map[string]interface{}, fields []*schema.Field) []*schema.Field {
	var missing []*schema.Field
	
	for _, field := range fields {
		if !field.Required() {
			continue // Skip optional fields
		}
		
		value, exists := extractedData[field.Name()]
		if !exists || value == nil || value == "" {
			missing = append(missing, field)
		}
	}
	
	return missing
}

// generateCollectionResponse creates a response asking for missing information
func (a *simpleAgentImpl) generateCollectionResponse(ctx context.Context, session Session, missingFields []*schema.Field) (*Response, error) {
	// Build prompt for generating collection response
	prompt := a.buildCollectionPrompt(session, missingFields)
	
	// Get the chat model from the core agent
	var chatModel ChatModel
	if basicAgent, ok := a.coreAgent.(*basicAgent); ok {
		chatModel = basicAgent.config.ChatModel
	} else {
		return nil, fmt.Errorf("unable to get chat model from core agent")
	}

	// Create a temporary agent for response generation
	responseAgent, err := NewBasicAgent(BasicAgentConfig{
		Name:         "response_generator",
		Instructions: prompt,
		Model:        a.options.model,
		ChatModel:    chatModel,
		SessionStore: NewInMemorySessionStore(),
		MaxTurns:     1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create response agent: %w", err)
	}
	
	// Create temporary session for response generation
	responseSession := NewSession("response-temp")
	
	// Generate response
	message, _, err := responseAgent.Chat(ctx, responseSession, "Generate a natural response asking for the missing information.")
	if err != nil {
		return nil, fmt.Errorf("response generation failed: %w", err)
	}
	
	// Build response object
	response := &Response{
		Message:  message.Content,
		Data:     nil,
		Session:  session,
		Metadata: map[string]interface{}{
			"schema_collection": true,
			"missing_fields":    getMissingFieldNames(missingFields),
			"timestamp":         timeNow(),
		},
	}
	
	return response, nil
}

// buildCollectionPrompt creates a prompt for generating information collection responses
func (a *simpleAgentImpl) buildCollectionPrompt(session Session, missingFields []*schema.Field) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are a helpful assistant that naturally asks for missing information. Based on the conversation context, generate a friendly response that asks for the missing required information.\n\n")
	
	prompt.WriteString("Missing required information:\n")
	for _, field := range missingFields {
		prompt.WriteString(fmt.Sprintf("- %s: %s\n", field.Name(), field.Prompt()))
	}
	
	prompt.WriteString("\nConversation context:\n")
	messages := session.Messages()
	for _, msg := range messages {
		if msg.Role == RoleUser || msg.Role == RoleAssistant {
			prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}
	
	prompt.WriteString("\nInstructions:\n")
	prompt.WriteString("1. Acknowledge what information you received from the user\n")
	prompt.WriteString("2. Naturally ask for the missing information using the provided prompts\n")
	prompt.WriteString("3. Be friendly and contextual\n")
	prompt.WriteString("4. Combine multiple questions when appropriate\n")
	prompt.WriteString("5. Keep the conversation flowing naturally\n")
	prompt.WriteString("6. Return only the response text, no additional formatting\n")
	
	return prompt.String()
}

// getMissingFieldNames extracts field names from missing fields for metadata
func getMissingFieldNames(fields []*schema.Field) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.Name()
	}
	return names
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }
