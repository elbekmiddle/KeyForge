package keyboard

import (
	"fmt"
	"log/slog"

	evdev "github.com/holoplot/go-evdev"
)

type Listener struct {
	logger *slog.Logger
}

func NewListener(logger *slog.Logger) *Listener {
	return &Listener{
		logger: logger,
	}
}

func (l *Listener) Listen(path string) error {
	l.logger.Info("opening keyboard",
		"path", path,
	)

	dev, err := evdev.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open keyboard: %w", err)
	}
	defer dev.Close()

	name, err := dev.Name()
	if err != nil {
		return fmt.Errorf("failed to get device name: %w", err)
	}

	l.logger.Info("keyboard listening",
		"name", name,
		"path", path,
	)

	for {
		event, err := dev.ReadOne()
		if err != nil {
			return fmt.Errorf("failed to read keyboard event: %w", err)
		}

		l.logger.Debug("keyboard event",
			"type", event.TypeName(),
			"code", event.CodeName(),
			"value", event.Value,
		)
	}
}
