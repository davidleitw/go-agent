package llm

// Config holds configuration for creating a Model instance
type Config struct {
	// API key for authentication
	APIKey string

	// Model name (e.g., "gpt-4", "gpt-3.5-turbo")
	Model string

	// Optional base URL for API endpoint (for proxies or custom endpoints)
	BaseURL string

	// TODO: Future configuration options
	// - Timeout settings
	// - Retry configuration
	// - Default temperature/max_tokens
	// - Organization ID
}