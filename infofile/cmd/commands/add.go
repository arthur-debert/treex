package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <path> <annotation>",
		Short: "Add a new annotation to .info file",
		Long: `Add creates a new annotation for the specified path in the .info file.
By default, it uses the .info file in the current directory, but you can specify
a different file using the --info-file flag.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := args[0]
			annotation := args[1]

			err := cli.api.Add(infoFile, targetPath, annotation)
			if err != nil {
				return err
			}

			if outputJSON {
				result := map[string]string{
					"status":     "success",
					"message":    "Annotation added successfully",
					"path":       targetPath,
					"annotation": annotation,
					"info_file":  infoFile,
				}
				return outputResult(result)
			} else {
				fmt.Printf("✓ Added annotation for '%s' to %s\n", targetPath, infoFile)
				return nil
			}
		},
	}

	return cmd
}
