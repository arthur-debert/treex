package query

import (
	"fmt"
	"strings"
	
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CLIIntegration handles CLI flag registration and query building
type CLIIntegration struct {
	registry *Registry
	flags    map[string]*string // flag name -> flag value pointer
}

// NewCLIIntegration creates a new CLI integration
func NewCLIIntegration(registry *Registry) *CLIIntegration {
	return &CLIIntegration{
		registry: registry,
		flags:    make(map[string]*string),
	}
}

// RegisterFlags dynamically registers query flags on a cobra command
func (c *CLIIntegration) RegisterFlags(cmd *cobra.Command) error {
	// Get all attributes
	attrs := c.registry.Attributes()
	
	for _, attr := range attrs {
		// Get operators valid for this attribute type
		operators := c.registry.OperatorsForType(attr.Type)
		
		for _, op := range operators {
			// Generate flag name
			flagName := GenerateFlagName(attr.Name, op.Name)
			
			// Create a new string pointer for this flag
			flagValue := new(string)
			c.flags[flagName] = flagValue
			
			// Create help text
			help := fmt.Sprintf("Filter by %s %s (type: %s)", attr.Description, op.Name, getTypeName(attr.Type))
			
			// Register the flag
			cmd.Flags().StringVar(flagValue, flagName, "", help)
			
			// Mark as hidden in help unless it's a primary operator
			if !isPrimaryOperator(op.Name) {
				_ = cmd.Flags().MarkHidden(flagName)
			}
		}
	}
	
	return nil
}

// BuildQuery builds a Query from the registered flags
func (c *CLIIntegration) BuildQuery(flags *pflag.FlagSet) (*Query, error) {
	query := &Query{
		Filters: []Filter{},
	}
	
	// Check each registered flag
	for flagName := range c.flags {
		// Check if this flag was actually set by the user
		flag := flags.Lookup(flagName)
		if flag == nil {
			continue
		}
		
		if !flag.Changed {
			// Flag wasn't set by user
			continue
		}
		
		// Get the actual value from the flag
		actualValue := flag.Value.String()
		
		// Skip if flag value is empty
		if actualValue == "" {
			continue
		}
		
		// Parse flag name to extract attribute and operator
		parts := strings.Split(flagName, "--")
		if len(parts) != 2 {
			continue // Skip malformed flags
		}
		
		attrName := parts[0]
		opName := parts[1]
		
		// Get attribute to determine type
		attr, exists := c.registry.GetAttribute(attrName)
		if !exists {
			return nil, fmt.Errorf("unknown attribute in flag %s", flagName)
		}
		
		// Parse the value based on attribute type
		parsedValue, err := ParseQueryValue(actualValue, attr.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", flagName, err)
		}
		
		// Add filter to query
		query.Filters = append(query.Filters, Filter{
			Attribute: attrName,
			Operator:  opName,
			Value:     parsedValue,
		})
	}
	
	return query, nil
}

// getTypeName returns a human-readable name for an attribute type
func getTypeName(t AttributeType) string {
	switch t {
	case StringType:
		return "text"
	case NumericType:
		return "number/size"
	case DateType:
		return "date"
	case BoolType:
		return "boolean"
	default:
		return "unknown"
	}
}

// isPrimaryOperator returns true for operators that should be shown in help
func isPrimaryOperator(opName string) bool {
	primaryOps := map[string]bool{
		"contains":        true,
		"not-contains":    true,
		"starts-with":     true,
		"ends-with":       true,
		"matches":         true,
		"eq":              true,
		"ne":              true,
		"gt":              true,
		"gte":             true,
		"lt":              true,
		"lte":             true,
		"between":         true,
		"not-matches":     true,
	}
	return primaryOps[opName]
}

var initialized bool

// InitializeQuerySystem sets up the global query system
func InitializeQuerySystem() error {
	// Ensure we only initialize once
	if initialized {
		return nil
	}
	
	// Register built-in operators
	if err := RegisterBuiltinOperators(); err != nil {
		return fmt.Errorf("failed to register operators: %w", err)
	}
	
	// Register built-in attributes
	if err := RegisterBuiltinAttributes(); err != nil {
		return fmt.Errorf("failed to register attributes: %w", err)
	}
	
	initialized = true
	return nil
}