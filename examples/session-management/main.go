package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidleitw/go-agent/pkg/agent"
	"github.com/davidleitw/go-agent/pkg/agent/mocks"
)

// This example demonstrates the new simplified Session interface
// and shows how to manage conversation sessions effectively.

func main() {
	fmt.Println("=== Session Management Example ===")
	fmt.Println("Demonstrating the new simplified Session interface in go-agent")
	fmt.Println()

	// Demo 1: Basic session operations
	fmt.Println("📚 Demo 1: Basic Session Operations")
	basicSessionDemo()
	fmt.Println()

	// Demo 2: Data storage and retrieval
	fmt.Println("💾 Demo 2: Session Data Storage")
	dataStorageDemo()
	fmt.Println()

	// Demo 3: Session cloning and independence
	fmt.Println("🔄 Demo 3: Session Cloning")
	sessionCloningDemo()
	fmt.Println()

	// Demo 4: Session with agent integration
	fmt.Println("🤖 Demo 4: Session with Agent Integration")
	agentIntegrationDemo()
	fmt.Println()

	// Demo 5: Concurrent session access
	fmt.Println("⚡ Demo 5: Concurrent Session Access")
	concurrentAccessDemo()
	fmt.Println()

	fmt.Println("✅ All session management demos completed!")
}

// basicSessionDemo demonstrates the core Session interface methods
func basicSessionDemo() {
	// Create a new session
	session := agent.NewSession("demo-session-1")
	
	fmt.Printf("Created session: %s\n", session.ID())
	fmt.Printf("Initial message count: %d\n", len(session.Messages()))
	
	// Add various types of messages
	userMsg := session.AddMessage(agent.RoleUser, "Hello! I need help with my account.")
	fmt.Printf("Added user message: %s\n", userMsg.Content)
	
	assistantMsg := session.AddMessage(agent.RoleAssistant, "I'd be happy to help you with your account. What specific issue are you experiencing?")
	fmt.Printf("Added assistant message: %s\n", assistantMsg.Content)
	
	systemMsg := session.AddMessage(agent.RoleSystem, "User has been authenticated successfully.")
	fmt.Printf("Added system message: %s\n", systemMsg.Content)
	
	// Display all messages
	fmt.Printf("\n📜 Conversation history (%d messages):\n", len(session.Messages()))
	for i, msg := range session.Messages() {
		fmt.Printf("  %d. [%s] %s (at %s)\n", 
			i+1, msg.Role, msg.Content, msg.Timestamp.Format("15:04:05"))
	}
}

// dataStorageDemo shows how to use session data storage
func dataStorageDemo() {
	session := agent.NewSession("demo-session-2")
	
	// Store various types of data
	session.SetData("user_id", "user_12345")
	session.SetData("login_attempts", 3)
	session.SetData("preferences", map[string]string{
		"language": "en",
		"theme":    "dark",
	})
	session.SetData("last_login", time.Now())
	
	// Retrieve and display data
	fmt.Printf("User ID: %v\n", session.GetData("user_id"))
	fmt.Printf("Login attempts: %v\n", session.GetData("login_attempts"))
	
	if prefs, ok := session.GetData("preferences").(map[string]string); ok {
		fmt.Printf("User preferences: %+v\n", prefs)
	}
	
	if lastLogin, ok := session.GetData("last_login").(time.Time); ok {
		fmt.Printf("Last login: %s\n", lastLogin.Format("2006-01-02 15:04:05"))
	}
	
	// Try to get non-existent data
	nonExistent := session.GetData("non_existent_key")
	fmt.Printf("Non-existent key: %v\n", nonExistent) // Should be nil
	
	// Update existing data
	session.SetData("login_attempts", 4)
	fmt.Printf("Updated login attempts: %v\n", session.GetData("login_attempts"))
}

// sessionCloningDemo demonstrates session cloning and independence
func sessionCloningDemo() {
	// Create original session with data and messages
	original := agent.NewSession("original-session")
	original.AddMessage(agent.RoleUser, "Original message 1")
	original.AddMessage(agent.RoleAssistant, "Original response 1")
	original.SetData("counter", 100)
	original.SetData("status", "active")
	
	fmt.Printf("Original session: %d messages, counter=%v\n", 
		len(original.Messages()), original.GetData("counter"))
	
	// Clone the session
	type cloneableSession interface {
		agent.Session
		Clone() agent.Session
	}
	
	if cloneable, ok := original.(cloneableSession); ok {
		cloned := cloneable.Clone()
		
		fmt.Printf("Cloned session: %d messages, counter=%v\n", 
			len(cloned.Messages()), cloned.GetData("counter"))
		
		// Modify original session
		original.AddMessage(agent.RoleUser, "Additional message in original")
		original.SetData("counter", 200)
		
		// Modify cloned session
		cloned.AddMessage(agent.RoleUser, "Additional message in clone")
		cloned.SetData("counter", 150)
		
		// Verify independence
		fmt.Printf("\nAfter modifications:")
		fmt.Printf("Original session: %d messages, counter=%v\n", 
			len(original.Messages()), original.GetData("counter"))
		fmt.Printf("Cloned session: %d messages, counter=%v\n", 
			len(cloned.Messages()), cloned.GetData("counter"))
	} else {
		fmt.Println("Session implementation doesn't support cloning")
	}
}

// agentIntegrationDemo shows how sessions work with agents
func agentIntegrationDemo() {
	// Create a mock chat model for testing
	mockChat := mocks.NewMockChatModel()
	mockChat.SetResponse(agent.Message{
		Role:      agent.RoleAssistant,
		Content:   "I understand you need help. Let me assist you with that.",
		Timestamp: time.Now(),
	})
	
	// Create a basic agent
	testAgent, err := agent.NewBasicAgent(agent.BasicAgentConfig{
		Name:         "session-demo-agent",
		Description:  "An agent for demonstrating session management",
		Instructions: "You are a helpful assistant that manages user sessions.",
		Model:        "test-model",
		ChatModel:    mockChat,
	})
	if err != nil {
		log.Printf("Failed to create agent: %v", err)
		return
	}
	
	// Create session and have conversation
	session := agent.NewSession("agent-demo-session")
	
	// Store some context data
	session.SetData("user_name", "Alice")
	session.SetData("session_start", time.Now())
	
	ctx := context.Background()
	response, structuredOutput, err := testAgent.Chat(ctx, session, "Hello, I need help with my session")
	if err != nil {
		log.Printf("Chat error: %v", err)
		return
	}
	
	fmt.Printf("User: Hello, I need help with my session\n")
	fmt.Printf("Agent: %s\n", response.Content)
	fmt.Printf("Structured output: %v\n", structuredOutput)
	fmt.Printf("Session now has %d messages\n", len(session.Messages()))
	
	// Display session data
	if userName := session.GetData("user_name"); userName != nil {
		fmt.Printf("Session user: %v\n", userName)
	}
	if startTime, ok := session.GetData("session_start").(time.Time); ok {
		fmt.Printf("Session duration: %v\n", time.Since(startTime).Round(time.Millisecond))
	}
}

// concurrentAccessDemo demonstrates thread-safe session operations
func concurrentAccessDemo() {
	session := agent.NewSession("concurrent-demo")
	
	// Add initial data
	session.SetData("shared_counter", 0)
	
	const numWorkers = 5
	const messagesPerWorker = 3
	
	done := make(chan bool, numWorkers)
	
	// Start multiple goroutines to access session concurrently
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer func() { done <- true }()
			
			for j := 0; j < messagesPerWorker; j++ {
				// Add messages
				message := fmt.Sprintf("Message from worker %d, iteration %d", workerID, j+1)
				session.AddMessage(agent.RoleUser, message)
				
				// Update shared data
				currentCounter := session.GetData("shared_counter")
				if counter, ok := currentCounter.(int); ok {
					session.SetData("shared_counter", counter+1)
				}
				
				// Add some variety
				time.Sleep(time.Millisecond * 10)
			}
		}(i)
	}
	
	// Wait for all workers to complete
	for i := 0; i < numWorkers; i++ {
		<-done
	}
	
	fmt.Printf("Concurrent operations completed:\n")
	fmt.Printf("Total messages: %d\n", len(session.Messages()))
	fmt.Printf("Final shared counter: %v\n", session.GetData("shared_counter"))
	
	// Verify message content
	userMessageCount := 0
	for _, msg := range session.Messages() {
		if msg.Role == agent.RoleUser {
			userMessageCount++
		}
	}
	fmt.Printf("User messages: %d (expected: %d)\n", userMessageCount, numWorkers*messagesPerWorker)
}