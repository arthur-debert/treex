package cmd

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/info"
	"github.com/spf13/cobra"
)

var addInfoCmd = &cobra.Command{
	Use:   "add <path> <description>",
	Short: "Add or update an entry in the current directory's .info file",
	Long: `Add or update an entry in the current directory's .info file.

This command will:
- Find the .info file in the current directory or create one if it doesn't exist
- Look for an existing entry for the specified path
- If an entry exists, prompt to replace, append, or abort (unless --replace is used)
- Add or update the entry with the provided description

Examples:
  treex add pkg "Main package containing core functionality"
  treex add config/ "Configuration files and settings"
  treex add --replace main.go "Application entry point"`,
	Args: cobra.ExactArgs(2),
	RunE: runAddInfoCmd,
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
	description := args[1]
	
	replace, err := cmd.Flags().GetBool("replace")
	if err != nil {
		return fmt.Errorf("failed to get replace flag: %w", err)
	}
	
	// Use current directory
	currentDir := "."
	
	// Delegate to business logic
	result, err := info.AddInfoEntry(currentDir, path, description, replace, promptUser)
	if err != nil {
		return err
	}
	
	// Handle the result
	switch result.Action {
	case info.ActionAdded:
		fmt.Printf("Added entry for '%s' to .info file\n", path)
	case info.ActionUpdated:
		fmt.Printf("Updated entry for '%s' in .info file\n", path)
	case info.ActionCancelled:
		fmt.Println("Operation cancelled.")
	}
	
	return nil
}

// promptUser handles user interaction for the CLI
func promptUser(path, currentDesc, newDesc string) (info.UserChoice, error) {
	fmt.Printf("Entry for '%s' already exists:\n", path)
	fmt.Printf("Current description: %s\n\n", currentDesc)
	fmt.Printf("New description: %s\n\n", newDesc)
	
	fmt.Print("Choose action: (r)eplace, (a)ppend, or (q)uit [r/a/q]: ")
	
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return info.UserChoiceQuit, fmt.Errorf("failed to read user input: %w", err)
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	
	switch response {
	case "r", "replace":
		return info.UserChoiceReplace, nil
	case "a", "append":
		return info.UserChoiceAppend, nil
	case "q", "quit", "abort":
		return info.UserChoiceQuit, nil
	default:
		return info.UserChoiceQuit, fmt.Errorf("invalid choice: %s", response)
	}
} 