package agent

import (
	"context"
	"testing"
	"time"
)

// Mock implementations for testing

type mockChatModel struct {
	response *Message
	error    error
}

func (m *mockChatModel) GenerateChatCompletion(ctx context.Context, messages []Message, model string, settings *ModelSettings, tools []Tool) (*Message, error) {
	if m.error != nil {
		return nil, m.error
	}
	if m.response != nil {
		return m.response, nil
	}
	return &Message{
		Role:      RoleAssistant,
		Content:   "Mock response",
		Timestamp: time.Now(),
	}, nil
}

func (m *mockChatModel) GetSupportedModels() []string {
	return []string{"mock-model", "test-model"}
}

func (m *mockChatModel) ValidateModel(model string) error {
	return nil
}

func (m *mockChatModel) GetModelInfo(model string) (*ModelInfo, error) {
	return &ModelInfo{
		ID:          model,
		Name:        "Mock Model",
		Description: "A mock model for testing",
		Provider:    "mock",
	}, nil
}

type mockTool struct {
	name        string
	description string
	schema      map[string]any
	executeFunc func(ctx context.Context, args map[string]any) (any, error)
}

func (m *mockTool) Name() string           { return m.name }
func (m *mockTool) Description() string    { return m.description }
func (m *mockTool) Schema() map[string]any { return m.schema }
func (m *mockTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, args)
	}
	return "mock result", nil
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
	mockChat := &mockChatModel{}
	mockStore := &mockSessionStore{}

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
				return NewBasicAgent(BasicAgentConfig{
					Name:      "test-agent",
					ChatModel: mockChat,
				})
			},
			wantErr: false,
			validate: func(t *testing.T, agent Agent) {
				if agent.Name() != "test-agent" {
					t.Errorf("Name() = %v, want %v", agent.Name(), "test-agent")
				}
			},
		},
		{
			name: "agent with all fields",
			build: func() (Agent, error) {
				tool := &mockTool{name: "test-tool"}
				return NewBasicAgent(BasicAgentConfig{
					Name:         "full-agent",
					Description:  "A test agent",
					Instructions: "You are a helpful assistant",
					Model:        "gpt-3.5-turbo",
					ModelSettings: &ModelSettings{
						Temperature: floatPtr(0.7),
						MaxTokens:   intPtr(1000),
					},
					Tools:        []Tool{tool},
					ChatModel:    mockChat,
					SessionStore: mockStore,
				})
			},
			wantErr: false,
			validate: func(t *testing.T, agent Agent) {
				if agent.Name() != "full-agent" {
					t.Errorf("Name() = %v, want %v", agent.Name(), "full-agent")
				}
				if agent.Description() != "A test agent" {
					t.Errorf("Description() = %v, want %v", agent.Description(), "A test agent")
				}
				if len(agent.GetTools()) != 1 {
					t.Errorf("GetTools() length = %v, want %v", len(agent.GetTools()), 1)
				}
			},
		},
		{
			name: "empty name",
			build: func() (Agent, error) {
				return NewBasicAgent(BasicAgentConfig{
					Name:      "",
					ChatModel: mockChat,
				})
			},
			wantErr: true,
			errMsg:  "agent name is required",
		},
		{
			name: "missing chat model",
			build: func() (Agent, error) {
				return NewBasicAgent(BasicAgentConfig{
					Name: "test-agent",
				})
			},
			wantErr: true,
			errMsg:  "chat model is required",
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

func TestAgentConfiguration(t *testing.T) {
	mockChat := &mockChatModel{}
	mockStore := &mockSessionStore{}

	// Test basic agent configuration
	agent, err := NewBasicAgent(BasicAgentConfig{
		Name:         "test-agent",
		Description:  "Test description",
		Instructions: "Test instructions",
		Model:        "gpt-4",
		MaxTurns:     5,
		ChatModel:    mockChat,
		SessionStore: mockStore,
	})

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent.Name() != "test-agent" {
		t.Errorf("Name() = %v, want %v", agent.Name(), "test-agent")
	}

	if agent.Description() != "Test description" {
		t.Errorf("Description() = %v, want %v", agent.Description(), "Test description")
	}
}

func TestAgentChat(t *testing.T) {
	mockChat := &mockChatModel{
		response: &Message{
			Role:      RoleAssistant,
			Content:   "Hello! How can I help you?",
			Timestamp: time.Now(),
		},
	}

	agent, err := NewBasicAgent(BasicAgentConfig{
		Name:      "test-agent",
		ChatModel: mockChat,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()
	session := NewSession("test-session")
	response, structuredOutput, err := agent.Chat(ctx, session, "Hello!")

	if err != nil {
		t.Errorf("Chat() error = %v, want nil", err)
	}

	if response == nil {
		t.Error("Chat() response is nil")
	} else {
		if response.Role != RoleAssistant {
			t.Errorf("Response role = %v, want %v", response.Role, RoleAssistant)
		}
		if response.Content != "Hello! How can I help you?" {
			t.Errorf("Response content = %v, want %v", response.Content, "Hello! How can I help you?")
		}
	}

	if structuredOutput != nil {
		t.Errorf("Structured output = %v, want nil (no output type set)", structuredOutput)
	}
}

func TestModelSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *ModelSettings
		wantErr  bool
	}{
		{
			name: "valid settings",
			settings: &ModelSettings{
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(1000),
				TopP:        floatPtr(0.9),
			},
			wantErr: false,
		},
		{
			name: "invalid temperature",
			settings: &ModelSettings{
				Temperature: floatPtr(3.0), // Too high
			},
			wantErr: true,
		},
		{
			name: "invalid max tokens",
			settings: &ModelSettings{
				MaxTokens: intPtr(-1), // Negative
			},
			wantErr: true,
		},
		{
			name: "invalid top_p",
			settings: &ModelSettings{
				TopP: floatPtr(1.5), // Too high
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelSettings.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMemorySessionStore(t *testing.T) {
	store := NewInMemorySessionStore()
	ctx := context.Background()

	// Test creating and saving a session
	session := NewSession("test-session")
	session.AddUserMessage("Hello")

	err := store.Save(ctx, session)
	if err != nil {
		t.Errorf("Save() error = %v, want nil", err)
	}

	// Test loading the session
	loadedSession, err := store.Load(ctx, "test-session")
	if err != nil {
		t.Errorf("Load() error = %v, want nil", err)
	}

	if loadedSession.ID() != "test-session" {
		t.Errorf("Loaded session ID = %v, want %v", loadedSession.ID(), "test-session")
	}

	if len(loadedSession.Messages()) != 1 {
		t.Errorf("Loaded session messages count = %v, want %v", len(loadedSession.Messages()), 1)
	}

	// Test deleting the session
	err = store.Delete(ctx, "test-session")
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Test loading deleted session
	_, err = store.Load(ctx, "test-session")
	if err != ErrSessionNotFound {
		t.Errorf("Load() after delete error = %v, want %v", err, ErrSessionNotFound)
	}
}

// Helper functions (floatPtr and intPtr are now in agent.go)

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
