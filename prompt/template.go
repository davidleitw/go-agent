package prompt

import (
	"context"
	"fmt"
	"strings"

	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/session"
)

// Template represents a prompt template that can render contexts into LLM messages
type Template interface {
	// Render converts contexts from providers into LLM messages using the template
	Render(ctx context.Context, providers []agentcontext.Provider, session session.Session, userInput string) ([]llm.Message, error)


	// Variables returns the list of variables used in the template
	Variables() []string

	// Explain returns a human-readable description of the template structure
	Explain() string

	// String returns the original template representation
	String() string
}

// section represents a part of the template
type section struct {
	Type     string            // "variable" or "text"
	Content  string            // variable name or text content
	Metadata map[string]string // additional metadata (e.g., for named references)
}

// promptTemplate implements Template interface
type promptTemplate struct {
	sections []section
	original string
}

// Render implements Template.Render
func (t *promptTemplate) Render(ctx context.Context, providers []agentcontext.Provider, s session.Session, userInput string) ([]llm.Message, error) {
	var messages []llm.Message

	for _, section := range t.sections {
		switch section.Type {
		case "variable":
			if section.Content == "user_input" {
				// Handle user_input specially
				if userInput != "" {
					messages = append(messages, llm.Message{
						Role:    "user",
						Content: userInput,
					})
				}
			} else {
				// Gather contexts for this variable
				contexts := t.gatherContexts(section, providers, ctx, s)

				// Render contexts to messages
				msgs := t.renderContexts(section.Content, contexts)
				messages = append(messages, msgs...)
			}

		case "text":
			// Static text becomes a system message
			if strings.TrimSpace(section.Content) != "" {
				messages = append(messages, llm.Message{
					Role:    "system",
					Content: strings.TrimSpace(section.Content),
				})
			}
		}
	}

	return messages, nil
}

// gatherContexts collects contexts from providers that match the variable
func (t *promptTemplate) gatherContexts(section section, providers []agentcontext.Provider, ctx context.Context, s session.Session) []agentcontext.Context {
	var contexts []agentcontext.Context

	providerType := section.Content
	providerName := section.Metadata["name"]

	for _, provider := range providers {
		// Check if provider type matches
		if provider.Type() == providerType {
			// If a specific name is required, check if provider supports it
			if providerName != "" {
				if namedProvider, ok := provider.(NamedProvider); ok {
					if namedProvider.Name() != providerName {
						continue
					}
				} else {
					continue // Skip providers that don't support naming
				}
			}

			// Collect contexts from this provider
			providerContexts := provider.Provide(ctx, s)
			contexts = append(contexts, providerContexts...)
		}
	}

	return contexts
}

// renderContexts converts contexts to LLM messages based on variable type
func (t *promptTemplate) renderContexts(variableType string, contexts []agentcontext.Context) []llm.Message {
	if len(contexts) == 0 {
		return nil
	}

	switch variableType {
	case "system":
		return t.renderSystemContexts(contexts)
	case "history":
		return t.renderHistoryContexts(contexts)
	case "user_input":
		// user_input is handled specially in main render loop
		return nil
	default:
		// Custom variables default to system role
		return t.renderCustomContexts(contexts)
	}
}

// renderSystemContexts combines all system contexts into system messages
func (t *promptTemplate) renderSystemContexts(contexts []agentcontext.Context) []llm.Message {
	var contents []string
	for _, ctx := range contexts {
		if ctx.Content != "" {
			contents = append(contents, ctx.Content)
		}
	}

	if len(contents) == 0 {
		return nil
	}

	return []llm.Message{{
		Role:    "system",
		Content: strings.Join(contents, "\n\n"),
	}}
}

// renderHistoryContexts preserves original message roles from history
func (t *promptTemplate) renderHistoryContexts(contexts []agentcontext.Context) []llm.Message {
	var messages []llm.Message

	for _, ctx := range contexts {
		// Try to get original role from metadata
		role := "user" // default role
		if originalRole, exists := ctx.Metadata["original_role"]; exists {
			if roleStr, ok := originalRole.(string); ok {
				role = roleStr
			}
		}

		if ctx.Content != "" {
			messages = append(messages, llm.Message{
				Role:    role,
				Content: ctx.Content,
			})
		}
	}

	return messages
}

// renderCustomContexts renders custom contexts as system messages
func (t *promptTemplate) renderCustomContexts(contexts []agentcontext.Context) []llm.Message {
	var contents []string
	for _, ctx := range contexts {
		if ctx.Content != "" {
			contents = append(contents, ctx.Content)
		}
	}

	if len(contents) == 0 {
		return nil
	}

	return []llm.Message{{
		Role:    "system",
		Content: strings.Join(contents, "\n"),
	}}
}

// Variables implements Template.Variables
func (t *promptTemplate) Variables() []string {
	var variables []string
	seen := make(map[string]bool)

	for _, section := range t.sections {
		if section.Type == "variable" && !seen[section.Content] {
			variables = append(variables, section.Content)
			seen[section.Content] = true
		}
	}

	return variables
}

// Explain implements Template.Explain
func (t *promptTemplate) Explain() string {
	var parts []string
	for i, section := range t.sections {
		if section.Type == "variable" {
			if name, hasName := section.Metadata["name"]; hasName {
				parts = append(parts, fmt.Sprintf("%d. {{%s:%s}}", i+1, section.Content, name))
			} else {
				parts = append(parts, fmt.Sprintf("%d. {{%s}}", i+1, section.Content))
			}
		} else {
			parts = append(parts, fmt.Sprintf("%d. Text: %q", i+1, section.Content))
		}
	}
	return fmt.Sprintf("Template sections:\n%s", strings.Join(parts, "\n"))
}

// String implements Template.String
func (t *promptTemplate) String() string {
	return t.original
}

// NamedProvider is an optional interface for providers that support named references
type NamedProvider interface {
	agentcontext.Provider
	Name() string
}
