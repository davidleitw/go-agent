package prompt

// Builder provides a fluent API for building prompt templates
type Builder interface {
	// Core variables for common use cases
	System() Builder
	History() Builder
	UserInput() Builder

	// Custom provider variables
	Provider(providerType string) Builder
	NamedProvider(providerType, name string) Builder

	// Static text content
	Text(content string) Builder
	Line(content string) Builder // Text with newline

	// Convenience methods
	DefaultFlow() Builder // Add standard flow: system -> history -> user_input
	Separator() Builder   // Add a visual separator

	// Build the template
	Build() Template
}

// TemplateBuilder implements the Builder interface
type TemplateBuilder struct {
	sections []section
}

// New creates a new template builder
func New() Builder {
	return &TemplateBuilder{
		sections: make([]section, 0),
	}
}

// System adds a system variable section
func (b *TemplateBuilder) System() Builder {
	return b.addVariable("system")
}

// History adds a history variable section
func (b *TemplateBuilder) History() Builder {
	return b.addVariable("history")
}

// UserInput adds a user_input variable section
func (b *TemplateBuilder) UserInput() Builder {
	return b.addVariable("user_input")
}

// Provider adds a custom provider variable section
func (b *TemplateBuilder) Provider(providerType string) Builder {
	return b.addVariable(providerType)
}

// NamedProvider adds a named provider variable section
func (b *TemplateBuilder) NamedProvider(providerType, name string) Builder {
	section := section{
		Type:     "variable",
		Content:  providerType,
		Metadata: map[string]string{"name": name},
	}
	b.sections = append(b.sections, section)
	return b
}

// Text adds static text content
func (b *TemplateBuilder) Text(content string) Builder {
	if content != "" {
		b.sections = append(b.sections, section{
			Type:    "text",
			Content: content,
		})
	}
	return b
}

// Line adds text content with a trailing newline
func (b *TemplateBuilder) Line(content string) Builder {
	return b.Text(content + "\n")
}

// DefaultFlow adds the standard prompt flow
func (b *TemplateBuilder) DefaultFlow() Builder {
	return b.System().
		History().
		Text("Context information:\n").
		Provider("context_providers").
		UserInput()
}

// Separator adds a visual separator
func (b *TemplateBuilder) Separator() Builder {
	return b.Text("\n---\n")
}

// Build creates the final Template
func (b *TemplateBuilder) Build() Template {
	return &promptTemplate{
		sections: b.sections,
		original: "", // Not needed for core functionality
	}
}

// addVariable is a helper method to add variable sections
func (b *TemplateBuilder) addVariable(varName string) Builder {
	b.sections = append(b.sections, section{
		Type:     "variable",
		Content:  varName,
		Metadata: make(map[string]string),
	})
	return b
}

