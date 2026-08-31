package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtensionsDir returns ~/.local/share/gnome-shell/extensions.
func ExtensionsDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "gnome-shell", "extensions"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("extension: failed to resolve home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share", "gnome-shell", "extensions"), nil
}

// InstallDir returns ~/.local/share/gnome-shell/extensions/keyforge@elbekmiddle.
func InstallDir() (string, error) {
	extensionsDir, err := ExtensionsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(extensionsDir, UUID), nil
}

// InstalledVersion returns the "version" from the currently installed
// metadata.json, or 0 if the extension isn't installed at all.
func InstalledVersion() (int, error) {
	dir, err := InstallDir()
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("extension: failed to read installed metadata.json: %w", err)
	}

	var meta metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return 0, fmt.Errorf("extension: failed to parse installed metadata.json: %w", err)
	}

	return meta.Version, nil
}

// IsInstalled reports whether the extension directory exists on disk.
func IsInstalled() (bool, error) {
	v, err := InstalledVersion()
	if err != nil {
		return false, err
	}

	return v > 0, nil
}

// NeedsUpdate reports whether the embedded extension is newer than what's
// currently installed (or nothing is installed yet).
func NeedsUpdate() (bool, error) {
	embeddedVersion, err := EmbeddedVersion()
	if err != nil {
		return false, err
	}

	installedVersion, err := InstalledVersion()
	if err != nil {
		return false, err
	}

	return installedVersion < embeddedVersion, nil
}

// Install writes the embedded extension files into InstallDir(),
// overwriting anything already there. It does not enable the extension —
// call Enable for that.
func Install() error {
	dir, err := InstallDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("extension: failed to create %s: %w", dir, err)
	}

	names, err := files()
	if err != nil {
		return err
	}

	for _, name := range names {
		data, err := readAsset(name)
		if err != nil {
			return err
		}

		dest := filepath.Join(dir, name)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("extension: failed to write %s: %w", dest, err)
		}
	}

	return nil
}

// IsEnabled reports whether GNOME currently has the extension enabled.
func IsEnabled() (bool, error) {
	out, err := exec.Command("gnome-extensions", "info", UUID).Output()
	if err != nil {
		// Not installed as far as gnome-extensions is concerned, or the
		// tool itself is unavailable — either way, not "enabled".
		return false, nil
	}

	// `gnome-extensions info` prints a "State: ENABLED" line.
	return containsEnabledState(string(out)), nil
}

// Enable asks GNOME Shell to enable the extension.
func Enable() error {
	if err := exec.Command("gnome-extensions", "enable", UUID).Run(); err != nil {
		return fmt.Errorf("extension: failed to enable %s: %w", UUID, err)
	}

	return nil
}

func containsEnabledState(info string) bool {
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "State:") && strings.Contains(line, "ENABLED") {
			return true
		}
	}

	return false
}
