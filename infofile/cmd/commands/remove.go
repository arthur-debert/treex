package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove an annotation from .info file",
		Long: `Remove deletes an annotation for the specified path from the .info file.
By default, it uses the .info file in the current directory, but you can specify
a different file using the --info-file flag.`,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"rm", "delete"},
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := args[0]

			err := cli.api.Remove(infoFile, targetPath)
			if err != nil {
				return err
			}

			if outputJSON {
				result := map[string]string{
					"status":    "success",
					"message":   "Annotation removed successfully",
					"path":      targetPath,
					"info_file": infoFile,
				}
				return outputResult(result)
			} else {
				fmt.Printf("✓ Removed annotation for '%s' from %s\n", targetPath, infoFile)
				return nil
			}
		},
	}

	return cmd
}
