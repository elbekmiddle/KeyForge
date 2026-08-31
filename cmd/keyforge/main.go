package main

import (
	"log/slog"
	"os"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/linux"
	"github.com/elbekmiddle/KeyForge/internal/logger"
	"github.com/elbekmiddle/KeyForge/internal/mapping"
)

func main() {
	log := logger.New()

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "devices":
			runDevices(log)
			return

		case "listen":
			runListen(log)
			return
		}
	}

	log.Info("⚒️ KeyForge")
	log.Info("usage",
		"commands", "devices | listen",
	)
}

func runDevices(log *slog.Logger) {
	log.Info("⚒️ KeyForge starting")
	log.Debug("scanning input devices")

	devices, err := linux.ListInputDevices()
	if err != nil {
		log.Error("failed to scan input devices",
			"error", err,
		)
		return
	}

	if len(devices) == 0 {
		log.Warn("no input devices found")
		return
	}

	for _, device := range devices {
		log.Info("device detected",
			"name", device.Name,
			"type", device.Type,
			"path", device.Path,
			"bus", device.Bus,
			"vendor_id", device.VendorID,
			"product_id", device.ProductID,
			"manufacturer", device.Manufacturer,
		)
	}

	log.Info("device scan completed",
		"count", len(devices),
	)
}

func runListen(log *slog.Logger) {
	path := "/dev/input/event6"

	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	engine := mapping.NewEngine()

	engine.SetMappings([]mapping.Mapping{
		{
			From: "KEY_A",
			To:   []string{"KEY_B"},
		},
	})

	listener := keyboard.NewListener(log)

	err := listener.Listen(path, func(event keyboard.KeyEvent) {
		mapping, ok := engine.Lookup(event)
		if !ok {
			return
		}

		log.Debug("mapping matched",
			"from", mapping.From,
			"to", mapping.To,
			"value", event.Value,
		)
	})

	if err != nil {
		log.Error("keyboard listener stopped",
			"error", err,
		)
	}
}
