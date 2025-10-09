package types

// Annotation represents a single file/directory annotation
type Annotation struct {
	Path  string
	Notes string // Complete notes for the file/directory
}

// GitStatus represents Git status information for a file
type GitStatus struct {
	Path      string // File path
	Staged    bool   // File has staged changes
	Unstaged  bool   // File has unstaged changes
	Untracked bool   // File is untracked
	Status    string // Human-readable status description
}
