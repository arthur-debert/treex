package builtin

import (
	"github.com/adebert/treex/pkg/core/plugins"
)

// RegisterBuiltinPlugins registers all built-in plugins with the given registry
func RegisterBuiltinPlugins(registry *plugins.Registry) error {
	builtinPlugins := []plugins.FileInfoPlugin{
		&SizePlugin{},
		&DateCreatedPlugin{},
		&DateModifiedPlugin{},
	}
	
	for _, plugin := range builtinPlugins {
		if err := registry.Register(plugin); err != nil {
			return err
		}
	}
	
	return nil
}