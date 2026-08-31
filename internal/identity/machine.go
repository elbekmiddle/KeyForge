package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// machineIDPaths are checked in order; the first readable one wins.
var machineIDPaths = []string{
	"/etc/machine-id",
	"/var/lib/dbus/machine-id",
}

// Fingerprint returns a one-way hash of the local machine-id.
//
// The raw machine-id is never persisted or transmitted (see doc section 5,
// 31): only this derived fingerprint is used to let the backend recognize
// "this is the same device" across re-installs, without exposing the
// underlying Linux machine identity.
func Fingerprint() (string, error) {
	raw, err := readMachineID()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte("keyforge:" + raw))

	return hex.EncodeToString(sum[:]), nil
}

func readMachineID() (string, error) {
	var lastErr error

	for _, path := range machineIDPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			lastErr = err
			continue
		}

		id := strings.TrimSpace(string(data))
		if id == "" {
			lastErr = fmt.Errorf("identity: %s is empty", path)
			continue
		}

		return id, nil
	}

	return "", fmt.Errorf("identity: failed to read machine id: %w", lastErr)
}
