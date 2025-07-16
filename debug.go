package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Create temp file like the test does
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	
	testFile := filepath.Join(tempDir, "test.info")
	testContent := `Dad Chill, dad
Mom Listen to your mother
kids/ Children
kids/Sam Little Sam
kids/Alex The smart one`
	
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Test file created at: %s\n", testFile)
	
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "treex-test", "./cmd/treex")
	buildCmd.Dir = "/workspace"
	err = buildCmd.Run()
	if err != nil {
		panic(err)
	}
	defer os.Remove("/workspace/treex-test")
	
	// Run the draw command
	drawCmd := exec.Command("./treex-test", "draw", "--info-file", testFile, "--format", "no-color")
	drawCmd.Dir = "/workspace"
	output, err := drawCmd.CombinedOutput()
	
	fmt.Printf("Command: %v\n", drawCmd.Args)
	fmt.Printf("Exit code: %v\n", err)
	fmt.Printf("Output length: %d\n", len(output))
	fmt.Printf("Output: %q\n", string(output))
	
	if err != nil {
		fmt.Printf("Error details: %v\n", err)
	}
}
