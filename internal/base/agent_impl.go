package base

import (
	"context"
	"fmt"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
)

// agentImpl is the concrete implementation of the Agent interface
type agentImpl struct {
	// Configuration
	name         string
	description  string
	instructions string
	model        string
	modelSettings *agent.ModelSettings
	tools        []agent.Tool
	outputType   agent.OutputType
	flowRules    []agent.FlowRule
	
	// Runtime dependencies
	chatModel    agent.ChatModel
	sessionStore agent.SessionStore
	maxTurns     int
	toolTimeout  time.Duration
	debugLogging bool
}

// NewAgent creates a new Agent implementation from the given configuration
func NewAgent(config *AgentConfig) agent.Agent {
	return &agentImpl{
		name:         config.Name,
		description:  config.Description,
		instructions: config.Instructions,
		model:        config.Model,
		modelSettings: config.ModelSettings,
		tools:        config.Tools,
		outputType:   config.OutputType,
		flowRules:    config.FlowRules,
		chatModel:    config.ChatModel,
		sessionStore: config.SessionStore,
		maxTurns:     config.MaxTurns,
		toolTimeout:  config.ToolTimeout,
		debugLogging: config.DebugLogging,
	}
}

// AgentConfig holds the configuration for creating an Agent implementation
type AgentConfig struct {
	Name         string
	Description  string
	Instructions string
	Model        string
	ModelSettings *agent.ModelSettings
	Tools        []agent.Tool
	OutputType   agent.OutputType
	FlowRules    []agent.FlowRule
	
	// Runtime dependencies
	ChatModel    agent.ChatModel
	SessionStore agent.SessionStore
	MaxTurns     int
	ToolTimeout  time.Duration
	DebugLogging bool
}

// Configuration interface methods
func (a *agentImpl) Name() string                         { return a.name }
func (a *agentImpl) Description() string                  { return a.description }
func (a *agentImpl) Instructions() string                 { return a.instructions }
func (a *agentImpl) Model() string                        { return a.model }
func (a *agentImpl) ModelSettings() *agent.ModelSettings  { return a.modelSettings }
func (a *agentImpl) Tools() []agent.Tool                  { return a.tools }
func (a *agentImpl) OutputType() agent.OutputType         { return a.outputType }

// Chat executes a conversation turn with the given user input
func (a *agentImpl) Chat(ctx context.Context, sessionID string, userInput string, options ...agent.ChatOption) (*agent.Message, interface{}, error) {
	// Get or create session
	session, err := a.getOrCreateSession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return a.ChatWithSession(ctx, session, userInput, options...)
}

// ChatWithSession executes a conversation turn using an existing session
func (a *agentImpl) ChatWithSession(ctx context.Context, session agent.Session, userInput string, options ...agent.ChatOption) (*agent.Message, interface{}, error) {
	// For now, use simple defaults - options will be applied later
	maxTurns := a.maxTurns
	modelSettings := a.modelSettings
	var additionalTools []agent.Tool
	instructions := a.instructions
	clearHistory := false
	
	// TODO: Apply ChatOptions - this requires exposing the chatOptions type
	// or using a different pattern
	
	// Add user message to session
	if !clearHistory {
		session.AddUserMessage(userInput)
	} else {
		session.Clear()
		session.AddUserMessage(userInput)
	}
	
	// Combine agent tools with additional tools
	allTools := append(a.tools, additionalTools...)
	
	// Execute conversation loop
	var finalMessage *agent.Message
	var structuredOutput interface{}
	
	for turn := 0; turn < maxTurns; turn++ {
		// Prepare messages for LLM
		messages := a.prepareMessages(session, instructions)
		
		// Call LLM
		response, err := a.chatModel.GenerateChatCompletion(
			ctx,
			messages,
			a.model,
			modelSettings,
			allTools,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("LLM call failed: %w", err)
		}
		
		finalMessage = response
		session.AddMessage(*response)
		
		// Check if response contains tool calls
		if len(response.ToolCalls) == 0 {
			// No tool calls, conversation is complete
			break
		}
		
		// Execute tool calls
		toolCallExecuted := false
		for _, toolCall := range response.ToolCalls {
			result, err := a.executeTool(ctx, toolCall, allTools)
			if err != nil {
				// Add error message to session
				session.AddToolMessage(toolCall.ID, toolCall.Function.Name, fmt.Sprintf("Error: %v", err))
			} else {
				// Add successful result to session
				resultStr := fmt.Sprintf("%v", result)
				session.AddToolMessage(toolCall.ID, toolCall.Function.Name, resultStr)
				toolCallExecuted = true
			}
		}
		
		if !toolCallExecuted {
			// All tool calls failed, stop here
			break
		}
	}
	
	// Check for structured output
	if a.outputType != nil && finalMessage != nil {
		parsed, err := a.parseStructuredOutput(finalMessage.Content)
		if err == nil {
			structuredOutput = parsed
		}
	}
	
	// Save session
	if err := a.sessionStore.Save(ctx, session); err != nil {
		// Log but don't fail the request
		if a.debugLogging {
			fmt.Printf("Warning: failed to save session: %v\n", err)
		}
	}
	
	return finalMessage, structuredOutput, nil
}

// Session management methods
func (a *agentImpl) GetSession(ctx context.Context, sessionID string) (agent.Session, error) {
	return a.sessionStore.Load(ctx, sessionID)
}

func (a *agentImpl) CreateSession(sessionID string) agent.Session {
	return agent.NewSession(sessionID)
}

func (a *agentImpl) SaveSession(ctx context.Context, session agent.Session) error {
	return a.sessionStore.Save(ctx, session)
}

func (a *agentImpl) DeleteSession(ctx context.Context, sessionID string) error {
	return a.sessionStore.Delete(ctx, sessionID)
}

func (a *agentImpl) ListSessions(ctx context.Context, filter agent.SessionFilter) ([]string, error) {
	return a.sessionStore.List(ctx, filter)
}

// Helper methods

func (a *agentImpl) getOrCreateSession(ctx context.Context, sessionID string) (agent.Session, error) {
	session, err := a.sessionStore.Load(ctx, sessionID)
	if err == agent.ErrSessionNotFound {
		// Create new session
		session = agent.NewSession(sessionID)
		return session, nil
	}
	return session, err
}

func (a *agentImpl) prepareMessages(session agent.Session, instructions string) []agent.Message {
	messages := []agent.Message{}
	
	// Add system message if we have instructions
	if instructions != "" {
		messages = append(messages, agent.Message{
			Role:      agent.RoleSystem,
			Content:   instructions,
			Timestamp: time.Now(),
		})
	}
	
	// Add session messages
	messages = append(messages, session.Messages()...)
	
	return messages
}

func (a *agentImpl) executeTool(ctx context.Context, toolCall agent.ToolCall, tools []agent.Tool) (interface{}, error) {
	// Find the tool
	var targetTool agent.Tool
	for _, tool := range tools {
		if tool.Name() == toolCall.Function.Name {
			targetTool = tool
			break
		}
	}
	
	if targetTool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}
	
	// Parse arguments
	args, err := parseToolArguments(toolCall.Function.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
	}
	
	// Create timeout context
	toolCtx, cancel := context.WithTimeout(ctx, a.toolTimeout)
	defer cancel()
	
	// Execute tool
	return targetTool.Execute(toolCtx, args)
}

func (a *agentImpl) parseStructuredOutput(content string) (interface{}, error) {
	if a.outputType == nil {
		return nil, fmt.Errorf("no output type defined")
	}
	
	// This will use the OutputType implementation to parse and validate
	instance := a.outputType.NewInstance()
	err := parseJSONString(content, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Validate using the output type
	err = a.outputType.Validate(instance)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	return instance, nil
}

// chatOptions functionality is handled directly in the ChatWithSession method
// to avoid import cycle issues with the agent package

// Utility functions (these would be implemented properly)
func parseToolArguments(argsJSON string) (map[string]interface{}, error) {
	// This would parse JSON string into map[string]interface{}
	// For now, return empty map
	return make(map[string]interface{}), nil
}

func parseJSONString(jsonStr string, target interface{}) error {
	// This would use json.Unmarshal to parse the JSON
	// For now, just return nil
	return nil
}