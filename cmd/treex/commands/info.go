package commands

import (
	"github.com/spf13/cobra"
)

// infoCmd represents the info command group
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Manage .info files and annotations",
	Long: `The info command group provides tools for creating, editing, and managing .info files
that contain file and directory annotations. These annotations are displayed when
using treex to show directory structures.

Use "treex info [command] --help" for more information about a specific subcommand.`,
}

func init() {
	// Register the info command with root
	rootCmd.AddCommand(infoCmd)
}