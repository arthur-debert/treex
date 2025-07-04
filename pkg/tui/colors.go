package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions using AdaptiveColor for automatic light/dark theme support
var (
	// Tree structure colors
	TreeConnectorColor = lipgloss.AdaptiveColor{
		Light: "#6B6B6B", // Medium gray for light backgrounds
		Dark:  "#6C7086", // Subtle gray for dark backgrounds
	}
	
	DirectoryColor = lipgloss.AdaptiveColor{
		Light: "#0969DA", // Darker blue for light backgrounds
		Dark:  "#89B4FA", // Lighter blue for dark backgrounds
	}
	
	FileColor = lipgloss.AdaptiveColor{
		Light: "#1F2328", // Dark gray for light backgrounds
		Dark:  "#CDD6F4", // Light gray for dark backgrounds
	}
	
	// Annotation colors
	AnnotationTitleColor = lipgloss.AdaptiveColor{
		Light: "#9A6700", // Dark yellow for light backgrounds
		Dark:  "#F9E2AF", // Light yellow for dark backgrounds
	}
	
	AnnotationDescriptionColor = lipgloss.AdaptiveColor{
		Light: "#1A7F37", // Dark green for light backgrounds
		Dark:  "#A6E3A1", // Light green for dark backgrounds
	}
	
	AnnotationBorderColor = lipgloss.AdaptiveColor{
		Light: "#D1D9E0", // Light gray for light backgrounds
		Dark:  "#585B70", // Darker gray for dark backgrounds
	}
	
	// Accent colors
	HighlightColor = lipgloss.AdaptiveColor{
		Light: "#CF222E", // Red for light backgrounds
		Dark:  "#F38BA8", // Pink for dark backgrounds
	}
	
	MutedColor = lipgloss.AdaptiveColor{
		Light: "#656D76", // Muted gray for light backgrounds
		Dark:  "#6C7086", // Muted gray for dark backgrounds
	}
)