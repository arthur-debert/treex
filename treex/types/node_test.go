package types

import "testing"

func TestNodeCreation(t *testing.T) {
	node := &Node{
		Name:  "test.go",
		Path:  "/path/to/test.go",
		IsDir: false,
	}

	if node.Name != "test.go" {
		t.Errorf("Expected name 'test.go', got '%s'", node.Name)
	}

	if node.IsDir {
		t.Error("Expected IsDir to be false")
	}
}
