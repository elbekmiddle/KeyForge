package main

import (
	"log/slog"
	"os"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/linux"
	"github.com/elbekmiddle/KeyForge/internal/logger"
)

func main() {
	log := logger.New()

	if len(os.Args) >= 2 && os.Args[1] == "listen" {
		listen(log)
		return
	}

	devices(log)
}

func devices(log *slog.Logger) {
	log.Info("⚒️ KeyForge starting")
	log.Debug("scanning input devices")

	items, err := linux.ListInputDevices()
	if err != nil {
		log.Error("failed to scan input devices",
			"error", err,
		)
		return
	}

	if len(items) == 0 {
		log.Warn("no input devices found")
		return
	}

	for _, d := range items {
		log.Info("device detected",
			"name", d.Name,
			"type", string(d.Type),
			"path", d.Path,
			"bus", d.Bus,
			"vendor_id", d.VendorID,
			"product_id", d.ProductID,
			"manufacturer", d.Manufacturer,
		)
	}

	log.Info("device scan completed",
		"count", len(items),
	)
}

func listen(log *slog.Logger) {
	path := "/dev/input/event6"

	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	listener := keyboard.NewListener(log)

	if err := listener.Listen(path); err != nil {
		log.Error("keyboard listener stopped",
			"error", err,
		)
	}
}
