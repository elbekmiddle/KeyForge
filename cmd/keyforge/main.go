package main

import (
	"log/slog"

	"github.com/elbekmiddle/KeyForge/internal/linux"
	"github.com/elbekmiddle/KeyForge/internal/logger"
)

func main() {
	log := logger.New()

	log.Info("⚒️ KeyForge starting")
	log.Debug("scanning input devices")

	devices, err := linux.ListInputDevices()
	if err != nil {
		log.Error("failed to scan input devices",
			slog.Any("error", err),
		)
		return
	}

	if len(devices) == 0 {
		log.Warn("no input devices found")
		return
	}

	for _, d := range devices {
		log.Info("device detected",
			slog.String("name", d.Name),
			slog.String("path", d.Path),
			slog.String("type", string(d.Type)),
		)
	}

	log.Info("device scan completed",
		slog.Int("count", len(devices)),
	)
}
