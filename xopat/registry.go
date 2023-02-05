package xopat

import (
	"sync"
)

type Registry struct {
	lock            sync.RWMutex
	registeredNames map[string]*Attribute
	allAttributes   []*Attribute
	errOnDuplicate  bool
}

var defaultRegistry = NewRegistry(true)

// NewRegistry is intended for use during replay of logs.
func NewRegistry(errOnDuplicate bool) *Registry {
	return &Registry{
		registeredNames: make(map[string]*Attribute),
		errOnDuplicate:  errOnDuplicate,
	}
}

