package builtin

import (
	"os"

	"github.com/dustin/go-humanize"
	"github.com/adebert/treex/pkg/core/types"
)

// SizePlugin collects file size information
type SizePlugin struct{}

// Name returns the plugin identifier
func (p *SizePlugin) Name() string {
	return "size"
}

// Description returns the plugin description
func (p *SizePlugin) Description() string {
	return "Display file sizes in human-readable format"
}

// AppliesTo returns true only for files (not directories)
func (p *SizePlugin) AppliesTo(node *types.Node) bool {
	return !node.IsDir
}

// Collect gathers file size information
func (p *SizePlugin) Collect(node *types.Node) (map[string]interface{}, error) {
	// Skip directories
	if node.IsDir {
		return nil, nil
	}
	
	// Get file info
	fileInfo, err := os.Stat(node.Path)
	if err != nil {
		// If we can't stat the file, return empty metadata rather than error
		// This allows the plugin to gracefully handle files that may not exist
		return make(map[string]interface{}), nil
	}
	
	return map[string]interface{}{
		"bytes":       fileInfo.Size(),
		"human_readable": humanize.Bytes(uint64(fileInfo.Size())),
	}, nil
}

// Format returns a formatted string representation of the file size
func (p *SizePlugin) Format(metadata map[string]interface{}) string {
	humanReadable, exists := metadata["human_readable"]
	if !exists {
		return ""
	}
	
	if humanReadableStr, ok := humanReadable.(string); ok {
		return humanReadableStr
	}
	
	return ""
}