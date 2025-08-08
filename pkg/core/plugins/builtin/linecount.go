package builtin

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adebert/treex/pkg/core/types"
)

// LineCountPlugin collects line count information for text files
type LineCountPlugin struct{}

// Name returns the plugin identifier
func (p *LineCountPlugin) Name() string {
	return "lc"
}

// Description returns the plugin description
func (p *LineCountPlugin) Description() string {
	return "Display line counts for text files"
}

// AppliesTo returns true for directories and text files
func (p *LineCountPlugin) AppliesTo(node *types.Node) bool {
	if node.IsDir {
		return true // Will aggregate from children
	}
	
	// Check if file extension suggests it's a text file
	return p.isTextFile(node.Name)
}

// isTextFile determines if a file is likely a text file based on extension
func (p *LineCountPlugin) isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Common text file extensions
	textExtensions := map[string]bool{
		".go":     true, // Go source
		".py":     true, // Python
		".js":     true, // JavaScript
		".ts":     true, // TypeScript
		".jsx":    true, // React JSX
		".tsx":    true, // React TSX
		".java":   true, // Java
		".c":      true, // C
		".cpp":    true, // C++
		".cc":     true, // C++
		".cxx":    true, // C++
		".h":      true, // C/C++ header
		".hpp":    true, // C++ header
		".cs":     true, // C#
		".php":    true, // PHP
		".rb":     true, // Ruby
		".rs":     true, // Rust
		".swift":  true, // Swift
		".kt":     true, // Kotlin
		".scala":  true, // Scala
		".clj":    true, // Clojure
		".hs":     true, // Haskell
		".ml":     true, // OCaml
		".fs":     true, // F#
		".elm":    true, // Elm
		".dart":   true, // Dart
		".r":      true, // R
		".m":      true, // Objective-C/MATLAB
		".pl":     true, // Perl
		".sh":     true, // Shell script
		".bash":   true, // Bash script
		".zsh":    true, // Zsh script
		".fish":   true, // Fish script
		".ps1":    true, // PowerShell
		".bat":    true, // Batch file
		".cmd":    true, // Command file
		".txt":    true, // Plain text
		".md":     true, // Markdown
		".rst":    true, // reStructuredText
		".tex":    true, // LaTeX
		".html":   true, // HTML
		".htm":    true, // HTML
		".xml":    true, // XML
		".yaml":   true, // YAML
		".yml":    true, // YAML
		".json":   true, // JSON
		".toml":   true, // TOML
		".ini":    true, // INI file
		".cfg":    true, // Config file
		".conf":   true, // Config file
		".config": true, // Config file
		".css":    true, // CSS
		".scss":   true, // SCSS
		".sass":   true, // Sass
		".less":   true, // Less
		".sql":    true, // SQL
		".dockerfile": true, // Dockerfile
		".gitignore":  true, // Git ignore
		".gitattributes": true, // Git attributes
		".editorconfig": true, // Editor config
		".env":    true, // Environment file
		".log":    true, // Log file
		".csv":    true, // CSV file
		".tsv":    true, // TSV file
		".proto":  true, // Protocol Buffers
		".graphql": true, // GraphQL
		".gql":    true, // GraphQL
		".vue":    true, // Vue.js
		".svelte": true, // Svelte
		".astro":  true, // Astro
	}
	
	// Also check for files without extensions that are commonly text files
	if ext == "" {
		baseName := strings.ToLower(filepath.Base(filename))
		textFiles := map[string]bool{
			"dockerfile":     true,
			"makefile":      true,
			"gemfile":       true,
			"rakefile":      true,
			"vagrantfile":   true,
			"readme":        true,
			"license":       true,
			"authors":       true,
			"contributors":  true,
			"changelog":     true,
			"changes":       true,
			"news":          true,
			"todo":          true,
			"copying":       true,
			"install":       true,
			"thanks":        true,
			"acknowledgments": true,
		}
		return textFiles[baseName]
	}
	
	return textExtensions[ext]
}

// Collect gathers line count information
func (p *LineCountPlugin) Collect(node *types.Node) (map[string]interface{}, error) {
	// For directories, the line count will be aggregated later
	if node.IsDir {
		return make(map[string]interface{}), nil
	}
	
	// Skip non-text files
	if !p.isTextFile(node.Name) {
		return nil, nil
	}
	
	// Count lines in the file
	lineCount, err := p.countLines(node.Path)
	if err != nil {
		// If we can't count lines, return empty metadata rather than error
		// This allows the plugin to gracefully handle files that may not be accessible
		return make(map[string]interface{}), nil
	}
	
	return map[string]interface{}{
		"lc_lines":        int64(lineCount),
		"lc_display_text": p.formatLineCount(lineCount),
	}, nil
}

// countLines counts the number of lines in a file
func (p *LineCountPlugin) countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = file.Close()
	}()
	
	scanner := bufio.NewScanner(file)
	lineCount := 0
	
	for scanner.Scan() {
		lineCount++
	}
	
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	
	return lineCount, nil
}

// formatLineCount formats the line count for display
func (p *LineCountPlugin) formatLineCount(lines int) string {
	if lines == 0 {
		return "0 lines"
	} else if lines == 1 {
		return "1 line"
	} else if lines < 1000 {
		return strconv.Itoa(lines) + " lines"
	} else if lines < 1000000 {
		// Show as "1.2K lines" for readability
		k := float64(lines) / 1000.0
		if k == float64(int(k)) {
			return strconv.Itoa(int(k)) + "K lines"
		}
		return strconv.FormatFloat(k, 'f', 1, 64) + "K lines"
	} else {
		// Show as "1.2M lines" for very large files
		m := float64(lines) / 1000000.0
		if m == float64(int(m)) {
			return strconv.Itoa(int(m)) + "M lines"
		}
		return strconv.FormatFloat(m, 'f', 1, 64) + "M lines"
	}
}

// Format returns a formatted string representation of the line count
func (p *LineCountPlugin) Format(metadata map[string]interface{}) string {
	// Check if we have pre-formatted display text
	if displayText, exists := metadata["lc_display_text"]; exists {
		if displayTextStr, ok := displayText.(string); ok {
			return displayTextStr
		}
	}
	
	// Otherwise, format from line count
	if lines, exists := metadata["lc_lines"]; exists {
		if linesInt64, ok := lines.(int64); ok {
			return p.formatLineCount(int(linesInt64))
		}
	}
	
	return ""
}