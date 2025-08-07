package session

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEntryJSONSerialization(t *testing.T) {
	// Test MessageEntry JSON serialization
	msgEntry := NewMessageEntry("user", "Hello world")

	// Serialize to JSON
	jsonData, err := json.Marshal(msgEntry)
	if err != nil {
		t.Fatalf("Failed to marshal MessageEntry: %v", err)
	}

	// Deserialize from JSON
	var deserializedEntry Entry
	err = json.Unmarshal(jsonData, &deserializedEntry)
	if err != nil {
		t.Fatalf("Failed to unmarshal MessageEntry: %v", err)
	}

	// Verify basic fields
	if deserializedEntry.ID != msgEntry.ID {
		t.Errorf("ID mismatch: expected %s, got %s", msgEntry.ID, deserializedEntry.ID)
	}

	if deserializedEntry.Type != msgEntry.Type {
		t.Errorf("Type mismatch: expected %s, got %s", msgEntry.Type, deserializedEntry.Type)
	}

	// Note: Content will be map[string]interface{} after JSON unmarshaling
	// This is expected behavior with interface{} fields
}

func TestMessageContentJSONSerialization(t *testing.T) {
	content := MessageContent{
		Role: "assistant",
		Text: "How can I help you today?",
	}

	jsonData, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Failed to marshal MessageContent: %v", err)
	}

	var deserializedContent MessageContent
	err = json.Unmarshal(jsonData, &deserializedContent)
	if err != nil {
		t.Fatalf("Failed to unmarshal MessageContent: %v", err)
	}

	if deserializedContent.Role != content.Role {
		t.Errorf("Role mismatch: expected %s, got %s", content.Role, deserializedContent.Role)
	}

	if deserializedContent.Text != content.Text {
		t.Errorf("Text mismatch: expected %s, got %s", content.Text, deserializedContent.Text)
	}
}

func TestToolCallContentJSONSerialization(t *testing.T) {
	content := ToolCallContent{
		Tool: "search_flights",
		Parameters: map[string]any{
			"destination": "Tokyo",
			"departure":   "New York",
			"passengers":  2,
		},
	}

	jsonData, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCallContent: %v", err)
	}

	var deserializedContent ToolCallContent
	err = json.Unmarshal(jsonData, &deserializedContent)
	if err != nil {
		t.Fatalf("Failed to unmarshal ToolCallContent: %v", err)
	}

	if deserializedContent.Tool != content.Tool {
		t.Errorf("Tool mismatch: expected %s, got %s", content.Tool, deserializedContent.Tool)
	}

	// Check parameters
	if len(deserializedContent.Parameters) != len(content.Parameters) {
		t.Errorf("Parameters length mismatch: expected %d, got %d",
			len(content.Parameters), len(deserializedContent.Parameters))
	}

	if deserializedContent.Parameters["destination"] != "Tokyo" {
		t.Errorf("Destination parameter mismatch: expected Tokyo, got %v",
			deserializedContent.Parameters["destination"])
	}
}

func TestToolResultContentJSONSerialization(t *testing.T) {
	// Test successful result
	successContent := ToolResultContent{
		Tool:    "search_flights",
		Success: true,
		Result:  []string{"Flight 1", "Flight 2"},
		Error:   "",
	}

	jsonData, err := json.Marshal(successContent)
	if err != nil {
		t.Fatalf("Failed to marshal ToolResultContent: %v", err)
	}

	var deserializedContent ToolResultContent
	err = json.Unmarshal(jsonData, &deserializedContent)
	if err != nil {
		t.Fatalf("Failed to unmarshal ToolResultContent: %v", err)
	}

	if deserializedContent.Tool != successContent.Tool {
		t.Errorf("Tool mismatch: expected %s, got %s", successContent.Tool, deserializedContent.Tool)
	}

	if deserializedContent.Success != successContent.Success {
		t.Errorf("Success mismatch: expected %v, got %v", successContent.Success, deserializedContent.Success)
	}

	if deserializedContent.Error != successContent.Error {
		t.Errorf("Error mismatch: expected %s, got %s", successContent.Error, deserializedContent.Error)
	}

	// Test failed result
	failContent := ToolResultContent{
		Tool:    "search_flights",
		Success: false,
		Result:  nil,
		Error:   "Connection timeout",
	}

	jsonData, err = json.Marshal(failContent)
	if err != nil {
		t.Fatalf("Failed to marshal failed ToolResultContent: %v", err)
	}

	err = json.Unmarshal(jsonData, &deserializedContent)
	if err != nil {
		t.Fatalf("Failed to unmarshal failed ToolResultContent: %v", err)
	}

	if deserializedContent.Success != false {
		t.Errorf("Expected Success to be false, got %v", deserializedContent.Success)
	}

	if deserializedContent.Error != "Connection timeout" {
		t.Errorf("Expected Error to be 'Connection timeout', got %s", deserializedContent.Error)
	}
}

func TestCompleteEntryJSONRoundtrip(t *testing.T) {
	// Create entries of different types
	entries := []Entry{
		NewMessageEntry("user", "Hello"),
		NewToolCallEntry("search", map[string]any{"query": "test"}),
		NewToolResultEntry("search", "result", nil),
	}

	for i, entry := range entries {
		// Add some metadata
		entry.Metadata["test_key"] = "test_value"
		entry.Metadata["index"] = i

		// Serialize
		jsonData, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal entry %d: %v", i, err)
		}

		// Deserialize
		var deserializedEntry Entry
		err = json.Unmarshal(jsonData, &deserializedEntry)
		if err != nil {
			t.Fatalf("Failed to unmarshal entry %d: %v", i, err)
		}

		// Verify basic fields
		if deserializedEntry.ID != entry.ID {
			t.Errorf("Entry %d ID mismatch: expected %s, got %s", i, entry.ID, deserializedEntry.ID)
		}

		if deserializedEntry.Type != entry.Type {
			t.Errorf("Entry %d Type mismatch: expected %s, got %s", i, entry.Type, deserializedEntry.Type)
		}

		// Verify metadata
		if deserializedEntry.Metadata["test_key"] != "test_value" {
			t.Errorf("Entry %d metadata mismatch: expected test_value, got %v",
				i, deserializedEntry.Metadata["test_key"])
		}

		// Index will be float64 after JSON unmarshaling (JSON number handling)
		if indexValue, ok := deserializedEntry.Metadata["index"].(float64); !ok || int(indexValue) != i {
			t.Errorf("Entry %d index metadata mismatch: expected %d, got %v",
				i, i, deserializedEntry.Metadata["index"])
		}
	}
}

func TestJSONFieldNames(t *testing.T) {
	// Test that JSON field names are as expected
	entry := NewMessageEntry("user", "test")

	jsonData, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal entry: %v", err)
	}

	// Parse as generic map to check field names
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Check expected field names exist
	expectedFields := []string{"id", "type", "timestamp", "content", "metadata"}
	for _, field := range expectedFields {
		if _, exists := jsonMap[field]; !exists {
			t.Errorf("Expected field '%s' not found in JSON", field)
		}
	}

	// Check content structure for MessageContent
	if content, ok := jsonMap["content"].(map[string]interface{}); ok {
		if _, exists := content["role"]; !exists {
			t.Error("Expected 'role' field in content")
		}
		if _, exists := content["text"]; !exists {
			t.Error("Expected 'text' field in content")
		}
	} else {
		t.Error("Expected content to be an object")
	}
}

func TestTimestampJSONHandling(t *testing.T) {
	// Create entry with known timestamp
	entry := NewMessageEntry("user", "test")
	originalTime := time.Now().Round(time.Second) // Round to avoid precision issues
	entry.Timestamp = originalTime

	jsonData, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal entry: %v", err)
	}

	var deserializedEntry Entry
	err = json.Unmarshal(jsonData, &deserializedEntry)
	if err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}

	// Check timestamp preservation (within reasonable precision)
	timeDiff := deserializedEntry.Timestamp.Sub(originalTime)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("Timestamp mismatch: expected %v, got %v (diff: %v)",
			originalTime, deserializedEntry.Timestamp, timeDiff)
	}
}
