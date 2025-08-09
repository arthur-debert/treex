package query

import (
	"fmt"
	"regexp"
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
	}
	
	// Register numeric operators
	for _, op := range numericOps {
		err := registry.RegisterOperator(&Operator{
			Name:       op.name,
			Aliases:    op.aliases,
			ValidTypes: []AttributeType{NumericType},
			Comparator: op.comparator,
		})
		if err != nil {
			// Skip duplicate registrations (eq and ne are shared with string ops)
			if !strings.Contains(err.Error(), "already registered") {
				return err
			}
		}
	}
	
	return nil
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