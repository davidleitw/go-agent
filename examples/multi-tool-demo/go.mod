module multi-tool-demo

go 1.22

toolchain go1.24.0

// Replace with local go-agent module for development
replace github.com/davidleitw/go-agent => ../../

require github.com/davidleitw/go-agent v0.0.0-00010101000000-000000000000

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/sashabaranov/go-openai v1.40.5 // indirect
)
