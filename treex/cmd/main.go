package main

import (
	"fmt"
	"os"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("treex version %s (commit: %s, built: %s)\n", version, commit, buildDate)
		return
	}
	fmt.Println("treex - modernized file tree viewer (stub implementation)")
}
