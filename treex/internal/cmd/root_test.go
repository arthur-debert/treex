package cmd

import (
	"testing"

	"github.com/jwaldrip/treex/treex"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestBuildTreeConfigFromFlags tests that command-line flags are correctly mapped to TreeConfig
func TestBuildTreeConfigFromFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected treex.TreeConfig
	}{
		{
			name: "default configuration",
			args: []string{},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{},
			},
		},
		{
			name: "depth flag set",
			args: []string{"-d", "3"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     3,
				ExcludeGlobs: []string{},
			},
		},
		{
			name: "depth flag long form",
			args: []string{"--depth", "5"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     5,
				ExcludeGlobs: []string{},
			},
		},
		{
			name: "depth zero (no limit)",
			args: []string{"-d", "0"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{},
			},
		},
		{
			name: "single exclude pattern",
			args: []string{"-e", "*.tmp"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{"*.tmp"},
			},
		},
		{
			name: "multiple exclude patterns short form",
			args: []string{"-e", "*.tmp", "-e", "node_modules"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{"*.tmp", "node_modules"},
			},
		},
		{
			name: "exclude pattern long form",
			args: []string{"--exclude", "*.log"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{"*.log"},
			},
		},
		{
			name: "depth and exclude combined",
			args: []string{"-d", "2", "-e", "*.tmp", "-e", ".git"},
			expected: treex.TreeConfig{
				Root:         "/test/path",
				Filesystem:   nil,
				MaxDepth:     2,
				ExcludeGlobs: []string{"*.tmp", ".git"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags before each test
			maxDepth = 0
			excludeGlobs = []string{}

			// Create a test command to parse flags
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {},
			}
			testCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "Maximum depth")
			testCmd.Flags().StringSliceVarP(&excludeGlobs, "exclude", "e", []string{}, "Exclude patterns")

			// Parse the test arguments
			testCmd.SetArgs(tt.args)
			err := testCmd.Execute()
			assert.NoError(t, err)

			// Build config from parsed flags
			config := buildTreeConfig("/test/path")

			// Assert the configuration matches expected
			assert.Equal(t, tt.expected.Root, config.Root)
			assert.Equal(t, tt.expected.MaxDepth, config.MaxDepth)
			assert.Equal(t, tt.expected.ExcludeGlobs, config.ExcludeGlobs)
			assert.Equal(t, tt.expected.Filesystem, config.Filesystem)
		})
	}
}

// TestRootPathArgument tests that the root path argument is correctly handled
func TestRootPathArgument(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedRoot string
	}{
		{
			name:         "no path argument defaults to current directory",
			args:         []string{},
			expectedRoot: ".",
		},
		{
			name:         "explicit path argument",
			args:         []string{"/custom/path"},
			expectedRoot: "/custom/path",
		},
		{
			name:         "relative path argument",
			args:         []string{"../relative"},
			expectedRoot: "../relative",
		},
		{
			name:         "path with depth flag",
			args:         []string{"/some/path", "-d", "2"},
			expectedRoot: "/some/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			maxDepth = 0

			// Extract path from args (simulate runTreeCommand logic)
			var rootPath string
			var flagArgs []string

			// Separate path argument from flags
			if len(tt.args) > 0 && tt.args[0][0] != '-' {
				rootPath = tt.args[0]
				flagArgs = tt.args[1:]
			} else {
				rootPath = "."
				flagArgs = tt.args
			}

			// Parse flags if any
			if len(flagArgs) > 0 {
				testCmd := &cobra.Command{
					Use: "test",
					Run: func(cmd *cobra.Command, args []string) {},
				}
				testCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "Maximum depth")
				testCmd.SetArgs(flagArgs)
				err := testCmd.Execute()
				assert.NoError(t, err)
			}

			// Assert root path is correctly extracted
			assert.Equal(t, tt.expectedRoot, rootPath)
		})
	}
}

// TestFlagValidation tests validation of flag combinations and values
func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid depth",
			args:        []string{"-d", "3"},
			expectError: false,
		},
		{
			name:        "zero depth is valid",
			args:        []string{"-d", "0"},
			expectError: false,
		},
		{
			name:        "negative depth should be handled by cobra",
			args:        []string{"-d", "-1"},
			expectError: false, // cobra will accept it, validation would happen at runtime
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			maxDepth = 0

			// Create test command
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {},
			}
			testCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "Maximum depth")

			// Parse arguments
			testCmd.SetArgs(tt.args)
			err := testCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCommandLineToAPIMapping tests the complete mapping from command line to API call
func TestCommandLineToAPIMapping(t *testing.T) {
	tests := []struct {
		name           string
		cmdLine        []string
		expectedConfig treex.TreeConfig
	}{
		{
			name:    "treex",
			cmdLine: []string{},
			expectedConfig: treex.TreeConfig{
				Root:         ".", // Would be converted to absolute path in real execution
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{},
			},
		},
		{
			name:    "treex -d 2",
			cmdLine: []string{"-d", "2"},
			expectedConfig: treex.TreeConfig{
				Root:         ".",
				Filesystem:   nil,
				MaxDepth:     2,
				ExcludeGlobs: []string{},
			},
		},
		{
			name:    "treex /usr/local -d 1",
			cmdLine: []string{"/usr/local", "-d", "1"},
			expectedConfig: treex.TreeConfig{
				Root:         "/usr/local",
				Filesystem:   nil,
				MaxDepth:     1,
				ExcludeGlobs: []string{},
			},
		},
		{
			name:    "treex -e *.tmp -e node_modules",
			cmdLine: []string{"-e", "*.tmp", "-e", "node_modules"},
			expectedConfig: treex.TreeConfig{
				Root:         ".",
				Filesystem:   nil,
				MaxDepth:     0,
				ExcludeGlobs: []string{"*.tmp", "node_modules"},
			},
		},
		{
			name:    "treex /home/user -d 3 -e *.log -e .git",
			cmdLine: []string{"/home/user", "-d", "3", "-e", "*.log", "-e", ".git"},
			expectedConfig: treex.TreeConfig{
				Root:         "/home/user",
				Filesystem:   nil,
				MaxDepth:     3,
				ExcludeGlobs: []string{"*.log", ".git"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			maxDepth = 0
			excludeGlobs = []string{}

			// Parse command line arguments
			var rootPath string
			var flagArgs []string

			// Extract path and flags (simulate cobra's argument parsing)
			if len(tt.cmdLine) > 0 && tt.cmdLine[0][0] != '-' {
				rootPath = tt.cmdLine[0]
				flagArgs = tt.cmdLine[1:]
			} else {
				rootPath = "."
				flagArgs = tt.cmdLine
			}

			// Parse flags
			if len(flagArgs) > 0 {
				testCmd := &cobra.Command{
					Use: "test",
					Run: func(cmd *cobra.Command, args []string) {},
				}
				testCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "Maximum depth")
				testCmd.Flags().StringSliceVarP(&excludeGlobs, "exclude", "e", []string{}, "Exclude patterns")
				testCmd.SetArgs(flagArgs)
				err := testCmd.Execute()
				assert.NoError(t, err)
			}

			// Build configuration (this is what would be passed to treex.BuildTree)
			config := buildTreeConfig(rootPath)

			// Assert the mapping is correct
			assert.Equal(t, tt.expectedConfig.Root, config.Root)
			assert.Equal(t, tt.expectedConfig.MaxDepth, config.MaxDepth)
			assert.Equal(t, tt.expectedConfig.ExcludeGlobs, config.ExcludeGlobs)
			assert.Equal(t, tt.expectedConfig.Filesystem, config.Filesystem)
		})
	}
}
