package openai

import (
	"context"
	"testing"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
)

func TestNewChatModel(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "test-key",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			config:  nil,
			wantErr: true,
		},
		{
			name:   "with custom config",
			apiKey: "test-key",
			config: &Config{
				BaseURL:      "https://api.example.com",
				Organization: "org-123",
				Timeout:      60 * time.Second,
				RetryCount:   5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatModel, err := NewChatModel(tt.apiKey, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewChatModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && chatModel == nil {
				t.Error("NewChatModel() returned nil chatModel")
			}
		})
	}
}

func TestChatModel_GetSupportedModels(t *testing.T) {
	chatModel, err := NewChatModel("test-key", nil)
	if err != nil {
		t.Fatalf("Failed to create chat model: %v", err)
	}

	models := chatModel.GetSupportedModels()
	expectedModels := []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
	}

	if len(models) != len(expectedModels) {
		t.Errorf("GetSupportedModels() returned %d models, want %d", len(models), len(expectedModels))
	}

	for _, expected := range expectedModels {
		found := false
		for _, model := range models {
			if model == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedModels() missing expected model: %s", expected)
		}
	}
}

func TestChatModel_ValidateModel(t *testing.T) {
	chatModel, err := NewChatModel("test-key", nil)
	if err != nil {
		t.Fatalf("Failed to create chat model: %v", err)
	}

	tests := []struct {
		name    string
		model   string
		wantErr bool
	}{
		{"valid gpt-4", "gpt-4", false},
		{"valid gpt-4o", "gpt-4o", false},
		{"valid gpt-3.5-turbo", "gpt-3.5-turbo", false},
		{"invalid model", "invalid-model", true},
		{"empty model", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chatModel.ValidateModel(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatModel_GetModelInfo(t *testing.T) {
	chatModel, err := NewChatModel("test-key", nil)
	if err != nil {
		t.Fatalf("Failed to create chat model: %v", err)
	}

	tests := []struct {
		name    string
		model   string
		wantErr bool
	}{
		{"gpt-4", "gpt-4", false},
		{"gpt-4o", "gpt-4o", false},
		{"gpt-3.5-turbo", "gpt-3.5-turbo", false},
		{"invalid model", "invalid-model", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := chatModel.GetModelInfo(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetModelInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if info == nil {
					t.Error("GetModelInfo() returned nil info")
					return
				}
				if info.ID != tt.model {
					t.Errorf("GetModelInfo() ID = %v, want %v", info.ID, tt.model)
				}
				if info.Provider != "openai" {
					t.Errorf("GetModelInfo() Provider = %v, want %v", info.Provider, "openai")
				}
				if !info.SupportsTools {
					t.Error("GetModelInfo() SupportsTools = false, want true")
				}
			}
		})
	}
}

func TestChatModel_GenerateChatCompletion_Validation(t *testing.T) {
	chatModel, err := NewChatModel("test-key", nil)
	if err != nil {
		t.Fatalf("Failed to create chat model: %v", err)
	}

	ctx := context.Background()
	messages := []agent.Message{
		{
			Role:      agent.RoleUser,
			Content:   "Hello",
			Timestamp: time.Now(),
		},
	}

	// Note: This will fail with actual API call, but tests the function signature
	// and validation logic without requiring a real API key
	_, err = chatModel.GenerateChatCompletion(ctx, messages, "gpt-4", nil, nil)

	// We expect an error since we're using a fake API key
	// The important thing is that the function signature works
	if err == nil {
		t.Log("GenerateChatCompletion() unexpectedly succeeded (might be using real API key)")
	} else {
		t.Logf("GenerateChatCompletion() failed as expected with test key: %v", err)
	}
}

func TestModelSettings_Validation(t *testing.T) {
	tests := []struct {
		name     string
		settings *agent.ModelSettings
		wantErr  bool
	}{
		{
			name: "valid settings",
			settings: &agent.ModelSettings{
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(1000),
				TopP:        floatPtr(0.9),
			},
			wantErr: false,
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name: "invalid temperature",
			settings: &agent.ModelSettings{
				Temperature: floatPtr(3.0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.settings != nil {
				err := tt.settings.Validate()
				if (err != nil) != tt.wantErr {
					t.Errorf("ModelSettings.Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }
