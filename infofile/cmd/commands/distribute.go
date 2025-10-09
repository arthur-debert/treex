package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *InfoCLI) newDistributeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distribute [path]",
		Short: "Redistribute annotations to optimal .info files",
		Long: `Distribute redistributes annotations to their optimal .info files based on path proximity.
This moves annotations to the .info file that is closest to the annotated path,
optimizing the file structure for better organization.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getPathArg(args, ".")

			err := cli.api.Distribute(path)
			if err != nil {
				return err
			}

			if outputJSON {
				return outputResult(map[string]string{"status": "success", "message": "Annotations redistributed successfully"})
			} else {
				fmt.Println("✓ Annotations redistributed successfully")
				return nil
			}
		},
	}

	return cmd
}
