package firstuse

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// CommonPath represents a common development path with suggested annotations
type CommonPath struct {
	Path       string
	Annotation string
	Priority   int // Higher priority means more likely to be selected
}

// GetCommonPaths returns a curated list of common development paths and their annotations
func GetCommonPaths() []CommonPath {
	return []CommonPath{
		// Source code directories
		{Path: "src/", Annotation: "Main source code directory", Priority: 10},
		{Path: "lib/", Annotation: "Library code and utilities", Priority: 9},
		{Path: "app/", Annotation: "Application main code", Priority: 9},

		// Test directories
		{Path: "tests/", Annotation: "Test files and test utilities", Priority: 9},
		{Path: "test/", Annotation: "Unit and integration tests", Priority: 9},
		{Path: "__tests__/", Annotation: "Jest test directory", Priority: 7},
		{Path: "spec/", Annotation: "RSpec or test specifications", Priority: 7},

		// Build and output
		{Path: "build/", Annotation: "Build output and artifacts", Priority: 8},
		{Path: "dist/", Annotation: "Distribution-ready files", Priority: 8},
		{Path: "out/", Annotation: "Compiled output directory", Priority: 7},
		{Path: "target/", Annotation: "Maven/Rust build output", Priority: 7},

		// Documentation
		{Path: "docs/", Annotation: "Project documentation", Priority: 8},
		{Path: "doc/", Annotation: "API and user documentation", Priority: 7},

		// Configuration
		{Path: "config/", Annotation: "Configuration files", Priority: 7},
		{Path: ".github/", Annotation: "GitHub workflows and configs", Priority: 7},

		// Common files
		{Path: "README.md", Annotation: "Project overview and setup guide", Priority: 9},
		{Path: "package.json", Annotation: "Node.js project configuration", Priority: 8},
		{Path: "Cargo.toml", Annotation: "Rust project manifest", Priority: 7},
		{Path: "go.mod", Annotation: "Go module definition", Priority: 7},
		{Path: "requirements.txt", Annotation: "Python dependencies", Priority: 7},
		{Path: "Gemfile", Annotation: "Ruby dependencies", Priority: 6},
		{Path: "Makefile", Annotation: "Build automation rules", Priority: 7},
		{Path: "Dockerfile", Annotation: "Container build instructions", Priority: 7},
		{Path: ".gitignore", Annotation: "Git ignore patterns", Priority: 6},
		{Path: ".env.example", Annotation: "Environment variables template", Priority: 6},

		// Frontend specific
		{Path: "components/", Annotation: "React/Vue/Angular components", Priority: 7},
		{Path: "pages/", Annotation: "Application pages or routes", Priority: 7},
		{Path: "public/", Annotation: "Static assets served directly", Priority: 7},
		{Path: "assets/", Annotation: "Images, fonts, and static files", Priority: 7},
		{Path: "styles/", Annotation: "CSS/SCSS stylesheets", Priority: 6},

		// Backend specific
		{Path: "api/", Annotation: "API endpoints and handlers", Priority: 8},
		{Path: "routes/", Annotation: "URL routing definitions", Priority: 7},
		{Path: "models/", Annotation: "Data models and schemas", Priority: 7},
		{Path: "controllers/", Annotation: "Request handlers and business logic", Priority: 7},
		{Path: "services/", Annotation: "Business logic and external services", Priority: 7},
		{Path: "utils/", Annotation: "Utility functions and helpers", Priority: 6},

		// Database
		{Path: "migrations/", Annotation: "Database migration files", Priority: 7},
		{Path: "seeds/", Annotation: "Database seed data", Priority: 6},

		// Scripts
		{Path: "scripts/", Annotation: "Utility and automation scripts", Priority: 6},
		{Path: "bin/", Annotation: "Executable scripts and binaries", Priority: 6},

		// Mobile specific
		{Path: "ios/", Annotation: "iOS platform-specific code", Priority: 6},
		{Path: "android/", Annotation: "Android platform-specific code", Priority: 6},

		// Dependencies
		{Path: "vendor/", Annotation: "Vendored dependencies", Priority: 5},
		{Path: "node_modules/", Annotation: "Node.js dependencies", Priority: 4},
		{Path: ".venv/", Annotation: "Python virtual environment", Priority: 4},
	}
}

// FindExamplesInPath searches for common paths in the given directory
// and returns up to maxExamples found paths with their suggested annotations
func FindExamplesInPath(basePath string, maxExamples int) []CommonPath {
	commonPaths := GetCommonPaths()
	var foundExamples []CommonPath

	// Check each common path
	for _, commonPath := range commonPaths {
		fullPath := filepath.Join(basePath, commonPath.Path)

		// Check if the path exists
		if _, err := os.Stat(fullPath); err == nil {
			foundExamples = append(foundExamples, commonPath)
		}
	}

	// Sort by priority (higher priority first)
	// Simple bubble sort for small dataset
	for i := 0; i < len(foundExamples); i++ {
		for j := i + 1; j < len(foundExamples); j++ {
			if foundExamples[j].Priority > foundExamples[i].Priority {
				foundExamples[i], foundExamples[j] = foundExamples[j], foundExamples[i]
			}
		}
	}

	// Return up to maxExamples
	if len(foundExamples) > maxExamples {
		return foundExamples[:maxExamples]
	}

	return foundExamples
}

// GetFallbackExamples returns random examples when no common paths are found
func GetFallbackExamples(basePath string, maxExamples int) []CommonPath {
	var examples []CommonPath

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return examples
	}

	// Collect all entries (excluding hidden files)
	var candidates []struct {
		name  string
		isDir bool
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".") || name == ".gitignore" || name == ".env" {
			candidates = append(candidates, struct {
				name  string
				isDir bool
			}{name: name, isDir: entry.IsDir()})
		}
	}

	// Shuffle and pick up to maxExamples
	for i := range candidates {
		j := rand.Intn(i + 1)
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}

	for i := 0; i < len(candidates) && i < maxExamples; i++ {
		path := candidates[i].name
		if candidates[i].isDir {
			path += "/"
		}

		annotation := generateGenericAnnotation(candidates[i].name, candidates[i].isDir)
		examples = append(examples, CommonPath{
			Path:       path,
			Annotation: annotation,
			Priority:   5,
		})
	}

	return examples
}

// generateGenericAnnotation creates a generic annotation based on file/dir name
func generateGenericAnnotation(name string, isDir bool) string {
	if isDir {
		// Try to guess based on common patterns
		nameLower := strings.ToLower(name)

		if strings.Contains(nameLower, "test") {
			return "Test files and utilities"
		}
		if strings.Contains(nameLower, "doc") {
			return "Documentation files"
		}
		if strings.Contains(nameLower, "src") || strings.Contains(nameLower, "source") {
			return "Source code files"
		}
		if strings.Contains(nameLower, "lib") {
			return "Library code"
		}
		if strings.Contains(nameLower, "asset") || strings.Contains(nameLower, "static") {
			return "Static assets"
		}
		if strings.Contains(nameLower, "config") {
			return "Configuration files"
		}

		return fmt.Sprintf("%s directory", name)
	}

	// For files, look at extension
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".md":
		return "Markdown documentation"
	case ".json":
		return "JSON configuration file"
	case ".yaml", ".yml":
		return "YAML configuration file"
	case ".js", ".ts", ".jsx", ".tsx":
		return "JavaScript/TypeScript source file"
	case ".py":
		return "Python source file"
	case ".go":
		return "Go source file"
	case ".rs":
		return "Rust source file"
	case ".java":
		return "Java source file"
	case ".cpp", ".cc", ".h", ".hpp":
		return "C++ source file"
	case ".c":
		return "C source file"
	case ".rb":
		return "Ruby source file"
	case ".sh":
		return "Shell script"
	case ".sql":
		return "SQL database script"
	case ".css", ".scss", ".sass":
		return "Stylesheet file"
	case ".html":
		return "HTML template"
	default:
		return fmt.Sprintf("%s file", strings.TrimPrefix(ext, "."))
	}
}
