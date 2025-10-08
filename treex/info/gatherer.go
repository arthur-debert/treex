package info

import (
	"io"

	"github.com/jwaldrip/treex/treex/logging"
)

// Gatherer coordinates the collection and merging of annotations
type Gatherer struct {
	parser *Parser
	merger *Merger
}

// NewGatherer creates a new gatherer instance
func NewGatherer() *Gatherer {
	return &Gatherer{
		parser: NewParser(),
		merger: NewMerger(),
	}
}

// GatherFromMap takes a map of info file paths to content and returns merged annotations.
// This method works purely in memory without file system access.
func (g *Gatherer) GatherFromMap(infoFiles InfoFileMap, pathExists func(string) bool) map[string]Annotation {
	var allAnnotations []Annotation

	for infoFilePath, content := range infoFiles {
		annotations := g.parser.Parse(content, infoFilePath)
		allAnnotations = append(allAnnotations, annotations...)
	}

	return g.merger.MergeAnnotations(allAnnotations, pathExists)
}

// GatherFromFileSystem walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
// This method still uses file system access for reading files.
func (g *Gatherer) GatherFromFileSystem(fs InfoFileSystem, root string) (map[string]Annotation, error) {
	// Read all .info files into a map
	infoFiles, err := fs.FindInfoFiles(root)
	if err != nil {
		return nil, err
	}

	infoFileMap := make(InfoFileMap)
	for _, infoFilePath := range infoFiles {
		file, err := fs.ReadInfoFile(infoFilePath)
		if err != nil {
			logging.Warn().Str("file", infoFilePath).Err(err).Msg("cannot open .info file")
			continue
		}

		// Read all content from reader
		content, err := io.ReadAll(file)
		if err != nil {
			logging.Warn().Str("file", infoFilePath).Err(err).Msg("cannot read .info file")
			continue
		}
		infoFileMap[infoFilePath] = string(content)
	}

	// Use pure function to gather annotations
	return g.GatherFromMap(infoFileMap, fs.PathExists), nil
}
