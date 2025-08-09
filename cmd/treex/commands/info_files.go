package commands

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed infofiles.txt
var infoFilesContent string

var infoFilesCmd = &cobra.Command{
	Use:     "info-help",
	Short:   "Show information about .info file format and usage",
	GroupID: "help",
	Long:    `Display compact reference information about .info files and their format.`,
	RunE:    runInfoFilesCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(infoFilesCmd)
}

// runInfoFilesCmd handles the CLI interface for info-files command
func runInfoFilesCmd(cmd *cobra.Command, args []string) error {
	fmt.Print(infoFilesContent)
	return nil
}
