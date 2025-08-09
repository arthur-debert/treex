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

// AppliesTo returns true for both files and directories
func (p *SizePlugin) AppliesTo(node *types.Node) bool {
	return true
}

// Collect gathers file size information
func (p *SizePlugin) Collect(node *types.Node) (map[string]interface{}, error) {
	// For directories, the size will be aggregated later
	if node.IsDir {
		return make(map[string]interface{}), nil
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
	// Check if we have pre-formatted human readable size
	if humanReadable, exists := metadata["human_readable"]; exists {
		if humanReadableStr, ok := humanReadable.(string); ok {
			return humanReadableStr
		}
	}
	
	// Otherwise, format from bytes
	if bytes, exists := metadata["bytes"]; exists {
		if bytesInt64, ok := bytes.(int64); ok {
			return humanize.Bytes(uint64(bytesInt64))
		}
	}
	
	return ""
}