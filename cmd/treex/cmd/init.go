package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a .info file for a directory",
	Long: `Generate a .info file for the specified directory (or current directory if not specified).

This command will:
- Scan the directory structure up to a specified depth (default: 3)
- Create a .info file with entries for all files and directories found
- Skip files that are typically not documented (like .git, node_modules, etc.)

The generated .info file will contain empty descriptions that you can fill in later.

Examples:
  treex init              # Initialize .info file for current directory
  treex init ./src        # Initialize .info file for src directory
  treex init --depth=2    # Initialize with depth limit of 2`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInitCmd,
}

func init() {
	// Add flags specific to init command
	initCmd.Flags().IntP("depth", "d", 3, "Maximum depth to scan (default: 3)")
	
	// Register the command with root
	rootCmd.AddCommand(initCmd)
}

// runInitCmd handles the CLI interface for init command
func runInitCmd(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}
	
	// Get depth flag
	depth, err := cmd.Flags().GetInt("depth")
	if err != nil {
		return fmt.Errorf("failed to get depth flag: %w", err)
	}
	
	// Delegate to business logic
	err = initializeInfoFile(targetPath, depth)
	if err != nil {
		return err
	}
	
	fmt.Printf("Initialized .info file for '%s' (depth: %d)\n", targetPath, depth)
	return nil
}

// initializeInfoFile creates a .info file for the given directory
func initializeInfoFile(targetPath string, depth int) error {
	// Ensure the target path exists and is a directory
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", targetPath)
	}
	
	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", targetPath)
	}
	
	// Build tree with depth limit and no ignore files (we want to see everything for init)
	builder, err := tree.NewBuilderWithOptions(absPath, make(map[string]*info.Annotation), "", depth)
	if err != nil {
		return fmt.Errorf("failed to create tree builder: %w", err)
	}
	
	root, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build directory tree: %w", err)
	}
	
	// Extract all files and directories from the tree that are direct children of root
	var entries []string
	for _, child := range root.Children {
		// Skip the "... more files" indicator nodes
		if child.Path == "" {
			continue
		}
		
		// Add the file/directory name with trailing slash for directories
		entryName := child.Name
		if child.IsDir {
			entryName += "/"
		}
		
		entries = append(entries, entryName)
	}
	
	// Check if .info file already exists
	infoPath := filepath.Join(absPath, ".info")
	if _, err := os.Stat(infoPath); err == nil {
		// File exists, ask for confirmation
		fmt.Printf(".info file already exists in %s. Overwrite? [y/N]: ", targetPath)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}
	
	// Create the .info file
	file, err := os.Create(infoPath)
	if err != nil {
		return fmt.Errorf("failed to create .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()
	
	// Write entries to the file
	for _, entry := range entries {
		if _, err := fmt.Fprintf(file, "%s\n", entry); err != nil {
			return fmt.Errorf("failed to write to .info file: %w", err)
		}
		
		// Add an empty description line
		if _, err := fmt.Fprintf(file, "\n\n"); err != nil {
			return fmt.Errorf("failed to write to .info file: %w", err)
		}
	}
	
	return nil
} 