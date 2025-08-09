package query

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/adebert/treex/pkg/core/types"
)

// MatchCollector extends Matcher to also collect matching lines
type MatchCollector struct {
	*Matcher
	collectMatches bool
}

// NewMatchCollector creates a new match collector
func NewMatchCollector(registry *Registry, query *Query, collectMatches bool) *MatchCollector {
	return &MatchCollector{
		Matcher:        NewMatcher(registry, query),
		collectMatches: collectMatches,
	}
}

// MatchesWithDetail returns whether node matches and collects matching lines for text queries
func (mc *MatchCollector) MatchesWithDetail(node *types.Node) (bool, []types.Match, error) {
	// First check if it matches at all
	matches, err := mc.Matches(node)
	if err != nil || !matches || !mc.collectMatches {
		return matches, nil, err
	}

	// If it matches and we're collecting matches, find the specific lines
	var allMatches []types.Match

	// Check each filter for text-based matches
	for _, filter := range mc.query.Filters {
		if filter.Attribute == "text" {
			// Check if operator exists
			_, exists := mc.registry.GetOperator(filter.Operator)
			if !exists {
				continue
			}

			// Read file content again (not ideal, but keeps separation of concerns)
			content, err := readFileContent(node.Path)
			if err != nil {
				continue
			}

			// Collect matches based on operator type
			switch filter.Operator {
			case "contains", "not-contains":
				searchStr := strings.ToLower(filter.Value.(string))
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					lowerLine := strings.ToLower(line)
					if strings.Contains(lowerLine, searchStr) {
						if filter.Operator == "contains" {
							// Trim the line first to get the actual content we'll display
							trimmedLine := strings.TrimSpace(line)
							// Find positions in the trimmed line
							positions := findAllPositions(trimmedLine, filter.Value.(string))
							allMatches = append(allMatches, types.Match{
								LineNumber:     i + 1,
								Line:           trimmedLine,
								MatchPositions: positions,
							})
						}
					}
				}

			case "matches", "not-matches":
				pattern := filter.Value.(string)
				re, err := regexp.Compile("(?i)" + pattern)
				if err != nil {
					continue
				}

				scanner := bufio.NewScanner(strings.NewReader(content))
				lineNum := 1
				for scanner.Scan() {
					line := scanner.Text()
					trimmedLine := strings.TrimSpace(line)
					matches := re.FindAllStringIndex(trimmedLine, -1)
					if len(matches) > 0 && filter.Operator == "matches" {
						// Convert matches to our position format
						positions := make([][]int, len(matches))
						for i, match := range matches {
							positions[i] = []int{match[0], match[1]}
						}
						allMatches = append(allMatches, types.Match{
							LineNumber:     lineNum,
							Line:           trimmedLine,
							MatchPositions: positions,
						})
					}
					lineNum++
				}

			case "starts-with", "ends-with":
				searchStr := strings.ToLower(filter.Value.(string))
				originalSearchStr := filter.Value.(string)
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					trimmedLine := strings.TrimSpace(line)
					lowerLine := strings.ToLower(trimmedLine)
					var positions [][]int

					if filter.Operator == "starts-with" && strings.HasPrefix(lowerLine, searchStr) {
						// Match is at the beginning
						positions = [][]int{{0, len(originalSearchStr)}}
					} else if filter.Operator == "ends-with" && strings.HasSuffix(lowerLine, searchStr) {
						// Match is at the end
						start := len(trimmedLine) - len(originalSearchStr)
						positions = [][]int{{start, len(trimmedLine)}}
					}

					if len(positions) > 0 {
						allMatches = append(allMatches, types.Match{
							LineNumber:     i + 1,
							Line:           trimmedLine,
							MatchPositions: positions,
						})
					}
				}
			}
		}
	}

	return matches, allMatches, nil
}

// findAllPositions finds all case-insensitive positions of searchStr in text
func findAllPositions(text, searchStr string) [][]int {
	var positions [][]int
	lowerText := strings.ToLower(text)
	lowerSearch := strings.ToLower(searchStr)

	start := 0
	for {
		index := strings.Index(lowerText[start:], lowerSearch)
		if index == -1 {
			break
		}

		realIndex := start + index
		positions = append(positions, []int{realIndex, realIndex + len(searchStr)})
		start = realIndex + 1
	}

	return positions
}
