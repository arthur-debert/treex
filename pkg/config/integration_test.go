package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigIntegration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "treex-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Test loading from a specific file
	t.Run("load from file", func(t *testing.T) {
		configContent := `version: "1"
styles:
  theme: dark
  themes:
    dark:
      colors:
        primary: "#89B4FA"
        text: "#CDD6F4"
        tree_connector: "#6C7086"
      text_style:
        annotated_bold: true
        unannotated_faint: true`

		configPath := filepath.Join(tempDir, "treex.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Validate the loaded config
		if config.Styles.Theme != "dark" {
			t.Errorf("Expected theme 'dark', got '%s'", config.Styles.Theme)
		}

		darkTheme := config.Styles.Themes["dark"]
		if darkTheme == nil {
			t.Fatal("Dark theme not loaded")
		}

		if darkTheme.Colors.Primary != "#89B4FA" {
			t.Errorf("Expected primary color '#89B4FA', got '%s'", darkTheme.Colors.Primary)
		}

		if !darkTheme.TextStyle.AnnotatedBold {
			t.Error("Expected annotated_bold to be true")
		}
	})

	// Test missing file returns defaults
	t.Run("missing file returns defaults", func(t *testing.T) {
		missingPath := filepath.Join(tempDir, "missing.yaml")
		config, err := LoadConfig(missingPath)
		if err != nil {
			t.Fatalf("Expected no error for missing file, got: %v", err)
		}

		if config.Version != "1" {
			t.Errorf("Expected default version '1', got '%s'", config.Version)
		}

		if config.Styles.Theme != "auto" {
			t.Errorf("Expected default theme 'auto', got '%s'", config.Styles.Theme)
		}
	})

	// Test invalid YAML
	t.Run("invalid yaml returns error", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid.yaml")
		if err := os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadConfig(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestFindConfigFile(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-find-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	t.Run("finds config in current directory", func(t *testing.T) {
		// Create treex.yaml in current directory
		if err := os.WriteFile("treex.yaml", []byte("version: \"1\""), 0644); err != nil {
			t.Fatal(err)
		}

		path, err := FindConfigFile()
		if err != nil {
			t.Fatalf("Failed to find config: %v", err)
		}

		if path != "treex.yaml" {
			t.Errorf("Expected 'treex.yaml', got '%s'", path)
		}

		// Clean up
		_ = os.Remove("treex.yaml")
	})

	t.Run("no config returns error", func(t *testing.T) {
		_, err := FindConfigFile()
		if err == nil {
			t.Error("Expected error when no config file exists")
		}
	})
}