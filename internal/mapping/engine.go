package mapping

import (
	"sync"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
)

type Engine struct {
	mu       sync.RWMutex
	mappings map[string]Mapping
}

func NewEngine() *Engine {
	return &Engine{
		mappings: make(map[string]Mapping),
	}
}

func (e *Engine) SetMappings(mappings []Mapping) {
	next := make(map[string]Mapping, len(mappings))

	for _, mapping := range mappings {
		next[mapping.From] = mapping
	}

	e.mu.Lock()
	e.mappings = next
	e.mu.Unlock()
}

func (e *Engine) Lookup(event keyboard.KeyEvent) (Mapping, bool) {
	e.mu.RLock()
	mapping, ok := e.mappings[event.Code]
	e.mu.RUnlock()

	return mapping, ok
}

func (e *Engine) Clear() {
	e.mu.Lock()
	e.mappings = make(map[string]Mapping)
	e.mu.Unlock()
}
