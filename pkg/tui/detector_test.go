package tui

import (
	"image/color"
	"os"
	"testing"
	"time"
)

func TestTerminalDetector_isDarkColor(t *testing.T) {
	detector := NewTerminalDetector(100*time.Millisecond, false)
	
	tests := []struct {
		name     string
		color    color.Color
		expected bool
	}{
		{
			name:     "Pure black",
			color:    color.RGBA{R: 0, G: 0, B: 0, A: 255},
			expected: true,
		},
		{
			name:     "Pure white",
			color:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
			expected: false,
		},
		{
			name:     "Dark gray",
			color:    color.RGBA{R: 64, G: 64, B: 64, A: 255},
			expected: true,
		},
		{
			name:     "Light gray",
			color:    color.RGBA{R: 192, G: 192, B: 192, A: 255},
			expected: false,
		},
		{
			name:     "Dark blue",
			color:    color.RGBA{R: 0, G: 0, B: 128, A: 255},
			expected: true,
		},
		{
			name:     "Light yellow",
			color:    color.RGBA{R: 255, G: 255, B: 200, A: 255},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isDarkColor(tt.color)
			if result != tt.expected {
				t.Errorf("isDarkColor(%v) = %v, want %v", tt.color, result, tt.expected)
			}
		})
	}
}

func TestTerminalDetector_parseColorResponse(t *testing.T) {
	detector := NewTerminalDetector(100*time.Millisecond, false)
	
	tests := []struct {
		name     string
		response string
		wantR    uint8
		wantG    uint8
		wantB    uint8
		wantErr  bool
	}{
		{
			name:     "Standard format - black",
			response: "\x1b]11;rgb:0000/0000/0000\x07",
			wantR:    0,
			wantG:    0,
			wantB:    0,
			wantErr:  false,
		},
		{
			name:     "Standard format - white",
			response: "\x1b]11;rgb:ffff/ffff/ffff\x07",
			wantR:    255,
			wantG:    255,
			wantB:    255,
			wantErr:  false,
		},
		{
			name:     "Standard format - color",
			response: "\x1b]11;rgb:8000/4000/c000\x07",
			wantR:    128,
			wantG:    64,
			wantB:    192,
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			response: "\x1b]11;invalid\x07",
			wantErr:  true,
		},
		{
			name:     "Missing values",
			response: "\x1b]11;rgb:ffff/ffff\x07",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := detector.parseColorResponse(tt.response)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseColorResponse() error = nil, wantErr = true")
				}
				return
			}
			
			if err != nil {
				t.Errorf("parseColorResponse() error = %v, wantErr = false", err)
				return
			}
			
			rgba, ok := c.(color.RGBA)
			if !ok {
				t.Errorf("parseColorResponse() returned non-RGBA color: %T", c)
				return
			}
			
			if rgba.R != tt.wantR || rgba.G != tt.wantG || rgba.B != tt.wantB {
				t.Errorf("parseColorResponse() = RGB(%d, %d, %d), want RGB(%d, %d, %d)",
					rgba.R, rgba.G, rgba.B, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestTerminalDetector_EnvironmentVariables(t *testing.T) {
	// Save original env vars
	origTheme := os.Getenv("TREEX_THEME")
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		_ = os.Setenv("TREEX_THEME", origTheme)
		_ = os.Setenv("NO_COLOR", origNoColor)
	}()
	
	tests := []struct {
		name       string
		themeEnv   string
		noColorEnv string
		wantDark   bool
	}{
		{
			name:     "TREEX_THEME=dark",
			themeEnv: "dark",
			wantDark: true,
		},
		{
			name:     "TREEX_THEME=light",
			themeEnv: "light",
			wantDark: false,
		},
		{
			name:       "NO_COLOR set",
			noColorEnv: "1",
			wantDark:   true,
		},
		{
			name:     "TREEX_THEME takes precedence over NO_COLOR",
			themeEnv: "light",
			noColorEnv: "1",
			wantDark: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			if tt.themeEnv != "" {
				_ = os.Setenv("TREEX_THEME", tt.themeEnv)
			} else {
				_ = os.Unsetenv("TREEX_THEME")
			}
			
			if tt.noColorEnv != "" {
				_ = os.Setenv("NO_COLOR", tt.noColorEnv)
			} else {
				_ = os.Unsetenv("NO_COLOR")
			}
			
			detector := NewTerminalDetector(100*time.Millisecond, false)
			isDark, err := detector.DetectTheme()
			
			if err != nil {
				t.Logf("DetectTheme() returned error (expected in test environment): %v", err)
			}
			
			if isDark != tt.wantDark {
				t.Errorf("DetectTheme() isDark = %v, want %v", isDark, tt.wantDark)
			}
		})
	}
}

func TestAutoSetTheme(t *testing.T) {
	// Save original theme and env
	origTheme := GetTheme()
	origEnv := os.Getenv("TREEX_THEME")
	defer func() {
		// Restore original theme
		SetTheme(origTheme.DirectoryColor == DarkTheme.DirectoryColor)
		_ = os.Setenv("TREEX_THEME", origEnv)
	}()
	
	// Test with explicit environment variable
	_ = os.Setenv("TREEX_THEME", "light")
	AutoSetTheme(false)
	
	currentTheme := GetTheme()
	if currentTheme.DirectoryColor != LightTheme.DirectoryColor {
		t.Error("AutoSetTheme did not set light theme when TREEX_THEME=light")
	}
	
	// Test with dark
	_ = os.Setenv("TREEX_THEME", "dark")
	AutoSetTheme(false)
	
	currentTheme = GetTheme()
	if currentTheme.DirectoryColor != DarkTheme.DirectoryColor {
		t.Error("AutoSetTheme did not set dark theme when TREEX_THEME=dark")
	}
}