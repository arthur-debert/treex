package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Get annotation for a specific path",
		Long: `Get retrieves the effective annotation for a specific path, taking into account
annotation precedence rules from multiple .info files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := args[0]

			annotation, err := cli.api.GetAnnotation(targetPath)
			if err != nil {
				if outputJSON {
					result := map[string]interface{}{
						"path":       targetPath,
						"annotation": nil,
						"found":      false,
						"error":      err.Error(),
					}
					return outputResult(result)
				} else {
					return fmt.Errorf("no annotation found for path '%s'", targetPath)
				}
			}

			if outputJSON {
				result := map[string]interface{}{
					"path":       annotation.Path,
					"annotation": annotation.Annotation,
					"info_file":  annotation.InfoFile,
					"line_num":   annotation.LineNum,
					"found":      true,
				}
				return outputResult(result)
			} else {
				return outputResult(annotation)
			}
		},
	}

	return cmd
}
