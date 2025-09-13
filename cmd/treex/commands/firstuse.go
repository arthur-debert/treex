package commands

import (
	"bytes"
	_ "embed"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/adebert/treex/pkg/core/firstuse"
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/types"
)

//go:embed firstuse-msg.txt
var firstUseTemplate string

// FirstUseTemplateData holds the data for rendering the first-use help message
type FirstUseTemplateData struct {
	InfoFileName       string
	ExamplePath        string
	ExampleAnnotation  string
	RenderedTree       string
}

// GenerateFirstUseMessage generates the first-time user help message using the template
func GenerateFirstUseMessage(root *types.Node, examples []firstuse.CommonPath, infoFileName string, outputFormat string) (string, error) {
	// Parse the template
	tmpl, err := template.New("firstuse").Parse(firstUseTemplate)
	if err != nil {
		return "", err
	}

	// Prepare template data
	data := FirstUseTemplateData{
		InfoFileName: infoFileName,
	}

	// Set example data
	if len(examples) > 0 {
		data.ExamplePath = examples[0].Path
		data.ExampleAnnotation = examples[0].Annotation
	} else {
		// Default examples if none found
		data.ExamplePath = "src/"
		data.ExampleAnnotation = "Source code directory"
	}

	// Generate simulated tree with only the first example
	var simulatedExamples []firstuse.CommonPath
	if len(examples) > 0 {
		simulatedExamples = []firstuse.CommonPath{examples[0]}
	}
	simulatedRoot := generateExampleTree(root, simulatedExamples)

	// Render the simulated tree
	renderRequest := format.RenderRequest{
		Tree:          simulatedRoot,
		Format:        parseFormatString(outputFormat),
		Verbose:       false,
		ShowStats:     false,
		SafeMode:      false,
		TerminalWidth: 80,
	}

	manager := format.GetDefaultManager()
	renderResponse, err := manager.RenderTree(renderRequest)
	if err != nil {
		return "", err
	}

	// Indent the tree output to match the command session style
	treeLines := strings.Split(renderResponse.Output, "\n")
	var indentedTree strings.Builder
	for _, line := range treeLines {
		if line != "" {
			indentedTree.WriteString("     " + line)
			if !strings.HasSuffix(line, "\n") {
				indentedTree.WriteString("\n")
			}
		}
	}
	data.RenderedTree = strings.TrimRight(indentedTree.String(), "\n")

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// parseFormatString safely converts a format string to OutputFormat
func parseFormatString(formatStr string) format.OutputFormat {
	if formatStr == "" {
		return "" // Let the manager use defaults
	}

	// Try to parse, but don't fail - let the manager handle validation
	if parsedFormat, err := format.ParseFormatString(formatStr); err == nil {
		return parsedFormat
	}

	// Return as-is and let the manager handle the error
	return format.OutputFormat(formatStr)
}

// generateExampleTree creates a simulated tree with example annotations
func generateExampleTree(root *types.Node, examples []firstuse.CommonPath) *types.Node {
	// Create a copy of the root node
	simulatedRoot := &types.Node{
		Name:     root.Name,
		Path:     root.Path,
		IsDir:    root.IsDir,
		Children: make([]*types.Node, len(root.Children)),
	}
	
	// Create a map of example paths for quick lookup
	exampleMap := make(map[string]string)
	for _, ex := range examples {
		// Normalize the path (remove trailing slash for comparison)
		normalizedPath := strings.TrimSuffix(ex.Path, "/")
		exampleMap[normalizedPath] = ex.Annotation
	}
	
	// Copy the tree and add annotations
	copyTreeWithAnnotations(root, simulatedRoot, exampleMap, "")
	
	return simulatedRoot
}

// copyTreeWithAnnotations recursively copies nodes and adds example annotations
func copyTreeWithAnnotations(src, dst *types.Node, exampleMap map[string]string, parentPath string) {
	// For root node, don't include it in path
	currentPath := src.Name
	if parentPath != "" {
		currentPath = filepath.Join(parentPath, src.Name)
	} else if src.Name == "." || strings.HasPrefix(src.Path, "/") {
		// This is the root, use empty path for children
		currentPath = ""
	}
	
	// Check if this node's path has an example annotation
	checkPath := src.Name
	if parentPath != "" && parentPath != "." {
		checkPath = filepath.Join(parentPath, src.Name)
	}
	
	// Normalize the path (remove trailing slash for comparison)
	normalizedPath := strings.TrimSuffix(checkPath, "/")
	if annotation, found := exampleMap[normalizedPath]; found {
		dst.Annotation = &types.Annotation{
			Notes: annotation,
		}
	}
	
	// Copy children
	if src.Children != nil {
		dst.Children = make([]*types.Node, len(src.Children))
		for i, child := range src.Children {
			dst.Children[i] = &types.Node{
				Name:  child.Name,
				Path:  child.Path,
				IsDir: child.IsDir,
			}
			copyTreeWithAnnotations(child, dst.Children[i], exampleMap, currentPath)
		}
	}
}