// Package schema provides structured input field definitions for intelligent information collection.
//
// This package allows agents to define expected input fields and automatically collect
// missing information from users through natural conversation.
//
// Basic usage:
//   schema.Define("email", "Please provide your email address")
//
// Optional fields:
//   schema.Define("phone", "Contact number (optional)").Optional()
package schema

// Field represents an expected input field that the agent should collect from user input.
// Each field has a name for identification, a prompt for collection, and a required flag.
type Field struct {
	name     string // Internal field name for identification
	prompt   string // Human-readable prompt to request this information
	required bool   // Whether this field is required (default: true)
}

// Define creates a new required input field with the given name and collection prompt.
// The prompt will be used by the LLM to naturally ask for this information when missing.
//
// Example:
//   email := schema.Define("email", "Please provide your email address")
//   issue := schema.Define("issue_description", "Please describe your issue in detail")
func Define(name, prompt string) *Field {
	return &Field{
		name:     name,
		prompt:   prompt,
		required: true,
	}
}

// Optional marks this field as optional, meaning the agent will ask for it once
// but won't insist if the user doesn't provide it.
//
// Example:
//   phone := schema.Define("phone", "Contact number for urgent matters").Optional()
func (f *Field) Optional() *Field {
	f.required = false
	return f
}

// Name returns the internal field name used for identification.
// This method is used internally by the agent package.
func (f *Field) Name() string {
	return f.name
}

// Prompt returns the human-readable prompt used to request this information.
// This method is used internally by the agent package.
func (f *Field) Prompt() string {
	return f.prompt
}

// Required returns whether this field is required.
// This method is used internally by the agent package.
func (f *Field) Required() bool {
	return f.required
}