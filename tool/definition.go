package tool

// Definition describes a tool for the model
type Definition struct {
	Type     string   `json:"type"` // always "function" for now
	Function Function `json:"function"`
}

// Function describes a callable function
type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

// Parameters uses a subset of JSON Schema
type Parameters struct {
	Type       string              `json:"type"` // "object"
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property describes a parameter property
type Property struct {
	Type        string `json:"type"` // string/number/boolean/array/object
	Description string `json:"description"`

	// TODO: Future JSON Schema features
	// - Enum for valid values
	// - Pattern for regex validation
	// - MinLength/MaxLength
	// - Minimum/Maximum for numbers
	// - Items for array type
	// - Properties for nested objects
}

// Call represents a tool invocation request from the model
type Call struct {
	ID       string       `json:"id"`
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the function name and arguments
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}
