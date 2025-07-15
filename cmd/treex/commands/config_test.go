package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigCommand(t *testing.T) {
	// Test that config command outputs valid YAML
	cmd := configCmd
	var out bytes.Buffer
	cmd.SetOut(&out)
	
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("config command failed: %v", err)
	}
	
	output := out.String()
	
	// Check that output contains expected content
	expectedStrings := []string{
		"version: \"1\"",
		"styles:",
		"theme: auto",
		"themes:",
		"light:",
		"dark:",
		"colors:",
		"primary:",
		"tree_connector:",
		"text_style:",
		"annotated_bold:",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Config output missing expected string: %s", expected)
		}
	}
	
	// Check it starts with a comment
	if !strings.HasPrefix(output, "#") {
		t.Error("Config output should start with a comment")
	}
	
	// Check it's not empty
	if len(output) < 1000 {
		t.Errorf("Config output seems too short: %d bytes", len(output))
	}
}