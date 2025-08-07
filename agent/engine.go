package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/prompt"
	"github.com/davidleitw/go-agent/session"
	"github.com/davidleitw/go-agent/session/memory"
	"github.com/davidleitw/go-agent/tool"
)

const (
	version = "v0.0.1"
)

// engine implements the Engine interface with pre-configured components
type engine struct {
	// Core components
	model            llm.Model
	sessionStore     session.SessionStore
	toolRegistry     *tool.Registry
	contextProviders []agentcontext.Provider

	// Prompt template
	promptTemplate prompt.Template

	// Configuration
	maxIterations int
	temperature   *float32
	maxTokens     *int

	// History configuration
	historyLimit       int
	historyInterceptor HistoryInterceptor

	// Session configuration
	sessionTTL       time.Duration
	cachedCreateOpts []session.CreateOption
}

// NewEngine creates a new engine with the provided configuration
func NewEngine(config EngineConfig) (Engine, error) {
	// Validate required components
	if config.Model == nil {
		return nil, fmt.Errorf("model is required")
	}

	// Set defaults for optional components
	if config.SessionStore == nil {
		config.SessionStore = memory.NewStore()
	}

	if config.ToolRegistry == nil {
		config.ToolRegistry = tool.NewRegistry()
	}

	if config.MaxIterations == 0 {
		config.MaxIterations = 5
	}

	// Set default session TTL if not specified
	sessionTTL := config.SessionTTL
	if sessionTTL == 0 {
		sessionTTL = 24 * time.Hour // Default 24 hours
	}

	// Pre-compute session create options for performance
	createOpts := []session.CreateOption{}
	if sessionTTL > 0 {
		createOpts = append(createOpts, session.WithTTL(sessionTTL))
	}

	// Add basic metadata
	createOpts = append(createOpts,
		session.WithMetadata("created_by", "agent"),
		session.WithMetadata("agent_version", version),
	)

	return &engine{
		model:              config.Model,
		sessionStore:       config.SessionStore,
		toolRegistry:       config.ToolRegistry,
		contextProviders:   config.ContextProviders,
		promptTemplate:     config.PromptTemplate,
		maxIterations:      config.MaxIterations,
		temperature:        config.Temperature,
		maxTokens:          config.MaxTokens,
		historyLimit:       config.HistoryLimit,
		historyInterceptor: config.HistoryInterceptor,
		sessionTTL:         sessionTTL,
		cachedCreateOpts:   createOpts,
	}, nil
}

// Execute implements the core agent execution logic
func (e *engine) Execute(ctx context.Context, request Request) (*Response, error) {
	// Validate input
	if request.Input == "" {
		return nil, ErrInvalidInput
	}

	// Step 1: Session Management
	agentSession, err := e.handleSession(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("session handling failed: %w", err)
	}

	// Step 2: Context Collection
	contexts, err := e.gatherContexts(ctx, request, agentSession)
	if err != nil {
		return nil, fmt.Errorf("context gathering failed: %w", err)
	}

	// Step 3: Main Execution Loop
	// TODO: Implement iterative agent thinking with tool calls
	result, err := e.executeIterations(ctx, request, contexts, agentSession)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// Step 4: Finalize Response
	response := &Response{
		Output:    result.FinalOutput,
		SessionID: result.SessionID,
		Session:   result.Session,
		Metadata:  result.Metadata,
		Usage:     result.Usage,
	}

	return response, nil
}

// handleSession manages session creation/retrieval
func (e *engine) handleSession(ctx context.Context, request Request) (session.Session, error) {
	if request.SessionID == "" {
		// Create new session with pre-cached options
		newSession := e.sessionStore.Create(ctx, e.cachedCreateOpts...)

		// Add some dynamic metadata based on request
		newSession.Set("initial_input_length", len(request.Input))
		newSession.Set("session_start_time", time.Now().Format(time.RFC3339))

		return newSession, nil
	}

	// Load existing session
	existingSession, err := e.sessionStore.Get(ctx, request.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	return existingSession, nil
}

// gatherContexts collects context from all providers and history
func (e *engine) gatherContexts(ctx context.Context, request Request, agentSession session.Session) ([]agentcontext.Context, error) {
	var allContexts []agentcontext.Context

	// 1. Collect contexts from providers (non-history)
	for _, provider := range e.contextProviders {
		// Check for cancellation before calling provider
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Call provider with context support
		contexts := provider.Provide(ctx, agentSession)
		allContexts = append(allContexts, contexts...)
	}

	// 2. Add history contexts if enabled
	if e.historyLimit > 0 {
		historyContexts, err := e.extractHistoryContexts(ctx, agentSession)
		if err != nil {
			return nil, fmt.Errorf("failed to extract history contexts: %w", err)
		}
		allContexts = append(allContexts, historyContexts...)
	}

	return allContexts, nil
}

// extractHistoryContexts extracts and processes history from session
func (e *engine) extractHistoryContexts(ctx context.Context, agentSession session.Session) ([]agentcontext.Context, error) {
	// 1. Get raw history entries from session
	entries := agentSession.GetHistory(e.historyLimit)
	if len(entries) == 0 {
		return nil, nil
	}

	// 2. Apply history interceptor if configured
	if e.historyInterceptor != nil {
		processedEntries, err := e.historyInterceptor.ProcessHistory(ctx, entries, e.model)
		if err != nil {
			return nil, fmt.Errorf("history interceptor failed: %w", err)
		}
		entries = processedEntries
	}

	// 3. Convert entries to contexts
	contexts := e.convertEntriesToContexts(entries)

	return contexts, nil
}

// convertEntriesToContexts converts session entries to context objects
func (e *engine) convertEntriesToContexts(entries []session.Entry) []agentcontext.Context {
	contexts := make([]agentcontext.Context, 0, len(entries))

	for _, entry := range entries {
		contextEntry := agentcontext.Context{
			Metadata: make(map[string]any),
		}

		// Copy entry metadata
		for k, v := range entry.Metadata {
			contextEntry.Metadata[k] = v
		}
		contextEntry.Metadata["entry_id"] = entry.ID
		contextEntry.Metadata["timestamp"] = entry.Timestamp

		// Convert based on entry type
		switch entry.Type {
		case session.EntryTypeMessage:
			if content, ok := session.GetMessageContent(entry); ok {
				contextEntry.Type = content.Role // "user", "assistant", or "system"
				contextEntry.Content = content.Text
			}

		case session.EntryTypeToolCall:
			if content, ok := session.GetToolCallContent(entry); ok {
				contextEntry.Type = agentcontext.TypeToolCall
				params, _ := json.Marshal(content.Parameters)
				contextEntry.Content = fmt.Sprintf("Tool: %s\nParameters: %s", content.Tool, string(params))
				contextEntry.Metadata["tool_name"] = content.Tool
			}

		case session.EntryTypeToolResult:
			if content, ok := session.GetToolResultContent(entry); ok {
				contextEntry.Type = agentcontext.TypeToolResult
				if content.Success {
					result, _ := json.Marshal(content.Result)
					contextEntry.Content = fmt.Sprintf("Tool: %s\nSuccess: true\nResult: %s", content.Tool, string(result))
				} else {
					contextEntry.Content = fmt.Sprintf("Tool: %s\nSuccess: false\nError: %s", content.Tool, content.Error)
				}
				contextEntry.Metadata["tool_name"] = content.Tool
				contextEntry.Metadata["success"] = content.Success
			}

		default:
			// Skip unknown entry types
			continue
		}

		contexts = append(contexts, contextEntry)
	}

	return contexts
}

// ExecutionResult holds the final execution result
type ExecutionResult struct {
	FinalOutput string
	SessionID   string
	Session     session.Session
	Metadata    map[string]any
	Usage       Usage
}

// executeIterations runs the main agent thinking loop
func (e *engine) executeIterations(ctx context.Context, request Request, contexts []agentcontext.Context, agentSession session.Session) (*ExecutionResult, error) {
	// Initialize execution state
	var totalUsage Usage
	var conversationMessages []llm.Message
	var finalResponse string

	// Step 1: Build initial messages from contexts and user input
	messages := e.buildLLMMessages(contexts, request)
	conversationMessages = append(conversationMessages, messages...)

	// Step 2: Main iteration loop
	for iteration := 0; iteration < e.maxIterations; iteration++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fmt.Printf("🤖 [Iteration %d] Agent thinking...\n", iteration+1)

		// Step 2a: Get available tools
		tools := e.toolRegistry.GetDefinitions()
		fmt.Printf("📚 Available tools: %d\n", len(tools))

		// Step 2b: Prepare LLM request
		llmRequest := llm.Request{
			Messages:    conversationMessages,
			Tools:       tools,
			Temperature: e.temperature,
			MaxTokens:   e.maxTokens,
		}

		// Step 2c: Call LLM
		fmt.Printf("💭 Calling LLM with %d messages...\n", len(conversationMessages))
		response, err := e.model.Complete(ctx, llmRequest)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed at iteration %d: %w", iteration, err)
		}

		// Step 2d: Update usage tracking
		totalUsage.LLMTokens.PromptTokens += response.Usage.PromptTokens
		totalUsage.LLMTokens.CompletionTokens += response.Usage.CompletionTokens
		totalUsage.LLMTokens.TotalTokens += response.Usage.TotalTokens

		fmt.Printf("📊 Token usage this iteration: %d tokens (prompt: %d, completion: %d)\n",
			response.Usage.TotalTokens, response.Usage.PromptTokens, response.Usage.CompletionTokens)

		// Step 2e: Process LLM response
		if len(response.ToolCalls) > 0 {
			fmt.Printf("🔧 Agent wants to use %d tool(s)\n", len(response.ToolCalls))
			// Add assistant's response that includes tool calls
			// Note: When there are tool calls, content might be empty, but we still need the assistant message
			assistantContent := response.Content
			if assistantContent == "" && len(response.ToolCalls) > 0 {
				assistantContent = " " // OpenAI API requires non-empty content
			}
			conversationMessages = append(conversationMessages, llm.Message{
				Role:      "assistant",
				Content:   assistantContent,
				ToolCalls: response.ToolCalls, // Include tool calls in the message
			})

			// Execute tools and get results
			toolResults := e.executeTools(ctx, response.ToolCalls)
			totalUsage.ToolCalls += len(response.ToolCalls)

			// Add tool results to conversation
			for _, result := range toolResults {
				toolMessage := e.formatToolResult(result)
				conversationMessages = append(conversationMessages, toolMessage)
			}

			// Continue iteration to let LLM process tool results
			continue
		}

		// Step 2f: Handle completion (no tool calls)
		if response.FinishReason == "stop" || response.FinishReason == "length" {
			fmt.Printf("✅ Agent completed task (reason: %s)\n", response.FinishReason)
			if response.Content != "" {
				fmt.Printf("💬 Final response: %s\n", response.Content)
			}
			// Agent has completed the task
			finalResponse = response.Content

			// Add final assistant response to conversation
			conversationMessages = append(conversationMessages, llm.Message{
				Role:    "assistant",
				Content: response.Content,
			})

			break
		}

		// Handle other finish reasons
		if response.FinishReason != "" {
			finalResponse = response.Content
			break
		}

		// If no clear finish reason, treat as completion
		finalResponse = response.Content
		conversationMessages = append(conversationMessages, llm.Message{
			Role:    "assistant",
			Content: response.Content,
		})
		break
	}

	// Check if we exceeded max iterations
	if finalResponse == "" {
		return nil, ErrMaxIterationsExceeded
	}

	// Step 3: Save conversation to session
	err := e.saveConversationToSession(agentSession, request.Input, finalResponse)
	if err != nil {
		// Log error but don't fail the entire execution
		fmt.Printf("Warning: failed to save conversation to session: %v\n", err)
	} else {
		totalUsage.SessionWrites = 1
	}

	// Step 4: Return execution result
	return &ExecutionResult{
		FinalOutput: finalResponse,
		SessionID:   agentSession.ID(),
		Session:     agentSession,
		Usage:       totalUsage,
		Metadata: map[string]any{
			"total_iterations": len(conversationMessages),
			"tools_called":     totalUsage.ToolCalls,
			"completion_time":  time.Now(),
		},
	}, nil
}

// executeTools handles tool execution within an iteration
func (e *engine) executeTools(ctx context.Context, toolCalls []tool.Call) []ToolResult {
	var results []ToolResult

	// Execute each tool call
	for i, call := range toolCalls {
		fmt.Printf("  🛠️  [Tool %d/%d] Calling: %s\n", i+1, len(toolCalls), call.Function.Name)
		fmt.Printf("  📋 Arguments: %s\n", call.Function.Arguments)

		// Execute tool using registry
		result, err := e.toolRegistry.Execute(ctx, call)

		if err != nil {
			fmt.Printf("  ❌ Tool execution failed: %v\n", err)
		} else {
			// Truncate result if too long for display
			resultStr := fmt.Sprintf("%v", result)
			if len(resultStr) > 200 {
				resultStr = resultStr[:200] + "..."
			}
			fmt.Printf("  ✅ Tool result: %s\n", resultStr)
		}

		// Create tool result
		toolResult := ToolResult{
			Call:   call,
			Result: result,
			Error:  err,
		}

		results = append(results, toolResult)
	}

	return results
}

// buildLLMMessages constructs the message array for LLM requests using PromptTemplate
func (e *engine) buildLLMMessages(contexts []agentcontext.Context, request Request) []llm.Message {
	// Use PromptTemplate if available
	if e.promptTemplate != nil {
		// Convert contexts to providers for template rendering
		providers := e.contextsToProviders(contexts)

		// Use template to render messages
		messages, err := e.promptTemplate.Render(context.Background(), providers, nil, request.Input)
		if err == nil {
			return messages
		}
		// Fall back to hardcoded format if template fails
		fmt.Printf("Warning: PromptTemplate render failed: %v, falling back to hardcoded format\n", err)
	}

	// Fallback: hardcoded format (original logic)
	var messages []llm.Message

	// Step 1: Add system message (hardcoded format)
	systemMessage := e.buildSystemMessage(contexts)
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: systemMessage,
	})

	// Step 2: Add conversation history from contexts
	historyMessages := e.buildHistoryMessages(contexts)
	messages = append(messages, historyMessages...)

	// Step 3: Add current user input
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: request.Input,
	})

	return messages
}

// contextsToProviders converts contexts to providers for template rendering
func (e *engine) contextsToProviders(contexts []agentcontext.Context) []agentcontext.Provider {
	// For now, create a simple provider that contains all contexts
	// This could be enhanced to convert different context types to different providers
	return []agentcontext.Provider{
		&contextProvider{contexts: contexts},
	}
}

// contextProvider wraps contexts as a provider
type contextProvider struct {
	contexts []agentcontext.Context
}

func (p *contextProvider) Type() string {
	return "aggregated_contexts"
}

func (p *contextProvider) Provide(ctx context.Context, session session.Session) []agentcontext.Context {
	return p.contexts
}

// buildSystemMessage creates a hardcoded system prompt with context information
func (e *engine) buildSystemMessage(contexts []agentcontext.Context) string {
	systemPrompt := `You are a helpful AI agent. Follow these guidelines:

1. Be concise and helpful in your responses
2. Use available tools when needed to provide accurate information
3. If you need to use tools, explain what you're doing
4. Always strive to be accurate and truthful

`

	// Check if we have history contexts and add history notice
	if e.hasHistoryContexts(contexts) {
		systemPrompt += `Note on Conversation History:
The conversation history provided may have been compressed or summarized to save space.
Key information and context have been preserved, but some details might be condensed.
Please use this history as reference for maintaining conversation continuity and context.

`
	}

	// Add system contexts (non-history contexts)
	var systemContexts []string
	for _, ctx := range contexts {
		// Skip history-type contexts as they'll be added as separate messages
		if ctx.Type == agentcontext.TypeUser || ctx.Type == agentcontext.TypeAssistant ||
			ctx.Type == agentcontext.TypeToolCall || ctx.Type == agentcontext.TypeToolResult {
			continue
		}

		// Add system, task, and other contextual information
		if ctx.Content != "" {
			systemContexts = append(systemContexts, ctx.Content)
		}
	}

	// Append additional context information to system prompt
	if len(systemContexts) > 0 {
		systemPrompt += "Additional Context:\n"
		for i, ctxContent := range systemContexts {
			systemPrompt += fmt.Sprintf("%d. %s\n", i+1, ctxContent)
		}
		systemPrompt += "\n"
	}

	systemPrompt += "Please provide helpful responses based on the above context."

	return systemPrompt
}

// hasHistoryContexts checks if there are any history-related contexts
func (e *engine) hasHistoryContexts(contexts []agentcontext.Context) bool {
	for _, ctx := range contexts {
		if ctx.Type == agentcontext.TypeUser || ctx.Type == agentcontext.TypeAssistant ||
			ctx.Type == agentcontext.TypeToolCall || ctx.Type == agentcontext.TypeToolResult {
			return true
		}
	}
	return false
}

// buildHistoryMessages extracts conversation history from contexts and converts to messages
func (e *engine) buildHistoryMessages(contexts []agentcontext.Context) []llm.Message {
	var messages []llm.Message

	for _, ctx := range contexts {
		// Process message-type contexts (from HistoryProvider)
		switch ctx.Type {
		case agentcontext.TypeUser, agentcontext.TypeAssistant, agentcontext.TypeSystem:
			// Direct message types
			messages = append(messages, llm.Message{
				Role:    ctx.Type,
				Content: ctx.Content,
			})

		case agentcontext.TypeToolCall:
			// Tool call context - convert to assistant message with tool calls
			// Note: This is a placeholder; actual tool call handling happens elsewhere
			messages = append(messages, llm.Message{
				Role:    "assistant",
				Content: " ", // OpenAI requires non-empty content
				// ToolCalls would be populated from metadata if needed
			})

		case agentcontext.TypeToolResult:
			// Tool result - convert to tool message
			toolCallID := ""
			if ctx.Metadata != nil {
				if id, ok := ctx.Metadata["tool_call_id"].(string); ok {
					toolCallID = id
				}
			}
			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    ctx.Content,
				ToolCallID: toolCallID,
			})

		default:
			// Skip other context types (they belong in system message)
			continue
		}
	}

	return messages
}

// formatToolResult converts a ToolResult to an LLM message
func (e *engine) formatToolResult(result ToolResult) llm.Message {
	var content string

	if result.Error != nil {
		// Format error result
		content = fmt.Sprintf("Tool '%s' execution failed: %s", result.Call.Function.Name, result.Error.Error())
	} else {
		// Format successful result
		resultStr := ""
		if result.Result != nil {
			resultStr = fmt.Sprintf("%v", result.Result)
		}
		content = fmt.Sprintf("Tool '%s' executed successfully. Result: %s", result.Call.Function.Name, resultStr)
	}

	return llm.Message{
		Role:       "tool",
		Content:    content,
		ToolCallID: result.Call.ID, // Link back to the tool call
	}
}

// saveConversationToSession saves the user input and agent response to session history
func (e *engine) saveConversationToSession(agentSession session.Session, userInput, agentResponse string) error {
	// Add user message entry
	userEntry := session.NewMessageEntry("user", userInput)
	agentSession.AddEntry(userEntry)

	// Add assistant response entry
	assistantEntry := session.NewMessageEntry("assistant", agentResponse)
	agentSession.AddEntry(assistantEntry)

	// Update session metadata
	agentSession.Set("last_interaction", time.Now().Format(time.RFC3339))
	agentSession.Set("total_messages", len(agentSession.GetHistory(1000))) // Get large history to count all

	return nil
}
