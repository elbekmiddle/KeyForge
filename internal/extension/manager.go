package extension

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// ErrNotGnome is returned by EnsureInstalled when the current session
// isn't running GNOME Shell — the extension has nothing to attach to.
var ErrNotGnome = fmt.Errorf("extension: current desktop session is not GNOME Shell")

// IsGnomeSession reports whether the current session looks like GNOME
// Shell, based on the same environment variables GNOME itself sets.
func IsGnomeSession() bool {
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if strings.Contains(desktop, "gnome") {
		return true
	}

	// Fallback: some session managers only set this one.
	return strings.ToLower(os.Getenv("DESKTOP_SESSION")) == "gnome"
}

// EnsureInstalled makes sure the KeyForge GNOME extension is installed,
// up to date, and enabled. It is safe to call on every `keyforge run` —
// each step is a cheap no-op once already satisfied.
//
// This mirrors doc section 20 (extension lifecycle):
//
//	Check GNOME -> Check extension -> install/update -> Enable -> D-Bus ready
func EnsureInstalled(log *slog.Logger) error {
	if !IsGnomeSession() {
		return ErrNotGnome
	}

	needsUpdate, err := NeedsUpdate()
	if err != nil {
		return err
	}

	if needsUpdate {
		installedVersion, _ := InstalledVersion()

		if err := Install(); err != nil {
			return err
		}

		log.Info(
			"gnome extension installed",
			"uuid", UUID,
			"previous_version", installedVersion,
		)
	} else {
		log.Debug("gnome extension already up to date", "uuid", UUID)
	}

	enabled, err := IsEnabled()
	if err != nil {
		return err
	}

	if !enabled {
		if err := Enable(); err != nil {
			return err
		}

		log.Info("gnome extension enabled", "uuid", UUID)
	} else {
		log.Debug("gnome extension already enabled", "uuid", UUID)
	}

	return nil
}

// Ping calls the extension's D-Bus Ping() method to confirm it is
// installed, enabled, and actually answering on the session bus — not
// just present on disk.
func Ping() error {
	out, err := exec.Command(
		"gdbus", "call", "--session",
		"--dest", "com.elbekmiddle.KeyForge",
		"--object-path", "/com/elbekmiddle/KeyForge",
		"--method", "com.elbekmiddle.KeyForge.Ping",
	).Output()
	if err != nil {
		return fmt.Errorf("extension: ping failed, extension not reachable on d-bus: %w", err)
	}

	if !strings.Contains(string(out), "pong") {
		return fmt.Errorf("extension: unexpected ping response: %s", strings.TrimSpace(string(out)))
	}

	return nil
}
