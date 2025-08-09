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
					if strings.Contains(strings.ToLower(line), searchStr) {
						if filter.Operator == "contains" {
							allMatches = append(allMatches, types.Match{
								LineNumber: i + 1,
								Line:       strings.TrimSpace(line),
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
					if re.MatchString(line) {
						if filter.Operator == "matches" {
							allMatches = append(allMatches, types.Match{
								LineNumber: lineNum,
								Line:       strings.TrimSpace(line),
							})
						}
					}
					lineNum++
				}
				
			case "starts-with", "ends-with":
				searchStr := strings.ToLower(filter.Value.(string))
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					lowerLine := strings.ToLower(strings.TrimSpace(line))
					matched := false
					
					if filter.Operator == "starts-with" && strings.HasPrefix(lowerLine, searchStr) {
						matched = true
					} else if filter.Operator == "ends-with" && strings.HasSuffix(lowerLine, searchStr) {
						matched = true
					}
					
					if matched {
						allMatches = append(allMatches, types.Match{
							LineNumber: i + 1,
							Line:       strings.TrimSpace(line),
						})
					}
				}
			}
		}
	}
	
	return matches, allMatches, nil
}