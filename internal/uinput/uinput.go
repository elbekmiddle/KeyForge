package uinput

import (
	"log/slog"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
)

type Output interface {
	Send(event keyboard.KeyEvent) error
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
