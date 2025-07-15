# Treex Configuration

Treex supports customization of its visual appearance through a YAML configuration file. This allows you to tailor the colors and text styles to match your preferences or terminal theme.

## Default Configuration

You can generate a fully documented default configuration file using the `treex config` command:

```bash
# View the default configuration
treex config

# Save to current directory
treex config > treex.yaml

# Save to user config directory
treex config > ~/.config/treex/treex.yaml
```

This outputs a complete configuration file with all available options documented, which you can use as a starting template for customization.

## Configuration File Locations

Treex looks for a configuration file named `treex.yaml` in the following locations (in order):

1. Current directory (`./treex.yaml`)
2. User config directory (`~/.config/treex/treex.yaml`)
3. Home directory (`~/.treex.yaml`)
4. XDG config directory (`$XDG_CONFIG_HOME/treex/treex.yaml`)

The first file found will be used. If no configuration file is found, treex uses its built-in defaults.

## Configuration Structure

The configuration file uses YAML format with the following structure:

```yaml
version: "1"  # Required

styles:
  theme: auto  # "dark", "light", or "auto"
  
  themes:
    light:
      colors:
        # Color definitions...
      text_style:
        # Text style options...
    
    dark:
      colors:
        # Color definitions...
      text_style:
        # Text style options...
```

## Available Options

### Theme Selection

- `theme`: Choose which theme to use
  - `"auto"` (default): Automatically detect based on terminal background
  - `"dark"`: Use dark theme
  - `"light"`: Use light theme

### Color Options

Colors can be specified in multiple formats:
- Hex colors: `"#FF0000"`
- ANSI color codes: `"255"` (0-255)
- Color names: `"red"`, `"blue"`, etc.

Available color settings:

#### Primary Colors
- `primary`: Main brand/action color
- `secondary`: Secondary actions and elements

#### Text Colors
- `text`: Regular text
- `text_muted`: De-emphasized text (faint)
- `text_subtle`: Subtle text (between regular and muted)
- `text_title`: Title text
- `text_bold`: Emphasized text

#### Structure Colors
- `border`: Borders and dividers
- `surface`: Background surfaces
- `overlay`: Overlays and modals

#### State Colors
- `success`: Success states
- `warning`: Warning states
- `error`: Error states
- `info`: Informational states

#### Tree-Specific Colors
- `tree_connector`: Tree structure lines (├── └──)
- `tree_directory`: Directory nodes
- `tree_file`: File nodes
- `tree_annotation`: Annotation text

### Text Style Options

- `annotated_bold`: Whether to use bold for annotated paths (default: true)
- `unannotated_faint`: Whether to use faint for unannotated paths (default: true)
- `root_bold`: Whether to use bold for root paths (default: true)

## Examples

### Minimal Configuration

Override just a few colors:

```yaml
version: "1"
styles:
  theme: dark
  themes:
    dark:
      colors:
        primary: "#FF79C6"        # Pink
        tree_directory: "#BD93F9" # Purple
        tree_annotation: "#50FA7B" # Green
```

### Disable Bold Text

```yaml
version: "1"
styles:
  theme: auto
  themes:
    light:
      text_style:
        annotated_bold: false
        root_bold: false
    dark:
      text_style:
        annotated_bold: false
        root_bold: false
```

### Custom Theme

Create a completely custom color scheme:

```yaml
version: "1"
styles:
  theme: dark
  themes:
    dark:
      colors:
        primary: "#E06C75"
        text: "#ABB2BF"
        text_muted: "#5C6370"
        tree_connector: "#4B5263"
        tree_directory: "#61AFEF"
        tree_annotation: "#98C379"
        error: "#E06C75"
        warning: "#E5C07B"
        success: "#98C379"
```

## Tips

1. Start with the provided `treex.yaml` example file and modify it to your needs
2. You only need to specify the colors you want to change - unspecified colors will use defaults
3. Use the `treex theme-demo` command to see how your colors look (if available)
4. For terminal compatibility, consider using ANSI color codes (0-255) instead of hex colors