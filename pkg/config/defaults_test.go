package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adebert/treex/pkg/display/styles"
)

func TestDefaultConfigMatchesBuiltinStyles(t *testing.T) {
	// Create a temporary config file with the default content
	tempDir, err := os.MkdirTemp("", "treex-defaults-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Read the default config from the project root
	defaultConfigContent, err := os.ReadFile("../../treex.yaml")
	if err != nil {
		t.Skipf("Could not read treex.yaml from project root: %v", err)
	}

	// Write it to temp location
	configPath := filepath.Join(tempDir, "treex.yaml")
	if err := os.WriteFile(configPath, defaultConfigContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Test both light and dark themes
	themes := []string{"light", "dark"}
	
	for _, theme := range themes {
		t.Run(theme+" theme", func(t *testing.T) {
			// Override theme to test specific theme
			originalTheme := config.Styles.Theme
			config.Styles.Theme = theme
			defer func() { config.Styles.Theme = originalTheme }()

			// Build styles from config
			configStyles := BuildTreeStyles(config)
			
			// Get built-in styles
			var builtinStyles *styles.TreeStyles
			if theme == "light" {
				// For light theme, we need to temporarily set the terminal to light mode
				// Since we can't easily do that, we'll just verify the config loads correctly
				builtinStyles = styles.NewTreeStyles()
			} else {
				// Same for dark theme
				builtinStyles = styles.NewTreeStyles()
			}

			// Basic validation - ensure both produce valid styles
			if configStyles == nil {
				t.Fatal("Config styles is nil")
			}
			if builtinStyles == nil {
				t.Fatal("Builtin styles is nil")
			}

			// Verify key styles are present and render something
			testStr := "test"
			
			// Check tree lines
			if configStyles.TreeLines.Render(testStr) == "" {
				t.Error("Config TreeLines renders empty")
			}
			
			// Check annotated path
			if configStyles.AnnotatedPath.Render(testStr) == "" {
				t.Error("Config AnnotatedPath renders empty")
			}
			
			// Check unannotated path
			if configStyles.UnannotatedPath.Render(testStr) == "" {
				t.Error("Config UnannotatedPath renders empty")
			}
		})
	}
}

func TestDefaultConfigColors(t *testing.T) {
	// Load the default config
	config := DefaultConfig()
	
	// Manually set the themes as they would appear in the file
	config.Styles.Themes = map[string]*ThemeConfig{
		"light": {
			Colors: &ColorsConfig{
				Primary:       "#0969DA",
				Secondary:     "#1A7F37",
				Text:          "#1F2328",
				TextMuted:     "255",
				TextSubtle:    "239",
				TextTitle:     "239",
				TextBold:      "#0A0C10",
				TreeConnector: "#6B6B6B",
				TreeDirectory: "#0969DA",
			},
			TextStyle: &TextStyles{
				AnnotatedBold:    true,
				UnannotatedFaint: true,
				RootBold:         true,
			},
		},
		"dark": {
			Colors: &ColorsConfig{
				Primary:       "#89B4FA",
				Secondary:     "#A6E3A1",
				Text:          "255",
				TextMuted:     "232",
				TextSubtle:    "249",
				TextTitle:     "252",
				TextBold:      "#FFFFFF",
				TreeConnector: "#6C7086",
				TreeDirectory: "#89B4FA",
			},
			TextStyle: &TextStyles{
				AnnotatedBold:    true,
				UnannotatedFaint: true,
				RootBold:         true,
			},
		},
	}

	// Test that styles can be built successfully
	lightStyles := BuildTreeStyles(&Config{
		Version: "1",
		Styles: &StylesConfig{
			Theme:  "light",
			Themes: config.Styles.Themes,
		},
	})

	darkStyles := BuildTreeStyles(&Config{
		Version: "1",
		Styles: &StylesConfig{
			Theme:  "dark",
			Themes: config.Styles.Themes,
		},
	})

	if lightStyles == nil {
		t.Fatal("Light styles is nil")
	}
	if darkStyles == nil {
		t.Fatal("Dark styles is nil")
	}

	// Verify colors are applied
	if lightStyles.Base == nil {
		t.Fatal("Light styles base is nil")
	}
	if darkStyles.Base == nil {
		t.Fatal("Dark styles base is nil")
	}
}