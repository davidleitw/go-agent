package context

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidleitw/go-agent/session"
)

type Provider interface {
	Provide(ctx context.Context, s session.Session) []Context
}

type SystemPromptProvider struct {
	systemPrompt string
}

func NewSystemPromptProvider(systemPrompt string) Provider {
	return &SystemPromptProvider{
		systemPrompt: systemPrompt,
	}
}

func (p *SystemPromptProvider) Provide(ctx context.Context, s session.Session) []Context {
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

func (p *HistoryProvider) Provide(ctx context.Context, s session.Session) []Context {
	history := s.GetHistory(p.limit)

	entries := make([]Context, len(history))
	for i, entry := range history {
		contextEntry := Context{
			Metadata: make(map[string]any),
		}

		// Copy entry metadata and add timestamp
		for k, v := range entry.Metadata {
			contextEntry.Metadata[k] = v
		}
		contextEntry.Metadata["entry_id"] = entry.ID
		contextEntry.Metadata["timestamp"] = entry.Timestamp

		// Convert based on entry type
		switch entry.Type {
		case session.EntryTypeMessage:
			if content, ok := session.GetMessageContent(entry); ok {
				contextEntry.Type = content.Role
				contextEntry.Content = content.Text
			}

		case session.EntryTypeToolCall:
			if content, ok := session.GetToolCallContent(entry); ok {
				contextEntry.Type = "tool_call"
				params, _ := json.Marshal(content.Parameters)
				contextEntry.Content = fmt.Sprintf("Tool: %s\nParameters: %s", content.Tool, string(params))
				contextEntry.Metadata["tool_name"] = content.Tool
			}

		case session.EntryTypeToolResult:
			if content, ok := session.GetToolResultContent(entry); ok {
				contextEntry.Type = "tool_result"
				if content.Success {
					result, _ := json.Marshal(content.Result)
					contextEntry.Content = fmt.Sprintf("Tool: %s\nSuccess: true\nResult: %s", content.Tool, string(result))
				} else {
					contextEntry.Content = fmt.Sprintf("Tool: %s\nSuccess: false\nError: %s", content.Tool, content.Error)
				}
				contextEntry.Metadata["tool_name"] = content.Tool
				contextEntry.Metadata["success"] = content.Success
			}

		case session.EntryTypeThinking:
			contextEntry.Type = "thinking"
			if str, ok := entry.Content.(string); ok {
				contextEntry.Content = str
			} else {
				content, _ := json.Marshal(entry.Content)
				contextEntry.Content = string(content)
			}

		default:
			// Fallback for unknown types
			contextEntry.Type = string(entry.Type)
			content, _ := json.Marshal(entry.Content)
			contextEntry.Content = string(content)
		}

		entries[i] = contextEntry
	}

	return entries
}
