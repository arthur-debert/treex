package commands

import (
	"fmt"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/display/rendering"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var themeDemoCmd = &cobra.Command{
	Use:    "theme-demo",
	Short:  "Demo dark and light themes",
	Long:   "Shows a sample tree with both dark and light themes for comparison",
	RunE:   runThemeDemo,
	Hidden: true, // Hidden command for testing
}

func init() {
	rootCmd.AddCommand(themeDemoCmd)
}

func runThemeDemo(cmd *cobra.Command, args []string) error {
	// Create a sample tree
	root := &types.Node{
		Name:  "project",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "src",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "main.go",
						IsDir: false,
						Annotation: &types.Annotation{
							Path:  "main.go",
							Notes: "Application entry point\nInitializes and starts the server",
						},
					},
					{
						Name:  "config.go",
						IsDir: false,
						Annotation: &types.Annotation{
							Path:  "config.go",
							Notes: "Configuration",
						},
					},
				},
			},
			{
				Name:  "tests",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "main_test.go",
						IsDir: false,
					},
				},
			},
			{
				Name:  "README.md",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "README.md",
					Notes: "Project documentation and setup instructions",
				},
			},
		},
	}

	fmt.Println("=== DARK THEME ===")
	lipgloss.SetHasDarkBackground(true)
	if err := rendering.RenderStyledTree(cmd.OutOrStdout(), root, true); err != nil {
		return err
	}

	fmt.Println("\n=== LIGHT THEME ===")
	lipgloss.SetHasDarkBackground(false)
	if err := rendering.RenderStyledTree(cmd.OutOrStdout(), root, true); err != nil {
		return err
	}

	fmt.Println("\n=== MINIMAL COLORS ===")
	renderer := rendering.NewMinimalStyledTreeRenderer(cmd.OutOrStdout(), true)
	if err := renderer.Render(root); err != nil {
		return err
	}

	fmt.Println("\n=== NO COLORS ===")
	noColorRenderer := rendering.NewNoColorStyledTreeRenderer(cmd.OutOrStdout(), true)
	if err := noColorRenderer.Render(root); err != nil {
		return err
	}

	// Reset to dark theme
	lipgloss.SetHasDarkBackground(true)

	return nil
}
