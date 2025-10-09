package main

import (
	"treex/treex/cmd"
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
