package event

import "sync"

// Listener func definition
type Listener func(e EventFace)

// Manager event manager definition
type Manager struct {
	name      string
	pool      sync.Pool
	events    map[string]EventFace
	listeners []Listener
}

// NewManager
func NewManager(name string) *Manager {
	return &Manager{name: name}
}
