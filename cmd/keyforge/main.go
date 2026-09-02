package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/elbekmiddle/KeyForge/internal/activeapp"
	"github.com/elbekmiddle/KeyForge/internal/daemon"
	"github.com/elbekmiddle/KeyForge/internal/device"
	"github.com/elbekmiddle/KeyForge/internal/extension"
	"github.com/elbekmiddle/KeyForge/internal/identity"
	"github.com/elbekmiddle/KeyForge/internal/keyboard"
	"github.com/elbekmiddle/KeyForge/internal/linux"
	"github.com/elbekmiddle/KeyForge/internal/logger"
	"github.com/elbekmiddle/KeyForge/internal/mapping"
	"github.com/elbekmiddle/KeyForge/internal/profile"
	"github.com/elbekmiddle/KeyForge/internal/sync"
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

		case "register":
			runRegisterAccount(log)
			return

		case "login":
			runLogin(log)
			return

		case "logout":
			runLogout(log)
			return

		case "sync":
			runSyncCmd(log)
			return
		}
	}

	log.Info("⚒️ KeyForge")
	log.Info(
		"usage",
		"commands", "run | devices | listen | active-app | extension | identity | register | login | logout | sync",
	)
}

// runDaemon boots the full pipeline: identity, GNOME extension, active-app
// tracking, profile matching, and keyboard + mouse remapping (doc section 21).
//
// Usage: keyforge run [keyboardPath] [mousePath]
// Either path may be omitted to auto-detect / read from keyboard.json /
// mouse.json.
func runDaemon(log *slog.Logger) {
	opts := daemon.Options{}

	if len(os.Args) >= 3 {
		opts.KeyboardPath = os.Args[2]
	}

	if len(os.Args) >= 4 {
		opts.MousePath = os.Args[3]
	}

	if err := daemon.Run(log, opts); err != nil {
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

// runRegisterAccount creates a new backend account (doc section 6:
// "email/password kiritadi" — POST /auth/register) and stores the
// resulting session locally, scoped to this device's identity.
func runRegisterAccount(log *slog.Logger) {
	email, password, err := promptCredentials()
	if err != nil {
		log.Error("failed to read credentials", "error", err)
		return
	}

	authenticate(log, "register", email, password)
}

// runLogin authenticates against an existing backend account (doc
// section 6-7: POST /auth/login) and stores the session locally.
// Password is never written to disk (doc section 6/31) — only the
// resulting tokens are.
func runLogin(log *slog.Logger) {
	email, password, err := promptCredentials()
	if err != nil {
		log.Error("failed to read credentials", "error", err)
		return
	}

	authenticate(log, "login", email, password)
}

func authenticate(log *slog.Logger, mode, email, password string) {
	id, err := device.Bootstrap()
	if err != nil {
		log.Error("failed to bootstrap device identity", "error", err)
		return
	}

	client := sync.NewClient()

	var session sync.Session

	if mode == "register" {
		session, err = client.Register(email, password)
	} else {
		session, err = client.Login(email, password)
	}

	if err != nil {
		log.Error("authentication failed", "mode", mode, "backend", sync.BaseURL(), "error", err)
		return
	}

	if err := sync.SaveSession(id.GUID, session); err != nil {
		log.Error("authenticated, but failed to save local session", "error", err)
		return
	}

	log.Info(
		"authenticated with backend",
		"mode", mode,
		"backend", sync.BaseURL(),
		"guid", id.GUID,
	)
	log.Info("run `keyforge run` to start syncing, or `keyforge sync` for a one-off sync")
}

// runLogout clears the locally stored session (doc section 31: tokens
// never belong in the client's regular config, so this just deletes the
// one file that holds them).
func runLogout(log *slog.Logger) {
	id, err := device.Bootstrap()
	if err != nil {
		log.Error("failed to bootstrap device identity", "error", err)
		return
	}

	if err := sync.ClearSession(id.GUID); err != nil {
		log.Error("failed to clear local session", "error", err)
		return
	}

	log.Info("logged out, local session cleared", "guid", id.GUID)
}

// runSyncCmd does a single push+pull against the backend without
// starting the full daemon — useful for testing connectivity and for
// forcing a sync outside of `keyforge run`.
func runSyncCmd(log *slog.Logger) {
	id, err := device.Bootstrap()
	if err != nil {
		log.Error("failed to bootstrap device identity", "error", err)
		return
	}

	session, ok := sync.LoadSession(id.GUID)
	if !ok {
		log.Error("not logged in, run `keyforge login` first")
		return
	}

	profiles, err := profile.LoadAll(id.GUID)
	if err != nil {
		log.Error("failed to load local profiles", "error", err)
		return
	}

	client := sync.NewClient()

	fingerprint, err := identity.Fingerprint()
	if err != nil {
		log.Warn("failed to compute machine fingerprint", "error", err)
	} else if err := client.RegisterDevice(session.AccessToken, id.GUID, fingerprint, id.Device.Name); err != nil {
		log.Warn("failed to register device with backend", "error", err)
	}

	accepted, skipped, remote, err := client.PushProfiles(session.AccessToken, id.GUID, profiles)
	if err != nil {
		log.Error("sync failed", "error", err)
		return
	}

	for _, p := range remote {
		if err := profile.Save(id.GUID, p); err != nil {
			log.Warn("failed to save synced profile locally", "profile", p.ID, "error", err)
		}
	}

	log.Info(
		"sync complete",
		"pushed_accepted", len(accepted),
		"pushed_skipped", len(skipped),
		"pulled", len(remote),
	)
}

func promptCredentials() (email, password string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	emailLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read email: %w", err)
	}

	// NOTE: password is read in plain text (no terminal echo suppression)
	// — keep this out of shell history / screen recordings. Hidden input
	// is a nice follow-up but isn't required for this to be safe: the
	// password is never persisted anywhere on the client (doc section 6).
	fmt.Print("Password: ")
	passwordLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read password: %w", err)
	}

	return strings.TrimSpace(emailLine), strings.TrimSpace(passwordLine), nil
}
