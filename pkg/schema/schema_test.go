package schema

import (
	"testing"
)

func TestDefine(t *testing.T) {
	// Test creating a new field
	field := Define("email", "Please provide your email address")
	
	if field.Name() != "email" {
		t.Errorf("Expected field name 'email', got '%s'", field.Name())
	}
	
	if field.Prompt() != "Please provide your email address" {
		t.Errorf("Expected prompt 'Please provide your email address', got '%s'", field.Prompt())
	}
	
	if !field.Required() {
		t.Error("Expected field to be required by default")
	}
}

func TestField_Optional(t *testing.T) {
	// Test making a field optional
	field := Define("phone", "Please provide your phone number").Optional()
	
	if field.Name() != "phone" {
		t.Errorf("Expected field name 'phone', got '%s'", field.Name())
	}
	
	if field.Prompt() != "Please provide your phone number" {
		t.Errorf("Expected prompt 'Please provide your phone number', got '%s'", field.Prompt())
	}
	
	if field.Required() {
		t.Error("Expected field to be optional after calling Optional()")
	}
}

func TestField_Chaining(t *testing.T) {
	// Test that Optional returns the same field for chaining
	field := Define("notes", "Any additional notes?")
	optionalField := field.Optional()
	
	// Should be the same field object
	if field != optionalField {
		t.Error("Expected Optional() to return the same field object for chaining")
	}
}

func TestField_EmptyValues(t *testing.T) {
	// Test behavior with empty strings
	field := Define("", "")
	
	if field.Name() != "" {
		t.Errorf("Expected empty field name, got '%s'", field.Name())
	}
	
	if field.Prompt() != "" {
		t.Errorf("Expected empty prompt, got '%s'", field.Prompt())
	}
	
	if !field.Required() {
		t.Error("Expected field to be required by default even with empty values")
	}
}

func TestMultipleFields(t *testing.T) {
	// Test creating multiple fields with different configurations
	email := Define("email", "Please provide your email address")
	phone := Define("phone", "Please provide your phone number").Optional()
	issue := Define("issue", "Please describe your issue")
	
	fields := []*Field{email, phone, issue}
	
	// Verify each field maintains its configuration
	if !fields[0].Required() {
		t.Error("Email field should be required")
	}
	
	if fields[1].Required() {
		t.Error("Phone field should be optional")
	}
	
	if !fields[2].Required() {
		t.Error("Issue field should be required")
	}
	
	// Verify names
	expectedNames := []string{"email", "phone", "issue"}
	for i, field := range fields {
		if field.Name() != expectedNames[i] {
			t.Errorf("Expected field name '%s', got '%s'", expectedNames[i], field.Name())
		}
	}
}