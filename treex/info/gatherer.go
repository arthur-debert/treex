package info

import (
	"log"
	"strings"
)

// Gatherer coordinates the collection and merging of annotations
type Gatherer struct {
	parser *Parser
	merger *Merger
	logger Logger
}

// NewGatherer creates a new gatherer instance
func NewGatherer() *Gatherer {
	return &Gatherer{
		parser: NewParser(),
		merger: NewMerger(),
	}
}

// NewGathererWithLogger creates a new gatherer instance with a custom logger
func NewGathererWithLogger(logger Logger) *Gatherer {
	return &Gatherer{
		parser: NewParserWithLogger(logger),
		merger: NewMergerWithLogger(logger),
		logger: logger,
	}
}

// GatherFromMap takes a map of info file paths to content and returns merged annotations.
// This method works purely in memory without file system access.
func (g *Gatherer) GatherFromMap(infoFiles InfoFileMap, pathExists func(string) bool) (map[string]Annotation, error) {
	var allAnnotations []Annotation

	for infoFilePath, content := range infoFiles {
		reader := strings.NewReader(content)
		annotations, err := g.parser.ParseWithLogger(reader, infoFilePath, g.logger)
		if err != nil {
			g.logf("info: cannot parse .info file %q: %v", infoFilePath, err)
			continue
		}
		allAnnotations = append(allAnnotations, annotations...)
	}

	return g.merger.MergeAnnotations(allAnnotations, pathExists), nil
}

// GatherFromFileSystem walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
func (g *Gatherer) GatherFromFileSystem(fs InfoFileSystem, root string) (map[string]Annotation, error) {
	var allAnnotations []Annotation

	infoFiles, err := fs.FindInfoFiles(root)
	if err != nil {
		return nil, err
	}

	for _, infoFilePath := range infoFiles {
		file, err := fs.ReadInfoFile(infoFilePath)
		if err != nil {
			g.logf("info: cannot open .info file %q: %v", infoFilePath, err)
			continue
		}

		annotations, err := g.parser.ParseWithLogger(file, infoFilePath, g.logger)
		if err != nil {
			g.logf("info: cannot parse .info file %q: %v", infoFilePath, err)
			continue
		}
		allAnnotations = append(allAnnotations, annotations...)
	}

	return g.merger.MergeAnnotations(allAnnotations, fs.PathExists), nil
}

// logf logs a warning message using the configured logger, or log.Printf if no logger is set
func (g *Gatherer) logf(format string, v ...interface{}) {
	if g.logger != nil {
		g.logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}
