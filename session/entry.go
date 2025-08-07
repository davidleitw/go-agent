package session

import (
	"time"

	"github.com/google/uuid"
)

// EntryType defines the type of history entry
type EntryType string

const (
	EntryTypeMessage    EntryType = "message"
	EntryTypeToolCall   EntryType = "tool_call"
	EntryTypeToolResult EntryType = "tool_result"
	EntryTypeThinking   EntryType = "thinking"
)

// Entry represents a unified history record structure
type Entry struct {
	ID        string         `json:"id"`
	Type      EntryType      `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Content   any            `json:"content"`
	Metadata  map[string]any `json:"metadata"`
}

// MessageContent represents a message entry content
type MessageContent struct {
	Role string `json:"role"` // user/assistant/system
	Text string `json:"text"`
}

// ToolCallContent represents a tool call entry content
type ToolCallContent struct {
	Tool       string         `json:"tool"`
	Parameters map[string]any `json:"parameters"`
}

// ToolResultContent represents a tool result entry content
type ToolResultContent struct {
	Tool    string `json:"tool"`
	Success bool   `json:"success"`
	Result  any    `json:"result"`
	Error   string `json:"error"` // error message if failed
}

// NewMessageEntry creates a new message entry
func NewMessageEntry(role, text string) Entry {
	return Entry{
		ID:        uuid.New().String(),
		Type:      EntryTypeMessage,
		Timestamp: time.Now(),
		Content: MessageContent{
			Role: role,
			Text: text,
		},
		Metadata: make(map[string]any),
	}
}

// NewToolCallEntry creates a new tool call entry
func NewToolCallEntry(tool string, params map[string]any) Entry {
	return Entry{
		ID:        uuid.New().String(),
		Type:      EntryTypeToolCall,
		Timestamp: time.Now(),
		Content: ToolCallContent{
			Tool:       tool,
			Parameters: params,
		},
		Metadata: make(map[string]any),
	}
}

// NewToolResultEntry creates a new tool result entry
func NewToolResultEntry(tool string, result any, err error) Entry {
	content := ToolResultContent{
		Tool:    tool,
		Success: err == nil,
		Result:  result,
	}

	if err != nil {
		content.Error = err.Error()
	}

	return Entry{
		ID:        uuid.New().String(),
		Type:      EntryTypeToolResult,
		Timestamp: time.Now(),
		Content:   content,
		Metadata:  make(map[string]any),
	}
}

// GetMessageContent extracts MessageContent from an entry
func GetMessageContent(entry Entry) (MessageContent, bool) {
	if entry.Type != EntryTypeMessage {
		return MessageContent{}, false
	}
	content, ok := entry.Content.(MessageContent)
	return content, ok
}

// GetToolCallContent extracts ToolCallContent from an entry
func GetToolCallContent(entry Entry) (ToolCallContent, bool) {
	if entry.Type != EntryTypeToolCall {
		return ToolCallContent{}, false
	}
	content, ok := entry.Content.(ToolCallContent)
	return content, ok
}

// GetToolResultContent extracts ToolResultContent from an entry
func GetToolResultContent(entry Entry) (ToolResultContent, bool) {
	if entry.Type != EntryTypeToolResult {
		return ToolResultContent{}, false
	}
	content, ok := entry.Content.(ToolResultContent)
	return content, ok
}
