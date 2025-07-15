package config

import (
	"strings"
	"testing"
)

func TestLoadConfigFromReader(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		validate  func(t *testing.T, config *Config)
	}{
		{
			name:      "empty config returns defaults",
			input:     "",
			wantError: false,
			validate: func(t *testing.T, config *Config) {
				if config.Version != "1" {
					t.Errorf("Expected version '1', got '%s'", config.Version)
				}
				if config.Styles.Theme != "auto" {
					t.Errorf("Expected theme 'auto', got '%s'", config.Styles.Theme)
				}
			},
		},
		{
			name: "basic config",
			input: `version: "1"
styles:
  theme: dark`,
			wantError: false,
			validate: func(t *testing.T, config *Config) {
				if config.Version != "1" {
					t.Errorf("Expected version '1', got '%s'", config.Version)
				}
				if config.Styles.Theme != "dark" {
					t.Errorf("Expected theme 'dark', got '%s'", config.Styles.Theme)
				}
			},
		},
		{
			name: "config with theme colors",
			input: `version: "1"
styles:
  theme: light
  themes:
    light:
      colors:
        primary: "#0969DA"
        text: "#1F2328"
        tree_connector: "#6B6B6B"`,
			wantError: false,
			validate: func(t *testing.T, config *Config) {
				if config.Styles.Themes == nil {
					t.Fatal("Expected themes to be defined")
				}
				lightTheme := config.Styles.Themes["light"]
				if lightTheme == nil {
					t.Fatal("Expected light theme to be defined")
				}
				if lightTheme.Colors == nil {
					t.Fatal("Expected light theme colors to be defined")
				}
				if lightTheme.Colors.Primary != "#0969DA" {
					t.Errorf("Expected primary color '#0969DA', got '%s'", lightTheme.Colors.Primary)
				}
			},
		},
		{
			name: "config with text styles",
			input: `version: "1"
styles:
  theme: dark
  themes:
    dark:
      text_style:
        annotated_bold: false
        unannotated_faint: true`,
			wantError: false,
			validate: func(t *testing.T, config *Config) {
				darkTheme := config.Styles.Themes["dark"]
				if darkTheme == nil || darkTheme.TextStyle == nil {
					t.Fatal("Expected dark theme with text styles")
				}
				if darkTheme.TextStyle.AnnotatedBold != false {
					t.Error("Expected annotated_bold to be false")
				}
				if darkTheme.TextStyle.UnannotatedFaint != true {
					t.Error("Expected unannotated_faint to be true")
				}
			},
		},
		{
			name: "invalid theme name",
			input: `version: "1"
styles:
  theme: invalid`,
			wantError: true,
		},
		{
			name: "missing version",
			input: `styles:
  theme: dark`,
			wantError: true,
		},
		{
			name: "unknown fields rejected",
			input: `version: "1"
unknown_field: value`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			config, err := LoadConfigFromReader(reader)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if config == nil {
				t.Fatal("Config is nil")
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Version: "1",
				Styles: &StylesConfig{
					Theme: "dark",
				},
			},
			wantError: false,
		},
		{
			name: "missing version",
			config: &Config{
				Styles: &StylesConfig{
					Theme: "dark",
				},
			},
			wantError: true,
		},
		{
			name: "invalid theme",
			config: &Config{
				Version: "1",
				Styles: &StylesConfig{
					Theme: "invalid",
				},
			},
			wantError: true,
		},
		{
			name: "auto theme is valid",
			config: &Config{
				Version: "1",
				Styles: &StylesConfig{
					Theme: "auto",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}