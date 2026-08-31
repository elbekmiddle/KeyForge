// Package identity generates and resolves the local, user-facing KeyForge
// device identifier (e.g. "KF-7A3M9X").
//
// This GUID is intentionally NOT a security credential. It is a short,
// human-shareable label for "this installation of KeyForge on this
// machine". Sensitive identity (machine-id, tokens, passwords) never
// touches this package — see internal/identity/machine.go.
package identity

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

const (
	// Prefix is prepended to every generated device GUID.
	Prefix = "KF-"

	// alphabet used for GUID bodies: A-Z0-9, 36^6 ≈ 2.17 billion values.
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// bodyLength is the number of characters generated after Prefix.
	bodyLength = 6

	// maxAttempts bounds collision-retry generation so we never loop forever.
	maxAttempts = 32
)

// Generate creates a new, locally-unique device GUID such as "KF-7A3M9X".
//
// existing should report whether a candidate GUID is already taken (e.g. a
// directory for it already exists under ~/.local/share/keyforge/). If nil,
// no collision check is performed.
func Generate(existing func(guid string) bool) (string, error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		candidate, err := generateOnce()
		if err != nil {
			return "", err
		}

		if existing == nil || !existing(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("identity: exhausted %d attempts generating a unique device guid", maxAttempts)
}

func generateOnce() (string, error) {
	body := make([]byte, bodyLength)

	for i := range body {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("identity: failed to generate random guid: %w", err)
		}

		body[i] = alphabet[n.Int64()]
	}

	return Prefix + string(body), nil
}

// Valid reports whether guid looks like a well-formed KeyForge device GUID.
func Valid(guid string) bool {
	if len(guid) != len(Prefix)+bodyLength {
		return false
	}

	if guid[:len(Prefix)] != Prefix {
		return false
	}

	for _, c := range guid[len(Prefix):] {
		if !containsRune(alphabet, c) {
			return false
		}
	}

	return true
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}

	return false
}

// ExistsLocally returns an "existing" predicate for Generate that checks
// whether a device directory already exists under the given KeyForge
// data root (~/.local/share/keyforge).
func ExistsLocally(dataRoot string) func(guid string) bool {
	return func(guid string) bool {
		_, err := os.Stat(filepath.Join(dataRoot, guid))
		return err == nil
	}
}
