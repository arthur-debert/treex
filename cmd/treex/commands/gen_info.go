package commands

import (
	"fmt"
	"os"

	"github.com/adebert/treex/pkg/edit/geninfo"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:     "import [file]",
	Short:   "Generate .info files from annotated tree structure",
	GroupID: "info",
	Long: `Generate .info files from a hand-written annotated tree structure.

The input can come from a file or be piped via stdin. Use "-" or omit the file argument to read from stdin.

The input should contain a tree-like structure with paths and descriptions.
Tree connectors are optional - you can use a simple format:

Simple format:
    myproject/cmd The go code for the cli utility
    myproject/docs All documentation  
    myproject/pkg The main parser code
    myproject/scripts Various utilities

Or with traditional tree connectors:
    myproject/
    ├── cmd/ The go code for the cli utility
    ├── docs/ All documentation
    │   └── dev/ Dev related, including technical topics
    ├── pkg/ The main parser code
    └── scripts/ Various utilities

Both formats work equally well. This will generate appropriate .info files 
in the corresponding directories.

Examples:
  treex import structure.txt           # Read from file
  treex import                         # Read from stdin
  treex import -                       # Read from stdin (explicit)
  echo "project/src Code" | treex import  # Pipe content`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImportCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(importCmd)
}

// runImportCmd handles the CLI interface for import command
func runImportCmd(cmd *cobra.Command, args []string) error {
	var err error

	// Check if reading from stdin or file
	if len(args) == 0 || args[0] == "-" {
		// Read from stdin
		err = geninfo.GenerateInfoFromReader(os.Stdin)
	} else {
		// Read from file
		inputFile := args[0]
		err = geninfo.GenerateInfoFromTree(inputFile)
	}

	if err != nil {
		return fmt.Errorf("failed to generate .info files: %w", err)
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Info files generated successfully"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}
