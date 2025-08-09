package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseSize parses human-readable size strings like "12mb", "1.5gb", "500k"
func ParseSize(sizeStr string) (int64, error) {
	// Normalize the string
	sizeStr = strings.TrimSpace(strings.ToLower(sizeStr))
	
	// Regular expression to match number and optional unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([kmgtpe]?b?)?$`)
	matches := re.FindStringSubmatch(sizeStr)
	
	if matches == nil {
		// Try parsing as plain number
		if num, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			return num, nil
		}
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	
	// Parse the numeric part
	numStr := matches[1]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", numStr)
	}
	
	// Parse the unit
	unit := matches[2]
	var multiplier float64
	
	switch unit {
	case "b", "":
		multiplier = 1
	case "k", "kb":
		multiplier = 1024
	case "m", "mb":
		multiplier = 1024 * 1024
	case "g", "gb":
		multiplier = 1024 * 1024 * 1024
	case "t", "tb":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "p", "pb":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	case "e", "eb":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
	
	// Calculate the size in bytes
	bytes := int64(num * multiplier)
	return bytes, nil
}

// ParseQueryValue parses a query value based on the attribute type
func ParseQueryValue(value string, attrType AttributeType) (interface{}, error) {
	switch attrType {
	case StringType:
		return value, nil
		
	case NumericType:
		// Check if this looks like a range (for between operator)
		if strings.Contains(value, "-") {
			// Return as string for between operator to parse
			return value, nil
		}
		
		// For size attributes, try parsing as human-readable first
		if bytes, err := ParseSize(value); err == nil {
			return bytes, nil
		}
		// Fall back to plain integer parsing
		if num, err := strconv.ParseInt(value, 10, 64); err == nil {
			return num, nil
		}
		return nil, fmt.Errorf("invalid numeric value: %s", value)
		
	case DateType:
		// TODO: Implement date parsing
		return nil, fmt.Errorf("date parsing not implemented yet")
		
	case BoolType:
		// Parse boolean values
		lower := strings.ToLower(value)
		if lower == "true" || lower == "yes" || lower == "1" {
			return true, nil
		}
		if lower == "false" || lower == "no" || lower == "0" {
			return false, nil
		}
		return nil, fmt.Errorf("invalid boolean value: %s", value)
		
	default:
		return nil, fmt.Errorf("unknown attribute type: %v", attrType)
	}
}