//go:build !linux

package uinput

import (
	"fmt"
	"runtime"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
)

// Keyboard is a no-op stand-in on non-Linux platforms. uinput (the
// kernel API this uses to create a virtual keyboard) is Linux-only; a
// Windows equivalent would be a low-level keyboard hook plus a virtual
// HID driver, which hasn't been built yet.
type Keyboard struct{}

func NewKeyboard() (*Keyboard, error) {
	return nil, fmt.Errorf(
		"uinput: virtual keyboard is not implemented on %s yet (uinput is Linux-only)",
		runtime.GOOS,
	)
}

func (k *Keyboard) Send(event keyboard.KeyEvent) error {
	return fmt.Errorf("uinput: virtual keyboard is not implemented on %s yet", runtime.GOOS)
}

func (k *Keyboard) Close() error {
	return nil
}
