// Package daemon wires together everything `keyforge run` needs on boot
// (doc section 21):
//
//	Bootstrap
//	   ├── Config / Identity
//	   ├── Extension
//	   ├── ActiveApp
//	   ├── Profiles
//	   └── Keyboard
//
// Authentication, backend sync, and mouse support are later phases (doc
// sections 5, 10, 29) and are not started here yet.
package daemon

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elbekmiddle/KeyForge/internal/activeapp"
	"github.com/elbekmiddle/KeyForge/internal/device"
	"github.com/elbekmiddle/KeyForge/internal/extension"
	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/mapping"
	"github.com/elbekmiddle/KeyForge/internal/profile"
	"github.com/elbekmiddle/KeyForge/internal/uinput"
)

// activeAppPollInterval controls how often we ask GNOME who's focused.
// This never touches the keyboard hot path (doc section 25) — it only
// ever swaps which mapping table the engine is using.
const activeAppPollInterval = 500 * time.Millisecond

// Run boots the full KeyForge daemon: identity, extension, active-app
// tracking, and the keyboard remap pipeline. It blocks until a shutdown
// signal (Ctrl+C / SIGTERM) is received.
func Run(log *slog.Logger, inputPath string) error {
	// ------------------------------------------------------------
	// Identity
	// ------------------------------------------------------------

	id, err := device.Bootstrap()
	if err != nil {
		log.Error("failed to bootstrap device identity", "error", err)
		return err
	}

	if id.Created {
		log.Info(
			"new device registered locally",
			"guid", id.GUID,
			"name", id.Device.Name,
		)
	} else {
		log.Info(
			"device identity loaded",
			"guid", id.GUID,
			"name", id.Device.Name,
		)
	}

	// ------------------------------------------------------------
	// GNOME extension
	// ------------------------------------------------------------

	extensionReady := true

	if err := extension.EnsureInstalled(log); err != nil {
		extensionReady = false

		log.Warn(
			"gnome extension unavailable, falling back to Shell.Eval for active-app detection",
			"error", err,
		)
	}

	// ------------------------------------------------------------
	// Profiles (static default for now — Phase 7 covers loading
	// profiles/*.json from disk per-application)
	// ------------------------------------------------------------

	profiles := profile.NewCache()

	profiles.Set(profile.Profile{
		ID:   "default",
		Name: "Default",
		Modes: []profile.Mode{
			{
				ID:   0,
				Name: "default",
				Mappings: []mapping.Mapping{
					{From: "KEY_A", To: []string{"KEY_B"}},
				},
			},
		},
	})
	profiles.SetActive("default")

	engine := mapping.NewEngine()
	applyActiveProfile(log, profiles, engine)

	// ------------------------------------------------------------
	// Active application watcher
	// ------------------------------------------------------------

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if extensionReady {
		detector := activeapp.NewAutoDetector()
		watcher := activeapp.NewWatcher(detector, activeAppPollInterval)

		go func() {
			err := watcher.Run(ctx, func(event activeapp.ChangeEvent) {
				log.Debug(
					"active application changed",
					"name", event.Current.Name,
					"class", event.Current.Class,
					"pid", event.Current.PID,
				)

				// Phase 8: map event.Current.Class -> a named profile.
				// Only "default" exists today, so this is a no-op switch,
				// but the wiring is in place for per-app profiles.
				if profiles.SetActive("default") {
					applyActiveProfile(log, profiles, engine)
				}
			})

			if err != nil && ctx.Err() == nil {
				log.Error("active application watcher stopped", "error", err)
			}
		}()
	}

	// ------------------------------------------------------------
	// Keyboard remap pipeline (hot path: RAM-only lookups)
	// ------------------------------------------------------------

	output, err := uinput.New(log)
	if err != nil {
		log.Error("failed to initialize virtual keyboard", "error", err)
		return err
	}
	defer func() {
		if err := output.Close(); err != nil {
			log.Error("failed to close virtual keyboard", "error", err)
		}
	}()

	listener := keyboard.NewListener(log)

	log.Info(
		"keyforge daemon ready",
		"guid", id.GUID,
		"input", inputPath,
		"gnome_extension", extensionReady,
	)

	done := make(chan error, 1)

	go func() {
		done <- listener.Listen(inputPath, func(event keyboard.KeyEvent) {
			mapped, ok := engine.Lookup(event)
			if !ok {
				return
			}

			for _, target := range mapped.To {
				if err := output.Send(keyboard.KeyEvent{Code: target, Value: event.Value}); err != nil {
					log.Error(
						"failed to send mapped key",
						"from", event.Code,
						"to", target,
						"error", err,
					)
					return
				}
			}
		})
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
		return nil

	case err := <-done:
		if err != nil {
			log.Error("keyboard listener stopped", "error", err)
			return err
		}

		return nil
	}
}

func applyActiveProfile(log *slog.Logger, profiles *profile.Cache, engine *mapping.Engine) {
	active, ok := profiles.Active()
	if !ok || len(active.Modes) == 0 {
		return
	}

	engine.SetMappings(active.Modes[0].Mappings)

	log.Debug(
		"mapping engine updated",
		"profile", active.Name,
		"mappings", len(active.Modes[0].Mappings),
	)
}
