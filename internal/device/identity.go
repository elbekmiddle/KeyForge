package device

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elbekmiddle/KeyForge/internal/config"
	"github.com/elbekmiddle/KeyForge/internal/identity"
)

// Identity is the bootstrapped local device identity: a stable, public
// GUID plus whatever user-facing metadata is attached to it. It carries
// nothing sensitive — see internal/identity for the machine fingerprint,
// which is derived, not stored.
type Identity struct {
	GUID     string
	Metadata config.Metadata
	Device   config.DeviceInfo
	// Created is true when this identity was generated during this call,
	// i.e. first launch on this machine.
	Created bool
}

// currentPointerName is a small marker file at the KeyForge data root that
// remembers which device directory is "this machine", so subsequent runs
// don't need to guess between multiple device dirs.
const currentPointerName = "current"

// Bootstrap loads the existing local device identity, or creates one on
// first launch: generates a collision-safe GUID, creates the on-disk
// layout, and writes keyforge.json + device.json.
func Bootstrap() (Identity, error) {
	root, err := config.DataRoot()
	if err != nil {
		return Identity{}, err
	}

	if err := os.MkdirAll(root, 0o700); err != nil {
		return Identity{}, fmt.Errorf("device: failed to create data root: %w", err)
	}

	if guid, ok := readCurrentPointer(root); ok && config.Exists(guid) {
		return load(guid, false)
	}

	guid, err := identity.Generate(identity.ExistsLocally(root))
	if err != nil {
		return Identity{}, err
	}

	if err := config.EnsureDirs(guid); err != nil {
		return Identity{}, err
	}

	meta := config.Metadata{Version: config.SchemaVersion, GUID: guid}
	if err := config.SaveMetadata(guid, meta); err != nil {
		return Identity{}, err
	}

	deviceInfo := config.DeviceInfo{Name: defaultDeviceName(), Registered: false}
	if err := config.SaveDeviceInfo(guid, deviceInfo); err != nil {
		return Identity{}, err
	}

	if err := writeCurrentPointer(root, guid); err != nil {
		return Identity{}, err
	}

	return Identity{GUID: guid, Metadata: meta, Device: deviceInfo, Created: true}, nil
}

func load(guid string, created bool) (Identity, error) {
	meta, err := config.LoadMetadata(guid)
	if err != nil {
		return Identity{}, err
	}

	deviceInfo, err := config.LoadDeviceInfo(guid)
	if err != nil {
		return Identity{}, err
	}

	return Identity{GUID: guid, Metadata: meta, Device: deviceInfo, Created: created}, nil
}

func readCurrentPointer(root string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(root, currentPointerName))
	if err != nil {
		return "", false
	}

	guid := strings.TrimSpace(string(data))

	return guid, identity.Valid(guid)
}

func writeCurrentPointer(root, guid string) error {
	if err := os.WriteFile(filepath.Join(root, currentPointerName), []byte(guid+"\n"), 0o600); err != nil {
		return fmt.Errorf("device: failed to persist current device pointer: %w", err)
	}

	return nil
}

func defaultDeviceName() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "My PC"
	}

	return hostname
}
