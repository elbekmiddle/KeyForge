//go:build !linux

package uinput

import (
	"fmt"
	"runtime"

	"github.com/elbekmiddle/KeyForge/internal/mouse"
)

// Mouse is a no-op stand-in on non-Linux platforms — see
// keyboard_other.go for why this exists.
type Mouse struct{}

func NewMouse() (*Mouse, error) {
	return nil, fmt.Errorf(
		"uinput: virtual mouse is not implemented on %s yet (uinput is Linux-only)",
		runtime.GOOS,
	)
}

func (m *Mouse) SendButton(event mouse.ButtonEvent) error {
	return fmt.Errorf("uinput: virtual mouse is not implemented on %s yet", runtime.GOOS)
}

func (m *Mouse) SendMotion(event mouse.MotionEvent) error {
	return fmt.Errorf("uinput: virtual mouse is not implemented on %s yet", runtime.GOOS)
}

func (m *Mouse) Close() error {
	return nil
}
