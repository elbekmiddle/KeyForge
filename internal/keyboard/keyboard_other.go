//go:build !linux

package keyboard

import (
	"fmt"
	"log/slog"
	"runtime"
)

// Listener is a no-op stand-in on non-Linux platforms. KeyForge's
// keyboard capture is built on Linux evdev; there's no Windows/macOS
// backend yet (see cmd/keyforge-gui's package doc for what that would
// take). This stub exists so packages that depend on keyboard.Listener —
// notably internal/daemon, and the cross-platform GUI that imports it —
// still compile everywhere. Calling Listen just returns an error.
type Listener struct {
	logger *slog.Logger
}

func NewListener(logger *slog.Logger) *Listener {
	return &Listener{logger: logger}
}

func (l *Listener) Listen(path string, handler func(KeyEvent)) error {
	return fmt.Errorf(
		"keyboard: remapping is not implemented on %s yet (evdev is Linux-only)",
		runtime.GOOS,
	)
}
