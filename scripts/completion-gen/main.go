package main

import (
	"os"
	"path/filepath"

	"github.com/adebert/treex/cmd/treex/cmd"
)

func main() {
	// Create completions directory
	if err := os.MkdirAll("completions", 0755); err != nil {
		panic(err)
	}

	// Get the root command
	rootCmd := cmd.GetRootCommand()

	// Generate bash completion
	bashFile, err := os.Create(filepath.Join("completions", "treex.bash"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := bashFile.Close(); err != nil {
			panic(err)
		}
	}()
	
	if err := rootCmd.GenBashCompletion(bashFile); err != nil {
		panic(err)
	}

	// Generate zsh completion
	zshFile, err := os.Create(filepath.Join("completions", "_treex"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := zshFile.Close(); err != nil {
			panic(err)
		}
	}()
	
	if err := rootCmd.GenZshCompletion(zshFile); err != nil {
		panic(err)
	}

	// Generate fish completion
	fishFile, err := os.Create(filepath.Join("completions", "treex.fish"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fishFile.Close(); err != nil {
			panic(err)
		}
	}()
	
	if err := rootCmd.GenFishCompletion(fishFile, true); err != nil {
		panic(err)
	}
} 