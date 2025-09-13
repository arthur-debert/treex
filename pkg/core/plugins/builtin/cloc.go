package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/adebert/treex/pkg/core/types"
)

// ClocPlugin collects code statistics using tokei
type ClocPlugin struct{
	tokeiChecked bool
	tokeiAvailable bool
}

// Name returns the plugin identifier
func (p *ClocPlugin) Name() string {
	return "cloc"
}

// Description returns the plugin description
func (p *ClocPlugin) Description() string {
	return "Display lines of code statistics (excluding blanks and comments)"
}

// AppliesTo returns true for all nodes
func (p *ClocPlugin) AppliesTo(node *types.Node) bool {
	return true
}

// tokeiOutput represents the structure of tokei's JSON output
type tokeiOutput map[string]struct {
	Code     int64 `json:"code"`
	Comments int64 `json:"comments"`
	Blanks   int64 `json:"blanks"`
}

// Collect gathers code statistics using tokei
func (p *ClocPlugin) Collect(node *types.Node) (map[string]interface{}, error) {
	// Check if tokei is available on first use
	if !p.tokeiChecked {
		p.tokeiChecked = true
		_, err := exec.LookPath("tokei")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: tokei is not installed. The cloc overlay requires tokei.\nPlease install it from https://github.com/XAMPPRocky/tokei\n")
			os.Exit(1)
		}
		p.tokeiAvailable = true
	}

	if node.IsDir {
		// For directories, return empty metadata
		// The aggregation system will sum up counts from children
		return make(map[string]interface{}), nil
	}

	// Run tokei on the specific file
	cmd := exec.Command("tokei", "--output", "json", node.Path)
	output, err := cmd.Output()
	if err != nil {
		// If tokei fails (e.g., binary file or unsupported type), return empty metadata
		return make(map[string]interface{}), nil
	}

	var result tokeiOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return make(map[string]interface{}), nil
	}

	// Calculate total lines of code (excluding blanks and comments)
	var totalCode int64
	for _, stats := range result {
		totalCode += stats.Code
	}

	if totalCode == 0 {
		return nil, nil
	}

	return map[string]interface{}{
		"cloc":         totalCode,
		"display_text": p.formatCodeLines(totalCode),
	}, nil
}

// formatCodeLines formats the code line count for display
func (p *ClocPlugin) formatCodeLines(lines int64) string {
	if lines == 0 {
		return "0 loc"
	} else if lines == 1 {
		return "1 loc"
	} else if lines < 1000 {
		return fmt.Sprintf("%d loc", lines)
	} else if lines < 1000000 {
		// Show as "1.2K loc" for readability
		k := float64(lines) / 1000.0
		if k == float64(int(k)) {
			return fmt.Sprintf("%dK loc", int(k))
		}
		return fmt.Sprintf("%.1fK loc", k)
	} else {
		// Show as "1.2M loc" for very large counts
		m := float64(lines) / 1000000.0
		if m == float64(int(m)) {
			return fmt.Sprintf("%dM loc", int(m))
		}
		return fmt.Sprintf("%.1fM loc", m)
	}
}

// Format formats the metadata for display
func (p *ClocPlugin) Format(metadata map[string]interface{}) string {
	// Check if we have pre-formatted display text
	if displayText, ok := metadata["display_text"].(string); ok {
		return displayText
	}
	
	// For aggregated directory values
	if cloc, ok := metadata["cloc"].(int64); ok {
		return p.formatCodeLines(cloc)
	}
	
	return ""
}