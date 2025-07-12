package internal

import (
	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/internal/base"
	"github.com/davidleitw/go-agent/internal/llm"
)

func init() {
	// Register implementation functions with the agent package
	agent.SetImplementationFunctions(
		createOpenAIChatModel,
		createOutputTypeFromStruct,
		createAgentImplementation,
	)
	
	// Set session factory
	agent.SetSessionFactory(createSession)
}

// createOpenAIChatModel creates an OpenAI ChatModel implementation
func createOpenAIChatModel(apiKey string, config *agent.OpenAIConfig) (agent.ChatModel, error) {
	llmConfig := &llm.OpenAIConfig{
		BaseURL:      config.BaseURL,
		Organization: config.Organization,
		Timeout:      config.Timeout,
		RetryCount:   config.RetryCount,
	}
	
	return llm.NewOpenAIChatModel(apiKey, llmConfig)
}

// createOutputTypeFromStruct creates an OutputType from a struct example
func createOutputTypeFromStruct(example interface{}) (agent.OutputType, error) {
	// This would use reflection to analyze the struct and generate JSON schema
	// For now, return a mock implementation
	return &mockOutputType{example: example}, nil
}

// createAgentImplementation creates an Agent implementation
func createAgentImplementation(config *agent.AgentConfig) agent.Agent {
	baseConfig := &base.AgentConfig{
		Name:         config.Name,
		Description:  config.Description,
		Instructions: config.Instructions,
		Model:        config.Model,
		ModelSettings: config.ModelSettings,
		Tools:        config.Tools,
		OutputType:   config.OutputType,
		FlowRules:    config.FlowRules,
		ChatModel:    config.ChatModel,
		SessionStore: config.SessionStore,
		MaxTurns:     config.MaxTurns,
		ToolTimeout:  config.ToolTimeout,
		DebugLogging: config.DebugLogging,
	}
	
	return base.NewAgent(baseConfig)
}

// createSession creates a new Session implementation
func createSession(id string) agent.Session {
	return base.NewSession(id)
}

// mockOutputType is a temporary implementation
type mockOutputType struct {
	example interface{}
}

func (m *mockOutputType) Name() string {
	return "MockOutputType"
}

func (m *mockOutputType) Description() string {
	return "A mock output type for testing"
}

func (m *mockOutputType) Schema() map[string]interface{} {
	// Return a basic schema
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"result": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func (m *mockOutputType) NewInstance() interface{} {
	// Return the example for now
	return m.example
}

func (m *mockOutputType) Validate(data interface{}) error {
	// No validation for now
	return nil
}