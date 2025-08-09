package query

import (
	"fmt"
	"sync"
)

// Registry manages attributes and operators
type Registry struct {
	mu         sync.RWMutex
	attributes map[string]*Attribute
	operators  map[string]*Operator
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		attributes: make(map[string]*Attribute),
		operators:  make(map[string]*Operator),
	}
}

// RegisterAttribute registers a new attribute
func (r *Registry) RegisterAttribute(attr *Attribute) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.attributes[attr.Name]; exists {
		return fmt.Errorf("attribute %s already registered", attr.Name)
	}

	r.attributes[attr.Name] = attr
	return nil
}

// RegisterOperator registers a new operator
func (r *Registry) RegisterOperator(op *Operator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.operators[op.Name]; exists {
		return fmt.Errorf("operator %s already registered", op.Name)
	}

	r.operators[op.Name] = op
	
	// Also register aliases
	for _, alias := range op.Aliases {
		if _, exists := r.operators[alias]; exists {
			return fmt.Errorf("operator alias %s already registered", alias)
		}
		r.operators[alias] = op
	}
	
	return nil
}

// GetAttribute returns an attribute by name
func (r *Registry) GetAttribute(name string) (*Attribute, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	attr, exists := r.attributes[name]
	return attr, exists
}

// GetOperator returns an operator by name or alias
func (r *Registry) GetOperator(name string) (*Operator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	op, exists := r.operators[name]
	return op, exists
}

// Attributes returns all registered attributes
func (r *Registry) Attributes() map[string]*Attribute {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Return a copy to prevent external modification
	result := make(map[string]*Attribute)
	for k, v := range r.attributes {
		result[k] = v
	}
	return result
}

// OperatorsForType returns operators valid for a given attribute type
func (r *Registry) OperatorsForType(attrType AttributeType) []*Operator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	seen := make(map[*Operator]bool)
	var result []*Operator
	
	for _, op := range r.operators {
		// Skip if we've already seen this operator (due to aliases)
		if seen[op] {
			continue
		}
		
		// Check if this operator is valid for the attribute type
		for _, validType := range op.ValidTypes {
			if validType == attrType {
				result = append(result, op)
				seen[op] = true
				break
			}
		}
	}
	
	return result
}

// GenerateFlagName creates a flag name from attribute and operator
func GenerateFlagName(attribute, operator string) string {
	return fmt.Sprintf("%s--%s", attribute, operator)
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetGlobalRegistry returns the global registry instance
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// RegisterBuiltinAttributes registers all built-in attributes
// This will be called by the attributes package
func RegisterBuiltinAttributes() error {
	// Implemented in init.go to avoid circular imports
	return registerAllAttributes()
}