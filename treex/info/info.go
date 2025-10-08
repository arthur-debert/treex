package info

import (
	"github.com/spf13/afero"
)

// Backward compatibility functions - delegate to new modules

// Collector manages the collection and merging of annotations.
// Deprecated: Use InfoAPI or Gatherer directly for new code.
type Collector struct {
	gatherer *Gatherer
}

// NewCollector creates a new annotation collector.
func NewCollector() *Collector {
	return &Collector{
		gatherer: NewGatherer(),
	}
}

// NewCollectorWithLogger creates a new annotation collector with a custom logger.
func NewCollectorWithLogger(logger Logger) *Collector {
	return &Collector{
		gatherer: NewGathererWithLogger(logger),
	}
}

// CollectAnnotations walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
func (c *Collector) CollectAnnotations(fsys afero.Fs, root string) (map[string]Annotation, error) {
	infoFS := NewAferoInfoFileSystem(fsys)
	return c.gatherer.GatherFromFileSystem(infoFS, root)
}

// OldValidator provides validation functionality for .info files
// Deprecated: Use InfoAPI.Validate() or Validator directly for new code.
type OldValidator struct {
	validator *Validator
}

// NewValidator creates a new info file validator
// Deprecated: Use NewInfoValidator() from validator.go or InfoAPI for new code.
func NewValidator(fs afero.Fs) *OldValidator {
	return &OldValidator{
		validator: NewInfoValidatorWithLogger(nil),
	}
}

// NewValidatorWithLogger creates a new info file validator with a custom logger
// Deprecated: Use NewInfoValidatorWithLogger() from validator.go or InfoAPI for new code.
func NewValidatorWithLogger(fs afero.Fs, logger Logger) *OldValidator {
	return &OldValidator{
		validator: NewInfoValidatorWithLogger(logger),
	}
}

// ValidateDirectory validates all .info files in a directory tree
func (v *OldValidator) ValidateDirectory(rootPath string) (*ValidationResult, error) {
	// Note: The original OldValidator had an fs field, but for backward compatibility
	// we'll need the filesystem to be passed in or use the OS filesystem
	// For now, this is a placeholder that would need the filesystem passed through
	panic("OldValidator.ValidateDirectory needs filesystem - use InfoAPI instead")
}
