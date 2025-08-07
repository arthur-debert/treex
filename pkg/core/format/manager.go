package format

import (
	"fmt"
	"os"

	"github.com/adebert/treex/pkg/core/types"
	"golang.org/x/term"
)

// RenderRequest encapsulates everything needed to render a tree
type RenderRequest struct {
	Tree          *types.Node
	Format        OutputFormat
	Verbose       bool
	ShowStats     bool
	SafeMode      bool
	TerminalWidth int
	// IsTTY indicates if output is going to a terminal
	// If not set (nil), it will be auto-detected
	IsTTY       *bool
	ShowPlugins []string
}

// RenderResponse contains the result of a render operation
type RenderResponse struct {
	Output       string
	Format       OutputFormat
	RendererUsed string
	Stats        *RenderStats
}

// RenderStats contains statistics about the rendering process
type RenderStats struct {
	NodesRendered    int
	AnnotationsFound int
	RenderTime       string
}

// RendererManager provides a high-level interface for tree rendering
type RendererManager struct {
	registry *RendererRegistry
}

// NewRendererManager creates a new renderer manager with the default registry
func NewRendererManager() *RendererManager {
	return &RendererManager{
		registry: GetDefaultRegistry(),
	}
}

// NewRendererManagerWithRegistry creates a manager with a custom registry
func NewRendererManagerWithRegistry(registry *RendererRegistry) *RendererManager {
	return &RendererManager{
		registry: registry,
	}
}

// RenderTree renders a tree using the best format for the given request
func (rm *RendererManager) RenderTree(request RenderRequest) (*RenderResponse, error) {
	// Determine the best format for this request
	selectedFormat, err := rm.selectFormat(request)
	if err != nil {
		return nil, fmt.Errorf("failed to select format: %w", err)
	}

	// Get the appropriate renderer
	renderer, err := rm.registry.GetRenderer(selectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get renderer for format %s: %w", selectedFormat, err)
	}

	// Create render options
	options := RenderOptions{
		Format:        selectedFormat,
		Verbose:       request.Verbose,
		ShowStats:     request.ShowStats,
		SafeMode:      request.SafeMode,
		TerminalWidth: request.TerminalWidth,
		ShowPlugins:   request.ShowPlugins,
	}

	// Render the tree
	output, err := renderer.Render(request.Tree, options)
	if err != nil {
		return nil, fmt.Errorf("failed to render tree: %w", err)
	}

	// Build response
	response := &RenderResponse{
		Output:       output,
		Format:       selectedFormat,
		RendererUsed: renderer.Description(),
		Stats: &RenderStats{
			NodesRendered:    rm.countNodes(request.Tree),
			AnnotationsFound: rm.countAnnotations(request.Tree),
			RenderTime:       "< 1ms", // Simplified for now
		},
	}

	return response, nil
}

// selectFormat determines the best format based on the request
func (rm *RendererManager) selectFormat(request RenderRequest) (OutputFormat, error) {
	// If a specific format is requested, validate and use it
	// This takes precedence over TTY detection
	if request.Format != "" {
		if err := rm.registry.ValidateFormat(request.Format); err != nil {
			return "", fmt.Errorf("invalid format %q: %w", request.Format, err)
		}
		return request.Format, nil
	}

	// Auto-detect format based on output destination
	var isTTY bool
	if request.IsTTY != nil {
		// Use the provided TTY status
		isTTY = *request.IsTTY
	} else {
		// Auto-detect: If stdout is not a TTY (e.g., piped or redirected)
		isTTY = term.IsTerminal(int(os.Stdout.Fd()))
	}

	// If not a TTY, use plain text format
	if !isTTY {
		return FormatNoColor, nil
	}

	// Use default format
	return rm.registry.DefaultFormat(), nil
}

// countNodes counts the total number of nodes in the tree
func (rm *RendererManager) countNodes(node *types.Node) int {
	if node == nil {
		return 0
	}

	count := 1 // Count this node
	for _, child := range node.Children {
		count += rm.countNodes(child)
	}
	return count
}

// countAnnotations counts the number of annotated nodes in the tree
func (rm *RendererManager) countAnnotations(node *types.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	if node.Annotation != nil {
		count = 1
	}

	for _, child := range node.Children {
		count += rm.countAnnotations(child)
	}
	return count
}

// GetAvailableFormats returns all available formats
func (rm *RendererManager) GetAvailableFormats() map[OutputFormat]string {
	return rm.registry.ListFormats()
}

// GetFormatHelp returns help text for format selection
func (rm *RendererManager) GetFormatHelp() string {
	return rm.registry.GetFormatHelp()
}

// ValidateFormat validates that a format is available
func (rm *RendererManager) ValidateFormat(format OutputFormat) error {
	return rm.registry.ValidateFormat(format)
}

// ParseFormat parses a format string into an OutputFormat
func (rm *RendererManager) ParseFormat(formatStr string) (OutputFormat, error) {
	return rm.registry.ParseFormat(formatStr)
}

// GetTerminalFormats returns formats suitable for terminal output
func (rm *RendererManager) GetTerminalFormats() []OutputFormat {
	return rm.registry.GetTerminalFormats()
}

// GetDataFormats returns formats suitable for data exchange
func (rm *RendererManager) GetDataFormats() []OutputFormat {
	return rm.registry.GetDataFormats()
}

// IsTerminalFormat checks if a format is designed for terminal output
func (rm *RendererManager) IsTerminalFormat(format OutputFormat) bool {
	renderer, err := rm.registry.GetRenderer(format)
	if err != nil {
		return false
	}
	return renderer.IsTerminalFormat()
}

// GetDefaultManager returns a shared renderer manager instance
var defaultManager *RendererManager

func GetDefaultManager() *RendererManager {
	if defaultManager == nil {
		defaultManager = NewRendererManager()
	}
	return defaultManager
}

// RenderTreeWithDefaults is a convenience function using the default manager
func RenderTreeWithDefaults(tree *types.Node, format OutputFormat, verbose bool, safeMode bool) (string, error) {
	manager := GetDefaultManager()

	request := RenderRequest{
		Tree:          tree,
		Format:        format,
		Verbose:       verbose,
		SafeMode:      safeMode,
		TerminalWidth: 80,
	}

	response, err := manager.RenderTree(request)
	if err != nil {
		return "", err
	}

	return response.Output, nil
}
