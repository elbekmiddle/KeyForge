package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// SchemaVersion is written to keyforge.json so future versions of KeyForge
// can detect and migrate older local layouts.
const SchemaVersion = 1

// Metadata is the minimal, stable identity record for this installation.
// It intentionally holds nothing sensitive: just the schema version and the
// public device GUID.
type Metadata struct {
	Version int    `json:"version"`
	GUID    string `json:"guid"`
}

// DeviceInfo is user-facing device configuration. Name is cosmetic and may
// change at any time; it is never used as an identity (the GUID is).
type DeviceInfo struct {
	Name       string `json:"name"`
	Registered bool   `json:"registered"`
}

// KeyboardConfig is keyboard.json (doc section 13): which physical
// keyboard to listen on.
type KeyboardConfig struct {
	Enabled bool   `json:"enabled"`
	Device  string `json:"device"`
}

// MouseConfig is mouse.json (doc section 14): which physical mice to
// listen on. Multiple devices are supported since a "keyboard" and its
// "mouse" often aren't a single USB device.
type MouseConfig struct {
	Enabled bool     `json:"enabled"`
	Devices []string `json:"devices"`
}

// LoadKeyboardConfig reads keyboard.json for the given guid. A missing
// file is not an error — it just means keyboard remapping hasn't been
// configured yet.
func LoadKeyboardConfig(guid string) (KeyboardConfig, error) {
	var cfg KeyboardConfig

	p, err := KeyboardJSONPath(guid)
	if err != nil {
		return cfg, err
	}

	if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
		return cfg, nil
	}

	if err := readJSON(p, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// SaveKeyboardConfig writes keyboard.json for the given guid.
func SaveKeyboardConfig(guid string, cfg KeyboardConfig) error {
	p, err := KeyboardJSONPath(guid)
	if err != nil {
		return err
	}

	return writeJSON(p, cfg)
}

// LoadMouseConfig reads mouse.json for the given guid. A missing file is
// not an error — it just means mouse remapping hasn't been configured yet.
func LoadMouseConfig(guid string) (MouseConfig, error) {
	var cfg MouseConfig

	p, err := MouseJSONPath(guid)
	if err != nil {
		return cfg, err
	}

	if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
		return cfg, nil
	}

	if err := readJSON(p, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// SaveMouseConfig writes mouse.json for the given guid.
func SaveMouseConfig(guid string, cfg MouseConfig) error {
	p, err := MouseJSONPath(guid)
	if err != nil {
		return err
	}

	return writeJSON(p, cfg)
}

// LoadMetadata reads keyforge.json for the given guid.
func LoadMetadata(guid string) (Metadata, error) {
	var meta Metadata

	p, err := KeyforgeJSONPath(guid)
	if err != nil {
		return meta, err
	}

	if err := readJSON(p, &meta); err != nil {
		return meta, err
	}

	return meta, nil
}

// SaveMetadata writes keyforge.json for the given guid.
func SaveMetadata(guid string, meta Metadata) error {
	p, err := KeyforgeJSONPath(guid)
	if err != nil {
		return err
	}

	return writeJSON(p, meta)
}

// LoadDeviceInfo reads device.json for the given guid.
func LoadDeviceInfo(guid string) (DeviceInfo, error) {
	var info DeviceInfo

	p, err := DeviceJSONPath(guid)
	if err != nil {
		return info, err
	}

	if err := readJSON(p, &info); err != nil {
		return info, err
	}

	return info, nil
}

// SaveDeviceInfo writes device.json for the given guid.
func SaveDeviceInfo(guid string, info DeviceInfo) error {
	p, err := DeviceJSONPath(guid)
	if err != nil {
		return err
	}

	return writeJSON(p, info)
}

// Exists reports whether keyforge.json already exists for guid, i.e.
// whether this device has already been bootstrapped once before.
func Exists(guid string) bool {
	p, err := KeyforgeJSONPath(guid)
	if err != nil {
		return false
	}

	_, statErr := os.Stat(p)

	return statErr == nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: failed to read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("config: failed to parse %s: %w", path, err)
	}

	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("config: failed to encode %s: %w", path, err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("config: failed to write %s: %w", path, err)
	}

	return nil
}

// ErrNotBootstrapped is returned by Bootstrap-aware callers when no local
// device metadata exists yet and none was requested to be created.
var ErrNotBootstrapped = errors.New("config: device not bootstrapped")
