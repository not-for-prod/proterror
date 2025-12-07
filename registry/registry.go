package registry

import (
	"reflect"
	"sync"
)

var registry = New()

// Instance returns the global registry instance.
func Instance() *Registry {
	return registry
}

// Registry stores a list of known error types.
// Safe for concurrent use.
type Registry struct {
	mu    sync.RWMutex
	types map[reflect.Type]bool
}

// New creates an empty registry.
func New() *Registry {
	return &Registry{
		types: make(map[reflect.Type]bool),
	}
}

// Add registers a new error type.
func (r *Registry) Add(_type any) {
	if _type == nil {
		return
	}

	reflectType := reflect.TypeOf(_type)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.types[reflectType] = true
}

// Has checks if the given error type is registered.
func (r *Registry) Has(_type any) bool {
	reflectType := reflect.TypeOf(_type)

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.types[reflectType]

	return exists
}
