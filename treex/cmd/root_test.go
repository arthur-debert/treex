package cmd

import (
	"testing"

	"github.com/jwaldrip/treex/treex"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// defaultExpectedConfig returns a base TreeConfig with default values for testing
func defaultExpectedConfig() treex.TreeConfig {
	return treex.TreeConfig{
		Root:            "/test/path",
		Filesystem:      nil,
		MaxDepth:        0,
		BuiltinIgnores:  true, // Built-in ignores enabled by default
		ExcludeGlobs:    []string{},
		IncludeHidden:   true,
		DirectoriesOnly: false,
		PluginFilters:   make(map[string]map[string]bool), // Empty plugin filters by default
	}
}

// TestBuildTreeConfigFromFlags tests that command-line flags are correctly mapped to TreeConfig
func TestBuildTreeConfigFromFlags(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		modify func(*treex.TreeConfig) // Function to modify the expected config
	}{
		{
			name:   "default configuration",
			args:   []string{},
			modify: nil, // No modifications needed - use defaults
		},
		{
			name: "level flag set",
			args: []string{"-l", "3"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.MaxDepth = 3
			},
		},
		{
			name: "level flag long form",
			args: []string{"--level", "5"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.MaxDepth = 5
			},
		},
		{
			name:   "level zero (no limit)",
			args:   []string{"-l", "0"},
			modify: nil, // Zero is already the default
		},
		{
			name: "single exclude pattern",
			args: []string{"-e", "*.tmp"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.ExcludeGlobs = []string{"*.tmp"}
			},
		},
		{
			name: "multiple exclude patterns short form",
			args: []string{"-e", "*.tmp", "-e", "node_modules"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.ExcludeGlobs = []string{"*.tmp", "node_modules"}
			},
		},
		{
			name: "exclude pattern long form",
			args: []string{"--exclude", "*.log"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.ExcludeGlobs = []string{"*.log"}
			},
		},
		{
			name: "level and exclude combined",
			args: []string{"-l", "2", "-e", "*.tmp", "-e", ".git"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.MaxDepth = 2
				cfg.ExcludeGlobs = []string{"*.tmp", ".git"}
			},
		},
		{
			name: "hidden files disabled",
			args: []string{"-h=false"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.IncludeHidden = false
			},
		},
		{
			name:   "hidden files enabled explicitly",
			args:   []string{"--hidden=true"},
			modify: nil, // True is already the default
		},
		{
			name: "directories only",
			args: []string{"-d"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.DirectoriesOnly = true
			},
		},
		{
			name: "directories only long form",
			args: []string{"--directory"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.DirectoriesOnly = true
			},
		},
		{
			name: "hidden files with level and exclude",
			args: []string{"-l", "3", "-e", "*.tmp", "-h=false"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.MaxDepth = 3
				cfg.ExcludeGlobs = []string{"*.tmp"}
				cfg.IncludeHidden = false
			},
		},
		{
			name: "no builtin ignores flag",
			args: []string{"--no-builtin-ignores"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.BuiltinIgnores = false
			},
		},
		{
			name: "no builtin ignores with other options",
			args: []string{"--no-builtin-ignores", "-l", "2", "-e", "*.tmp"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.BuiltinIgnores = false
				cfg.MaxDepth = 2
				cfg.ExcludeGlobs = []string{"*.tmp"}
			},
		},
		{
			name: "all options combined",
			args: []string{"-l", "2", "-e", "*.tmp", "-e", ".git", "-h=false", "-d"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.MaxDepth = 2
				cfg.ExcludeGlobs = []string{"*.tmp", ".git"}
				cfg.IncludeHidden = false
				cfg.DirectoriesOnly = true
			},
		},
		{
			name: "all options including no builtin ignores",
			args: []string{"--no-builtin-ignores", "-l", "1", "-e", "*.test", "-h=false", "-d"},
			modify: func(cfg *treex.TreeConfig) {
				cfg.BuiltinIgnores = false
				cfg.MaxDepth = 1
				cfg.ExcludeGlobs = []string{"*.test"}
				cfg.IncludeHidden = false
				cfg.DirectoriesOnly = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags before each test
			maxLevel = 0
			noBuiltinIgnores = false
			excludeGlobs = []string{}
			includeHidden = true
			directoriesOnly = false

			// Create a test command to parse flags
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {},
			}
			// Add our flags first, then remove help shorthand
			testCmd.Flags().IntVarP(&maxLevel, "level", "l", 0, "Maximum level")
			testCmd.Flags().BoolVar(&noBuiltinIgnores, "no-builtin-ignores", false, "Disable built-in ignores")
			testCmd.Flags().StringSliceVarP(&excludeGlobs, "exclude", "e", []string{}, "Exclude patterns")
			testCmd.Flags().BoolVarP(&includeHidden, "hidden", "h", true, "Include hidden files")
			testCmd.Flags().BoolVarP(&directoriesOnly, "directory", "d", false, "Show directories only")

			// Override help flag without shorthand to avoid conflict
			testCmd.Flags().Bool("help", false, "help for test")
			testCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {})

			// Parse the test arguments
			testCmd.SetArgs(tt.args)
			err := testCmd.Execute()
			assert.NoError(t, err)

			// Build config from parsed flags
			config := buildTreeConfig("/test/path")

			// Generate expected config using fixture
			expected := defaultExpectedConfig()
			if tt.modify != nil {
				tt.modify(&expected)
			}

			// Assert the configuration matches expected
			assert.Equal(t, expected, config)
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
			name:         "path with level flag",
			args:         []string{"/some/path", "-l", "2"},
			expectedRoot: "/some/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			maxLevel = 0
			includeHidden = true
			directoriesOnly = false

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
				testCmd.Flags().IntVarP(&maxLevel, "level", "l", 0, "Maximum level")
				testCmd.Flags().Bool("help", false, "help for test")
				testCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {})
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
			name:        "valid level",
			args:        []string{"-l", "3"},
			expectError: false,
		},
		{
			name:        "zero level is valid",
			args:        []string{"-l", "0"},
			expectError: false,
		},
		{
			name:        "negative level should be handled by cobra",
			args:        []string{"-l", "-1"},
			expectError: false, // cobra will accept it, validation would happen at runtime
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			maxLevel = 0
			includeHidden = true
			directoriesOnly = false

			// Create test command
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {},
			}
			testCmd.Flags().IntVarP(&maxLevel, "level", "l", 0, "Maximum level")
			testCmd.Flags().Bool("help", false, "help for test")
			testCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {})

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
				Root:            ".", // Would be converted to absolute path in real execution
				Filesystem:      nil,
				MaxDepth:        0,
				ExcludeGlobs:    []string{},
				IncludeHidden:   true,
				DirectoriesOnly: false,
				PluginFilters:   make(map[string]map[string]bool),
			},
		},
		{
			name:    "treex -l 2",
			cmdLine: []string{"-l", "2"},
			expectedConfig: treex.TreeConfig{
				Root:            ".",
				Filesystem:      nil,
				MaxDepth:        2,
				ExcludeGlobs:    []string{},
				IncludeHidden:   true,
				DirectoriesOnly: false,
				PluginFilters:   make(map[string]map[string]bool),
			},
		},
		{
			name:    "treex /usr/local -l 1",
			cmdLine: []string{"/usr/local", "-l", "1"},
			expectedConfig: treex.TreeConfig{
				Root:            "/usr/local",
				Filesystem:      nil,
				MaxDepth:        1,
				ExcludeGlobs:    []string{},
				IncludeHidden:   true,
				DirectoriesOnly: false,
				PluginFilters:   make(map[string]map[string]bool),
			},
		},
		{
			name:    "treex -e *.tmp -e node_modules",
			cmdLine: []string{"-e", "*.tmp", "-e", "node_modules"},
			expectedConfig: treex.TreeConfig{
				Root:            ".",
				Filesystem:      nil,
				MaxDepth:        0,
				ExcludeGlobs:    []string{"*.tmp", "node_modules"},
				IncludeHidden:   true,
				DirectoriesOnly: false,
				PluginFilters:   make(map[string]map[string]bool),
			},
		},
		{
			name:    "treex /home/user -l 3 -e *.log -e .git",
			cmdLine: []string{"/home/user", "-l", "3", "-e", "*.log", "-e", ".git"},
			expectedConfig: treex.TreeConfig{
				Root:            "/home/user",
				Filesystem:      nil,
				MaxDepth:        3,
				ExcludeGlobs:    []string{"*.log", ".git"},
				IncludeHidden:   true,
				DirectoriesOnly: false,
				PluginFilters:   make(map[string]map[string]bool),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			maxLevel = 0
			includeHidden = true
			directoriesOnly = false
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
				testCmd.Flags().IntVarP(&maxLevel, "level", "l", 0, "Maximum level")
				testCmd.Flags().StringSliceVarP(&excludeGlobs, "exclude", "e", []string{}, "Exclude patterns")
				testCmd.Flags().BoolVarP(&includeHidden, "hidden", "h", true, "Include hidden files")
				testCmd.Flags().Bool("help", false, "help for test")
				testCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {})
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
			assert.Equal(t, tt.expectedConfig.IncludeHidden, config.IncludeHidden)
			assert.Equal(t, tt.expectedConfig.DirectoriesOnly, config.DirectoriesOnly)
			assert.Equal(t, tt.expectedConfig.Filesystem, config.Filesystem)
		})
	}
}
