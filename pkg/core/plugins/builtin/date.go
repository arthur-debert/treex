package builtin

import (
	"os"

	"github.com/dustin/go-humanize"
	"github.com/adebert/treex/pkg/core/types"
)

// DateCreatedPlugin collects file creation date information
type DateCreatedPlugin struct{}

// Name returns the plugin identifier
func (p *DateCreatedPlugin) Name() string {
	return "date-created"
}

// Description returns the plugin description
func (p *DateCreatedPlugin) Description() string {
	return "Display file creation dates in human-readable format"
}

// AppliesTo returns true only for files (not directories)
func (p *DateCreatedPlugin) AppliesTo(node *types.Node) bool {
	return !node.IsDir
}

// Collect gathers file creation date information
func (p *DateCreatedPlugin) Collect(node *types.Node) (map[string]interface{}, error) {
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
	
	// Use ModTime as creation time (most reliable cross-platform approach)
	// Note: On some systems, this might be modification time rather than creation time
	modTime := fileInfo.ModTime()
	
	return map[string]interface{}{
		"timestamp":      modTime.Unix(),
		"time":           modTime,
		"human_readable": humanize.Time(modTime),
	}, nil
}

// Format returns a formatted string representation of the creation date
func (p *DateCreatedPlugin) Format(metadata map[string]interface{}) string {
	humanReadable, exists := metadata["human_readable"]
	if !exists {
		return ""
	}
	
	if humanReadableStr, ok := humanReadable.(string); ok {
		return humanReadableStr
	}
	
	return ""
}

// DateModifiedPlugin collects file modification date information  
type DateModifiedPlugin struct{}

// Name returns the plugin identifier
func (p *DateModifiedPlugin) Name() string {
	return "date-modified"
}

// Description returns the plugin description
func (p *DateModifiedPlugin) Description() string {
	return "Display file modification dates in human-readable format"
}

// AppliesTo returns true only for files (not directories)
func (p *DateModifiedPlugin) AppliesTo(node *types.Node) bool {
	return !node.IsDir
}

// Collect gathers file modification date information
func (p *DateModifiedPlugin) Collect(node *types.Node) (map[string]interface{}, error) {
	// Skip directories
	if node.IsDir {
		return nil, nil
	}
	
	// Get file info
	fileInfo, err := os.Stat(node.Path)
	if err != nil {
		// If we can't stat the file, return empty metadata rather than error
		return make(map[string]interface{}), nil
	}
	
	modTime := fileInfo.ModTime()
	
	return map[string]interface{}{
		"timestamp":      modTime.Unix(),
		"time":           modTime,
		"human_readable": humanize.Time(modTime),
	}, nil
}

// Format returns a formatted string representation of the modification date
func (p *DateModifiedPlugin) Format(metadata map[string]interface{}) string {
	humanReadable, exists := metadata["human_readable"]
	if !exists {
		return ""
	}
	
	if humanReadableStr, ok := humanReadable.(string); ok {
		return humanReadableStr
	}
	
	return ""
}