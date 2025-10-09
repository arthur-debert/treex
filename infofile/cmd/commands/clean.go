package commands

import (
	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean [path]",
		Short: "Clean invalid and redundant annotations",
		Long: `Clean removes invalid or redundant annotations from .info files in the specified
directory tree. This includes annotations for non-existent paths, duplicate entries,
and other problematic annotations.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getPathArg(args, ".")

			result, err := cli.api.Clean(path)
			if err != nil {
				return err
			}

			return outputResult(result)
		},
	}

	return cmd
}
