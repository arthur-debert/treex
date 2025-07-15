package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/adebert/treex/pkg/app"
)

func TestInfoFileFlagBehavior(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-info-file-test-*")
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
	
	// Create actual files that will be annotated
	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# README"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "guide.md"), []byte("# Guide"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create .info files
	rootInfoContent := `README.md: Root readme from .info
docs/guide.md: Guide from root .info`
	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfoContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	docsInfoContent := `guide.md: Guide from docs .info`
	if err := os.WriteFile(filepath.Join(docsDir, ".info"), []byte(docsInfoContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create other.txt files
	rootOtherContent := `README.md: Root readme from other.txt
docs/guide.md: Guide from root other.txt`
	if err := os.WriteFile(filepath.Join(tempDir, "other.txt"), []byte(rootOtherContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	docsOtherContent := `guide.md: Guide from docs other.txt`
	if err := os.WriteFile(filepath.Join(docsDir, "other.txt"), []byte(docsOtherContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test 1: Default behavior (should use .info files)
	t.Run("default uses .info files", func(t *testing.T) {
		options := app.RenderOptions{
			InfoFileName: ".info",
			ViewMode:    "annotated",
		}
		
		result, err := app.RenderAnnotatedTree(tempDir, options)
		if err != nil {
			t.Fatalf("RenderAnnotatedTree failed: %v", err)
		}
		
		// Should see annotations from .info files
		if !strings.Contains(result.Output, "Root readme from .info") {
			t.Error("Expected to see annotation from root .info file")
		}
		if !strings.Contains(result.Output, "Guide from docs .info") {
			t.Error("Expected to see annotation from docs .info file (precedence)")
		}
		
		// Should NOT see annotations from other.txt files
		if strings.Contains(result.Output, "from other.txt") {
			t.Error("Should not see annotations from other.txt files when using default")
		}
	})
	
	// Test 2: With --info-file other.txt (should ONLY use other.txt files)
	t.Run("--info-file other.txt uses only other.txt files", func(t *testing.T) {
		options := app.RenderOptions{
			InfoFileName: "other.txt",
			ViewMode:    "annotated",
		}
		
		result, err := app.RenderAnnotatedTree(tempDir, options)
		if err != nil {
			t.Fatalf("RenderAnnotatedTree failed: %v", err)
		}
		
		// Should see annotations from other.txt files
		if !strings.Contains(result.Output, "Root readme from other.txt") {
			t.Error("Expected to see annotation from root other.txt file")
		}
		if !strings.Contains(result.Output, "Guide from docs other.txt") {
			t.Error("Expected to see annotation from docs other.txt file (precedence)")
		}
		
		// Should NOT see annotations from .info files
		if strings.Contains(result.Output, "from .info") {
			t.Error("Should not see annotations from .info files when using --info-file other.txt")
		}
		
		// Verify we have the expected number of annotations
		if result.Stats.AnnotationsFound != 2 {
			t.Errorf("Expected 2 annotations from other.txt files, got %d", result.Stats.AnnotationsFound)
		}
	})
	
	// Test 3: With a non-existent info file name
	t.Run("--info-file nonexistent.txt finds no annotations", func(t *testing.T) {
		options := app.RenderOptions{
			InfoFileName: "nonexistent.txt",
			ViewMode:    "annotated",
		}
		
		result, err := app.RenderAnnotatedTree(tempDir, options)
		if err != nil {
			t.Fatalf("RenderAnnotatedTree failed: %v", err)
		}
		
		// Should find no annotations
		if result.Stats.AnnotationsFound != 0 {
			t.Errorf("Expected 0 annotations, got %d", result.Stats.AnnotationsFound)
		}
		
		// Should not see any annotations
		if strings.Contains(result.Output, "from .info") || strings.Contains(result.Output, "from other.txt") {
			t.Error("Should not see any annotations with non-existent info file name")
		}
	})
}

// Test that verifies the specific bug scenario mentioned
func TestInfoFileFlagDoesNotMixFiles(t *testing.T) {
	// This test specifically checks the bug described where using --info-file
	// might be picking up both the custom file AND .info files
	
	tempDir, err := os.MkdirTemp("", "treex-nomix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	
	// Create a file to annotate
	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create both .info and custom.txt with different annotations
	infoContent := `test.txt: This is from .info file`
	customContent := `test.txt: This is from custom.txt file`
	
	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "custom.txt"), []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Use custom.txt as info file
	options := app.RenderOptions{
		InfoFileName: "custom.txt",
		ViewMode:    "annotated",
	}
	
	result, err := app.RenderAnnotatedTree(tempDir, options)
	if err != nil {
		t.Fatalf("RenderAnnotatedTree failed: %v", err)
	}
	
	// Should ONLY see annotation from custom.txt
	if !strings.Contains(result.Output, "This is from custom.txt file") {
		t.Error("Expected to see annotation from custom.txt")
	}
	
	// Should NOT see annotation from .info
	if strings.Contains(result.Output, "This is from .info file") {
		t.Error("Bug: Should not see annotation from .info file when using --info-file custom.txt")
	}
	
	// Should have exactly 1 annotation
	if result.Stats.AnnotationsFound != 1 {
		t.Errorf("Expected exactly 1 annotation, got %d", result.Stats.AnnotationsFound)
	}
}