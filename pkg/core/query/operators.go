package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// RegisterBuiltinOperators registers all built-in operators
func RegisterBuiltinOperators() error {
	registry := GetGlobalRegistry()
	
	// String operators
	stringOps := []struct {
		name       string
		aliases    []string
		comparator func(nodeValue, queryValue interface{}) (bool, error)
	}{
		{
			name:    "contains",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return strings.Contains(strings.ToLower(nodeStr), strings.ToLower(queryStr)), nil
			},
		},
		{
			name:    "not-contains",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return !strings.Contains(strings.ToLower(nodeStr), strings.ToLower(queryStr)), nil
			},
		},
		{
			name:    "starts-with",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return strings.HasPrefix(strings.ToLower(nodeStr), strings.ToLower(queryStr)), nil
			},
		},
		{
			name:    "not-starts-with",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return !strings.HasPrefix(strings.ToLower(nodeStr), strings.ToLower(queryStr)), nil
			},
		},
		{
			name:    "ends-with",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return strings.HasSuffix(strings.ToLower(nodeStr), strings.ToLower(queryStr)), nil
			},
		},
		{
			name:    "not-ends-with",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return !strings.HasSuffix(strings.ToLower(nodeStr), strings.ToLower(queryStr)), nil
			},
		},
		{
			name:    "matches",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Compile regex with case-insensitive flag
				re, err := regexp.Compile("(?i)" + queryStr)
				if err != nil {
					return false, fmt.Errorf("invalid regex: %w", err)
				}
				return re.MatchString(nodeStr), nil
			},
		},
		{
			name:    "not-matches",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Compile regex with case-insensitive flag
				re, err := regexp.Compile("(?i)" + queryStr)
				if err != nil {
					return false, fmt.Errorf("invalid regex: %w", err)
				}
				return !re.MatchString(nodeStr), nil
			},
		},
		{
			name:    "eq",
			aliases: []string{"equals", "="},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return strings.EqualFold(nodeStr, queryStr), nil
			},
		},
		{
			name:    "ne",
			aliases: []string{"not-equals", "!="},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeStr, ok := nodeValue.(string)
				if !ok {
					return false, fmt.Errorf("node value is not a string")
				}
				queryStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("query value is not a string")
				}
				// Case-insensitive by default
				return !strings.EqualFold(nodeStr, queryStr), nil
			},
		},
	}
	
	// Register string operators
	for _, op := range stringOps {
		err := registry.RegisterOperator(&Operator{
			Name:       op.name,
			Aliases:    op.aliases,
			ValidTypes: []AttributeType{StringType},
			Comparator: op.comparator,
		})
		if err != nil {
			return err
		}
	}
	
	// Numeric operators
	numericOps := []struct {
		name       string
		aliases    []string
		comparator func(nodeValue, queryValue interface{}) (bool, error)
	}{
		{
			name:    "eq",
			aliases: []string{"equals", "="},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				queryNum, err := toInt64(queryValue)
				if err != nil {
					return false, err
				}
				return nodeNum == queryNum, nil
			},
		},
		{
			name:    "ne",
			aliases: []string{"not-equals", "!="},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				queryNum, err := toInt64(queryValue)
				if err != nil {
					return false, err
				}
				return nodeNum != queryNum, nil
			},
		},
		{
			name:    "gt",
			aliases: []string{">"},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				queryNum, err := toInt64(queryValue)
				if err != nil {
					return false, err
				}
				return nodeNum > queryNum, nil
			},
		},
		{
			name:    "gte",
			aliases: []string{">=", "ge"},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				queryNum, err := toInt64(queryValue)
				if err != nil {
					return false, err
				}
				return nodeNum >= queryNum, nil
			},
		},
		{
			name:    "lt",
			aliases: []string{"<"},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				queryNum, err := toInt64(queryValue)
				if err != nil {
					return false, err
				}
				return nodeNum < queryNum, nil
			},
		},
		{
			name:    "lte",
			aliases: []string{"<=", "le"},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				queryNum, err := toInt64(queryValue)
				if err != nil {
					return false, err
				}
				return nodeNum <= queryNum, nil
			},
		},
		{
			name:    "between",
			aliases: []string{},
			comparator: func(nodeValue, queryValue interface{}) (bool, error) {
				nodeNum, err := toInt64(nodeValue)
				if err != nil {
					return false, err
				}
				// Parse range value (e.g., "100-500" or "1kb-10kb")
				rangeStr, ok := queryValue.(string)
				if !ok {
					return false, fmt.Errorf("between operator requires a range string (e.g., '100-500')")
				}
				
				// Split on dash
				parts := strings.Split(rangeStr, "-")
				if len(parts) != 2 {
					return false, fmt.Errorf("invalid range format, expected 'min-max'")
				}
				
				// Parse min and max values
				// For now, parse as plain integers
				// TODO: Use ParseSize when we reorganize to avoid circular imports
				minVal, err := toInt64(strings.TrimSpace(parts[0]))
				if err != nil {
					// Try parsing as size
					if val, err2 := parseSimpleSize(strings.TrimSpace(parts[0])); err2 == nil {
						minVal = val
					} else {
						return false, fmt.Errorf("invalid min value in range: %w", err)
					}
				}
				
				maxVal, err := toInt64(strings.TrimSpace(parts[1]))
				if err != nil {
					// Try parsing as size
					if val, err2 := parseSimpleSize(strings.TrimSpace(parts[1])); err2 == nil {
						maxVal = val
					} else {
						return false, fmt.Errorf("invalid max value in range: %w", err)
					}
				}
				
				return nodeNum >= minVal && nodeNum <= maxVal, nil
			},
		},
	}
	
	// Register numeric operators
	for _, op := range numericOps {
		// For eq and ne, we need to update the existing operator to support numeric types too
		if op.name == "eq" || op.name == "ne" {
			if existingOp, exists := registry.GetOperator(op.name); exists {
				// Add numeric type to valid types
				existingOp.ValidTypes = append(existingOp.ValidTypes, NumericType)
				// We need to wrap the comparators to handle both string and numeric
				stringComparator := existingOp.Comparator
				numericComparator := op.comparator
				existingOp.Comparator = func(nodeValue, queryValue interface{}) (bool, error) {
					// Try numeric comparison first
					if _, err := toInt64(nodeValue); err == nil {
						return numericComparator(nodeValue, queryValue)
					}
					// Fall back to string comparison
					return stringComparator(nodeValue, queryValue)
				}
				continue
			}
		}
		
		err := registry.RegisterOperator(&Operator{
			Name:       op.name,
			Aliases:    op.aliases,
			ValidTypes: []AttributeType{NumericType},
			Comparator: op.comparator,
		})
		if err != nil {
			return err
		}
	}
	
	return nil
}

// parseSimpleSize parses simple size strings like "1k", "2mb"
func parseSimpleSize(s string) (int64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	
	// Extract number and unit
	var numStr string
	var unit string
	
	for i, ch := range s {
		if (ch >= '0' && ch <= '9') || ch == '.' {
			continue
		}
		numStr = s[:i]
		unit = s[i:]
		break
	}
	
	if numStr == "" {
		numStr = s
	}
	
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}
	
	multiplier := int64(1)
	switch unit {
	case "", "b":
		multiplier = 1
	case "k", "kb":
		multiplier = 1024
	case "m", "mb":
		multiplier = 1024 * 1024
	case "g", "gb":
		multiplier = 1024 * 1024 * 1024
	}
	
	return int64(num * float64(multiplier)), nil
}

// toInt64 converts various numeric types to int64
func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}