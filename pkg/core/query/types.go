package query

import (
	"github.com/adebert/treex/pkg/core/types"
)

// AttributeType represents the data type of an attribute
type AttributeType int

const (
	StringType AttributeType = iota
	NumericType
	DateType
	BoolType
)

// Attribute defines a queryable property of a file or directory
type Attribute struct {
	Name        string
	Type        AttributeType
	Description string
	// Extractor function to get the value from a node
	Extractor func(*types.Node) (interface{}, error)
}

// Operator defines a comparison operation
type Operator struct {
	Name       string
	Aliases    []string // e.g., "gte" -> [">=", "ge"]
	ValidTypes []AttributeType
	// Comparator returns true if the comparison matches
	// nodeValue is the value from the node, queryValue is what the user is searching for
	Comparator func(nodeValue, queryValue interface{}) (bool, error)
}

// Filter represents a single query filter
type Filter struct {
	Attribute string
	Operator  string
	Value     interface{}
}

// Query represents a collection of filters (AND logic)
type Query struct {
	Filters []Filter
}