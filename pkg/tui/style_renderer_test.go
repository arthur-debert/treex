package tui

import (
	"bytes"
	"testing"
	
	"github.com/muesli/termenv"
)

func TestStyleRenderer(t *testing.T) {
	// Create a buffer to write to
	buf := &bytes.Buffer{}
	
	// Create a style renderer
	sr := NewStyleRenderer(buf)
	
	// Test that we can access the renderer
	if sr.Renderer() == nil {
		t.Error("Renderer should not be nil")
	}
	
	// Test that we can access the styles
	if sr.Styles() == nil {
		t.Error("Styles should not be nil")
	}
	
	// Test setting color profile
	sr.SetColorProfile(termenv.ANSI256)
	
	// Test setting dark background
	sr.SetHasDarkBackground(true)
	if !sr.HasDarkBackground() {
		t.Error("HasDarkBackground should return true after setting")
	}
	
	sr.SetHasDarkBackground(false)
	if sr.HasDarkBackground() {
		t.Error("HasDarkBackground should return false after setting")
	}
}

func TestNewStyledTreeRendererWithRenderer(t *testing.T) {
	// Create a buffer to write to
	buf := &bytes.Buffer{}
	
	// Create a styled tree renderer with renderer
	str := NewStyledTreeRendererWithRenderer(buf, true)
	
	// Test that the style renderer was created
	if str.styleRenderer == nil {
		t.Error("Style renderer should not be nil")
	}
	
	// Test that styles are available
	if str.styles == nil {
		t.Error("Styles should not be nil")
	}
}

func TestAutoDetectTheme(t *testing.T) {
	// Create a buffer to write to
	buf := &bytes.Buffer{}
	
	// Create a style renderer with auto theme detection
	sr := NewStyleRendererWithAutoTheme(buf, false)
	
	// Test that the renderer was created
	if sr.Renderer() == nil {
		t.Error("Renderer should not be nil")
	}
	
	// The theme detection might fail in test environment, but that's ok
	// We're just testing that the API works
}

func TestNewStyledTreeRendererWithAutoTheme(t *testing.T) {
	// Create a buffer to write to
	buf := &bytes.Buffer{}
	
	// Create a styled tree renderer with auto theme
	str := NewStyledTreeRendererWithAutoTheme(buf, true, false)
	
	// Test that the style renderer was created
	if str.styleRenderer == nil {
		t.Error("Style renderer should not be nil")
	}
	
	// Test that styles are available
	if str.styles == nil {
		t.Error("Styles should not be nil")
	}
}

func TestFormatSpecificRenderers(t *testing.T) {
	buf := &bytes.Buffer{}
	
	t.Run("MinimalStyleRenderer", func(t *testing.T) {
		sr := NewMinimalStyleRenderer(buf)
		
		if sr.Renderer() == nil {
			t.Error("Renderer should not be nil")
		}
		
		if sr.Styles() == nil {
			t.Error("Styles should not be nil")
		}
		
		// Test that minimal styles are created
		if sr.Styles().Base == nil {
			t.Error("Base styles should not be nil")
		}
	})
	
	t.Run("NoColorStyleRenderer", func(t *testing.T) {
		sr := NewNoColorStyleRenderer(buf)
		
		if sr.Renderer() == nil {
			t.Error("Renderer should not be nil")
		}
		
		if sr.Styles() == nil {
			t.Error("Styles should not be nil")
		}
		
		// Test that no-color styles are created
		if sr.Styles().Base == nil {
			t.Error("Base styles should not be nil")
		}
	})
}