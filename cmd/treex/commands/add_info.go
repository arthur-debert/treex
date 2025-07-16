package commands

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/edit/addinfo"
	"github.com/spf13/cobra"
)

//go:embed add_info.help.txt
var addInfoHelp string

var addInfoCmd = &cobra.Command{
	Use:     "add <path> <description>",
	Short:   "Add or update an entry in the current directory's .info file",
	GroupID: "info",
	Long:    addInfoHelp,
	Args:    cobra.MinimumNArgs(2),
	RunE:    runAddInfoCmd,
}

func init() {
	// Add flags specific to add command
	addInfoCmd.Flags().Bool("replace", false, "Replace existing entry without prompting")

	// Register the command with root
	rootCmd.AddCommand(addInfoCmd)
}

// runAddInfoCmd handles the CLI interface for add command
func runAddInfoCmd(cmd *cobra.Command, args []string) error {
	path := args[0]
	// Join all remaining arguments as the description
	description := strings.Join(args[1:], " ")

	replace, err := cmd.Flags().GetBool("replace")
	if err != nil {
		return fmt.Errorf("failed to get replace flag: %w", err)
	}

	// Use current directory
	currentDir := "."

	// Delegate to business logic
	result, err := addinfo.AddInfoEntry(currentDir, path, description, replace, promptUser)
	if err != nil {
		return err
	}

	// Handle the result
	switch result.Action {
	case addinfo.ActionAdded:
		fmt.Printf("Added entry for '%s' to .info file\n", path)
	case addinfo.ActionUpdated:
		fmt.Printf("Updated entry for '%s' in .info file\n", path)
	case addinfo.ActionCancelled:
		fmt.Println("Operation cancelled.")
	}

	return nil
}

// promptUser handles user interaction for the CLI
func promptUser(path, currentDesc, newDesc string) (addinfo.UserChoice, error) {
	fmt.Printf("Entry for '%s' already exists:\n", path)
	fmt.Printf("Current description: %s\n\n", currentDesc)
	fmt.Printf("New description: %s\n\n", newDesc)

	fmt.Print("Choose action: (r)eplace, (a)ppend, or (q)uit [r/a/q]: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return addinfo.UserChoiceQuit, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "r", "replace":
		return addinfo.UserChoiceReplace, nil
	case "a", "append":
		return addinfo.UserChoiceAppend, nil
	case "q", "quit", "abort":
		return addinfo.UserChoiceQuit, nil
	default:
		return addinfo.UserChoiceQuit, fmt.Errorf("invalid choice: %s", response)
	}
}
