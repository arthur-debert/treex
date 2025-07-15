package config

// Config represents the main treex configuration
type Config struct {
	Version string        `yaml:"version"`
	Styles  *StylesConfig `yaml:"styles,omitempty"`
}

// StylesConfig represents the styles configuration
type StylesConfig struct {
	Theme  string                  `yaml:"theme,omitempty"`  // "dark", "light", "auto"
	Themes map[string]*ThemeConfig `yaml:"themes,omitempty"`
}

// ThemeConfig represents a theme configuration
type ThemeConfig struct {
	Colors    *ColorsConfig `yaml:"colors,omitempty"`
	TextStyle *TextStyles   `yaml:"text_style,omitempty"`
}

// ColorsConfig represents color settings for a theme
type ColorsConfig struct {
	// Primary colors
	Primary   string `yaml:"primary,omitempty"`
	Secondary string `yaml:"secondary,omitempty"`

	// Text colors
	Text       string `yaml:"text,omitempty"`
	TextMuted  string `yaml:"text_muted,omitempty"`
	TextSubtle string `yaml:"text_subtle,omitempty"`
	TextTitle  string `yaml:"text_title,omitempty"`
	TextBold   string `yaml:"text_bold,omitempty"`

	// Structure colors
	Border  string `yaml:"border,omitempty"`
	Surface string `yaml:"surface,omitempty"`
	Overlay string `yaml:"overlay,omitempty"`

	// State colors
	Success string `yaml:"success,omitempty"`
	Warning string `yaml:"warning,omitempty"`
	Error   string `yaml:"error,omitempty"`
	Info    string `yaml:"info,omitempty"`

	// Tree-specific colors
	TreeConnector  string `yaml:"tree_connector,omitempty"`
	TreeDirectory  string `yaml:"tree_directory,omitempty"`
	TreeFile       string `yaml:"tree_file,omitempty"`
	TreeAnnotation string `yaml:"tree_annotation,omitempty"`
}

// TextStyles represents text styling options
type TextStyles struct {
	// Whether to use bold for annotated paths
	AnnotatedBold bool `yaml:"annotated_bold,omitempty"`
	// Whether to use faint for unannotated paths
	UnannotatedFaint bool `yaml:"unannotated_faint,omitempty"`
	// Whether to use bold for root paths
	RootBold bool `yaml:"root_bold,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1",
		Styles: &StylesConfig{
			Theme: "auto",
		},
	}
}