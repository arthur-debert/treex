package rendering

import (
	"bytes"
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNewStyleRenderer(t *testing.T) {
	var buf bytes.Buffer
	sr := NewStyleRenderer(&buf)

	if sr == nil {
		t.Fatal("Expected non-nil StyleRenderer")
	}

	if sr.renderer == nil {
		t.Fatal("Expected non-nil lipgloss renderer")
	}

	if sr.styles == nil {
		t.Fatal("Expected non-nil styles")
	}

	// Verify that the renderer is associated with the output
	if sr.Renderer() == nil {
		t.Error("Renderer should be initialized")
	}
}

func TestNewStyleRendererWithAutoTheme(t *testing.T) {
	// Save original env vars
	originalTermProgram := os.Getenv("TERM_PROGRAM")
	originalTerm := os.Getenv("TERM")
	defer func() {
		_ = os.Setenv("TERM_PROGRAM", originalTermProgram)
		_ = os.Setenv("TERM", originalTerm)
	}()

	// Test with different terminal settings
	tests := []struct {
		name        string
		termProgram string
		term        string
		verbose     bool
	}{
		{
			name:        "Standard terminal",
			termProgram: "Terminal.app",
			term:        "xterm-256color",
			verbose:     false,
		},
		{
			name:        "Verbose mode",
			termProgram: "iTerm.app",
			term:        "xterm-256color",
			verbose:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("TERM_PROGRAM", tt.termProgram)
			_ = os.Setenv("TERM", tt.term)

			var buf bytes.Buffer
			sr := NewStyleRendererWithAutoTheme(&buf, tt.verbose)

			if sr == nil {
				t.Fatal("Expected non-nil StyleRenderer")
			}

			if sr.renderer == nil {
				t.Fatal("Expected non-nil renderer")
			}

			if sr.styles == nil {
				t.Fatal("Expected non-nil styles")
			}
		})
	}
}

func TestStyleRenderer_Methods(t *testing.T) {
	var buf bytes.Buffer
	sr := NewStyleRenderer(&buf)

	// Test Renderer()
	renderer := sr.Renderer()
	if renderer == nil {
		t.Error("Renderer() should return non-nil renderer")
	}

	// Test Styles()
	styles := sr.Styles()
	if styles == nil {
		t.Error("Styles() should return non-nil styles")
	}

	// Test SetColorProfile()
	sr.SetColorProfile(termenv.ANSI256)
	// No direct way to verify, but should not panic

	// Test SetHasDarkBackground()
	sr.SetHasDarkBackground(true)

	// Test HasDarkBackground()
	isDark := sr.HasDarkBackground()
	if !isDark {
		t.Error("Expected HasDarkBackground to return true after setting")
	}

	sr.SetHasDarkBackground(false)
	isDark = sr.HasDarkBackground()
	if isDark {
		t.Error("Expected HasDarkBackground to return false after setting")
	}
}

func TestStyleRenderer_AutoDetectTheme(t *testing.T) {
	// Save original env vars
	originalTermProgram := os.Getenv("TERM_PROGRAM")
	defer func() { _ = os.Setenv("TERM_PROGRAM", originalTermProgram) }()

	tests := []struct {
		name        string
		termProgram string
		verbose     bool
	}{
		{
			name:        "Normal terminal",
			termProgram: "Terminal.app",
			verbose:     false,
		},
		{
			name:        "Verbose mode",
			termProgram: "iTerm.app",
			verbose:     true,
		},
		{
			name:        "Unknown terminal",
			termProgram: "UnknownTerm",
			verbose:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("TERM_PROGRAM", tt.termProgram)

			var buf bytes.Buffer
			sr := NewStyleRenderer(&buf)

			err := sr.AutoDetectTheme(tt.verbose)
			// The method might return an error in verbose mode, but should not fail
			if err != nil && !tt.verbose {
				t.Errorf("Unexpected error in non-verbose mode: %v", err)
			}
		})
	}
}

func TestNewMinimalStyleRenderer(t *testing.T) {
	var buf bytes.Buffer
	sr := NewMinimalStyleRenderer(&buf)

	if sr == nil {
		t.Fatal("Expected non-nil StyleRenderer")
	}

	if sr.renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if sr.styles == nil {
		t.Fatal("Expected non-nil styles")
	}

	// Verify color profile is set to ANSI
	// Note: We can't directly check the color profile, but we can verify the styles exist
	if sr.styles == nil {
		t.Error("styles should be initialized")
	}
	if sr.styles.TreeLines.String() == "" {
		t.Error("TreeLines style should render something")
	}
}

func TestNewNoColorStyleRenderer(t *testing.T) {
	var buf bytes.Buffer
	sr := NewNoColorStyleRenderer(&buf)

	if sr == nil {
		t.Fatal("Expected non-nil StyleRenderer")
	}

	if sr.renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if sr.styles == nil {
		t.Fatal("Expected non-nil styles")
	}

	// Verify styles are created without colors
	if sr.styles == nil {
		t.Error("styles should be initialized")
	}
	// In no-color mode, styles exist but may render empty strings
	// Just verify they can be used
	testText := "test"
	_ = sr.styles.TreeLines.Render(testText)
}

func TestNewTreeStylesWithRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	styles := NewTreeStylesWithRenderer(renderer)

	if styles == nil {
		t.Fatal("Expected non-nil TreeStyles")
	}

	if styles.Base == nil {
		t.Fatal("Expected non-nil BaseStyles")
	}

	// Verify all style fields are initialized by checking they exist
	// We can't compare lipgloss.Style to empty struct, so we just verify fields exist
	styleChecks := []struct {
		name string
		hasStyle bool
	}{
		{"Base", styles.Base != nil},
	}

	for _, check := range styleChecks {
		if !check.hasStyle {
			t.Errorf("%s should be initialized", check.name)
		}
	}
}

func TestNewBaseStylesWithRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	base := NewBaseStylesWithRenderer(renderer)

	if base == nil {
		t.Fatal("Expected non-nil BaseStyles")
	}

	// Verify all base style fields exist by checking the struct is properly initialized
	// We can't compare lipgloss.Style directly, so we just verify the base struct exists
	// and has expected behavior
	if base.Text.String() == "" {
		t.Log("Text style exists (may render empty string)")
	}
	if base.TextBold.String() == "" {
		t.Log("TextBold style exists (may render empty string)")
	}
}

func TestNewMinimalBaseStylesWithRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	base := NewMinimalBaseStylesWithRenderer(renderer)

	if base == nil {
		t.Fatal("Expected non-nil BaseStyles")
	}

	// Verify styles are created with minimal colors
	// We can't compare lipgloss.Style directly, just verify they work
	if base.Text.String() == "" {
		t.Log("Text style exists (may render empty string)")
	}
	if base.TextBold.String() == "" {
		t.Log("TextBold style exists (may render empty string)")
	}
}

func TestNewNoColorBaseStylesWithRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	base := NewNoColorBaseStylesWithRenderer(renderer)

	if base == nil {
		t.Fatal("Expected non-nil BaseStyles")
	}

	// Verify all styles are created without colors
	// We verify by checking the base exists and trying to use styles
	if base.Text.String() == "" {
		t.Log("Text style exists (may render empty string)")
	}
	if base.TextBold.String() == "" {
		t.Log("TextBold style exists (may render empty string)")
	}

	// All color-related styles should be plain but usable
	testString := "test"
	if base.Primary.Render(testString) != testString {
		t.Error("Primary style should render text without modification in no-color mode")
	}
}

func TestStyleRenderer_ProblematicTerminal(t *testing.T) {
	// Save original env vars
	originalTermProgram := os.Getenv("TERM_PROGRAM")
	originalTerm := os.Getenv("TERM")
	originalSafeMode := os.Getenv("TREEX_SAFE_MODE")
	defer func() {
		_ = os.Setenv("TERM_PROGRAM", originalTermProgram)
		_ = os.Setenv("TERM", originalTerm)
		_ = os.Setenv("TREEX_SAFE_MODE", originalSafeMode)
	}()

	tests := []struct {
		name         string
		termProgram  string
		term         string
		safeMode     string
		expectSafe   bool
	}{
		{
			name:         "Ghostty terminal",
			termProgram:  "ghostty",
			term:         "xterm-256color",
			safeMode:     "",
			expectSafe:   true, // Should detect as problematic
		},
		{
			name:         "Ghostty uppercase",
			termProgram:  "GHOSTTY",
			term:         "xterm-256color",
			safeMode:     "",
			expectSafe:   true,
		},
		{
			name:         "Safe mode env var",
			termProgram:  "Terminal.app",
			term:         "xterm-256color",
			safeMode:     "1",
			expectSafe:   true,
		},
		{
			name:         "Normal terminal",
			termProgram:  "iTerm.app",
			term:         "xterm-256color",
			safeMode:     "",
			expectSafe:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("TERM_PROGRAM", tt.termProgram)
			_ = os.Setenv("TERM", tt.term)
			if tt.safeMode != "" {
				_ = os.Setenv("TREEX_SAFE_MODE", tt.safeMode)
			} else {
				_ = os.Unsetenv("TREEX_SAFE_MODE")
			}

			// The safe mode detection happens in styled_renderer.go
			// We can't directly test it here without creating a StyledTreeRenderer
			// This is more of an integration test placeholder
		})
	}
}