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
	return e.LookupCode(event.Code)
}

// LookupCode looks up a mapping by raw evdev code name (e.g. "KEY_A",
// "BTN_EXTRA"), independent of which device it came from — keyboard and
// mouse both funnel through this.
func (e *Engine) LookupCode(code string) (Mapping, bool) {
	mapping, ok := e.mappings[code]

	return mapping, ok
}
