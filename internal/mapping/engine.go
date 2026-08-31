package mapping

import (
	"github.com/elbekmiddle/KeyForge/internal/keyboard"
)

type Engine struct {
	mappings map[string]Mapping
}

func NewEngine() *Engine {
	return &Engine{
		mappings: make(map[string]Mapping),
	}
}

func (e *Engine) SetMappings(mappings []Mapping) {
	e.mappings = make(map[string]Mapping, len(mappings))

	for _, mapping := range mappings {
		e.mappings[mapping.From] = mapping
	}
}

func (e *Engine) Lookup(event keyboard.KeyEvent) (Mapping, bool) {
	mapping, ok := e.mappings[event.Code]

	return mapping, ok
}
