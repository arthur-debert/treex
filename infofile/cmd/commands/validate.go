package commands

import (
	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate .info files for errors",
		Long: `Validate checks all .info files in the specified directory tree for syntax errors,
duplicate paths, invalid formats, and other issues. It reports any problems found
and provides suggestions for fixing them.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getPathArg(args, ".")

			result, err := cli.api.Validate(path)
			if err != nil {
				return err
			}

			return outputResult(result)
		},
	}

	return cmd
}
