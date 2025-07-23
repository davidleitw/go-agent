package context

// Context Type constants for commonly used types
const (
	// Message types (for conversation history)
	TypeUser      = "user"
	TypeAssistant = "assistant"
	TypeSystem    = "system"
	TypeTool      = "tool"
	
	// Tool-related types (for history tracking)
	TypeToolCall   = "tool_call"
	TypeToolResult = "tool_result"
	
	// Special types (for advanced use cases)
	TypeThinking   = "thinking"
	TypeSummary    = "summary"
)

type Context struct {
	Type     string
	Content  string
	Metadata map[string]any
}
