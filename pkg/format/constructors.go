package format


// Terminal renderer constructors
func NewColorRenderer() *ColorRenderer {
	return &ColorRenderer{}
}

func NewMinimalRenderer() *MinimalRenderer {
	return &MinimalRenderer{}
}

func NewNoColorRenderer() *NoColorRenderer {
	return &NoColorRenderer{}
}

// Data format renderer constructors
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

func NewYAMLRenderer() *YAMLRenderer {
	return &YAMLRenderer{}
}

func NewCompactJSONRenderer() *CompactJSONRenderer {
	return &CompactJSONRenderer{}
}

func NewFlatJSONRenderer() *FlatJSONRenderer {
	return &FlatJSONRenderer{}
}

// Markdown renderer constructors
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

func NewNestedMarkdownRenderer() *NestedMarkdownRenderer {
	return &NestedMarkdownRenderer{}
}

func NewTableMarkdownRenderer() *TableMarkdownRenderer {
	return &TableMarkdownRenderer{}
}

// HTML renderer constructors
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

func NewCompactHTMLRenderer() *CompactHTMLRenderer {
	return &CompactHTMLRenderer{}
}

func NewTableHTMLRenderer() *TableHTMLRenderer {
	return &TableHTMLRenderer{}
}

// Simple list renderer constructor
func NewSimpleListRenderer() *SimpleListRenderer {
	return &SimpleListRenderer{}
}