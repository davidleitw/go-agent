package agent

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// OutputType defines the expected structured output format for an Agent.
type OutputType interface {
	// Name returns the name of this output type
	Name() string

	// Description returns a description of this output type
	Description() string

	// Schema returns the JSON Schema for this output type
	Schema() map[string]interface{}

	// NewInstance returns a new, empty instance for deserialization
	NewInstance() interface{}

	// Validate validates that the given data matches this output type
	Validate(data interface{}) error
}

// OutputTypeBuilder provides a fluent API for creating OutputType instances
type OutputTypeBuilder struct {
	name        string
	description string
	schema      map[string]interface{}
	example     interface{}
	validator   func(interface{}) error
	typeInfo    reflect.Type
	err         error
}

// NewOutputType creates a new OutputTypeBuilder with the given name
func NewOutputType(name string) *OutputTypeBuilder {
	if name == "" {
		return &OutputTypeBuilder{err: fmt.Errorf("output type name cannot be empty")}
	}
	return &OutputTypeBuilder{
		name: name,
	}
}

// WithDescription sets the output type's description
func (b *OutputTypeBuilder) WithDescription(description string) *OutputTypeBuilder {
	if b.err != nil {
		return b
	}
	b.description = description
	return b
}

// WithSchema sets the JSON Schema for this output type
func (b *OutputTypeBuilder) WithSchema(schema map[string]interface{}) *OutputTypeBuilder {
	if b.err != nil {
		return b
	}
	if schema == nil {
		b.err = fmt.Errorf("schema cannot be nil")
		return b
	}
	b.schema = schema
	return b
}

// WithExample sets an example instance for this output type
// The schema will be automatically generated from the struct if not already set
func (b *OutputTypeBuilder) WithExample(example interface{}) *OutputTypeBuilder {
	if b.err != nil {
		return b
	}
	if example == nil {
		b.err = fmt.Errorf("example cannot be nil")
		return b
	}

	b.example = example
	b.typeInfo = reflect.TypeOf(example)

	// If schema not set, generate it from the struct
	if b.schema == nil {
		schema, err := generateSchemaFromStruct(b.typeInfo)
		if err != nil {
			b.err = fmt.Errorf("failed to generate schema from example: %w", err)
			return b
		}
		b.schema = schema
	}

	return b
}

// WithValidator sets a custom validation function
func (b *OutputTypeBuilder) WithValidator(validator func(interface{}) error) *OutputTypeBuilder {
	if b.err != nil {
		return b
	}
	b.validator = validator
	return b
}

// Build creates the OutputType instance
func (b *OutputTypeBuilder) Build() (OutputType, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	// Import cycle prevention: use factory function
	if outputTypeFactory == nil {
		return nil, fmt.Errorf("output type factory not initialized")
	}

	return outputTypeFactory(
		b.name,
		b.description,
		b.schema,
		b.example,
		b.typeInfo,
		b.validator,
	)
}

// outputTypeFactory is set by the internal package to avoid import cycles
var outputTypeFactory func(
	name string,
	description string,
	schema map[string]interface{},
	example interface{},
	typeInfo reflect.Type,
	validator func(interface{}) error,
) (OutputType, error)

// SetOutputTypeFactory is used by internal packages to register the output type factory
func SetOutputTypeFactory(factory func(
	name string,
	description string,
	schema map[string]interface{},
	example interface{},
	typeInfo reflect.Type,
	validator func(interface{}) error,
) (OutputType, error)) {
	outputTypeFactory = factory
}

// StructuredOutput is a convenience function that creates an OutputType from a struct instance
func StructuredOutput(name string, example interface{}) (OutputType, error) {
	return NewOutputType(name).
		WithExample(example).
		Build()
}

// generateSchemaFromStruct generates a JSON Schema from a Go struct type
func generateSchemaFromStruct(t reflect.Type) (map[string]interface{}, error) {
	// Dereference pointer types
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", t.Kind())
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	var required []string
	properties := schema["properties"].(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			// Parse JSON tag (name,options)
			parts := []string{jsonTag}
			if len(parts) > 0 && parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// Generate property schema
		propSchema, err := generatePropertySchema(field.Type, field.Tag)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for field %s: %w", field.Name, err)
		}

		properties[fieldName] = propSchema

		// Check if field is required
		if isRequired(field.Tag) {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// generatePropertySchema generates a JSON Schema for a single property
func generatePropertySchema(t reflect.Type, tag reflect.StructTag) (map[string]interface{}, error) {
	// Dereference pointer types
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := make(map[string]interface{})

	switch t.Kind() {
	case reflect.String:
		schema["type"] = "string"

		// Check for enum values
		if enumTag := tag.Get("enum"); enumTag != "" {
			// Parse comma-separated enum values
			var enumValues []interface{}
			for _, val := range parseCommaSeparated(enumTag) {
				enumValues = append(enumValues, val)
			}
			schema["enum"] = enumValues
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"

	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"

	case reflect.Bool:
		schema["type"] = "boolean"

	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		itemSchema, err := generatePropertySchema(t.Elem(), "")
		if err != nil {
			return nil, fmt.Errorf("failed to generate array item schema: %w", err)
		}
		schema["items"] = itemSchema

	case reflect.Map:
		schema["type"] = "object"
		if t.Key().Kind() == reflect.String {
			valueSchema, err := generatePropertySchema(t.Elem(), "")
			if err != nil {
				return nil, fmt.Errorf("failed to generate map value schema: %w", err)
			}
			schema["additionalProperties"] = valueSchema
		}

	case reflect.Struct:
		// Recursive struct handling
		structSchema, err := generateSchemaFromStruct(t)
		if err != nil {
			return nil, fmt.Errorf("failed to generate nested struct schema: %w", err)
		}
		return structSchema, nil

	case reflect.Interface:
		// For interface{}, allow any type
		schema = map[string]interface{}{}

	default:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind())
	}

	// Add description from tag
	if desc := tag.Get("description"); desc != "" {
		schema["description"] = desc
	}

	return schema, nil
}

// isRequired checks if a field is marked as required in struct tags
func isRequired(tag reflect.StructTag) bool {
	if validateTag := tag.Get("validate"); validateTag != "" {
		tags := parseCommaSeparated(validateTag)
		for _, t := range tags {
			if t == "required" {
				return true
			}
		}
	}

	if jsonTag := tag.Get("json"); jsonTag != "" {
		tags := parseCommaSeparated(jsonTag)
		for _, t := range tags {
			if t == "required" {
				return true
			}
		}
	}

	return false
}

// parseCommaSeparated parses a comma-separated string into a slice
func parseCommaSeparated(str string) []string {
	if str == "" {
		return nil
	}

	var result []string
	for _, part := range splitAndTrim(str, ",") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(str, sep string) []string {
	parts := []string{}
	for _, part := range parseCommaSeparatedHelper(str, sep) {
		trimmed := trimSpace(part)
		parts = append(parts, trimmed)
	}
	return parts
}

// Helper functions to avoid importing strings package
func parseCommaSeparatedHelper(str, sep string) []string {
	if str == "" {
		return nil
	}

	var parts []string
	start := 0
	for i := 0; i < len(str); i++ {
		if str[i:i+len(sep)] == sep && i+len(sep) <= len(str) {
			parts = append(parts, str[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, str[start:])
	return parts
}

func trimSpace(str string) string {
	start := 0
	end := len(str)

	// Trim from start
	for start < end && isSpace(str[start]) {
		start++
	}

	// Trim from end
	for end > start && isSpace(str[end-1]) {
		end--
	}

	return str[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// ValidateAgainstSchema validates data against a JSON Schema
func ValidateAgainstSchema(data interface{}, schema map[string]interface{}) error {
	// Convert data to JSON and back to ensure it's JSON-serializable
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	var unmarshaled interface{}
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		return fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	return validateValue(unmarshaled, schema)
}

// validateValue validates a value against a schema
func validateValue(value interface{}, schema map[string]interface{}) error {
	schemaType, ok := schema["type"].(string)
	if !ok {
		// If no type specified, validation passes
		return nil
	}

	switch schemaType {
	case "object":
		return validateObject(value, schema)
	case "array":
		return validateArray(value, schema)
	case "string":
		return validateString(value, schema)
	case "number":
		return validateNumber(value, schema)
	case "integer":
		return validateInteger(value, schema)
	case "boolean":
		return validateBoolean(value, schema)
	default:
		return fmt.Errorf("unknown schema type: %s", schemaType)
	}
}

func validateObject(value interface{}, schema map[string]interface{}) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", value)
	}

	// Check required fields
	if required, ok := schema["required"].([]interface{}); ok {
		for _, field := range required {
			fieldName, ok := field.(string)
			if !ok {
				continue
			}
			if _, exists := obj[fieldName]; !exists {
				return fmt.Errorf("required field '%s' is missing", fieldName)
			}
		}
	}

	// Validate properties
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for fieldName, fieldValue := range obj {
			if propSchema, exists := properties[fieldName]; exists {
				if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
					if err := validateValue(fieldValue, propSchemaMap); err != nil {
						return fmt.Errorf("field '%s': %w", fieldName, err)
					}
				}
			}
		}
	}

	return nil
}

func validateArray(value interface{}, schema map[string]interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected array, got %T", value)
	}

	// Validate items if schema is specified
	if itemSchema, ok := schema["items"].(map[string]interface{}); ok {
		for i, item := range arr {
			if err := validateValue(item, itemSchema); err != nil {
				return fmt.Errorf("array item %d: %w", i, err)
			}
		}
	}

	return nil
}

func validateString(value interface{}, schema map[string]interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	// Check enum values
	if enum, ok := schema["enum"].([]interface{}); ok {
		for _, enumValue := range enum {
			if enumStr, ok := enumValue.(string); ok && enumStr == str {
				return nil
			}
		}
		return fmt.Errorf("value '%s' is not in allowed enum values", str)
	}

	return nil
}

func validateNumber(value interface{}, schema map[string]interface{}) error {
	switch value.(type) {
	case float64, float32, int, int32, int64:
		return nil
	default:
		return fmt.Errorf("expected number, got %T", value)
	}
}

func validateInteger(value interface{}, schema map[string]interface{}) error {
	switch value.(type) {
	case int, int32, int64, float64:
		// JSON unmarshaling often produces float64 for integers
		if f, ok := value.(float64); ok {
			if f != float64(int64(f)) {
				return fmt.Errorf("expected integer, got float %f", f)
			}
		}
		return nil
	default:
		return fmt.Errorf("expected integer, got %T", value)
	}
}

func validateBoolean(value interface{}, schema map[string]interface{}) error {
	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("expected boolean, got %T", value)
	}
	return nil
}
