package main

import (
	"log/slog"
	"os"

	"github.com/elbekmiddle/KeyForge/internal/activeapp"
	"github.com/elbekmiddle/KeyForge/internal/daemon"
	"github.com/elbekmiddle/KeyForge/internal/device"
	"github.com/elbekmiddle/KeyForge/internal/extension"
	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/linux"
	"github.com/elbekmiddle/KeyForge/internal/logger"
	"github.com/elbekmiddle/KeyForge/internal/mapping"
	"github.com/elbekmiddle/KeyForge/internal/uinput"
)

func main() {
	log := logger.New()

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "run":
			runDaemon(log)
			return

		case "devices":
			runDevices(log)
			return

		case "listen":
			runListen(log)
			return

		case "active-app":
			runActiveApp(log)
			return

		case "extension":
			runExtension(log)
			return

		case "identity":
			runIdentity(log)
			return
		}
	}

	log.Info("⚒️ KeyForge")
	log.Info(
		"usage",
		"commands", "run | devices | listen | active-app | extension | identity",
	)
}

// runDaemon boots the full pipeline: identity, GNOME extension, active-app
// tracking, and keyboard remapping (doc section 21).
func runDaemon(log *slog.Logger) {
	path := "/dev/input/event6"

	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	if err := daemon.Run(log, path); err != nil {
		os.Exit(1)
	}
}

// runExtension installs/updates and enables the KeyForge GNOME extension
// without starting the rest of the daemon — useful for debugging Phase 2/3
// in isolation.
func runExtension(log *slog.Logger) {
	log.Info("ensuring gnome extension is installed and enabled")

	if err := extension.EnsureInstalled(log); err != nil {
		log.Error("failed to ensure gnome extension", "error", err)
		return
	}

	if err := extension.Ping(); err != nil {
		log.Warn(
			"extension installed and enabled, but not yet reachable on d-bus (GNOME may need a reload)",
			"error", err,
		)
		return
	}

	log.Info("gnome extension installed, enabled, and reachable on d-bus")
}

// runIdentity bootstraps (or loads) the local device identity and prints
// it — useful for debugging Phase 4/6 in isolation.
func runIdentity(log *slog.Logger) {
	id, err := device.Bootstrap()
	if err != nil {
		log.Error("failed to bootstrap device identity", "error", err)
		return
	}

	log.Info(
		"device identity",
		"guid", id.GUID,
		"name", id.Device.Name,
		"registered", id.Device.Registered,
		"created_this_run", id.Created,
	)
}

func runDevices(log *slog.Logger) {
	log.Info("⚒️ KeyForge starting")
	log.Debug("scanning input devices")

	devices, err := linux.ListInputDevices()
	if err != nil {
		log.Error(
			"failed to scan input devices",
			"error", err,
		)
		return
	}

	if len(devices) == 0 {
		log.Warn("no input devices found")
		return
	}

	for _, device := range devices {
		log.Info(
			"device detected",
			"name", device.Name,
			"type", device.Type,
			"path", device.Path,
			"bus", device.Bus,
			"vendor_id", device.VendorID,
			"product_id", device.ProductID,
			"manufacturer", device.Manufacturer,
		)
	}

	log.Info(
		"device scan completed",
		"count", len(devices),
	)
}

func runListen(log *slog.Logger) {
	path := "/dev/input/event6"

	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	log.Info(
		"starting keyboard remapper",
		"path", path,
	)

	// ------------------------------------------------------------
	// Virtual keyboard
	// ------------------------------------------------------------

	output, err := uinput.New(log)
	if err != nil {
		log.Error(
			"failed to initialize virtual keyboard",
			"error", err,
		)
		return
	}

	defer func() {
		if err := output.Close(); err != nil {
			log.Error(
				"failed to close virtual keyboard",
				"error", err,
			)
		}
	}()

	// ------------------------------------------------------------
	// Mapping engine
	// ------------------------------------------------------------

	engine := mapping.NewEngine()

	engine.SetMappings([]mapping.Mapping{
		{
			From: "KEY_A",
			To:   []string{"KEY_B"},
		},
	})

	log.Info(
		"mapping engine initialized",
		"mappings", 1,
	)

	// ------------------------------------------------------------
	// Keyboard listener
	// ------------------------------------------------------------

	listener := keyboard.NewListener(log)

	log.Info(
		"keyboard remapper ready",
		"input", path,
		"output", "KeyForge Virtual Keyboard",
	)

	err = listener.Listen(path, func(event keyboard.KeyEvent) {
		// Lookup faqat RAM'dagi mapping table'dan foydalanadi.
		mapped, ok := engine.Lookup(event)

		if !ok {
			return
		}

		log.Debug(
			"mapping matched",
			"from", mapped.From,
			"to", mapped.To,
			"value", event.Value,
		)

		// Har bir target key'ni virtual keyboard orqali OS'ga yuboramiz.
		for _, target := range mapped.To {
			err := output.Send(keyboard.KeyEvent{
				Code:  target,
				Value: event.Value,
			})

			if err != nil {
				log.Error(
					"failed to send mapped key",
					"from", event.Code,
					"to", target,
					"value", event.Value,
					"error", err,
				)

				return
			}

			log.Debug(
				"output event sent",
				"code", target,
				"value", event.Value,
			)
		}
	})

	if err != nil {
		log.Error(
			"keyboard listener stopped",
			"error", err,
		)

		return
	}

	log.Info("keyboard remapper stopped")
}

func runActiveApp(log *slog.Logger) {
	log.Info("detecting active application")

	detector := activeapp.NewAutoDetector()

	app, err := detector.Current()
	if err != nil {
		log.Error(
			"failed to detect active application",
			"error", err,
		)
		return
	}

	log.Info(
		"active application detected",
		"name", app.Name,
		"class", app.Class,
		"executable", app.Executable,
		"pid", app.PID,
	)
}
