package format

import (
	"encoding/json"

	"github.com/adebert/treex/pkg/core/tree"
	"gopkg.in/yaml.v3"
)

// TreeData represents the tree structure for JSON/YAML output
type TreeData struct {
	Name        string      `json:"name" yaml:"name"`
	Path        string      `json:"path,omitempty" yaml:"path,omitempty"`
	IsDirectory bool        `json:"is_directory" yaml:"is_directory"`
	Annotation  *Annotation `json:"annotation,omitempty" yaml:"annotation,omitempty"`
	Children    []TreeData  `json:"children,omitempty" yaml:"children,omitempty"`
}

// Annotation represents annotation data for JSON/YAML output
type Annotation struct {
	Notes string `json:"notes,omitempty" yaml:"notes,omitempty"`
	
	// Deprecated fields for backwards compatibility
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// JSONRenderer renders trees as JSON
type JSONRenderer struct{}


func (r *JSONRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	data := r.convertToTreeData(root, "")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (r *JSONRenderer) Format() OutputFormat {
	return FormatJSON
}

func (r *JSONRenderer) Description() string {
	return "JSON structured data format"
}

func (r *JSONRenderer) IsTerminalFormat() bool {
	return false
}

func (r *JSONRenderer) convertToTreeData(node *tree.Node, parentPath string) TreeData {
	var currentPath string
	if parentPath == "" {
		currentPath = node.Name
	} else {
		currentPath = parentPath + "/" + node.Name
	}

	data := TreeData{
		Name:        node.Name,
		Path:        currentPath,
		IsDirectory: node.IsDir,
	}

	// Convert annotation if present
	if node.Annotation != nil {
		data.Annotation = &Annotation{
			Title:       node.Annotation.Title,
			Description: node.Annotation.Description,
		}
	}

	// Convert children
	if len(node.Children) > 0 {
		data.Children = make([]TreeData, len(node.Children))
		for i, child := range node.Children {
			data.Children[i] = r.convertToTreeData(child, currentPath)
		}
	}

	return data
}

// YAMLRenderer renders trees as YAML
type YAMLRenderer struct{}


func (r *YAMLRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	data := r.convertToTreeData(root, "")

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

func (r *YAMLRenderer) Format() OutputFormat {
	return FormatYAML
}

func (r *YAMLRenderer) Description() string {
	return "YAML structured data format"
}

func (r *YAMLRenderer) IsTerminalFormat() bool {
	return false
}

func (r *YAMLRenderer) convertToTreeData(node *tree.Node, parentPath string) TreeData {
	var currentPath string
	if parentPath == "" {
		currentPath = node.Name
	} else {
		currentPath = parentPath + "/" + node.Name
	}

	data := TreeData{
		Name:        node.Name,
		Path:        currentPath,
		IsDirectory: node.IsDir,
	}

	// Convert annotation if present
	if node.Annotation != nil {
		data.Annotation = &Annotation{
			Title:       node.Annotation.Title,
			Description: node.Annotation.Description,
		}
	}

	// Convert children
	if len(node.Children) > 0 {
		data.Children = make([]TreeData, len(node.Children))
		for i, child := range node.Children {
			data.Children[i] = r.convertToTreeData(child, currentPath)
		}
	}

	return data
}

// CompactJSONRenderer renders trees as compact JSON (single line)
type CompactJSONRenderer struct{}

func (r *CompactJSONRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	jsonRenderer := &JSONRenderer{}
	data := jsonRenderer.convertToTreeData(root, "")

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (r *CompactJSONRenderer) Format() OutputFormat {
	return "compact-json"
}

func (r *CompactJSONRenderer) Description() string {
	return "Compact JSON format (single line)"
}

func (r *CompactJSONRenderer) IsTerminalFormat() bool {
	return false
}

// FlatJSONRenderer renders trees as a flat array of paths with metadata
type FlatJSONRenderer struct{}


type FlatPath struct {
	Path        string      `json:"path"`
	Name        string      `json:"name"`
	IsDirectory bool        `json:"is_directory"`
	Depth       int         `json:"depth"`
	Annotation  *Annotation `json:"annotation,omitempty"`
}

func (r *FlatJSONRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	var paths []FlatPath
	r.collectPaths(root, "", 0, &paths)

	jsonBytes, err := json.MarshalIndent(paths, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (r *FlatJSONRenderer) collectPaths(node *tree.Node, parentPath string, depth int, paths *[]FlatPath) {
	var currentPath string
	if parentPath == "" {
		currentPath = node.Name
	} else {
		currentPath = parentPath + "/" + node.Name
	}

	flatPath := FlatPath{
		Path:        currentPath,
		Name:        node.Name,
		IsDirectory: node.IsDir,
		Depth:       depth,
	}

	if node.Annotation != nil {
		// Use Notes if available, otherwise fall back to Description for compatibility
		notes := node.Annotation.Notes
		if notes == "" && node.Annotation.Description != "" {
			notes = node.Annotation.Description
		}
		
		flatPath.Annotation = &Annotation{
			Notes:       notes,
			Title:       node.Annotation.Title,       // For backwards compatibility
			Description: node.Annotation.Description, // For backwards compatibility
		}
	}

	*paths = append(*paths, flatPath)

	// Process children
	for _, child := range node.Children {
		r.collectPaths(child, currentPath, depth+1, paths)
	}
}

func (r *FlatJSONRenderer) Format() OutputFormat {
	return "flat-json"
}

func (r *FlatJSONRenderer) Description() string {
	return "Flat JSON array of paths with metadata"
}

func (r *FlatJSONRenderer) IsTerminalFormat() bool {
	return false
}
