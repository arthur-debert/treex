package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/internal/info"
	"github.com/adebert/treex/internal/tree"
	"github.com/adebert/treex/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	verbose    bool
	path       string
	noColor    bool
	minimal    bool
	ignoreFile string
	maxDepth   int
	safeMode   bool
	version    string // Holds the version string
)

// SetVersion allows the main package to set the version
func SetVersion(v string) {
	version = v
	rootCmd.Version = v // Set the version on the root command
}

var rootCmd = &cobra.Command{
	Use:   "treex [path]",
	Short: "treex is a CLI file viewer for annotated file trees",
	Long: `treex displays directory trees with annotations from .info files.
	
Annotations are read from .info files in directories and displayed
alongside the file tree structure, similar to the unix tree command
but with additional context and descriptions for files and directories.

.INFO FILES:

.info files are simple text files that describe the contents of directories.
Each directory can contain its own .info file to document files and subdirectories.

Basic format:
    filename
    Description of the file

    directory/
    Description of the directory

Example .info file:
    README.md
    Main project documentation

    src/main.go
    Application Entry Point
    Handles command line arguments and initializes the application.

    config/
    Configuration files and settings

NESTED .INFO FILES:

treex supports nested .info files - any directory can have its own .info file:
    project/.info          # Describes project/ contents
    project/src/.info      # Describes src/ contents  
    project/docs/.info     # Describes docs/ contents

Each .info file can only describe paths within its own directory for security.

GENERATING .INFO FILES:

Use 'treex gen-info <file>' to generate .info files from annotated tree structures.
The input can be simple:
    myproject/cmd The CLI utilities
    myproject/docs Documentation

Use 'treex info-files' for a quick reference guide.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine the target path
		targetPath := path
		if len(args) > 0 {
			targetPath = args[0]
		}
		if targetPath == "" {
			targetPath = "."
		}

		// Run the main treex logic
		if err := runTreex(targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate completion script for your shell",
	Long: `To load completions:

Bash:
  $ source <(treex completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ treex completion bash > /etc/bash_completion.d/treex
  # macOS:
  $ treex completion bash > $(brew --prefix)/etc/bash_completion.d/treex

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ treex completion zsh > "${fpath[1]}/_treex"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ treex completion fish | source

  # To load completions for each session, execute once:
  $ treex completion fish > ~/.config/fish/completions/treex.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		}
	},
}

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages for treex",
	Long:  `This command generates man pages for the treex CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		manPath, _ := cmd.Flags().GetString("path")
		if manPath == "" {
			manPath = "./"
		}
		
		header := &doc.GenManHeader{
			Title:   "TREEX",
			Section: "1",
			Source:  "Treex CLI " + version,
		}
		
		// Ensure the directory exists
		if err := os.MkdirAll(manPath, 0755); err != nil {
			log.Fatalf("Failed to create man page directory: %v", err)
		}
		
		err := doc.GenManTree(rootCmd, header, manPath)
		if err != nil {
			log.Fatalf("Failed to generate man pages: %v", err)
		}
		
		fmt.Printf("Man page generated successfully in %s\n", manPath)
	},
}

var genInfoCmd = &cobra.Command{
	Use:   "gen-info <file>",
	Short: "Generate .info files from annotated tree structure",
	Long: `Generate .info files from a hand-written annotated tree structure.

The input file should contain a tree-like structure with paths and descriptions.
Tree connectors are optional - you can use a simple format:

Simple format:
    myproject/cmd The go code for the cli utility
    myproject/docs All documentation  
    myproject/pkg The main parser code
    myproject/scripts Various utilities

Or with traditional tree connectors:
    myproject/
    ├── cmd/ The go code for the cli utility
    ├── docs/ All documentation
    │   └── dev/ Dev related, including technical topics
    ├── pkg/ The main parser code
    └── scripts/ Various utilities

Both formats work equally well. This will generate appropriate .info files 
in the corresponding directories.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		
		if err := runGenInfo(inputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Info files generated successfully")
	},
}

var infoFilesCmd = &cobra.Command{
	Use:   "info-files",
	Short: "Show information about .info file format and usage",
	Long:  `Display compact reference information about .info files and their format.`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfoFilesHelp()
	},
}

var addInfoCmd = &cobra.Command{
	Use:   "add-info <path> <description>",
	Short: "Add or update an entry in the current directory's .info file",
	Long: `Add or update an entry in the current directory's .info file.

This command will:
- Find the .info file in the current directory or create one if it doesn't exist
- Look for an existing entry for the specified path
- If an entry exists, prompt to replace, append, or abort (unless --replace is used)
- Add or update the entry with the provided description

Examples:
  treex add-info pkg "Main package containing core functionality"
  treex add-info config/ "Configuration files and settings"
  treex add-info --replace main.go "Application entry point"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		description := args[1]
		
		replace, _ := cmd.Flags().GetBool("replace")
		
		if err := runAddInfo(path, description, replace); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
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
	fmt.Println("   Use 'treex gen-info <tree-file>' to generate .info files")
	fmt.Println("   from a simple annotated tree structure.")
	fmt.Println()
	fmt.Println("   Example input:")
	fmt.Println("   myproject/cmd The CLI utilities")
	fmt.Println("   myproject/docs All documentation")
	fmt.Println()
	fmt.Println("📖 For complete documentation: see docs/INFO-FILES.md")
	fmt.Println()
}

func init() {
	// Add our flags
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVar(&minimal, "minimal", false, "Use minimal styling (fewer colors)")
	rootCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	rootCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")

	// Add flags for man command
	manCmd.Flags().String("path", "./", "Directory to generate man pages in")
	
	// Add flags for add-info command
	addInfoCmd.Flags().Bool("replace", false, "Replace existing entry without prompting")
	
	// Add subcommands
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(manCmd)
	rootCmd.AddCommand(genInfoCmd)
	rootCmd.AddCommand(infoFilesCmd)
	rootCmd.AddCommand(addInfoCmd)
}

// runGenInfo contains the main logic for the gen-info command
func runGenInfo(inputFile string) error {
	return info.GenerateInfoFromTree(inputFile)
}

// runAddInfo contains the main logic for the add-info command
func runAddInfo(path, description string, replace bool) error {
	// Use current directory
	currentDir := "."
	infoFilePath := filepath.Join(currentDir, ".info")
	
	// Parse existing .info file if it exists
	annotations, err := info.ParseDirectory(currentDir)
	if err != nil {
		return fmt.Errorf("failed to parse existing .info file: %w", err)
	}
	
	// Check if entry already exists
	existingAnnotation, exists := annotations[path]
	
	if exists && !replace {
		// Entry exists and we haven't been told to replace - ask user
		fmt.Printf("Entry for '%s' already exists:\n", path)
		fmt.Printf("Current description: %s\n\n", existingAnnotation.Description)
		fmt.Printf("New description: %s\n\n", description)
		
		fmt.Print("Choose action: (r)eplace, (a)ppend, or (q)uit [r/a/q]: ")
		
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		
		response = strings.ToLower(strings.TrimSpace(response))
		
		switch response {
		case "r", "replace":
			// Replace - do nothing, we'll overwrite below
		case "a", "append":
			// Append to existing description
			description = existingAnnotation.Description + "\n" + description
		case "q", "quit", "abort":
			fmt.Println("Operation cancelled.")
			return nil
		default:
			return fmt.Errorf("invalid choice: %s", response)
		}
	}
	
	// Update or add the annotation
	annotations[path] = &info.Annotation{
		Path:        path,
		Description: description,
	}
	
	// Write the updated .info file
	if err := writeInfoFile(infoFilePath, annotations); err != nil {
		return fmt.Errorf("failed to write .info file: %w", err)
	}
	
	if exists {
		fmt.Printf("Updated entry for '%s' in .info file\n", path)
	} else {
		fmt.Printf("Added entry for '%s' to .info file\n", path)
	}
	
	return nil
}

// writeInfoFile writes annotations to a .info file
func writeInfoFile(filePath string, annotations map[string]*info.Annotation) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create .info file: %w", err)
	}
	defer file.Close()
	
	// Write each annotation
	for path, annotation := range annotations {
		// Write the path
		if _, err := fmt.Fprintf(file, "%s\n", path); err != nil {
			return fmt.Errorf("failed to write path: %w", err)
		}
		
		// Write the description
		if _, err := fmt.Fprintf(file, "%s\n", annotation.Description); err != nil {
			return fmt.Errorf("failed to write description: %w", err)
		}
		
		// Add blank line between entries
		if _, err := fmt.Fprintf(file, "\n"); err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}
	}
	
	return nil
}

// runTreex contains the main logic for the treex command
func runTreex(targetPath string) error {
	if verbose {
		fmt.Printf("Analyzing directory: %s\n", targetPath)
		fmt.Println("Verbose mode enabled - will show parsed .info structure")
		fmt.Println()
	}

	// Phase 1 - Parse .info files (nested)
	annotations, err := info.ParseDirectoryTree(targetPath)
	if err != nil {
		return fmt.Errorf("failed to parse .info files: %w", err)
	}

	if verbose {
		fmt.Println("=== Parsed Annotations ===")
		if len(annotations) == 0 {
			fmt.Println("No annotations found (no .info file or empty file)")
		} else {
			for path, annotation := range annotations {
				fmt.Printf("Path: %s\n", path)
				if annotation.Title != "" {
					fmt.Printf("  Title: %s\n", annotation.Title)
				}
				fmt.Printf("  Description: %s\n", annotation.Description)
				fmt.Println()
			}
		}
		fmt.Println("=== End Annotations ===")
		fmt.Println()
	}

	// Phase 2 - Build file tree (using nested annotations with filtering options)
	var root *tree.Node
	if ignoreFile != "" || maxDepth != -1 {
		// Build tree with filtering options
		root, err = tree.BuildTreeNestedWithOptions(targetPath, ignoreFile, maxDepth)
		if err != nil {
			return fmt.Errorf("failed to build file tree with options: %w", err)
		}
	} else {
		// Build tree without filtering
		root, err = tree.BuildTreeNested(targetPath)
		if err != nil {
			return fmt.Errorf("failed to build file tree: %w", err)
		}
	}

	if verbose {
		fmt.Println("=== File Tree Structure ===")
		err = tree.WalkTree(root, func(node *tree.Node, depth int) error {
			indent := ""
			for i := 0; i < depth; i++ {
				indent += "  "
			}
			
			nodeType := "file"
			if node.IsDir {
				nodeType = "dir"
			}
			
			annotationInfo := ""
			if node.Annotation != nil {
				if node.Annotation.Title != "" {
					annotationInfo = fmt.Sprintf(" [%s]", node.Annotation.Title)
				} else {
					annotationInfo = " [annotated]"
				}
			}
			
			fmt.Printf("%s%s (%s)%s\n", indent, node.Name, nodeType, annotationInfo)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk tree: %w", err)
		}
		fmt.Println("=== End Tree Structure ===")
		fmt.Println()
	}

	// Phase 3 - Render tree with beautiful styling
	if verbose {
		fmt.Printf("treex analysis of: %s\n", targetPath)
		fmt.Printf("Found %d annotations\n", len(annotations))
		fmt.Println()
	}
	
	// Choose the appropriate renderer based on flags
	if noColor {
		// Use plain renderer without colors
		err = tui.RenderPlainTree(os.Stdout, root, true)
	} else if minimal {
		// Use minimal styling
		err = tui.RenderMinimalStyledTreeWithSafeMode(os.Stdout, root, true, safeMode)
	} else {
		// Use full beautiful styling
		err = tui.RenderStyledTreeWithSafeMode(os.Stdout, root, true, safeMode)
	}
	
	if err != nil {
		return fmt.Errorf("failed to render tree: %w", err)
	}
	
	return nil
}

