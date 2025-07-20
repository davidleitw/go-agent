package context

import (
	"encoding/json"
	"fmt"

	"github.com/davidleitw/go-agent/session"
)

type Provider interface {
	Provide(s session.Session) []Context
}

type SystemPromptProvider struct {
	systemPrompt string
}

func NewSystemPromptProvider(systemPrompt string) Provider {
	return &SystemPromptProvider{
		systemPrompt: systemPrompt,
	}
}

func (p *SystemPromptProvider) Provide(s session.Session) []Context {
	return []Context{
		{
			Type:     "system",
			Content:  p.systemPrompt,
			Metadata: map[string]any{},
		},
	}
}

type HistoryProvider struct {
	limit int
}

func NewHistoryProvider(limit int) Provider {
	return &HistoryProvider{
		limit: limit,
	}
}

func (p *HistoryProvider) Provide(s session.Session) []Context {
	history := s.GetHistory(p.limit)

	entries := make([]Context, len(history))
	for i, entry := range history {
		ctx := Context{
			Metadata: make(map[string]any),
		}

		// Copy entry metadata and add timestamp
		for k, v := range entry.Metadata {
			ctx.Metadata[k] = v
		}
		ctx.Metadata["entry_id"] = entry.ID
		ctx.Metadata["timestamp"] = entry.Timestamp

		// Convert based on entry type
		switch entry.Type {
		case session.EntryTypeMessage:
			if content, ok := session.GetMessageContent(entry); ok {
				ctx.Type = content.Role
				ctx.Content = content.Text
			}

		case session.EntryTypeToolCall:
			if content, ok := session.GetToolCallContent(entry); ok {
				ctx.Type = "tool_call"
				params, _ := json.Marshal(content.Parameters)
				ctx.Content = fmt.Sprintf("Tool: %s\nParameters: %s", content.Tool, string(params))
				ctx.Metadata["tool_name"] = content.Tool
			}

		case session.EntryTypeToolResult:
			if content, ok := session.GetToolResultContent(entry); ok {
				ctx.Type = "tool_result"
				if content.Success {
					result, _ := json.Marshal(content.Result)
					ctx.Content = fmt.Sprintf("Tool: %s\nSuccess: true\nResult: %s", content.Tool, string(result))
				} else {
					ctx.Content = fmt.Sprintf("Tool: %s\nSuccess: false\nError: %s", content.Tool, content.Error)
				}
				ctx.Metadata["tool_name"] = content.Tool
				ctx.Metadata["success"] = content.Success
			}

		case session.EntryTypeThinking:
			ctx.Type = "thinking"
			if str, ok := entry.Content.(string); ok {
				ctx.Content = str
			} else {
				content, _ := json.Marshal(entry.Content)
				ctx.Content = string(content)
			}

		default:
			// Fallback for unknown types
			ctx.Type = string(entry.Type)
			content, _ := json.Marshal(entry.Content)
			ctx.Content = string(content)
		}

		entries[i] = ctx
	}

	return entries
}
