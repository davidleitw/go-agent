package prompt

import (
	"regexp"
	"strings"
)

// Parse parses a string template and returns a Template
func Parse(template string) Template {
	sections := parseTemplate(template)
	return &promptTemplate{
		sections: sections,
		original: template,
	}
}

// parseTemplate parses a template string into sections
// Supports two variable formats:
// {{provider_type}} - type reference
// {{provider_type:name}} - named reference
func parseTemplate(template string) []section {
	var sections []section

	// Regular expression to match {{variable}} and {{variable:name}}
	// Group 1: variable type, Group 2: optional name
	re := regexp.MustCompile(`\{\{([^}:]+)(?::([^}]+))?\}\}`)

	lastEnd := 0
	matches := re.FindAllStringSubmatchIndex(template, -1)

	for _, match := range matches {
		start, end := match[0], match[1]
		typeStart, typeEnd := match[2], match[3]
		nameStart, nameEnd := match[4], match[5]

		// Add text before the variable as a text section
		if start > lastEnd {
			text := template[lastEnd:start]
			if strings.TrimSpace(text) != "" {
				sections = append(sections, section{
					Type:    "text",
					Content: text,
				})
			}
		}

		// Parse the variable
		providerType := template[typeStart:typeEnd]
		section := section{
			Type:     "variable",
			Content:  providerType,
			Metadata: make(map[string]string),
		}

		// Handle named reference if present
		if nameStart >= 0 && nameEnd >= 0 {
			providerName := template[nameStart:nameEnd]
			section.Metadata["name"] = providerName
		}

		sections = append(sections, section)
		lastEnd = end
	}

	// Add remaining text after the last variable
	if lastEnd < len(template) {
		text := template[lastEnd:]
		if strings.TrimSpace(text) != "" {
			sections = append(sections, section{
				Type:    "text",
				Content: text,
			})
		}
	}

	return sections
}

