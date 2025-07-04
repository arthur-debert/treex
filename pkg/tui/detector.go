package tui

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"os"
	"time"

	"golang.org/x/term"
)

// TerminalDetector handles terminal background color detection
type TerminalDetector struct {
	timeout time.Duration
	verbose bool
}

// NewTerminalDetector creates a new terminal detector with the specified timeout
func NewTerminalDetector(timeout time.Duration, verbose bool) *TerminalDetector {
	return &TerminalDetector{
		timeout: timeout,
		verbose: verbose,
	}
}

// DefaultTerminalDetector creates a detector with sensible defaults
func DefaultTerminalDetector(verbose bool) *TerminalDetector {
	return NewTerminalDetector(100*time.Millisecond, verbose)
}

// DetectTheme attempts to detect if the terminal has a dark or light background
// Returns true for dark theme, false for light theme
func (td *TerminalDetector) DetectTheme() (isDark bool, err error) {
	if td.verbose {
		fmt.Fprintf(os.Stderr, "[treex] Attempting terminal theme detection...\n")
	}

	// Check environment variable override first
	if themeEnv := os.Getenv("TREEX_THEME"); themeEnv != "" {
		if td.verbose {
			fmt.Fprintf(os.Stderr, "[treex] Using theme from TREEX_THEME environment variable: %s\n", themeEnv)
		}
		switch themeEnv {
		case "light":
			return false, nil
		case "dark":
			return true, nil
		default:
			if td.verbose {
				fmt.Fprintf(os.Stderr, "[treex] Unknown TREEX_THEME value '%s', proceeding with detection\n", themeEnv)
			}
		}
	}

	// Check for NO_COLOR environment variable (indicates user prefers no colors)
	if os.Getenv("NO_COLOR") != "" {
		if td.verbose {
			fmt.Fprintf(os.Stderr, "[treex] NO_COLOR is set, defaulting to dark theme\n")
		}
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), td.timeout)
	defer cancel()

	bgColor, err := td.getTerminalBackgroundColor(ctx)
	if err != nil {
		if td.verbose {
			fmt.Fprintf(os.Stderr, "[treex] Terminal detection failed: %v. Using dark theme as default.\n", err)
		}
		// Default to dark theme on error
		return true, err
	}

	isDarkBg := td.isDarkColor(bgColor)
	if td.verbose {
		r, g, b, _ := bgColor.RGBA()
		// Convert from 16-bit to 8-bit for display
		fmt.Fprintf(os.Stderr, "[treex] Detected background color: RGB(%d, %d, %d) - Theme: %s\n", 
			r>>8, g>>8, b>>8, 
			map[bool]string{true: "dark", false: "light"}[isDarkBg])
	}

	return isDarkBg, nil
}

// getTerminalBackgroundColor queries the terminal for its background color
func (td *TerminalDetector) getTerminalBackgroundColor(ctx context.Context) (color.Color, error) {
	// Check if we're in a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, fmt.Errorf("not in a terminal")
	}

	// Check if stdout is also a terminal (we write the query there)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, fmt.Errorf("stdout is not a terminal")
	}

	// Save the current terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to enter raw mode: %w", err)
	}
	// Ensure terminal state is restored
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	// Send OSC 11 query for background color
	// Format: ESC ] 11 ; ? BEL
	fmt.Print("\x1b]11;?\x07")

	// Read the response
	reader := bufio.NewReader(os.Stdin)
	
	// Channel for response or error
	type result struct {
		resp string
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		// Read until BEL character
		resp, err := reader.ReadString('\x07')
		ch <- result{resp, err}
	}()

	// Wait for response or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for terminal response")
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("failed to read response: %w", res.err)
		}
		return td.parseColorResponse(res.resp)
	}
}

// parseColorResponse parses the OSC 11 response format
// Expected format: ESC ] 11 ; rgb : RRRR / GGGG / BBBB BEL
func (td *TerminalDetector) parseColorResponse(response string) (color.Color, error) {
	var r, g, b uint16

	// Try to parse the response
	// The response format can vary slightly between terminals
	n, err := fmt.Sscanf(response, "\x1b]11;rgb:%04x/%04x/%04x", &r, &g, &b)
	if err != nil || n != 3 {
		// Try alternative format with lowercase hex
		n, err = fmt.Sscanf(response, "\x1b]11;rgb:%4x/%4x/%4x", &r, &g, &b)
		if err != nil || n != 3 {
			return nil, fmt.Errorf("failed to parse color response: %q", response)
		}
	}

	// Convert 16-bit color values to 8-bit
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: 255,
	}, nil
}

// isDarkColor determines if a color is dark based on perceived luminance
func (td *TerminalDetector) isDarkColor(c color.Color) bool {
	// Convert to grayscale to get luminance
	// This uses the standard luminance formula: Y = 0.299*R + 0.587*G + 0.114*B
	gray := color.GrayModel.Convert(c).(color.Gray)
	
	// Threshold for dark vs light
	// Values below 128 are considered dark
	return gray.Y < 128
}

// AutoSetTheme detects the terminal theme and sets the appropriate color theme
func AutoSetTheme(verbose bool) {
	detector := DefaultTerminalDetector(verbose)
	isDark, err := detector.DetectTheme()
	
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "[treex] Theme detection error: %v\n", err)
	}
	
	SetTheme(isDark)
	
	if verbose {
		theme := map[bool]string{true: "dark", false: "light"}[isDark]
		fmt.Fprintf(os.Stderr, "[treex] Using %s theme\n", theme)
	}
}