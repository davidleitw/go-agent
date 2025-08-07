package prompt

import (
	"testing"
)

func TestParseTemplate_SimpleVariables(t *testing.T) {
	template := "{{system}}\n{{history}}\n{{user_input}}"
	sections := parseTemplate(template)

	// The parser filters out whitespace-only text sections
	expected := []section{
		{Type: "variable", Content: "system", Metadata: map[string]string{}},
		{Type: "variable", Content: "history", Metadata: map[string]string{}},
		{Type: "variable", Content: "user_input", Metadata: map[string]string{}},
	}

	if len(sections) != len(expected) {
		t.Fatalf("Expected %d sections, got %d", len(expected), len(sections))
	}

	for i, section := range sections {
		if section.Type != expected[i].Type {
			t.Errorf("Section %d type: expected %s, got %s", i, expected[i].Type, section.Type)
		}
		if section.Content != expected[i].Content {
			t.Errorf("Section %d content: expected %s, got %s", i, expected[i].Content, section.Content)
		}
	}
}

func TestParseTemplate_NamedVariables(t *testing.T) {
	template := "{{provider:name1}}\n{{provider:name2}}"
	sections := parseTemplate(template)

	// Only variables, whitespace is filtered out
	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	// Check first named variable
	if sections[0].Type != "variable" || sections[0].Content != "provider" {
		t.Errorf("First variable incorrect: %+v", sections[0])
	}
	if sections[0].Metadata["name"] != "name1" {
		t.Errorf("First variable name incorrect: %s", sections[0].Metadata["name"])
	}

	// Check second named variable
	if sections[1].Type != "variable" || sections[1].Content != "provider" {
		t.Errorf("Second variable incorrect: %+v", sections[1])
	}
	if sections[1].Metadata["name"] != "name2" {
		t.Errorf("Second variable name incorrect: %s", sections[1].Metadata["name"])
	}
}

func TestParseTemplate_MixedContent(t *testing.T) {
	template := "Instructions:\n{{system}}\nHistory:\n{{history:recent}}\nQuestion: {{user_input}}"
	sections := parseTemplate(template)

	expected := []section{
		{Type: "text", Content: "Instructions:\n"},
		{Type: "variable", Content: "system"},
		{Type: "text", Content: "\nHistory:\n"},
		{Type: "variable", Content: "history", Metadata: map[string]string{"name": "recent"}},
		{Type: "text", Content: "\nQuestion: "},
		{Type: "variable", Content: "user_input"},
	}

	if len(sections) != len(expected) {
		t.Fatalf("Expected %d sections, got %d", len(expected), len(sections))
	}

	for i, section := range sections {
		if section.Type != expected[i].Type {
			t.Errorf("Section %d type: expected %s, got %s", i, expected[i].Type, section.Type)
		}
		if section.Content != expected[i].Content {
			t.Errorf("Section %d content: expected %s, got %s", i, expected[i].Content, section.Content)
		}
		if expected[i].Metadata != nil {
			for k, v := range expected[i].Metadata {
				if section.Metadata[k] != v {
					t.Errorf("Section %d metadata[%s]: expected %s, got %s", i, k, v, section.Metadata[k])
				}
			}
		}
	}
}

func TestParseTemplate_OnlyText(t *testing.T) {
	template := "This is only text with no variables"
	sections := parseTemplate(template)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	if sections[0].Type != "text" {
		t.Errorf("Expected text section, got %s", sections[0].Type)
	}

	if sections[0].Content != template {
		t.Errorf("Text content incorrect: expected %s, got %s", template, sections[0].Content)
	}
}

func TestParseTemplate_OnlyVariables(t *testing.T) {
	template := "{{var1}}{{var2}}{{var3}}"
	sections := parseTemplate(template)

	if len(sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(sections))
	}

	expectedVars := []string{"var1", "var2", "var3"}
	for i, section := range sections {
		if section.Type != "variable" {
			t.Errorf("Section %d should be variable, got %s", i, section.Type)
		}
		if section.Content != expectedVars[i] {
			t.Errorf("Variable %d: expected %s, got %s", i, expectedVars[i], section.Content)
		}
	}
}

func TestParseTemplate_EmptyTemplate(t *testing.T) {
	template := ""
	sections := parseTemplate(template)

	if len(sections) != 0 {
		t.Fatalf("Expected 0 sections for empty template, got %d", len(sections))
	}
}

func TestParseTemplate_WhitespaceHandling(t *testing.T) {
	template := "   \n\n   {{system}}   \n\n   "
	sections := parseTemplate(template)

	// Whitespace-only sections are filtered out
	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	// Only the variable should remain
	if sections[0].Type != "variable" || sections[0].Content != "system" {
		t.Errorf("Variable section incorrect: %+v", sections[0])
	}
}

func TestParseTemplate_ComplexNames(t *testing.T) {
	template := "{{project_info:main_config}}\n{{user_preferences:ui_settings}}"
	sections := parseTemplate(template)

	// Only variables, no whitespace sections
	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	// Check first variable with underscore in both type and name
	if sections[0].Content != "project_info" {
		t.Errorf("First variable type incorrect: %s", sections[0].Content)
	}
	if sections[0].Metadata["name"] != "main_config" {
		t.Errorf("First variable name incorrect: %s", sections[0].Metadata["name"])
	}

	// Check second variable
	if sections[1].Content != "user_preferences" {
		t.Errorf("Second variable type incorrect: %s", sections[1].Content)
	}
	if sections[1].Metadata["name"] != "ui_settings" {
		t.Errorf("Second variable name incorrect: %s", sections[1].Metadata["name"])
	}
}

func TestParseTemplate_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		expected int // expected number of sections
	}{
		{
			name:     "Variable at start",
			template: "{{system}}text after",
			expected: 2,
		},
		{
			name:     "Variable at end",
			template: "text before{{system}}",
			expected: 2,
		},
		{
			name:     "Adjacent variables",
			template: "{{system}}{{history}}",
			expected: 2,
		},
		{
			name:     "Variable with spaces around name",
			template: "{{ system }}",
			expected: 1, // Should still parse as single variable
		},
		{
			name:     "Empty variable name",
			template: "{{}}",
			expected: 1, // Should be treated as text since regex won't match
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sections := parseTemplate(tc.template)
			if len(sections) != tc.expected {
				t.Errorf("Expected %d sections, got %d for template: %s", tc.expected, len(sections), tc.template)
			}
		})
	}
}

func TestParseTemplate_SpecialCharacters(t *testing.T) {
	template := "{{provider-with-dashes}}\n{{provider_with_underscores}}\n{{provider123}}"
	sections := parseTemplate(template)

	// Only variables, whitespace filtered out
	if len(sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(sections))
	}

	// Check variable names are preserved
	vars := []string{"provider-with-dashes", "provider_with_underscores", "provider123"}

	for i, expectedVar := range vars {
		if sections[i].Type != "variable" {
			t.Errorf("Section %d should be variable", i)
		}
		if sections[i].Content != expectedVar {
			t.Errorf("Variable %d: expected %s, got %s", i, expectedVar, sections[i].Content)
		}
	}
}

func TestParse_ReturnsTemplate(t *testing.T) {
	template := "{{system}}\n{{user_input}}"
	parsed := Parse(template)

	// Check that Parse returns a Template interface
	if parsed == nil {
		t.Fatal("Parse returned nil")
	}

	// Check that the template can return its original string
	if parsed.String() != template {
		t.Errorf("Template string mismatch: expected %s, got %s", template, parsed.String())
	}

	// Check that variables are correctly identified
	variables := parsed.Variables()
	expected := []string{"system", "user_input"}

	if len(variables) != len(expected) {
		t.Fatalf("Expected %d variables, got %d", len(expected), len(variables))
	}

	for i, v := range variables {
		if v != expected[i] {
			t.Errorf("Variable %d: expected %s, got %s", i, expected[i], v)
		}
	}
}

// TestValidateTemplate_ValidTemplates has been removed as validateTemplate() function
// was deleted during cleanup. Template validation is now done implicitly by Parse()

// TestValidateTemplate_InvalidTemplates has been removed as validateTemplate() function
// was deleted during cleanup. Template validation is now done implicitly by Parse()

// TestTemplateError has been removed as TemplateError type was deleted during cleanup
