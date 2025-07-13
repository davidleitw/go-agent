module multi-tool-agent

go 1.22.0

replace github.com/davidleitw/go-agent => ../../../

require (
	github.com/davidleitw/go-agent v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
)

require github.com/sashabaranov/go-openai v1.40.5 // indirect
