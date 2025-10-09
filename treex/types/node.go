package types

// Node represents a file or directory in the tree
type Node struct {
	Name       string                 // Just the filename/dirname, e.g., "main.go"
	Path       string                 // The unique, relative path from the tree root, e.g., "src/main.go"
	IsDir      bool                   // Whether this is a directory
	Size       int64                  // File size in bytes (0 for directories)
	Annotation *Annotation            // Associated annotation if any (DEPRECATED: use Data["info"])
	Children   []*Node                // Child nodes (for directories)
	Parent     *Node                  // Parent node (nil for root)
	Data       map[string]interface{} // Plugin-specific data storage
}

// SetPluginData stores data for a specific plugin namespace
func (n *Node) SetPluginData(pluginName string, data interface{}) {
	if n.Data == nil {
		n.Data = make(map[string]interface{})
	}
	n.Data[pluginName] = data
}

// GetPluginData retrieves data for a specific plugin namespace
func (n *Node) GetPluginData(pluginName string) (interface{}, bool) {
	if n.Data == nil {
		return nil, false
	}
	data, exists := n.Data[pluginName]
	return data, exists
}

// GetAnnotation returns the annotation for this node, checking both the legacy
// Annotation field and the new Data["info"] storage for backward compatibility
func (n *Node) GetAnnotation() *Annotation {
	// Check new storage first
	if data, exists := n.GetPluginData("info"); exists {
		if annotation, ok := data.(*Annotation); ok {
			return annotation
		}
	}
	// Fall back to legacy field
	return n.Annotation
}

// SetAnnotation stores an annotation using the new Data storage system
// This is the preferred way to set annotations going forward
func (n *Node) SetAnnotation(annotation *Annotation) {
	n.SetPluginData("info", annotation)
}
