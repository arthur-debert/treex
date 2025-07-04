package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoFilesCmd = &cobra.Command{
	Use:     "info-files",
	Short:   "Show information about .info file format and usage",
	GroupID: "help",
	Long:    `Display compact reference information about .info files and their format.`,
	RunE:    runInfoFilesCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(infoFilesCmd)
}

// runInfoFilesCmd handles the CLI interface for info-files command
func runInfoFilesCmd(cmd *cobra.Command, args []string) error {
	showInfoFilesHelp()
	return nil
}

// showInfoFilesHelp displays compact information about .info files
func showInfoFilesHelp() {
	fmt.Println("=== .info Files Quick Reference ===")
	fmt.Println()
	fmt.Println("📁 What are .info files?")
	fmt.Println("   .info files contain descriptions for files and directories.")
	fmt.Println("   They make your codebase self-documenting and easier to understand.")
	fmt.Println()
	fmt.Println("📝 Basic Format:")
	fmt.Println("   path/to/file")
	fmt.Println("   Description of what this file does")
	fmt.Println()
	fmt.Println("   another-file.js")
	fmt.Println("   Single line description")
	fmt.Println()
	fmt.Println("💡 Example .info file:")
	fmt.Println("   README.md")
	fmt.Println("   Main project documentation")
	fmt.Println()
	fmt.Println("   src/main.go")
	fmt.Println("   Application Entry Point")
	fmt.Println("   Handles command line arguments and initializes the app.")
	fmt.Println()
	fmt.Println("   config/")
	fmt.Println("   Configuration files and settings")
	fmt.Println()
	fmt.Println("🏗️  Nested Structure:")
	fmt.Println("   • Each directory can have its own .info file")
	fmt.Println("   • Files describe only their directory's contents")
	fmt.Println("   • treex automatically merges all .info files")
	fmt.Println()
	fmt.Println("⚡ Auto-Generation:")
	fmt.Println("   Use 'treex import <tree-file>' to generate .info files")
	fmt.Println("   from a simple annotated tree structure.")
	fmt.Println()
	fmt.Println("   Example input:")
	fmt.Println("   myproject/cmd The CLI utilities")
	fmt.Println("   myproject/docs All documentation")
	fmt.Println()
	fmt.Println("📖 For complete documentation: see docs/INFO-FILES.md")
	fmt.Println()
}
