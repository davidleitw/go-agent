package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// simpleAgent implements the Agent interface
type simpleAgent struct {
	name          string
	description   string
	instructions  string
	model         string
	modelSettings *ModelSettings
	tools         []Tool
	outputType    OutputType
	flowRules     []FlowRule

	// Runtime dependencies
	chatModel    ChatModel
	sessionStore SessionStore
	maxTurns     int
	toolTimeout  time.Duration
	debugLogging bool
}

// newSimpleAgent creates a new simple agent implementation
func newSimpleAgent(config *AgentConfig) Agent {
	return &simpleAgent{
		name:          config.Name,
		description:   config.Description,
		instructions:  config.Instructions,
		model:         config.Model,
		modelSettings: config.ModelSettings,
		tools:         config.Tools,
		outputType:    config.OutputType,
		flowRules:     config.FlowRules,
		chatModel:     config.ChatModel,
		sessionStore:  config.SessionStore,
		maxTurns:      config.MaxTurns,
		toolTimeout:   config.ToolTimeout,
		debugLogging:  config.DebugLogging,
	}
}

// Configuration methods
func (a *simpleAgent) Name() string                  { return a.name }
func (a *simpleAgent) Description() string           { return a.description }
func (a *simpleAgent) Instructions() string          { return a.instructions }
func (a *simpleAgent) Model() string                 { return a.model }
func (a *simpleAgent) ModelSettings() *ModelSettings { return a.modelSettings }
func (a *simpleAgent) Tools() []Tool                 { return a.tools }
func (a *simpleAgent) OutputType() OutputType        { return a.outputType }

// Chat executes a conversation turn with the given session ID and user input
func (a *simpleAgent) Chat(ctx context.Context, sessionID string, userInput string, options ...ChatOption) (*Message, interface{}, error) {
	// Load or create session
	session, err := a.sessionStore.Load(ctx, sessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			session = newSimpleSession(sessionID)
		} else {
			return nil, nil, fmt.Errorf("failed to load session: %w", err)
		}
	}

	return a.ChatWithSession(ctx, session, userInput, options...)
}

// ChatWithSession executes a conversation turn with the given session and user input
func (a *simpleAgent) ChatWithSession(ctx context.Context, session Session, userInput string, options ...ChatOption) (*Message, interface{}, error) {
	// Process chat options
	opts := &chatOptions{}
	for _, option := range options {
		option(opts)
	}

	// Clear history if requested
	if opts.clearHistory {
		session.Clear()
	}

	// Add user message to session
	userMessage := Message{
		Role:      RoleUser,
		Content:   userInput,
		Timestamp: time.Now(),
	}
	session.AddMessage(userMessage)

	// Prepare messages for chat model
	messages := session.Messages()

	// Add system message
	instructions := a.instructions
	if opts.systemMessage != "" {
		instructions = opts.systemMessage
	}

	if instructions != "" {
		systemMessage := Message{
			Role:      RoleSystem,
			Content:   instructions,
			Timestamp: time.Now(),
		}
		// Insert at the beginning
		allMessages := []Message{systemMessage}
		allMessages = append(allMessages, messages...)
		messages = allMessages
	}

	// Use model settings from options or agent default
	modelSettings := a.modelSettings
	if opts.modelSettings != nil {
		modelSettings = opts.modelSettings
	}

	// Combine tools
	tools := a.tools
	if len(opts.additionalTools) > 0 {
		tools = append(tools, opts.additionalTools...)
	}

	// Apply flow rules based on session data and structured output
	var flowData map[string]interface{}
	
	// Try to extract data from the last message if it contains structured output
	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		if lastMessage.Content != "" && a.outputType != nil {
			// Try to parse JSON from last message content
			var parsedData map[string]interface{}
			if err := json.Unmarshal([]byte(lastMessage.Content), &parsedData); err == nil {
				flowData = parsedData
			}
		}
	}
	
	// If no structured data, create empty map
	if flowData == nil {
		flowData = make(map[string]interface{})
	}

	// Evaluate each flow rule
	for _, rule := range a.flowRules {
		shouldTrigger, err := rule.Condition.Evaluate(ctx, session, flowData)
		if err != nil {
			if a.debugLogging {
				fmt.Printf("Warning: flow rule '%s' evaluation failed: %v\n", rule.Name, err)
			}
			continue
		}

		if shouldTrigger {
			// Apply the flow rule
			if rule.Action.NewInstructionsTemplate != "" {
				// For simplicity, we'll just add the new instructions to the system prompt
				// In a more sophisticated implementation, we might replace the instructions
				systemMsg := NewSystemMessage(rule.Action.NewInstructionsTemplate)
				messages = append([]Message{systemMsg}, messages...)
			}

			// Prioritize recommended tools
			if len(rule.Action.RecommendedToolNames) > 0 {
				var recommendedTools []Tool
				var otherTools []Tool
				
				for _, tool := range tools {
					isRecommended := false
					for _, recName := range rule.Action.RecommendedToolNames {
						if tool.Name() == recName {
							isRecommended = true
							break
						}
					}
					
					if isRecommended {
						recommendedTools = append(recommendedTools, tool)
					} else {
						otherTools = append(otherTools, tool)
					}
				}
				
				// Put recommended tools first
				tools = append(recommendedTools, otherTools...)
			}

			if a.debugLogging {
				fmt.Printf("Flow rule '%s' triggered: %s\n", rule.Name, rule.Description)
			}
		}
	}

	// Call chat model
	response, err := a.chatModel.GenerateChatCompletion(ctx, messages, a.model, modelSettings, tools)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate chat completion: %w", err)
	}

	// Add response to session
	session.AddMessage(*response)

	// Handle tool calls if present
	if len(response.ToolCalls) > 0 {
		// Execute each tool call
		for _, toolCall := range response.ToolCalls {
			// Find the matching tool
			var matchingTool Tool
			for _, tool := range tools {
				if tool.Name() == toolCall.Function.Name {
					matchingTool = tool
					break
				}
			}

			if matchingTool == nil {
				// Tool not found, add error message
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: tool '%s' not found", toolCall.Function.Name),
				)
				session.AddMessage(errorMsg)
				continue
			}

			// Parse tool arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: invalid arguments - %v", err),
				)
				session.AddMessage(errorMsg)
				continue
			}

			// Execute the tool
			result, err := matchingTool.Execute(ctx, args)
			if err != nil {
				errorMsg := NewToolMessage(
					toolCall.ID,
					toolCall.Function.Name,
					fmt.Sprintf("Error: %v", err),
				)
				session.AddMessage(errorMsg)
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
				session.AddMessage(errorMsg)
				continue
			}

			// Add successful tool result
			toolMsg := NewToolMessage(
				toolCall.ID,
				toolCall.Function.Name,
				string(resultJSON),
			)
			session.AddMessage(toolMsg)
		}

		// Get updated messages for next API call
		messages = session.Messages()

		// Call chat model again with tool results
		finalResponse, err := a.chatModel.GenerateChatCompletion(ctx, messages, a.model, modelSettings, tools)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate final chat completion: %w", err)
		}

		// Add final response to session
		session.AddMessage(*finalResponse)
		response = finalResponse
	}

	// Save session
	if err := a.sessionStore.Save(ctx, session); err != nil {
		// Log but don't fail
		if a.debugLogging {
			fmt.Printf("Warning: failed to save session: %v\n", err)
		}
	}

	// Handle structured output if needed
	var structuredOutput interface{}
	if a.outputType != nil && response.Content != "" {
		// Try to parse JSON from response
		instance := a.outputType.NewInstance()
		if err := json.Unmarshal([]byte(response.Content), instance); err == nil {
			if a.outputType.Validate(instance) == nil {
				structuredOutput = instance
			}
		}
	}

	return response, structuredOutput, nil
}

// Session management methods
func (a *simpleAgent) GetSession(ctx context.Context, sessionID string) (Session, error) {
	return a.sessionStore.Load(ctx, sessionID)
}

func (a *simpleAgent) CreateSession(sessionID string) Session {
	return newSimpleSession(sessionID)
}

func (a *simpleAgent) SaveSession(ctx context.Context, session Session) error {
	return a.sessionStore.Save(ctx, session)
}

func (a *simpleAgent) DeleteSession(ctx context.Context, sessionID string) error {
	return a.sessionStore.Delete(ctx, sessionID)
}

func (a *simpleAgent) ListSessions(ctx context.Context, filter SessionFilter) ([]string, error) {
	return a.sessionStore.List(ctx, filter)
}

// inMemoryStore implements SessionStore using in-memory storage
type inMemoryStore struct {
	sessions map[string]Session
	mutex    sync.RWMutex
}

// newInMemoryStore creates a new in-memory SessionStore implementation
func newInMemoryStore() SessionStore {
	return &inMemoryStore{
		sessions: make(map[string]Session),
	}
}

// Save persists a session in memory
func (s *inMemoryStore) Save(ctx context.Context, session Session) error {
	if session == nil {
		return ErrInvalidSession
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clone the session to prevent external modification
	s.sessions[session.ID()] = session.Clone()
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

	// Return a clone to prevent external modification
	return session.Clone(), nil
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

	for id, session := range s.sessions {
		if s.matchesFilter(session, filter) {
			sessionIDs = append(sessionIDs, id)
		}
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

// Count returns the total number of sessions
func (s *inMemoryStore) Count(ctx context.Context) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.sessions), nil
}

// Clear removes all sessions from memory
func (s *inMemoryStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessions = make(map[string]Session)
	return nil
}

// GetSessions returns all sessions, optionally filtered
func (s *inMemoryStore) GetSessions(ctx context.Context, filter SessionFilter) ([]Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var sessions []Session

	for _, session := range s.sessions {
		if s.matchesFilter(session, filter) {
			// Return clone to prevent external modification
			sessions = append(sessions, session.Clone())
		}
	}

	return sessions, nil
}

// GetSessionsSince returns sessions created or updated since the given time
func (s *inMemoryStore) GetSessionsSince(ctx context.Context, since time.Time) ([]Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var sessions []Session

	for _, session := range s.sessions {
		if session.CreatedAt().After(since) || session.UpdatedAt().After(since) {
			sessions = append(sessions, session.Clone())
		}
	}

	return sessions, nil
}

// Helper methods

func (s *inMemoryStore) matchesFilter(session Session, filter SessionFilter) bool {
	// Filter by prefix
	if filter.IDPrefix != "" && !strings.HasPrefix(session.ID(), filter.IDPrefix) {
		return false
	}

	// Filter by created after
	if !filter.CreatedAfter.IsZero() && session.CreatedAt().Before(filter.CreatedAfter) {
		return false
	}

	// Filter by created before
	if !filter.CreatedBefore.IsZero() && session.CreatedAt().After(filter.CreatedBefore) {
		return false
	}

	// Filter by updated after
	if !filter.UpdatedAfter.IsZero() && session.UpdatedAt().Before(filter.UpdatedAfter) {
		return false
	}

	// Filter by updated before
	if !filter.UpdatedBefore.IsZero() && session.UpdatedAt().After(filter.UpdatedBefore) {
		return false
	}

	// Filter by minimum message count
	if filter.MinMessages > 0 && len(session.Messages()) < filter.MinMessages {
		return false
	}

	// Filter by maximum message count
	if filter.MaxMessages > 0 && len(session.Messages()) > filter.MaxMessages {
		return false
	}

	return true
}

// simpleOutputType implements OutputType for structured output
type simpleOutputType struct {
	example interface{}
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

func (s *simpleOutputType) Schema() map[string]interface{} {
	// Generate basic JSON schema from struct
	t := reflect.TypeOf(s.example)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})

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

		properties[fieldName] = map[string]interface{}{
			"type": fieldType,
		}
	}

	return schema
}

func (s *simpleOutputType) NewInstance() interface{} {
	t := reflect.TypeOf(s.example)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t).Interface()
}

func (s *simpleOutputType) Validate(data interface{}) error {
	// Basic validation - check if types match
	expectedType := reflect.TypeOf(s.example)
	actualType := reflect.TypeOf(data)

	if expectedType != actualType {
		return fmt.Errorf("type mismatch: expected %v, got %v", expectedType, actualType)
	}

	return nil
}

// simpleSession implements the Session interface
type simpleSession struct {
	id        string
	messages  []Message
	createdAt time.Time
	updatedAt time.Time
}

// newSimpleSession creates a new simple session
func newSimpleSession(id string) Session {
	now := time.Now()
	return &simpleSession{
		id:        id,
		messages:  make([]Message, 0),
		createdAt: now,
		updatedAt: now,
	}
}

// NewSession is the public constructor for creating sessions
func NewSession(id string) Session {
	return newSimpleSession(id)
}

// ID returns the session identifier
func (s *simpleSession) ID() string {
	return s.id
}

// Messages returns all messages in the session
func (s *simpleSession) Messages() []Message {
	return s.messages
}

// AddMessage adds a message to the session
func (s *simpleSession) AddMessage(msg Message) {
	s.messages = append(s.messages, msg)
	s.updatedAt = time.Now()
}

// AddUserMessage adds a user message to the session
func (s *simpleSession) AddUserMessage(content string) {
	msg := Message{
		Role:      RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.AddMessage(msg)
}

// AddAssistantMessage adds an assistant message to the session
func (s *simpleSession) AddAssistantMessage(content string) {
	msg := Message{
		Role:      RoleAssistant,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.AddMessage(msg)
}

// AddSystemMessage adds a system message to the session
func (s *simpleSession) AddSystemMessage(content string) {
	msg := Message{
		Role:      RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.AddMessage(msg)
}

// AddToolMessage adds a tool response message to the session
func (s *simpleSession) AddToolMessage(toolCallID, toolName, content string) {
	msg := Message{
		Role:       RoleTool,
		Content:    content,
		ToolCallID: toolCallID,
		Name:       toolName,
		Timestamp:  time.Now(),
	}
	s.AddMessage(msg)
}

// CreatedAt returns when the session was created
func (s *simpleSession) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns when the session was last updated
func (s *simpleSession) UpdatedAt() time.Time {
	return s.updatedAt
}

// Clear removes all messages from the session
func (s *simpleSession) Clear() {
	s.messages = make([]Message, 0)
	s.updatedAt = time.Now()
}

// Clone creates a copy of the session
func (s *simpleSession) Clone() Session {
	clone := &simpleSession{
		id:        s.id,
		messages:  make([]Message, len(s.messages)),
		createdAt: s.createdAt,
		updatedAt: s.updatedAt,
	}
	copy(clone.messages, s.messages)
	return clone
}
