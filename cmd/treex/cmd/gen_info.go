package cmd

import (
	"fmt"

	"github.com/adebert/treex/pkg/info"
	"github.com/spf13/cobra"
)

var genInfoCmd = &cobra.Command{
	Use:   "gen-info <file>",
	Short: "Generate .info files from annotated tree structure",
	Long: `Generate .info files from a hand-written annotated tree structure.

The input file should contain a tree-like structure with paths and descriptions.
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
in the corresponding directories.`,
	Args: cobra.ExactArgs(1),
	RunE: runGenInfoCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(genInfoCmd)
}

// runGenInfoCmd handles the CLI interface for gen-info command
func runGenInfoCmd(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	
	// Delegate to business logic
	if err := info.GenerateInfoFromTree(inputFile); err != nil {
		return fmt.Errorf("failed to generate .info files: %w", err)
	}
	
	fmt.Println("Info files generated successfully")
	return nil
} 