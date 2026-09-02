package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elbekmiddle/KeyForge/internal/config"
)

// Session holds the tokens issued by POST /auth/login (doc section 7).
//
// This is stored as plain JSON at
// ~/.local/share/keyforge/<guid>/session.json, mode 0600. That is a
// deliberate, documented gap against doc section 31 ("Access token:
// plain config ❌") — the doc itself defers the real fix ("Sensitive
// credential'lar uchun keyingi bosqichda OS secure storage... tanlaymiz").
// Until that lands (e.g. via the Secret Service / libsecret on Linux),
// this file is the weak point: anything that can read this user's home
// directory can read these tokens. The password itself is never stored
// here or anywhere on the client (doc section 6).
type Session struct {
	UserID       string `json:"userId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func sessionPath(guid string) (string, error) {
	dir, err := config.DeviceDir(guid)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "session.json"), nil
}

// LoadSession reads the local session, if one exists.
func LoadSession(guid string) (Session, bool) {
	var s Session

	path, err := sessionPath(guid)
	if err != nil {
		return s, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return s, false
	}

	if err := json.Unmarshal(data, &s); err != nil {
		return s, false
	}

	return s, s.AccessToken != ""
}

// SaveSession persists the session (mode 0600 — owner read/write only).
func SaveSession(guid string, s Session) error {
	path, err := sessionPath(guid)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("sync: failed to encode session: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("sync: failed to write session: %w", err)
	}

	return nil
}

// ClearSession removes the local session (logout).
func ClearSession(guid string) error {
	path, err := sessionPath(guid)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("sync: failed to clear session: %w", err)
	}

	return nil
}
