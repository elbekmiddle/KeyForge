//go:build !linux

package mouse

import (
	"fmt"
	"log/slog"
	"runtime"
)

// Listener is a no-op stand-in on non-Linux platforms — see
// keyboard/keyboard_other.go for why this exists.
type Listener struct {
	logger *slog.Logger
}

func NewListener(logger *slog.Logger) *Listener {
	return &Listener{logger: logger}
}

func (l *Listener) Listen(
	path string,
	onButton func(ButtonEvent),
	onMotion func(MotionEvent),
) error {
	return fmt.Errorf(
		"mouse: remapping is not implemented on %s yet (evdev is Linux-only)",
		runtime.GOOS,
	)
}
