package main

import (
	"github.com/jwaldrip/treex/treex/internal/cmd"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Set version information for the CLI
	cmd.Version = version
	cmd.Commit = commit
	cmd.BuildDate = buildDate

	// Execute the root command
	cmd.Execute()
}
