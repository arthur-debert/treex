package builtin

import (
	"bufio"
	"os"
	"strconv"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/core/utils"
)

// LineCountPlugin collects line count information for text files
type LineCountPlugin struct{}

// Name returns the plugin identifier
func (p *LineCountPlugin) Name() string {
	return "lc"
}

// Description returns the plugin description
func (p *LineCountPlugin) Description() string {
	return "Display line counts for text files"
}

// AppliesTo returns true for directories and text files
func (p *LineCountPlugin) AppliesTo(node *types.Node) bool {
	// Always return true - we'll check file type in Collect
	// This allows us to attempt line counting on any file
	return true
}

// Collect gathers line count information for the node
func (p *LineCountPlugin) Collect(node *types.Node) (map[string]interface{}, error) {
	if node.IsDir {
		// For directories, return empty metadata
		// The aggregation plugin will sum up line counts from children
		return make(map[string]interface{}), nil
	}

	// Check if this is a text file using content-based detection
	isText, err := utils.IsTextFile(node.Path)
	if err != nil {
		// If we can't read the file, return empty metadata
		return make(map[string]interface{}), nil
	}

	if !isText {
		// Binary file, don't count lines
		return nil, nil
	}

	// Count lines in the file
	lineCount, err := p.countLines(node.Path)
	if err != nil {
		// If we can't count lines, return empty metadata rather than error
		// This allows the plugin to gracefully handle files that may not be accessible
		return make(map[string]interface{}), nil
	}

	return map[string]interface{}{
		"lines":        int64(lineCount),
		"display_text": p.formatLineCount(lineCount),
	}, nil
}

// countLines counts the number of lines in a file
func (p *LineCountPlugin) countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}

// formatLineCount formats the line count for display
func (p *LineCountPlugin) formatLineCount(lines int) string {
	if lines == 0 {
		return "0 lines"
	} else if lines == 1 {
		return "1 line"
	} else if lines < 1000 {
		return strconv.Itoa(lines) + " lines"
	} else if lines < 1000000 {
		// Show as "1.2K lines" for readability
		k := float64(lines) / 1000.0
		if k == float64(int(k)) {
			return strconv.Itoa(int(k)) + "K lines"
		}
		return strconv.FormatFloat(k, 'f', 1, 64) + "K lines"
	} else {
		// Show as "1.2M lines" for very large files
		m := float64(lines) / 1000000.0
		if m == float64(int(m)) {
			return strconv.Itoa(int(m)) + "M lines"
		}
		return strconv.FormatFloat(m, 'f', 1, 64) + "M lines"
	}
}

// Format formats the metadata for display
func (p *LineCountPlugin) Format(metadata map[string]interface{}) string {
	if displayText, ok := metadata["display_text"].(string); ok {
		return displayText
	}
	return ""
}