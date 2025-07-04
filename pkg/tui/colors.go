package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// SemanticColors provides a semantic color palette for the application
type SemanticColors struct {
	// Primary colors for main UI elements
	Primary   lipgloss.TerminalColor // Main brand/action color
	Secondary lipgloss.TerminalColor // Secondary actions and elements
	
	// Content colors
	Text       lipgloss.TerminalColor // Regular text
	TextMuted  lipgloss.TerminalColor // De-emphasized text (faint)
	TextSubtle lipgloss.TerminalColor // Subtle text (between regular and muted)
	TextTitle  lipgloss.TerminalColor // Title text
	TextBold   lipgloss.TerminalColor // Emphasized text
	
	// Structure colors
	Border    lipgloss.TerminalColor // Borders and dividers
	Surface   lipgloss.TerminalColor // Background surfaces
	Overlay   lipgloss.TerminalColor // Overlays and modals
	
	// State colors
	Success lipgloss.TerminalColor // Success states
	Warning lipgloss.TerminalColor // Warning states
	Error   lipgloss.TerminalColor // Error states
	Info    lipgloss.TerminalColor // Informational states
	
	// Tree-specific colors
	TreeConnector  lipgloss.TerminalColor // Tree structure lines
	TreeDirectory  lipgloss.TerminalColor // Directory nodes
	TreeFile       lipgloss.TerminalColor // File nodes
	TreeAnnotation lipgloss.TerminalColor // Annotation text
}

// Colors is the semantic color palette for treex
var Colors = SemanticColors{
	// Primary colors - using CompleteAdaptiveColor for better terminal support
	Primary: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#0969DA", // GitHub blue for light mode
			ANSI256:   "26",      // Blue
			ANSI:      "4",       // Blue
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#89B4FA", // Catppuccin blue for dark mode
			ANSI256:   "111",     // Light blue
			ANSI:      "12",      // Light blue
		},
	},
	Secondary: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#1A7F37", // GitHub green for light mode
			ANSI256:   "28",      // Green
			ANSI:      "2",       // Green
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#A6E3A1", // Catppuccin green for dark mode
			ANSI256:   "151",     // Light green
			ANSI:      "10",      // Light green
		},
	},
	
	// Content colors
	Text: lipgloss.AdaptiveColor{
		Light: "#1F2328", // Near black for light mode
		Dark:  "255",     // Brightest for dark mode (regular text)
	},
	TextMuted: lipgloss.AdaptiveColor{
		Light: "255", // Faintest gray for light mode (255)
		Dark:  "232", // Faintest gray for dark mode (232) - for items without annotations
	},
	TextSubtle: lipgloss.AdaptiveColor{
		Light: "239", // Subtle gray for light mode (239)
		Dark:  "249", // Bright gray for dark mode (249) - for descriptions
	},
	TextTitle: lipgloss.AdaptiveColor{
		Light: "239", // Same as subtle for light mode
		Dark:  "252", // Brighter gray for dark mode (252) - for titles
	},
	TextBold: lipgloss.AdaptiveColor{
		Light: "#0A0C10", // Full black for light mode
		Dark:  "#FFFFFF", // Full white for dark mode
	},
	
	// Structure colors
	Border: lipgloss.AdaptiveColor{
		Light: "#D1D9E0", // Light gray border for light mode
		Dark:  "#585B70", // Dark gray border for dark mode
	},
	Surface: lipgloss.AdaptiveColor{
		Light: "#F6F8FA", // Light surface for light mode
		Dark:  "#1E1E2E", // Dark surface for dark mode
	},
	Overlay: lipgloss.AdaptiveColor{
		Light: "#FFFFFF", // White overlay for light mode
		Dark:  "#313244", // Dark overlay for dark mode
	},
	
	// State colors - using CompleteAdaptiveColor for visibility
	Success: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#1A7F37", // Green for light mode
			ANSI256:   "28",      // Green
			ANSI:      "2",       // Green
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#A6E3A1", // Green for dark mode
			ANSI256:   "151",     // Light green
			ANSI:      "10",      // Light green
		},
	},
	Warning: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#9A6700", // Yellow for light mode
			ANSI256:   "136",     // Dark yellow
			ANSI:      "3",       // Yellow
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#F9E2AF", // Yellow for dark mode
			ANSI256:   "223",     // Light yellow
			ANSI:      "11",      // Bright yellow
		},
	},
	Error: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#CF222E", // Red for light mode
			ANSI256:   "196",     // Red
			ANSI:      "1",       // Red
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#F38BA8", // Red for dark mode
			ANSI256:   "211",     // Light red
			ANSI:      "9",       // Bright red
		},
	},
	Info: lipgloss.AdaptiveColor{
		Light: "#0969DA", // Blue for light mode
		Dark:  "#89DCEB", // Cyan for dark mode
	},
	
	// Tree-specific colors - using CompleteAdaptiveColor for core tree elements
	TreeConnector: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#6B6B6B", // Medium gray for light mode
			ANSI256:   "242",     // Gray
			ANSI:      "8",       // Bright black (gray)
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#6C7086", // Subtle gray for dark mode
			ANSI256:   "243",     // Light gray
			ANSI:      "7",       // White (appears as gray on dark bg)
		},
	},
	TreeDirectory: lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#0969DA", // Blue for light mode
			ANSI256:   "26",      // Blue
			ANSI:      "4",       // Blue
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#89B4FA", // Blue for dark mode
			ANSI256:   "111",     // Light blue
			ANSI:      "12",      // Light blue
		},
	},
	TreeFile: lipgloss.AdaptiveColor{
		Light: "#1F2328", // Dark gray for light mode
		Dark:  "#CDD6F4", // Light gray for dark mode
	},
	TreeAnnotation: lipgloss.AdaptiveColor{
		Light: "#0969DA", // Blue for light mode (matches directory)
		Dark:  "#89B4FA", // Blue for dark mode (matches directory)
	},
}

// Legacy color aliases for backward compatibility
var (
	TreeConnectorColor         lipgloss.TerminalColor = Colors.TreeConnector
	DirectoryColor             lipgloss.TerminalColor = Colors.TreeDirectory
	FileColor                  lipgloss.TerminalColor = Colors.TreeFile
	AnnotationTitleColor       lipgloss.TerminalColor = Colors.Warning  // Using warning color for titles
	AnnotationDescriptionColor lipgloss.TerminalColor = Colors.Success  // Using success color for descriptions
	AnnotationBorderColor      lipgloss.TerminalColor = Colors.Border
	HighlightColor             lipgloss.TerminalColor = Colors.Error    // Using error color for highlights
	MutedColor                 lipgloss.TerminalColor = Colors.TextMuted
)