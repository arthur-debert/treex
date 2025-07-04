package main

import (
	"fmt"
	"os"

	"github.com/adebert/treex/cmd/treex/commands"
)

// version is set by the build system via ldflags
var version = "dev"

func main() {
	// Pass the version to the command package
	commands.SetVersion(version)
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

