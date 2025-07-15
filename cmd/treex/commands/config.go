package commands

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed config.yaml
var defaultConfigYAML string

// configCmd outputs the default configuration to stdout
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Output default configuration file",
	GroupID: "help",
	Long: `Output the default treex configuration file to stdout.

This command prints a fully documented treex.yaml configuration file
that you can use as a starting point for customization.

Examples:
  # Save default config to a file
  treex config > treex.yaml
  
  # Save to home directory config
  treex config > ~/.config/treex/treex.yaml
  
  # View the default configuration
  treex config | less`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), defaultConfigYAML)
		return err
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}