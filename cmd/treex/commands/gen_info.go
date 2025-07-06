package commands

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/adebert/treex/pkg/edit/geninfo"
	"github.com/spf13/cobra"
)

//go:embed gen_info.help.txt
var genInfoHelp string

var importCmd = &cobra.Command{
	Use:     "import [file]",
	Short:   "Generate .info files from annotated tree structure",
	GroupID: "info",
	Long: genInfoHelp,
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
