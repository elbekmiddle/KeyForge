// Package config owns KeyForge's local-first storage layout:
//
//	~/.local/share/keyforge/
//	└── KF-7A3M9X/
//	    ├── keyforge.json   (minimal metadata: version, guid)
//	    ├── device.json     (user-facing device name, registration state)
//	    ├── config.json     (general runtime configuration)
//	    ├── keyboard.json   (keyboard mappings)
//	    ├── mouse.json      (mouse mappings)
//	    └── profiles/
//	        ├── default.json
//	        └── ...
//
// Nothing sensitive lives here: no password, no email, no access token, no
// plain machine-id (see doc section 31). Those belong to the backend and,
// once introduced, OS secure storage.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DataRoot returns ~/.local/share/keyforge, honoring XDG_DATA_HOME.
func DataRoot() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "keyforge"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: failed to resolve home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share", "keyforge"), nil
}

// DeviceDir returns ~/.local/share/keyforge/<guid>.
func DeviceDir(guid string) (string, error) {
	root, err := DataRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, guid), nil
}

// ProfilesDir returns ~/.local/share/keyforge/<guid>/profiles.
func ProfilesDir(guid string) (string, error) {
	deviceDir, err := DeviceDir(guid)
	if err != nil {
		return "", err
	}

	return filepath.Join(deviceDir, "profiles"), nil
}

func path(guid, file string) (string, error) {
	deviceDir, err := DeviceDir(guid)
	if err != nil {
		return "", err
	}

	return filepath.Join(deviceDir, file), nil
}

func KeyforgeJSONPath(guid string) (string, error) { return path(guid, "keyforge.json") }
func DeviceJSONPath(guid string) (string, error)   { return path(guid, "device.json") }
func ConfigJSONPath(guid string) (string, error)   { return path(guid, "config.json") }
func KeyboardJSONPath(guid string) (string, error) { return path(guid, "keyboard.json") }
func MouseJSONPath(guid string) (string, error)    { return path(guid, "mouse.json") }

// EnsureDirs creates the device directory and its profiles subdirectory.
func EnsureDirs(guid string) error {
	profilesDir, err := ProfilesDir(guid)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(profilesDir, 0o700); err != nil {
		return fmt.Errorf("config: failed to create local storage: %w", err)
	}

	return nil
}
