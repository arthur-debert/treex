package main

import (
	"os"

	"github.com/adebert/treex/cmd/treex/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Create man directory
	if err := os.MkdirAll("man/man1", 0755); err != nil {
		panic(err)
	}

	// Get the root command
	rootCmd := cmd.GetRootCommand()

	// Generate man pages
	header := &doc.GenManHeader{
		Title:   "TREEX",
		Section: "1",
		Source:  "Treex CLI",
	}

	if err := doc.GenManTree(rootCmd, header, "man/man1"); err != nil {
		panic(err)
	}
} 