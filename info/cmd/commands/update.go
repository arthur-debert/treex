package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <path> <annotation>",
		Short: "Update an existing annotation in .info file",
		Long: `Update modifies an existing annotation for the specified path in the .info file.
By default, it uses the .info file in the current directory, but you can specify
a different file using the --info-file flag.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := args[0]
			newAnnotation := args[1]

			err := cli.api.Update(infoFile, targetPath, newAnnotation)
			if err != nil {
				return err
			}

			if outputJSON {
				result := map[string]string{
					"status":     "success",
					"message":    "Annotation updated successfully",
					"path":       targetPath,
					"annotation": newAnnotation,
					"info_file":  infoFile,
				}
				return outputResult(result)
			} else {
				fmt.Printf("✓ Updated annotation for '%s' in %s\n", targetPath, infoFile)
				return nil
			}
		},
	}

	return cmd
}
