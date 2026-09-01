// Package daemon wires together everything `keyforge run` needs on boot
// (doc section 21):
//
//	Bootstrap
//	   ├── Config / Identity
//	   ├── Extension
//	   ├── ActiveApp
//	   ├── Profiles
//	   ├── Keyboard
//	   └── Mouse
//
// Authentication, backend sync, and conflict resolution are later phases
// (doc sections 5, 10, 27-30) and are not started here yet.
package daemon

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elbekmiddle/KeyForge/internal/activeapp"
	"github.com/elbekmiddle/KeyForge/internal/config"
	"github.com/elbekmiddle/KeyForge/internal/device"
	"github.com/elbekmiddle/KeyForge/internal/extension"
	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/linux"
	"github.com/elbekmiddle/KeyForge/internal/mapping"
	"github.com/elbekmiddle/KeyForge/internal/mouse"
	"github.com/elbekmiddle/KeyForge/internal/profile"
	"github.com/elbekmiddle/KeyForge/internal/uinput"
)

// activeAppPollInterval controls how often we ask GNOME who's focused.
// This never touches the keyboard/mouse hot path (doc section 25) — it
// only ever swaps which mapping table the engine is using.
const activeAppPollInterval = 500 * time.Millisecond

// Options lets callers override auto-detected device paths. Empty fields
// fall back to config.json / auto-detection.
type Options struct {
	KeyboardPath string
	MousePath    string // "" disables mouse remapping for this run
}

// Run boots the full KeyForge daemon: identity, extension, active-app
// tracking, profile matching, and the keyboard + mouse remap pipelines.
// It blocks until a shutdown signal (Ctrl+C / SIGTERM) is received.
func Run(log *slog.Logger, opts Options) error {
	// ------------------------------------------------------------
	// Identity
	// ------------------------------------------------------------

	id, err := device.Bootstrap()
	if err != nil {
		log.Error("failed to bootstrap device identity", "error", err)
		return err
	}

	if id.Created {
		log.Info("new device registered locally", "guid", id.GUID, "name", id.Device.Name)
	} else {
		log.Info("device identity loaded", "guid", id.GUID, "name", id.Device.Name)
	}

	keyboardPath, mousePath := resolveDevicePaths(log, id.GUID, opts)

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
	// Profiles (doc section 15/23): load profiles/*.json from disk,
	// falling back to a starter "default" profile on first run.
	// ------------------------------------------------------------

	profiles, err := loadProfiles(log, id.GUID)
	if err != nil {
		log.Error("failed to load profiles", "error", err)
		return err
	}

	cache := profile.NewCache()
	for _, p := range profiles {
		cache.Set(p)
	}

	keyboardEngine := mapping.NewEngine()
	mouseEngine := mapping.NewEngine()

	applyProfile(log, cache, keyboardEngine, mouseEngine, pickDefault(profiles))

	// ------------------------------------------------------------
	// Active application watcher -> profile switch (doc section 8/23)
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

				matched, ok := profile.Match(cache.All(), event.Current.Class)
				if !ok {
					return
				}

				if active, isActive := cache.Active(); isActive && active.ID == matched.ID {
					return
				}

				applyProfile(log, cache, keyboardEngine, mouseEngine, matched)
			})

			if err != nil && ctx.Err() == nil {
				log.Error("active application watcher stopped", "error", err)
			}
		}()
	}

	// ------------------------------------------------------------
	// Keyboard remap pipeline (hot path: RAM-only lookups)
	// ------------------------------------------------------------

	keyboardOutput, err := uinput.New(log)
	if err != nil {
		log.Error("failed to initialize virtual keyboard", "error", err)
		return err
	}
	defer func() {
		if err := keyboardOutput.Close(); err != nil {
			log.Error("failed to close virtual keyboard", "error", err)
		}
	}()

	done := make(chan error, 2)
	pipelines := 0

	if keyboardPath != "" {
		pipelines++
		go runKeyboardPipeline(log, keyboardPath, keyboardEngine, keyboardOutput, done)
	} else {
		log.Warn("no keyboard device configured, keyboard remapping disabled")
	}

	// ------------------------------------------------------------
	// Mouse remap pipeline (doc section 9 / Phase 9)
	// ------------------------------------------------------------

	if mousePath != "" {
		mouseOutput, err := uinput.NewMouse()
		if err != nil {
			log.Error("failed to initialize virtual mouse", "error", err)
			return err
		}
		defer func() {
			if err := mouseOutput.Close(); err != nil {
				log.Error("failed to close virtual mouse", "error", err)
			}
		}()

		pipelines++
		go runMousePipeline(log, mousePath, mouseEngine, mouseOutput, done)
	} else {
		log.Info("no mouse device configured, mouse remapping disabled")
	}

	log.Info(
		"keyforge daemon ready",
		"guid", id.GUID,
		"keyboard", keyboardPath,
		"mouse", mousePath,
		"gnome_extension", extensionReady,
	)

	if pipelines == 0 {
		log.Warn("nothing to listen on, waiting for shutdown signal")
	}

	for i := 0; i < pipelines; i++ {
		select {
		case <-ctx.Done():
			log.Info("shutdown signal received")
			return nil

		case err := <-done:
			if err != nil {
				log.Error("a remap pipeline stopped", "error", err)
				return err
			}
		}
	}

	<-ctx.Done()
	log.Info("shutdown signal received")

	return nil
}

func runKeyboardPipeline(
	log *slog.Logger,
	path string,
	engine *mapping.Engine,
	output uinput.Output,
	done chan<- error,
) {
	listener := keyboard.NewListener(log)

	done <- listener.Listen(path, func(event keyboard.KeyEvent) {
		mapped, ok := engine.Lookup(event)
		if !ok {
			return
		}

		for _, target := range mapped.To {
			if err := output.Send(keyboard.KeyEvent{Code: target, Value: event.Value}); err != nil {
				log.Error("failed to send mapped key", "from", event.Code, "to", target, "error", err)
				return
			}
		}
	})
}

func runMousePipeline(
	log *slog.Logger,
	path string,
	engine *mapping.Engine,
	output uinput.MouseOutput,
	done chan<- error,
) {
	listener := mouse.NewListener(log)

	done <- listener.Listen(
		path,
		func(event mouse.ButtonEvent) {
			mapped, ok := engine.LookupCode(event.Code)
			if !ok {
				// No remap configured for this button — pass it through
				// unchanged so the mouse still works normally.
				if err := output.SendButton(event); err != nil {
					log.Error("failed to pass through button", "code", event.Code, "error", err)
				}
				return
			}

			for _, target := range mapped.To {
				if err := output.SendButton(mouse.ButtonEvent{Code: target, Value: event.Value}); err != nil {
					log.Error("failed to send mapped button", "from", event.Code, "to", target, "error", err)
					return
				}
			}
		},
		func(event mouse.MotionEvent) {
			// Movement/scroll is always passed through untouched (doc
			// section 14) — there's no "remap" concept for continuous
			// relative motion.
			if err := output.SendMotion(event); err != nil {
				log.Error("failed to pass through motion", "axis", event.Axis, "error", err)
			}
		},
	)
}

func loadProfiles(log *slog.Logger, guid string) ([]profile.Profile, error) {
	profiles, err := profile.LoadAll(guid)
	if err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		def, err := profile.EnsureDefault(guid)
		if err != nil {
			return nil, err
		}

		log.Info("no profiles found, created starter default profile", "path_hint", "profiles/default.json")

		return []profile.Profile{def}, nil
	}

	log.Info("profiles loaded", "count", len(profiles))

	return profiles, nil
}

func pickDefault(profiles []profile.Profile) profile.Profile {
	for _, p := range profiles {
		if p.ID == "default" {
			return p
		}
	}

	if len(profiles) > 0 {
		return profiles[0]
	}

	return profile.Profile{ID: "default", Name: "Default"}
}

func applyProfile(
	log *slog.Logger,
	cache *profile.Cache,
	keyboardEngine *mapping.Engine,
	mouseEngine *mapping.Engine,
	p profile.Profile,
) {
	cache.Set(p)
	cache.SetActive(p.ID)

	var mappings []mapping.Mapping
	if len(p.Modes) > 0 {
		mappings = p.Modes[0].Mappings
	}

	// A single mapping table is split across two engines by code prefix
	// so keyboard and mouse each only see what applies to them; harmless
	// either way since KEY_* and BTN_* never collide, but keeps intent
	// clear if profiles grow separate keyboard/mouse sections later.
	keyboardEngine.SetMappings(mappings)
	mouseEngine.SetMappings(mappings)

	log.Debug("mapping engine updated", "profile", p.Name, "mappings", len(mappings))
}

// resolveDevicePaths applies, in order: explicit Options, saved
// keyboard.json/mouse.json config, then auto-detection via /proc scan
// (doc section 13-14).
func resolveDevicePaths(log *slog.Logger, guid string, opts Options) (keyboardPath, mousePath string) {
	keyboardPath = opts.KeyboardPath
	mousePath = opts.MousePath

	kbCfg, err := config.LoadKeyboardConfig(guid)
	if err != nil {
		log.Warn("failed to load keyboard.json", "error", err)
	} else if keyboardPath == "" && kbCfg.Enabled {
		keyboardPath = kbCfg.Device
	}

	mouseCfg, err := config.LoadMouseConfig(guid)
	if err != nil {
		log.Warn("failed to load mouse.json", "error", err)
	} else if mousePath == "" && mouseCfg.Enabled && len(mouseCfg.Devices) > 0 {
		mousePath = mouseCfg.Devices[0]
	}

	if keyboardPath == "" {
		if found, ok := firstDeviceOfType(device.TypeKeyboard); ok {
			keyboardPath = found
		}
	}

	if mousePath == "" {
		if found, ok := firstDeviceOfType(device.TypeMouse); ok {
			mousePath = found
		}
	}

	return keyboardPath, mousePath
}

func firstDeviceOfType(t device.Type) (string, bool) {
	devices, err := linux.ListInputDevices()
	if err != nil {
		return "", false
	}

	for _, d := range devices {
		if d.Type == t {
			return d.Path, true
		}
	}

	return "", false
}
