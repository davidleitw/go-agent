// Package agent provides the core interfaces and types for building AI agents.
package agent

import (
	"fmt"
)

// ModelSettings contains configuration parameters for LLM models.
type ModelSettings struct {
	// Temperature controls randomness in generation (0.0 to 2.0)
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens limits the maximum tokens in the response
	MaxTokens *int `json:"max_tokens,omitempty"`

	// TopP is nucleus sampling parameter (0.0 to 1.0)
	TopP *float64 `json:"top_p,omitempty"`

	// Stop sequences that will halt generation
	Stop []string `json:"stop,omitempty"`

	// FrequencyPenalty reduces repetition (-2.0 to 2.0)
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// PresencePenalty encourages new topics (-2.0 to 2.0)
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// Seed for deterministic generation
	Seed *int `json:"seed,omitempty"`

	// ResponseFormat specifies the output format (e.g., "json_object")
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ResponseFormat specifies the desired output format
type ResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// Merge combines two ModelSettings, with values from 'other' taking precedence
func (ms *ModelSettings) Merge(other *ModelSettings) *ModelSettings {
	if ms == nil {
		return other
	}
	if other == nil {
		return ms
	}

	result := &ModelSettings{
		Temperature:      ms.Temperature,
		MaxTokens:        ms.MaxTokens,
		TopP:             ms.TopP,
		Stop:             ms.Stop,
		FrequencyPenalty: ms.FrequencyPenalty,
		PresencePenalty:  ms.PresencePenalty,
		Seed:             ms.Seed,
		ResponseFormat:   ms.ResponseFormat,
	}

	// Override with values from 'other' if they are set
	if other.Temperature != nil {
		result.Temperature = other.Temperature
	}
	if other.MaxTokens != nil {
		result.MaxTokens = other.MaxTokens
	}
	if other.TopP != nil {
		result.TopP = other.TopP
	}
	if len(other.Stop) > 0 {
		result.Stop = other.Stop
	}
	if other.FrequencyPenalty != nil {
		result.FrequencyPenalty = other.FrequencyPenalty
	}
	if other.PresencePenalty != nil {
		result.PresencePenalty = other.PresencePenalty
	}
	if other.Seed != nil {
		result.Seed = other.Seed
	}
	if other.ResponseFormat != nil {
		result.ResponseFormat = other.ResponseFormat
	}

	return result
}

// Validate checks if the model settings are within valid ranges
func (ms *ModelSettings) Validate() error {
	if ms == nil {
		return nil
	}

	if ms.Temperature != nil && (*ms.Temperature < 0 || *ms.Temperature > 2.0) {
		return fmt.Errorf("temperature must be between 0 and 2.0, got %f", *ms.Temperature)
	}

	if ms.MaxTokens != nil && *ms.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive, got %d", *ms.MaxTokens)
	}

	if ms.TopP != nil && (*ms.TopP < 0 || *ms.TopP > 1.0) {
		return fmt.Errorf("top_p must be between 0 and 1.0, got %f", *ms.TopP)
	}

	if ms.FrequencyPenalty != nil && (*ms.FrequencyPenalty < -2.0 || *ms.FrequencyPenalty > 2.0) {
		return fmt.Errorf("frequency_penalty must be between -2.0 and 2.0, got %f", *ms.FrequencyPenalty)
	}

	if ms.PresencePenalty != nil && (*ms.PresencePenalty < -2.0 || *ms.PresencePenalty > 2.0) {
		return fmt.Errorf("presence_penalty must be between -2.0 and 2.0, got %f", *ms.PresencePenalty)
	}

	if ms.ResponseFormat != nil && ms.ResponseFormat.Type != "text" && ms.ResponseFormat.Type != "json_object" {
		return fmt.Errorf("response_format.type must be 'text' or 'json_object', got %s", ms.ResponseFormat.Type)
	}

	return nil
}

// Common errors returned by the agent package
var (
	// ErrSessionNotFound is returned when a session ID doesn't exist
	ErrSessionNotFound = fmt.Errorf("session not found")

	// ErrInvalidSession is returned when a session is invalid
	ErrInvalidSession = fmt.Errorf("invalid session")

	// ErrInvalidSessionID is returned when a session ID is invalid
	ErrInvalidSessionID = fmt.Errorf("invalid session ID")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = fmt.Errorf("invalid input")

	// ErrToolNotFound is returned when a requested tool doesn't exist
	ErrToolNotFound = fmt.Errorf("tool not found")

	// ErrToolExecutionFailed is returned when a tool fails to execute
	ErrToolExecutionFailed = fmt.Errorf("tool execution failed")

	// ErrOutputValidationFailed is returned when structured output validation fails
	ErrOutputValidationFailed = fmt.Errorf("output validation failed")
)
