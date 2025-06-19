package parser

import (
	"testing"
)

func TestParseString(t *testing.T) {
	jsonStr := `{
		"name": "test-app",
		"version": "1.0.0",
		"config": {
			"debug": "true",
			"port": "8080"
		}
	}`
	
	data, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Error parsing JSON string: %v", err)
	}
	
	if data.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", data.Name)
	}
	
	if data.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", data.Version)
	}
	
	if data.Config["debug"] != "true" {
		t.Errorf("Expected debug config 'true', got '%s'", data.Config["debug"])
	}
}

func TestParseStringInvalidJSON(t *testing.T) {
	invalidJSON := `{"name": "test", "invalid": }`
	
	_, err := ParseString(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
} 