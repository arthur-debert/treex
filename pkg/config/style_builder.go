package config

import (
	"github.com/adebert/treex/pkg/display/styles"
	"github.com/charmbracelet/lipgloss"
)

// BuildTreeStyles builds TreeStyles from configuration
func BuildTreeStyles(config *Config) *styles.TreeStyles {
	if config == nil || config.Styles == nil {
		return styles.NewTreeStyles()
	}

	theme := determineTheme(config.Styles.Theme)
	themeConfig := getThemeConfig(config.Styles, theme)

	if themeConfig == nil {
		return styles.NewTreeStyles()
	}

	// Build custom styles from configuration
	base := buildBaseStyles(themeConfig)
	return buildTreeStylesFromBase(base, themeConfig)
}

func determineTheme(configTheme string) string {
	switch configTheme {
	case "dark", "light":
		return configTheme
	case "auto", "":
		// Auto-detect based on terminal background
		if lipgloss.HasDarkBackground() {
			return "dark"
		}
		return "light"
	default:
		return "auto"
	}
}

func getThemeConfig(stylesConfig *StylesConfig, theme string) *ThemeConfig {
	if stylesConfig.Themes == nil {
		return nil
	}

	if config, exists := stylesConfig.Themes[theme]; exists {
		return config
	}

	return nil
}

func buildBaseStyles(themeConfig *ThemeConfig) *styles.BaseStyles {
	base := styles.NewBaseStyles()

	if themeConfig.Colors == nil {
		return base
	}

	colors := themeConfig.Colors

	// Apply primary colors
	if colors.Primary != "" {
		base.Primary = lipgloss.NewStyle().Foreground(parseColor(colors.Primary))
	}
	if colors.Secondary != "" {
		base.Secondary = lipgloss.NewStyle().Foreground(parseColor(colors.Secondary))
	}

	// Apply text colors
	if colors.Text != "" {
		base.Text = lipgloss.NewStyle().Foreground(parseColor(colors.Text))
	}
	if colors.TextBold != "" {
		base.TextBold = lipgloss.NewStyle().Foreground(parseColor(colors.TextBold)).Bold(true)
	}
	if colors.TextMuted != "" {
		base.TextFaint = lipgloss.NewStyle().Foreground(parseColor(colors.TextMuted)).Faint(true)
	}
	if colors.TextSubtle != "" {
		base.TextSubtle = lipgloss.NewStyle().Foreground(parseColor(colors.TextSubtle))
	}
	if colors.TextTitle != "" {
		base.TextTitle = lipgloss.NewStyle().Foreground(parseColor(colors.TextTitle)).Bold(true)
	}

	// Apply state colors
	if colors.Success != "" {
		base.Success = lipgloss.NewStyle().Foreground(parseColor(colors.Success))
	}
	if colors.Warning != "" {
		base.Warning = lipgloss.NewStyle().Foreground(parseColor(colors.Warning))
	}
	if colors.Error != "" {
		base.Error = lipgloss.NewStyle().Foreground(parseColor(colors.Error))
	}
	if colors.Info != "" {
		base.Info = lipgloss.NewStyle().Foreground(parseColor(colors.Info))
	}

	// Apply structure colors
	if colors.Border != "" {
		base.Border = lipgloss.NewStyle().Foreground(parseColor(colors.Border))
	}

	return base
}

func buildTreeStylesFromBase(base *styles.BaseStyles, themeConfig *ThemeConfig) *styles.TreeStyles {
	treeStyles := &styles.TreeStyles{
		Base: base,
	}

	// Set default styles from base
	treeStyles.TreeLines = base.Structure.Faint(true)
	treeStyles.RootPath = base.TextBold
	treeStyles.AnnotatedPath = base.TextBold
	treeStyles.UnannotatedPath = base.TextFaint
	treeStyles.AnnotationText = base.Text
	treeStyles.AnnotationNotes = base.Text
	treeStyles.AnnotationDescription = base.TextSubtle
	treeStyles.AnnotationContainer = base.Text
	treeStyles.AnnotationSeparator = base.TextFaint.SetString("  ")
	treeStyles.MultiLineIndent = base.Border.Faint(true).PaddingLeft(1)

	// Apply tree-specific colors if provided
	if themeConfig.Colors != nil {
		colors := themeConfig.Colors

		if colors.TreeConnector != "" {
			treeStyles.TreeLines = lipgloss.NewStyle().
				Foreground(parseColor(colors.TreeConnector)).
				Faint(true)
		}

		if colors.TreeDirectory != "" {
			// This affects the root path style
			treeStyles.RootPath = lipgloss.NewStyle().
				Foreground(parseColor(colors.TreeDirectory)).
				Bold(true)
		}

		if colors.TreeAnnotation != "" {
			treeStyles.AnnotationText = lipgloss.NewStyle().
				Foreground(parseColor(colors.TreeAnnotation))
		}
	}

	// Apply text style overrides
	if themeConfig.TextStyle != nil {
		textStyle := themeConfig.TextStyle

		if !textStyle.AnnotatedBold {
			// Remove bold from annotated paths
			treeStyles.AnnotatedPath = base.Text
		}

		if !textStyle.UnannotatedFaint {
			// Remove faint from unannotated paths
			treeStyles.UnannotatedPath = base.Text
		}

		if !textStyle.RootBold {
			// Remove bold from root path
			treeStyles.RootPath = lipgloss.NewStyle().
				Foreground(treeStyles.RootPath.GetForeground())
		}
	}

	return treeStyles
}

func parseColor(color string) lipgloss.TerminalColor {
	// lipgloss.Color handles all color formats (hex, ansi, names)
	return lipgloss.Color(color)
}