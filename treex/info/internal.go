package info

// Logger interface for warning reporting during info processing
type Logger interface {
	Printf(format string, v ...interface{})
}

// Annotation represents a single entry in an .info file.
type Annotation struct {
	Path       string
	Annotation string
	InfoFile   string // The path to the .info file this annotation came from.
	LineNum    int
}
