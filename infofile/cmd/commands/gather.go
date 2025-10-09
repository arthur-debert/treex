package commands

import (
	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newGatherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gather [path]",
		Short: "Collect and display all annotations from .info files",
		Long: `Gather collects all annotations from .info files in the specified directory tree
and displays them. This shows the effective annotations after resolving conflicts
between multiple .info files.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getPathArg(args, ".")

			annotations, err := cli.api.Gather(path)
			if err != nil {
				return err
			}

			return outputResult(annotations)
		},
	}

	return cmd
}

func (cli *InfoCLI) newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [path]",
		Short:   "Show all annotations (alias for gather)",
		Long:    "Show displays all annotations from .info files. This is an alias for the gather command.",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"list"},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getPathArg(args, ".")

			annotations, err := cli.api.List(path)
			if err != nil {
				return err
			}

			return outputResult(annotations)
		},
	}

	return cmd
}
