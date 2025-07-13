package agent

import (
	"context"
	"fmt"
	"reflect"
)

// NewTool creates a simple function-based tool
func NewTool(name, description string, fn interface{}) Tool {
	return &funcTool{
		name:        name,
		description: description,
		function:    fn,
		schema:      generateSchema(fn),
	}
}

// ToolBuilder provides a fluent interface for creating tools
type ToolBuilder struct {
	name        string
	description string
	params      []paramDef
	function    interface{}
	err         error
}

type paramDef struct {
	name        string
	paramType   string
	description string
	required    bool
}

// NewToolBuilder creates a new tool builder
func NewToolBuilder(name string) *ToolBuilder {
	return &ToolBuilder{
		name: name,
	}
}

// WithDescription sets the tool description
func (t *ToolBuilder) WithDescription(desc string) *ToolBuilder {
	if t.err != nil {
		return t
	}
	t.description = desc
	return t
}

// WithParams adds a parameter definition
func (t *ToolBuilder) WithParams(name, paramType, description string) *ToolBuilder {
	if t.err != nil {
		return t
	}
	t.params = append(t.params, paramDef{
		name:        name,
		paramType:   paramType,
		description: description,
		required:    true,
	})
	return t
}

// WithOptionalParams adds an optional parameter definition
func (t *ToolBuilder) WithOptionalParams(name, paramType, description string) *ToolBuilder {
	if t.err != nil {
		return t
	}
	t.params = append(t.params, paramDef{
		name:        name,
		paramType:   paramType,
		description: description,
		required:    false,
	})
	return t
}

// WithFunc sets the function to execute
func (t *ToolBuilder) WithFunc(fn interface{}) *ToolBuilder {
	if t.err != nil {
		return t
	}
	t.function = fn
	return t
}

// Build creates the tool
func (t *ToolBuilder) Build() (Tool, error) {
	if t.err != nil {
		return nil, t.err
	}
	
	if t.name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	
	if t.function == nil {
		return nil, fmt.Errorf("tool function is required")
	}

	schema := t.buildSchema()
	
	return &funcTool{
		name:        t.name,
		description: t.description,
		function:    t.function,
		schema:      schema,
	}, nil
}

// buildSchema creates JSON schema from parameter definitions
func (t *ToolBuilder) buildSchema() map[string]interface{} {
	schema := map[string]interface{}{
		"type": "object",
		"properties": make(map[string]interface{}),
	}
	
	properties := schema["properties"].(map[string]interface{})
	var required []string
	
	for _, param := range t.params {
		paramSchema := map[string]interface{}{
			"type":        param.paramType,
			"description": param.description,
		}
		properties[param.name] = paramSchema
		
		if param.required {
			required = append(required, param.name)
		}
	}
	
	if len(required) > 0 {
		schema["required"] = required
	}
	
	return schema
}

// funcTool implements Tool using a function
type funcTool struct {
	name        string
	description string
	function    interface{}
	schema      map[string]interface{}
}

func (f *funcTool) Name() string {
	return f.name
}

func (f *funcTool) Description() string {
	return f.description
}

func (f *funcTool) Schema() map[string]interface{} {
	return f.schema
}

func (f *funcTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return callFunction(f.function, args)
}

// generateSchema automatically generates JSON schema from function signature
func generateSchema(fn interface{}) map[string]interface{} {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		}
	}

	properties := make(map[string]interface{})
	var required []string

	// Analyze function parameters (skip context if first param)
	startIdx := 0
	if fnType.NumIn() > 0 {
		firstParam := fnType.In(0)
		// Skip context.Context parameter
		if firstParam.String() == "context.Context" {
			startIdx = 1
		}
	}

	for i := startIdx; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		paramName := fmt.Sprintf("param%d", i-startIdx+1)
		
		// Try to infer parameter name from struct tags if available
		jsonType := mapGoTypeToJSON(paramType)
		properties[paramName] = map[string]interface{}{
			"type": jsonType,
		}
		required = append(required, paramName)
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// mapGoTypeToJSON maps Go types to JSON schema types
func mapGoTypeToJSON(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}

// callFunction dynamically calls a function with provided arguments
func callFunction(fn interface{}, args map[string]interface{}) (interface{}, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("provided value is not a function")
	}

	// Prepare arguments
	var callArgs []reflect.Value
	startIdx := 0

	// Handle context parameter if present
	if fnType.NumIn() > 0 {
		firstParam := fnType.In(0)
		if firstParam.String() == "context.Context" {
			callArgs = append(callArgs, reflect.ValueOf(context.Background()))
			startIdx = 1
		}
	}

	// Map arguments to function parameters
	for i := startIdx; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		paramName := fmt.Sprintf("param%d", i-startIdx+1)
		
		// Try to find argument by name or position
		var argValue interface{}
		if val, ok := args[paramName]; ok {
			argValue = val
		} else {
			// Try to find by parameter index in args
			keys := make([]string, 0, len(args))
			for k := range args {
				keys = append(keys, k)
			}
			if len(keys) > i-startIdx {
				argValue = args[keys[i-startIdx]]
			}
		}

		if argValue == nil {
			return nil, fmt.Errorf("missing argument for parameter %d", i-startIdx+1)
		}

		// Convert argument to correct type
		convertedArg, err := convertToType(argValue, paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert argument %d: %w", i-startIdx+1, err)
		}
		
		callArgs = append(callArgs, convertedArg)
	}

	// Call the function
	results := fnValue.Call(callArgs)

	// Handle return values
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		result := results[0]
		if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !result.IsNil() {
				return nil, result.Interface().(error)
			}
			return nil, nil
		}
		return result.Interface(), nil
	case 2:
		result := results[0]
		errResult := results[1]
		
		var err error
		if !errResult.IsNil() {
			err = errResult.Interface().(error)
		}
		
		if result.IsValid() {
			return result.Interface(), err
		}
		return nil, err
	default:
		return nil, fmt.Errorf("functions with more than 2 return values are not supported")
	}
}

// convertToType converts an interface{} value to a specific reflect.Type
func convertToType(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	valueType := reflect.TypeOf(value)
	if valueType.ConvertibleTo(targetType) {
		return reflect.ValueOf(value).Convert(targetType), nil
	}

	// Handle special cases
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(fmt.Sprintf("%v", value)), nil
	case reflect.Float64:
		if f, ok := value.(float64); ok {
			return reflect.ValueOf(f), nil
		}
		if i, ok := value.(int); ok {
			return reflect.ValueOf(float64(i)), nil
		}
	case reflect.Int:
		if f, ok := value.(float64); ok {
			return reflect.ValueOf(int(f)), nil
		}
		if i, ok := value.(int); ok {
			return reflect.ValueOf(i), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", value, targetType)
}