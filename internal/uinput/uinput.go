package uinput

import (
	"log/slog"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/mouse"
)

type Output interface {
	Send(event keyboard.KeyEvent) error
	Close() error
}

// MouseOutput is the mouse counterpart of Output: it accepts remapped
// button events and passed-through motion/scroll events and re-emits them
// as a virtual mouse.
type MouseOutput interface {
	SendButton(event mouse.ButtonEvent) error
	SendMotion(event mouse.MotionEvent) error
	Close() error
}

func New(logger *slog.Logger) (Output, error) {
	logger.Debug("creating virtual keyboard")

	keyboardDevice, err := NewKeyboard()
	if err != nil {
		logger.Error("failed to create virtual keyboard", "error", err)
		return nil, err
	}

	logger.Info(
		"virtual keyboard created",
		"name", "KeyForge Virtual Keyboard",
	)

	return keyboardDevice, nil
}
