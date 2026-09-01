package mouse

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

// Listen reads from an evdev mouse device at path. onButton fires for
// EV_KEY (button) events — these go through the mapping engine. onMotion
// fires for EV_REL (movement/scroll) events — these are always passed
// through untouched; there is no "remap" concept for continuous motion.
func (l *Listener) Listen(
	path string,
	onButton func(ButtonEvent),
	onMotion func(MotionEvent),
) error {
	l.logger.Info("opening mouse", "path", path)

	dev, err := evdev.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open mouse: %w", err)
	}
	defer dev.Close()

	name, err := dev.Name()
	if err != nil {
		return fmt.Errorf("failed to get device name: %w", err)
	}

	l.logger.Info("mouse listening", "name", name, "path", path)

	for {
		event, err := dev.ReadOne()
		if err != nil {
			return fmt.Errorf("failed to read mouse event: %w", err)
		}

		switch event.TypeName() {
		case "EV_KEY":
			if onButton == nil {
				continue
			}

			onButton(ButtonEvent{
				Code:  event.CodeName(),
				Value: ButtonValue(event.Value),
			})

		case "EV_REL":
			if onMotion == nil {
				continue
			}

			onMotion(MotionEvent{
				Axis:  event.CodeName(),
				Value: event.Value,
			})

		default:
			continue
		}
	}
}
