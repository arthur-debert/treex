package main

import (
	"fmt"
	"os"

	"github.com/adebert/treex/cmd/treex/cmd"
)

// version is set by the build system via ldflags
var version = "dev"

func main() {
	// Pass the version to the command package
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

