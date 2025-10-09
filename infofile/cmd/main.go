package main

import (
	"os"

	"github.com/jwaldrip/treex/infofile/cmd/commands"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	rootCmd := commands.NewRootCommand()

	// Set version information
	commands.SetVersionInfo(version, commit, buildDate)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
