package config

import (
	"testing"
)

func TestBuildTreeStyles(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "nil config returns default styles",
			config: nil,
		},
		{
			name:   "empty config returns default styles",
			config: &Config{},
		},
		{
			name: "config with custom colors",
			config: &Config{
				Version: "1",
				Styles: &StylesConfig{
					Theme: "light",
					Themes: map[string]*ThemeConfig{
						"light": {
							Colors: &ColorsConfig{
								Primary:       "#0969DA",
								Text:          "#1F2328",
								TreeConnector: "#6B6B6B",
							},
						},
					},
				},
			},
		},
		{
			name: "config with text style overrides",
			config: &Config{
				Version: "1",
				Styles: &StylesConfig{
					Theme: "dark",
					Themes: map[string]*ThemeConfig{
						"dark": {
							TextStyle: &TextStyles{
								AnnotatedBold:    false,
								UnannotatedFaint: false,
								RootBold:         false,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := BuildTreeStyles(tt.config)

			// Basic validation - ensure we get a valid TreeStyles
			if styles == nil {
				t.Fatal("BuildTreeStyles returned nil")
			}

			if styles.Base == nil {
				t.Fatal("TreeStyles.Base is nil")
			}

			// Basic validation - styles should have renderable content
			// We can't compare styles directly, but we can check they render something
			testStr := "test"
			if styles.TreeLines.Render(testStr) == "" {
				t.Error("TreeLines style renders empty")
			}
			if styles.RootPath.Render(testStr) == "" {
				t.Error("RootPath style renders empty")
			}
			if styles.AnnotatedPath.Render(testStr) == "" {
				t.Error("AnnotatedPath style renders empty")
			}
			if styles.UnannotatedPath.Render(testStr) == "" {
				t.Error("UnannotatedPath style renders empty")
			}
		})
	}
}

func TestDetermineTheme(t *testing.T) {
	tests := []struct {
		name        string
		configTheme string
		want        string
	}{
		{
			name:        "explicit dark",
			configTheme: "dark",
			want:        "dark",
		},
		{
			name:        "explicit light",
			configTheme: "light",
			want:        "light",
		},
		{
			name:        "auto detection",
			configTheme: "auto",
			want:        "light", // Will be either "light" or "dark" based on terminal
		},
		{
			name:        "empty defaults to auto",
			configTheme: "",
			want:        "light", // Will be either "light" or "dark" based on terminal
		},
		{
			name:        "invalid defaults to auto",
			configTheme: "invalid",
			want:        "auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineTheme(tt.configTheme)

			// For auto detection, we accept either light or dark
			if tt.configTheme == "auto" || tt.configTheme == "" {
				if got != "light" && got != "dark" {
					t.Errorf("determineTheme() = %v, want 'light' or 'dark'", got)
				}
			} else if got != tt.want {
				t.Errorf("determineTheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
	}{
		{
			name:  "hex color",
			color: "#FF0000",
		},
		{
			name:  "ansi color number",
			color: "255",
		},
		{
			name:  "color name",
			color: "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseColor(tt.color)

			// Basic check - ensure we get a valid color
			if result == nil {
				t.Errorf("parseColor(%s) returned nil", tt.color)
			}
		})
	}
}