//go:build ignore
// +build ignore

// Package main provides artifact generation for packaging treex distribution.
// This generates shell completions that are included in goreleaser builds.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"treex/treex/cmd"
)

func main() {
	// Use .dist-artifacts directory for artifacts that goreleaser will package
	artifactsDir := ".dist-artifacts"
	completionsDir := filepath.Join(artifactsDir, "completions")

	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		log.Fatalf("Failed to create completions directory: %v", err)
	}

	// Get the root command
	rootCmd := cmd.NewRootCommand()

	// Generate shell completions
	fmt.Println("Generating shell completions...")

	// Bash completion
	bashFile := filepath.Join(completionsDir, "treex.bash")
	if err := rootCmd.GenBashCompletionFile(bashFile); err != nil {
		log.Fatalf("Failed to generate bash completion: %v", err)
	}
	fmt.Printf("✓ Bash completion: %s\n", bashFile)

	// Zsh completion
	zshFile := filepath.Join(completionsDir, "_treex")
	if err := rootCmd.GenZshCompletionFile(zshFile); err != nil {
		log.Fatalf("Failed to generate zsh completion: %v", err)
	}
	fmt.Printf("✓ Zsh completion: %s\n", zshFile)

	// Fish completion
	fishFile := filepath.Join(completionsDir, "treex.fish")
	if err := rootCmd.GenFishCompletionFile(fishFile, true); err != nil {
		log.Fatalf("Failed to generate fish completion: %v", err)
	}
	fmt.Printf("✓ Fish completion: %s\n", fishFile)

	fmt.Println("✅ Shell completions generated successfully!")
}
