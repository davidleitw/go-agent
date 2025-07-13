module basic-chat

go 1.22

require (
	github.com/davidleitw/go-agent v0.0.0
	github.com/joho/godotenv v1.5.1
)

require github.com/sashabaranov/go-openai v1.40.5 // indirect

replace github.com/davidleitw/go-agent => ../../../
