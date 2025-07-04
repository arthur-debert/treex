package detector

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"testing"
	"time"
)


func TestNewTerminalDetector(t *testing.T) {
	timeout := 200 * time.Millisecond
	verbose := true
	
	td := NewTerminalDetector(timeout, verbose)
	
	if td.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, td.timeout)
	}
	if td.verbose != verbose {
		t.Errorf("Expected verbose %v, got %v", verbose, td.verbose)
	}
}

func TestDefaultTerminalDetector(t *testing.T) {
	td := DefaultTerminalDetector(false)
	
	if td.timeout != 100*time.Millisecond {
		t.Errorf("Expected default timeout 100ms, got %v", td.timeout)
	}
	if td.verbose != false {
		t.Errorf("Expected verbose false, got %v", td.verbose)
	}
}

func TestDetectTheme_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name      string
		envVar    string
		envValue  string
		wantDark  bool
		wantError bool
	}{
		{
			name:     "TREEX_THEME=light",
			envVar:   "TREEX_THEME",
			envValue: "light",
			wantDark: false,
		},
		{
			name:     "TREEX_THEME=dark",
			envVar:   "TREEX_THEME",
			envValue: "dark",
			wantDark: true,
		},
		{
			name:      "TREEX_THEME=invalid",
			envVar:    "TREEX_THEME",
			envValue:  "invalid",
			wantDark:  true, // Should proceed with detection, which defaults to dark
			wantError: true, // Will error because we're not in a terminal
		},
		{
			name:     "NO_COLOR set",
			envVar:   "NO_COLOR",
			envValue: "1",
			wantDark: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env var
			oldValue, exists := os.LookupEnv(tt.envVar)
			defer func() {
				if exists {
					_ = os.Setenv(tt.envVar, oldValue)
				} else {
					_ = os.Unsetenv(tt.envVar)
				}
			}()

			// Set test env var
			_ = os.Setenv(tt.envVar, tt.envValue)

			td := NewTerminalDetector(100*time.Millisecond, false)
			isDark, err := td.DetectTheme()

			if (err != nil) != tt.wantError {
				t.Errorf("DetectTheme() error = %v, wantError %v", err, tt.wantError)
			}
			if isDark != tt.wantDark {
				t.Errorf("DetectTheme() = %v, want %v", isDark, tt.wantDark)
			}
		})
	}
}

func TestDetectTheme_WithVerbose(t *testing.T) {
	// Save original env vars
	oldTreexTheme, existsTreex := os.LookupEnv("TREEX_THEME")
	oldNoColor, existsNoColor := os.LookupEnv("NO_COLOR")
	defer func() {
		if existsTreex {
			_ = os.Setenv("TREEX_THEME", oldTreexTheme)
		} else {
			_ = os.Unsetenv("TREEX_THEME")
		}
		if existsNoColor {
			_ = os.Setenv("NO_COLOR", oldNoColor)
		} else {
			_ = os.Unsetenv("NO_COLOR")
		}
	}()

	// Test with TREEX_THEME set and verbose mode
	_ = os.Setenv("TREEX_THEME", "light")
	_ = os.Unsetenv("NO_COLOR")

	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	td := NewTerminalDetector(100*time.Millisecond, true)
	isDark, _ := td.DetectTheme()

	// Restore stderr
	_ = w.Close()
	os.Stderr = oldStderr

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "[treex] Attempting terminal theme detection") {
		t.Error("Expected verbose output for theme detection attempt")
	}
	if !strings.Contains(output, "Using theme from TREEX_THEME environment variable") {
		t.Error("Expected verbose output for TREEX_THEME usage")
	}
	if isDark {
		t.Error("Expected light theme when TREEX_THEME=light")
	}
}

func TestParseColorResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantR    uint8
		wantG    uint8
		wantB    uint8
		wantErr  bool
	}{
		{
			name:     "standard format",
			response: "\x1b]11;rgb:0000/0000/0000\x07",
			wantR:    0,
			wantG:    0,
			wantB:    0,
		},
		{
			name:     "white color",
			response: "\x1b]11;rgb:ffff/ffff/ffff\x07",
			wantR:    255,
			wantG:    255,
			wantB:    255,
		},
		{
			name:     "mid-range color",
			response: "\x1b]11;rgb:8000/4000/c000\x07",
			wantR:    128,
			wantG:    64,
			wantB:    192,
		},
		{
			name:     "lowercase hex",
			response: "\x1b]11;rgb:00ff/00ff/00ff\x07",
			wantR:    0,
			wantG:    0,
			wantB:    0,
		},
		{
			name:     "invalid format",
			response: "\x1b]11;invalid\x07",
			wantErr:  true,
		},
		{
			name:     "missing values",
			response: "\x1b]11;rgb:0000/0000\x07",
			wantErr:  true,
		},
		{
			name:     "empty response",
			response: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewTerminalDetector(100*time.Millisecond, false)
			got, err := td.parseColorResponse(tt.response)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseColorResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				rgba, ok := got.(color.RGBA)
				if !ok {
					t.Fatalf("Expected color.RGBA, got %T", got)
				}
				if rgba.R != tt.wantR || rgba.G != tt.wantG || rgba.B != tt.wantB {
					t.Errorf("parseColorResponse() = RGB(%d, %d, %d), want RGB(%d, %d, %d)",
						rgba.R, rgba.G, rgba.B, tt.wantR, tt.wantG, tt.wantB)
				}
				if rgba.A != 255 {
					t.Errorf("Expected alpha = 255, got %d", rgba.A)
				}
			}
		})
	}
}

func TestIsDarkColor(t *testing.T) {
	tests := []struct {
		name     string
		color    color.Color
		wantDark bool
	}{
		{
			name:     "pure black",
			color:    color.RGBA{R: 0, G: 0, B: 0, A: 255},
			wantDark: true,
		},
		{
			name:     "pure white",
			color:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
			wantDark: false,
		},
		{
			name:     "dark gray",
			color:    color.RGBA{R: 64, G: 64, B: 64, A: 255},
			wantDark: true,
		},
		{
			name:     "light gray",
			color:    color.RGBA{R: 192, G: 192, B: 192, A: 255},
			wantDark: false,
		},
		{
			name:     "dark blue",
			color:    color.RGBA{R: 0, G: 0, B: 128, A: 255},
			wantDark: true,
		},
		{
			name:     "light yellow",
			color:    color.RGBA{R: 255, G: 255, B: 128, A: 255},
			wantDark: false,
		},
		{
			name:     "threshold edge - just below",
			color:    color.RGBA{R: 127, G: 127, B: 127, A: 255},
			wantDark: true,
		},
		{
			name:     "threshold edge - just above",
			color:    color.RGBA{R: 128, G: 128, B: 128, A: 255},
			wantDark: false,
		},
		{
			name:     "dark red",
			color:    color.RGBA{R: 128, G: 0, B: 0, A: 255},
			wantDark: true,
		},
		{
			name:     "bright green",
			color:    color.RGBA{R: 0, G: 255, B: 0, A: 255},
			wantDark: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewTerminalDetector(100*time.Millisecond, false)
			if got := td.isDarkColor(tt.color); got != tt.wantDark {
				t.Errorf("isDarkColor() = %v, want %v", got, tt.wantDark)
			}
		})
	}
}

func TestDetectTheme_Timeout(t *testing.T) {
	// This test simulates a timeout scenario
	td := NewTerminalDetector(1*time.Millisecond, false) // Very short timeout

	// Clear environment variables to force actual detection
	oldTreexTheme, existsTreex := os.LookupEnv("TREEX_THEME")
	oldNoColor, existsNoColor := os.LookupEnv("NO_COLOR")
	defer func() {
		if existsTreex {
			_ = os.Setenv("TREEX_THEME", oldTreexTheme)
		} else {
			_ = os.Unsetenv("TREEX_THEME")
		}
		if existsNoColor {
			_ = os.Setenv("NO_COLOR", oldNoColor)
		} else {
			_ = os.Unsetenv("NO_COLOR")
		}
	}()
	_ = os.Unsetenv("TREEX_THEME")
	_ = os.Unsetenv("NO_COLOR")

	// The detection should timeout and return dark theme as default
	isDark, err := td.DetectTheme()
	
	if !isDark {
		t.Error("Expected dark theme as default on timeout")
	}
	if err == nil {
		t.Error("Expected error on timeout, but detection might have succeeded (not in a real terminal)")
	}
}

func TestAutoSetTheme(t *testing.T) {
	// Save original env vars
	oldTreexTheme, existsTreex := os.LookupEnv("TREEX_THEME")
	defer func() {
		if existsTreex {
			_ = os.Setenv("TREEX_THEME", oldTreexTheme)
		} else {
			_ = os.Unsetenv("TREEX_THEME")
		}
	}()

	tests := []struct {
		name    string
		theme   string
		verbose bool
	}{
		{
			name:    "light theme",
			theme:   "light",
			verbose: false,
		},
		{
			name:    "dark theme",
			theme:   "dark",
			verbose: false,
		},
		{
			name:    "with verbose",
			theme:   "light",
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("TREEX_THEME", tt.theme)

			// Capture stderr if verbose
			if tt.verbose {
				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				AutoSetTheme(tt.verbose)

				_ = w.Close()
				os.Stderr = oldStderr

				buf := make([]byte, 1024)
				n, _ := r.Read(buf)
				output := string(buf[:n])

				if !strings.Contains(output, "[treex]") {
					t.Error("Expected verbose output")
				}
				if !strings.Contains(output, fmt.Sprintf("Using %s theme", tt.theme)) {
					t.Errorf("Expected theme usage message for %s theme", tt.theme)
				}
			} else {
				// Just run without checking output
				AutoSetTheme(tt.verbose)
			}
		})
	}
}

