package activeapp

import (
	"context"
	"time"
)

type Watcher struct {
	detector Detector
	interval time.Duration
}

func NewWatcher(detector Detector, interval time.Duration) *Watcher {
	return &Watcher{
		detector: detector,
		interval: interval,
	}
}

func (w *Watcher) Run(
	ctx context.Context,
	onChange func(ChangeEvent),
) error {
	current, err := w.detector.Current()
	if err == nil {
		onChange(ChangeEvent{
			Current: current,
		})
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	var previous Application = current

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			next, err := w.detector.Current()
			if err != nil {
				continue
			}

			if sameApplication(previous, next) {
				continue
			}

			event := ChangeEvent{
				Previous: previous,
				Current:  next,
			}

			previous = next
			onChange(event)
		}
	}
}

func sameApplication(a, b Application) bool {
	return a.Name == b.Name &&
		a.Class == b.Class &&
		a.Executable == b.Executable
}
