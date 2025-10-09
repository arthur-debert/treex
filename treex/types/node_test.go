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

func TestPluginDataStorage(t *testing.T) {
	node := &Node{
		Name: "test.go",
		Path: "/test.go",
	}

	// Test setting and getting plugin data
	testData := map[string]string{"key": "value"}
	node.SetPluginData("testplugin", testData)

	retrieved, exists := node.GetPluginData("testplugin")
	if !exists {
		t.Error("Expected plugin data to exist")
	}

	retrievedMap, ok := retrieved.(map[string]string)
	if !ok {
		t.Error("Expected retrieved data to be map[string]string")
	}

	if retrievedMap["key"] != "value" {
		t.Errorf("Expected 'value', got '%s'", retrievedMap["key"])
	}

	// Test getting non-existent plugin data
	_, exists = node.GetPluginData("nonexistent")
	if exists {
		t.Error("Expected non-existent plugin data to not exist")
	}
}

func TestPluginDataInitialization(t *testing.T) {
	node := &Node{
		Name: "test.go",
		Path: "/test.go",
	}

	// Data map should be initialized when first plugin data is set
	if node.Data != nil {
		t.Error("Expected Data map to be nil initially")
	}

	node.SetPluginData("plugin1", "data1")

	if node.Data == nil {
		t.Error("Expected Data map to be initialized after setting plugin data")
	}

	if len(node.Data) != 1 {
		t.Errorf("Expected Data map to have 1 entry, got %d", len(node.Data))
	}
}

func TestAnnotationBackwardCompatibility(t *testing.T) {
	node := &Node{
		Name: "test.go",
		Path: "/test.go",
	}

	// Test setting annotation via new method
	annotation := &Annotation{
		Path:  "/test.go",
		Notes: "Test annotation",
	}

	node.SetAnnotation(annotation)

	// Should be retrievable via new method
	retrieved := node.GetAnnotation()
	if retrieved == nil {
		t.Error("Expected annotation to be retrievable via GetAnnotation")
		return
	}

	if retrieved.Notes != "Test annotation" {
		t.Errorf("Expected 'Test annotation', got '%s'", retrieved.Notes)
	}

	// Should also be stored in plugin data
	data, exists := node.GetPluginData("info")
	if !exists {
		t.Error("Expected annotation to be stored in plugin data")
	}

	if data != annotation {
		t.Error("Expected plugin data to contain the same annotation")
	}
}

func TestAnnotationFallback(t *testing.T) {
	node := &Node{
		Name: "test.go",
		Path: "/test.go",
		Annotation: &Annotation{
			Path:  "/test.go",
			Notes: "Legacy annotation",
		},
	}

	// GetAnnotation should return legacy annotation when no plugin data exists
	retrieved := node.GetAnnotation()
	if retrieved == nil {
		t.Error("Expected annotation to be retrievable from legacy field")
		return
	}

	if retrieved.Notes != "Legacy annotation" {
		t.Errorf("Expected 'Legacy annotation', got '%s'", retrieved.Notes)
	}

	// Now set a new annotation via plugin data - it should take precedence
	newAnnotation := &Annotation{
		Path:  "/test.go",
		Notes: "New annotation",
	}
	node.SetAnnotation(newAnnotation)

	retrieved = node.GetAnnotation()
	if retrieved.Notes != "New annotation" {
		t.Errorf("Expected 'New annotation', got '%s'", retrieved.Notes)
	}
}
