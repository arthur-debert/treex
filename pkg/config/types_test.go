package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.Version != "1" {
		t.Errorf("Expected version '1', got '%s'", config.Version)
	}

	if config.Styles == nil {
		t.Fatal("DefaultConfig styles is nil")
	}

	if config.Styles.Theme != "auto" {
		t.Errorf("Expected theme 'auto', got '%s'", config.Styles.Theme)
	}
}