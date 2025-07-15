package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestInfoFileBehavior tests the current behavior of --info-file flag
func TestInfoFileBehavior(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-infofile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create directory structure
	// root/
	//   .info
	//   other.txt
	//   docs/
	//     .info
	//     other.txt
	
	// Create docs directory
	docsDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create root .info file
	rootInfoContent := `docs/ Documentation directory from .info
README.md Main readme from .info`
	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create root other.txt file
	rootOtherContent := `docs/ Documentation directory from other.txt
CHANGELOG.md Change log from other.txt`
	if err := os.WriteFile(filepath.Join(tempDir, "other.txt"), []byte(rootOtherContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create docs/.info file
	docsInfoContent := `api.md API docs from .info
guide.md User guide from .info`
	if err := os.WriteFile(filepath.Join(docsDir, ".info"), []byte(docsInfoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create docs/other.txt file
	docsOtherContent := `tutorial.md Tutorial from other.txt
faq.md FAQ from other.txt`
	if err := os.WriteFile(filepath.Join(docsDir, "other.txt"), []byte(docsOtherContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Default behavior (should use .info files)
	t.Run("DefaultBehaviorUsesInfoFiles", func(t *testing.T) {
		// Reset global flags
		resetGlobalFlags()
		
		// Execute show command
		output, err := executeShowCommandForBehaviorTest(tempDir, "--format=no-color")
		if err != nil {
			t.Fatalf("Show command failed: %v", err)
		}

		// Should see content from .info files
		if !strings.Contains(output, "Main readme from .info") {
			t.Errorf("Expected to see content from root .info file, got:\n%s", output)
		}
		if !strings.Contains(output, "Documentation directory from .info") {
			t.Errorf("Expected to see docs annotation from .info file, got:\n%s", output)
		}
		
		// Should NOT see content from other.txt files
		if strings.Contains(output, "from other.txt") {
			t.Errorf("Should not see content from other.txt files, got:\n%s", output)
		}
	})

	// Test 2: With --info-file other.txt (should use other.txt files)
	t.Run("InfoFileFlagUsesCustomFiles", func(t *testing.T) {
		// Reset global flags
		resetGlobalFlags()
		
		// Execute show command with --info-file
		output, err := executeShowCommandForBehaviorTest(tempDir, "--format=no-color", "--info-file", "other.txt")
		if err != nil {
			t.Fatalf("Show command failed: %v", err)
		}

		// Should see content from other.txt files
		if !strings.Contains(output, "Change log from other.txt") {
			t.Errorf("Expected to see content from root other.txt file, got:\n%s", output)
		}
		if !strings.Contains(output, "Documentation directory from other.txt") {
			t.Errorf("Expected to see docs annotation from other.txt file, got:\n%s", output)
		}
		
		// Should NOT see content from .info files
		if strings.Contains(output, "from .info") {
			t.Errorf("Should not see content from .info files when using --info-file other.txt, got:\n%s", output)
		}
	})

	// Test 3: Verify that nested files are properly handled
	t.Run("NestedFilesWithCustomInfoFile", func(t *testing.T) {
		// Reset global flags
		resetGlobalFlags()
		
		// Change to docs directory and run from there
		oldDir, _ := os.Getwd()
		if err := os.Chdir(docsDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(oldDir) }()

		// Execute show command with --info-file from docs directory
		output, err := executeShowCommandForBehaviorTest(".", "--format=no-color", "--info-file", "other.txt")
		if err != nil {
			t.Fatalf("Show command failed: %v", err)
		}

		// Should see content from docs/other.txt
		if !strings.Contains(output, "Tutorial from other.txt") || !strings.Contains(output, "FAQ from other.txt") {
			t.Errorf("Expected to see content from docs/other.txt file, got:\n%s", output)
		}
		
		// Should NOT see content from .info files
		if strings.Contains(output, "from .info") {
			t.Errorf("Should not see content from .info files when using --info-file other.txt, got:\n%s", output)
		}
	})
}

// executeShowCommandForBehaviorTest helper function for testing show command
func executeShowCommandForBehaviorTest(args ...string) (string, error) {
	// Create a fresh show command for testing
	testShowCmd := setupShowCmd()
	
	// Create a root command and add show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(testShowCmd)
	
	// Capture output
	output := &bytes.Buffer{}
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	
	// Set arguments (prepend "show" to the args)
	testRootCmd.SetArgs(append([]string{"show"}, args...))
	
	// Execute command
	err := testRootCmd.Execute()
	return output.String(), err
}