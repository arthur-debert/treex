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