package tree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
)

// ViewBuilder extends Builder with view mode capabilities
type ViewBuilder struct {
	*Builder
	viewOptions ViewOptions
}

// NewViewBuilder creates a new view-aware tree builder
func NewViewBuilder(rootPath string, annotations map[string]*info.Annotation, viewOptions ViewOptions) *ViewBuilder {
	return &ViewBuilder{
		Builder:     NewBuilder(rootPath, annotations),
		viewOptions: viewOptions,
	}
}

// NewViewBuilderWithOptions creates a new view-aware tree builder with all options
func NewViewBuilderWithOptions(rootPath string, annotations map[string]*info.Annotation, ignoreFilePath string, maxDepth int, viewOptions ViewOptions) (*ViewBuilder, error) {
	builder, err := NewBuilderWithOptions(rootPath, annotations, ignoreFilePath, maxDepth)
	if err != nil {
		return nil, err
	}

	return &ViewBuilder{
		Builder:     builder,
		viewOptions: viewOptions,
	}, nil
}

// Build constructs the file tree with view mode filtering applied
func (vb *ViewBuilder) Build() (*Node, error) {
	// For ViewModeAll, we need to build without the MAX_FILES_PER_DIR limit
	if vb.viewOptions.Mode == ViewModeAll {
		// Temporarily set a high limit
		oldBuilder := vb.Builder
		vb.Builder = &Builder{
			rootPath:      oldBuilder.rootPath,
			annotations:   oldBuilder.annotations,
			ignoreMatcher: oldBuilder.ignoreMatcher,
			maxDepth:      oldBuilder.maxDepth,
		}
	}

	// First build the full tree using the parent builder
	root, err := vb.Builder.Build()
	if err != nil {
		return nil, err
	}

	// Apply view mode filtering
	vb.applyViewMode(root)

	return root, nil
}

// applyViewMode applies the view mode filtering to the tree
func (vb *ViewBuilder) applyViewMode(node *Node) {
	switch vb.viewOptions.Mode {
	case ViewModeAll:
		// No filtering needed, show everything
		return
	case ViewModeAnnotated:
		vb.filterAnnotatedOnly(node)
	case ViewModeMix:
		vb.applyMixMode(node, 0)
	}
}

// filterAnnotatedOnly recursively filters to show only annotated paths
func (vb *ViewBuilder) filterAnnotatedOnly(node *Node) {
	if node == nil || len(node.Children) == 0 {
		return
	}

	// Filter children to keep only annotated ones or directories containing annotations
	var filteredChildren []*Node
	for _, child := range node.Children {
		if child.Annotation != nil {
			// Keep annotated files/directories
			filteredChildren = append(filteredChildren, child)
			if child.IsDir {
				vb.filterAnnotatedOnly(child)
			}
		} else if child.IsDir && vb.hasAnnotatedDescendants(child) {
			// Keep directories that contain annotated descendants
			filteredChildren = append(filteredChildren, child)
			vb.filterAnnotatedOnly(child)
		}
		// Skip unannotated files and empty directories
	}

	node.Children = filteredChildren

	// Add message if we're at the root and have filtered content
	if node.Parent == nil && len(filteredChildren) > 0 {
		node.Children = append(node.Children, &Node{
			Name:         "",
			Path:         "",
			RelativePath: "",
			IsDir:        false,
			Annotation: &info.Annotation{
				Path:  "",
				Notes: "treex --show all to see all paths",
			},
			Children: []*Node{},
			Parent:   node,
		})
	}
}

// hasAnnotatedDescendants checks if a node has any annotated descendants
func (vb *ViewBuilder) hasAnnotatedDescendants(node *Node) bool {
	if node.Annotation != nil {
		return true
	}

	for _, child := range node.Children {
		if vb.hasAnnotatedDescendants(child) {
			return true
		}
	}

	return false
}

// applyMixMode applies the mix mode logic
func (vb *ViewBuilder) applyMixMode(node *Node, depth int) {
	if node == nil || len(node.Children) == 0 {
		return
	}

	// Separate annotated and unannotated children
	var annotatedChildren []*Node
	var unannotatedChildren []*Node

	for _, child := range node.Children {
		// Skip "more files" indicators
		if strings.HasPrefix(child.Name, "... ") && strings.HasSuffix(child.Name, " more files not shown") {
			continue
		}

		if child.Annotation != nil || (child.IsDir && vb.hasAnnotatedDescendants(child)) {
			annotatedChildren = append(annotatedChildren, child)
		} else {
			unannotatedChildren = append(unannotatedChildren, child)
		}
	}

	// Determine how many context paths to show
	contextCount := vb.calculateContextPaths(node, len(annotatedChildren), len(unannotatedChildren), depth)

	// Build the final children list
	var finalChildren []*Node
	
	// Add all annotated children
	finalChildren = append(finalChildren, annotatedChildren...)
	
	// Add context paths
	if contextCount > 0 && len(unannotatedChildren) > 0 {
		if contextCount >= len(unannotatedChildren) {
			// Show all unannotated children
			finalChildren = append(finalChildren, unannotatedChildren...)
		} else {
			// Show limited context paths
			finalChildren = append(finalChildren, unannotatedChildren[:contextCount]...)
			
			// Add "more items" indicator
			hiddenCount := len(unannotatedChildren) - contextCount
			moreNode := &Node{
				Name:         fmt.Sprintf("... %d more items", hiddenCount),
				Path:         "",
				RelativePath: "",
				IsDir:        false,
				Annotation:   nil,
				Children:     []*Node{},
				Parent:       node,
			}
			finalChildren = append(finalChildren, moreNode)
		}
	}

	// Sort to maintain consistent order
	sort.Slice(finalChildren, func(i, j int) bool {
		// "more items" indicators go last
		if strings.HasPrefix(finalChildren[i].Name, "... ") {
			return false
		}
		if strings.HasPrefix(finalChildren[j].Name, "... ") {
			return true
		}
		
		// Directories first
		if finalChildren[i].IsDir != finalChildren[j].IsDir {
			return finalChildren[i].IsDir
		}
		
		return finalChildren[i].Name < finalChildren[j].Name
	})

	node.Children = finalChildren

	// Recursively apply to child directories
	for _, child := range node.Children {
		if child.IsDir && !strings.HasPrefix(child.Name, "... ") {
			vb.applyMixMode(child, depth+1)
		}
	}
}

// calculateContextPaths determines how many unannotated paths to show as context
func (vb *ViewBuilder) calculateContextPaths(node *Node, annotatedCount, unannotatedCount int, depth int) int {
	// Top-level directory (root)
	if depth == 0 {
		if unannotatedCount <= 6 {
			return unannotatedCount // Show all if 6 or fewer
		}
		return 6 // Show max 6 context paths
	}

	// For directories with 2+ annotated files, show 2 context paths
	if annotatedCount >= 2 {
		return min(2, unannotatedCount)
	}

	// For directories with 1 annotated file, show 1 context path
	if annotatedCount == 1 {
		return min(1, unannotatedCount)
	}

	// No annotated files, don't show any context
	return 0
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}