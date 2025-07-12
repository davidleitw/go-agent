package agent

import (
	"context"
	"testing"
	"time"
)

// Mock implementations for testing

type mockAgent struct {
	name          string
	description   string
	instructions  string
	model         string
	modelSettings *ModelSettings
	tools         []Tool
	outputType    OutputType
}

// Configuration methods
func (m *mockAgent) Name() string                  { return m.name }
func (m *mockAgent) Description() string           { return m.description }
func (m *mockAgent) Instructions() string          { return m.instructions }
func (m *mockAgent) Model() string                 { return m.model }
func (m *mockAgent) ModelSettings() *ModelSettings { return m.modelSettings }
func (m *mockAgent) Tools() []Tool                 { return m.tools }
func (m *mockAgent) OutputType() OutputType        { return m.outputType }

// Execution methods - mock implementations
func (m *mockAgent) Chat(ctx context.Context, sessionID string, userInput string, options ...ChatOption) (*Message, interface{}, error) {
	return &Message{
		Role:      RoleAssistant,
		Content:   "Mock response",
		Timestamp: time.Now(),
	}, nil, nil
}

func (m *mockAgent) ChatWithSession(ctx context.Context, session Session, userInput string, options ...ChatOption) (*Message, interface{}, error) {
	return &Message{
		Role:      RoleAssistant,
		Content:   "Mock response",
		Timestamp: time.Now(),
	}, nil, nil
}

// Session management methods - mock implementations
func (m *mockAgent) GetSession(ctx context.Context, sessionID string) (Session, error) {
	return NewSession(sessionID), nil
}

func (m *mockAgent) CreateSession(sessionID string) Session {
	return NewSession(sessionID)
}

func (m *mockAgent) SaveSession(ctx context.Context, session Session) error {
	return nil
}

func (m *mockAgent) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockAgent) ListSessions(ctx context.Context, filter SessionFilter) ([]string, error) {
	return []string{}, nil
}

type mockTool struct {
	name        string
	description string
	schema      map[string]interface{}
	executeFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

func (m *mockTool) Name() string                   { return m.name }
func (m *mockTool) Description() string            { return m.description }
func (m *mockTool) Schema() map[string]interface{} { return m.schema }
func (m *mockTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, args)
	}
	return nil, nil
}

type mockSessionStore struct{}

func (m *mockSessionStore) Save(ctx context.Context, session Session) error {
	return nil
}

func (m *mockSessionStore) Load(ctx context.Context, sessionID string) (Session, error) {
	return NewSession(sessionID), nil
}

func (m *mockSessionStore) Delete(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockSessionStore) List(ctx context.Context, filter SessionFilter) ([]string, error) {
	return []string{}, nil
}

func (m *mockSessionStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	return true, nil
}

func TestAgentCreation(t *testing.T) {
	// Mock implementations for the registry
	originalChatModelImpl := newOpenAIChatModelImpl
	originalAgentImpl := createAgentImplementation
	originalSessionFactory := sessionFactory
	
	defer func() {
		newOpenAIChatModelImpl = originalChatModelImpl
		createAgentImplementation = originalAgentImpl
		sessionFactory = originalSessionFactory
	}()
	
	// Set up mock implementations
	newOpenAIChatModelImpl = func(apiKey string, config *OpenAIConfig) (ChatModel, error) {
		return NewMockChatModel(), nil
	}
	
	createAgentImplementation = func(config *AgentConfig) Agent {
		return &mockAgent{
			name:          config.Name,
			description:   config.Description,
			instructions:  config.Instructions,
			model:         config.Model,
			modelSettings: config.ModelSettings,
			tools:         config.Tools,
			outputType:    config.OutputType,
		}
	}
	
	sessionFactory = func(id string) Session {
		return &basicSession{id: id}
	}

	tests := []struct {
		name     string
		build    func() (Agent, error)
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, agent Agent)
	}{
		{
			name: "basic agent",
			build: func() (Agent, error) {
				return New(
					WithName("test-agent"),
					WithOpenAI("test-key"),
					WithSessionStore(&mockSessionStore{}),
				)
			},
			wantErr: false,
			validate: func(t *testing.T, agent Agent) {
				if agent.Name() != "test-agent" {
					t.Errorf("Name() = %v, want %v", agent.Name(), "test-agent")
				}
				if agent.Model() != "gpt-4" {
					t.Errorf("Model() = %v, want %v", agent.Model(), "gpt-4")
				}
			},
		},
		{
			name: "agent with all fields",
			build: func() (Agent, error) {
				tool := &mockTool{name: "test-tool"}
				return New(
					WithName("full-agent"),
					WithDescription("A test agent"),
					WithInstructions("You are a helpful assistant"),
					WithModel("gpt-3.5-turbo"),
					WithModelSettings(&ModelSettings{
						Temperature: floatPtr(0.7),
						MaxTokens:   intPtr(1000),
					}),
					WithTools(tool),
					WithOpenAI("test-key"),
					WithSessionStore(&mockSessionStore{}),
				)
			},
			wantErr: false,
			validate: func(t *testing.T, agent Agent) {
				if agent.Name() != "full-agent" {
					t.Errorf("Name() = %v, want %v", agent.Name(), "full-agent")
				}
				if agent.Description() != "A test agent" {
					t.Errorf("Description() = %v, want %v", agent.Description(), "A test agent")
				}
				if agent.Instructions() != "You are a helpful assistant" {
					t.Errorf("Instructions() = %v, want %v", agent.Instructions(), "You are a helpful assistant")
				}
				if agent.Model() != "gpt-3.5-turbo" {
					t.Errorf("Model() = %v, want %v", agent.Model(), "gpt-3.5-turbo")
				}
				if len(agent.Tools()) != 1 {
					t.Errorf("Tools() length = %v, want %v", len(agent.Tools()), 1)
				}
			},
		},
		{
			name: "empty name",
			build: func() (Agent, error) {
				return New(
					WithName(""),
					WithOpenAI("test-key"),
					WithSessionStore(&mockSessionStore{}),
				)
			},
			wantErr: true,
			errMsg:  "agent name cannot be empty",
		},
		{
			name: "missing chat model",
			build: func() (Agent, error) {
				return New(
					WithName("test-agent"),
					WithSessionStore(&mockSessionStore{}),
				)
			},
			wantErr: true,
			errMsg:  "chat model is required",
		},
		{
			name: "missing session store",
			build: func() (Agent, error) {
				return New(
					WithName("test-agent"),
					WithOpenAI("test-key"),
				)
			},
			wantErr: true,
			errMsg:  "session store is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := tt.build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Build() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
			if err == nil && tt.validate != nil {
				tt.validate(t, agent)
			}
		})
	}
}

func TestAgentFunctionalOptions(t *testing.T) {
	// Set up mocks
	originalChatModelImpl := newOpenAIChatModelImpl
	originalAgentImpl := createAgentImplementation
	originalSessionFactory := sessionFactory
	
	defer func() {
		newOpenAIChatModelImpl = originalChatModelImpl
		createAgentImplementation = originalAgentImpl
		sessionFactory = originalSessionFactory
	}()
	
	newOpenAIChatModelImpl = func(apiKey string, config *OpenAIConfig) (ChatModel, error) {
		return NewMockChatModel(), nil
	}
	
	createAgentImplementation = func(config *AgentConfig) Agent {
		return &mockAgent{
			name:          config.Name,
			description:   config.Description,
			instructions:  config.Instructions,
			model:         config.Model,
			modelSettings: config.ModelSettings,
			tools:         config.Tools,
			outputType:    config.OutputType,
		}
	}
	
	sessionFactory = func(id string) Session {
		return &basicSession{id: id}
	}

	// Test individual options
	agent, err := New(
		WithName("test-agent"),
		WithDescription("Test description"),
		WithInstructions("Test instructions"),
		WithModel("gpt-4"),
		WithMaxTurns(5),
		WithDebugLogging(),
		WithOpenAI("test-key"),
		WithSessionStore(&mockSessionStore{}),
	)
	
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	
	if agent.Name() != "test-agent" {
		t.Errorf("Name() = %v, want %v", agent.Name(), "test-agent")
	}
	
	if agent.Description() != "Test description" {
		t.Errorf("Description() = %v, want %v", agent.Description(), "Test description")
	}
	
	if agent.Instructions() != "Test instructions" {
		t.Errorf("Instructions() = %v, want %v", agent.Instructions(), "Test instructions")
	}
	
	if agent.Model() != "gpt-4" {
		t.Errorf("Model() = %v, want %v", agent.Model(), "gpt-4")
	}
}

func TestOpenAIOptions(t *testing.T) {
	// Set up mock implementation
	originalImpl := newOpenAIChatModelImpl
	defer func() { newOpenAIChatModelImpl = originalImpl }()
	
	newOpenAIChatModelImpl = func(apiKey string, config *OpenAIConfig) (ChatModel, error) {
		return NewMockChatModel(), nil
	}
	
	tests := []struct {
		name      string
		options   []OpenAIOption
		expectErr bool
	}{
		{
			name:    "default options",
			options: []OpenAIOption{},
		},
		{
			name: "with custom options",
			options: []OpenAIOption{
				WithOpenAIBaseURL("https://api.example.com"),
				WithOpenAIOrganization("org-123"),
				WithOpenAITimeout(60 * time.Second),
				WithOpenAIRetryCount(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOpenAIChatModel("test-key", tt.options...)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewOpenAIChatModel() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findString(s, substr) >= 0
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}