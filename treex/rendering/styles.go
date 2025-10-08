// Package rendering provides styling definitions using the two-layer approach:
// Semantic styles (what it represents) -> Presentation styles (how it looks)
package rendering

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// StyleManager manages the two-layer styling system
type StyleManager struct {
	enabled            bool // Whether styling is enabled
	presentationStyles *PresentationStyles
}

// PresentationStyles defines the visual properties (CSS-like properties)
type PresentationStyles struct {
	// Text strength variations
	StrongText lipgloss.Style
	NormalText lipgloss.Style
	WeakText   lipgloss.Style

	// State-based text styles
	ActiveText   lipgloss.Style
	InactiveText lipgloss.Style

	// Semantic color variations
	SuccessText lipgloss.Style
	ErrorText   lipgloss.Style
	WarningText lipgloss.Style
	InfoText    lipgloss.Style

	// UI element styles
	HeaderText lipgloss.Style
	SubtleText lipgloss.Style
}

// NewStyleManager creates a new style manager
func NewStyleManager(enableColors bool) *StyleManager {
	return &StyleManager{
		enabled:            enableColors,
		presentationStyles: newPresentationStyles(enableColors),
	}
}

// newPresentationStyles creates presentation styles with adaptive theming
func newPresentationStyles(enableColors bool) *PresentationStyles {
	// For now, start with empty styles as instructed
	// "at first we can use empty styles, just to get the structure right"
	emptyStyle := lipgloss.NewStyle()
	return &PresentationStyles{
		StrongText:   emptyStyle,
		NormalText:   emptyStyle,
		WeakText:     emptyStyle,
		ActiveText:   emptyStyle,
		InactiveText: emptyStyle,
		SuccessText:  emptyStyle,
		ErrorText:    emptyStyle,
		WarningText:  emptyStyle,
		InfoText:     emptyStyle,
		HeaderText:   emptyStyle,
		SubtleText:   emptyStyle,
	}
}

// Semantic Style Methods
// These represent what the content means, not how it looks

// TreeConnector styles tree drawing characters (├─, └─, etc.)
func (sm *StyleManager) TreeConnector(text string) string {
	return sm.presentationStyles.StrongText.Render(text)
}

// FileName styles file and directory names
func (sm *StyleManager) FileName(text string) string {
	return sm.presentationStyles.NormalText.Render(text)
}

// DirectoryName styles directory names specifically
func (sm *StyleManager) DirectoryName(text string) string {
	return sm.presentationStyles.ActiveText.Render(text)
}

// Annotation styles annotation text from .info files
func (sm *StyleManager) Annotation(text string) string {
	return sm.presentationStyles.InfoText.Render(text)
}

// ErrorMessage styles error messages
func (sm *StyleManager) ErrorMessage(text string) string {
	return sm.presentationStyles.ErrorText.Render(text)
}

// WarningMessage styles warning messages
func (sm *StyleManager) WarningMessage(text string) string {
	return sm.presentationStyles.WarningText.Render(text)
}

// SuccessMessage styles success messages
func (sm *StyleManager) SuccessMessage(text string) string {
	return sm.presentationStyles.SuccessText.Render(text)
}

// StatsHeader styles statistics section headers
func (sm *StyleManager) StatsHeader(text string) string {
	return sm.presentationStyles.HeaderText.Render(text)
}

// StatsItem styles statistics item labels
func (sm *StyleManager) StatsItem(text string) string {
	return sm.presentationStyles.WeakText.Render(text)
}

// StatsValue styles statistics values
func (sm *StyleManager) StatsValue(text string) string {
	return sm.presentationStyles.StrongText.Render(text)
}

// HiddenFile styles hidden files/directories
func (sm *StyleManager) HiddenFile(text string) string {
	return sm.presentationStyles.SubtleText.Render(text)
}

// PluginResult styles plugin-generated content
func (sm *StyleManager) PluginResult(text string) string {
	return sm.presentationStyles.InfoText.Render(text)
}

// Utility methods for common formatting needs

// FormatNumber formats numbers with appropriate styling
func (sm *StyleManager) FormatNumber(n int) string {
	return sm.presentationStyles.StrongText.Render(strconv.Itoa(n))
}

// FormatPath formats file paths with appropriate styling
func (sm *StyleManager) FormatPath(path string) string {
	return sm.presentationStyles.SubtleText.Render(path)
}

// FormatSize formats file sizes with appropriate styling
func (sm *StyleManager) FormatSize(size int64) string {
	return sm.presentationStyles.WeakText.Render(formatBytes(size))
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + "B"
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + units[exp]
}
