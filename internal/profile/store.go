package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elbekmiddle/KeyForge/internal/config"
)

// LoadAll reads every profiles/*.json file for the given device guid.
// A missing profiles directory is not an error — it just means no
// profiles have been created yet.
func LoadAll(guid string) ([]Profile, error) {
	dir, err := config.ProfilesDir(guid)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("profile: failed to list %s: %w", dir, err)
	}

	var profiles []Profile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		p, err := load(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}

// Save writes a profile to profiles/<id>.json, creating the profiles
// directory if needed.
func Save(guid string, p Profile) error {
	dir, err := config.ProfilesDir(guid)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("profile: failed to create profiles dir: %w", err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("profile: failed to encode %s: %w", p.ID, err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, p.ID+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("profile: failed to write %s: %w", path, err)
	}

	return nil
}

// EnsureDefault writes a starter "default" profile if none exists yet,
// so a fresh install always has something to remap with.
func EnsureDefault(guid string) (Profile, error) {
	profiles, err := LoadAll(guid)
	if err != nil {
		return Profile{}, err
	}

	for _, p := range profiles {
		if p.ID == "default" {
			return p, nil
		}
	}

	def := Profile{
		ID:   "default",
		Name: "Default",
		Modes: []Mode{
			{ID: 0, Name: "default"},
		},
	}

	if err := Save(guid, def); err != nil {
		return Profile{}, err
	}

	return def, nil
}

// ModTime returns the last-modified time of profiles/<id>.json — used as
// the "updatedAt" clock for last-write-wins sync (Phase 10). Local file
// mtime is good enough here since the daemon is the only writer.
func ModTime(guid, id string) (time.Time, error) {
	dir, err := config.ProfilesDir(guid)
	if err != nil {
		return time.Time{}, err
	}

	info, err := os.Stat(filepath.Join(dir, id+".json"))
	if err != nil {
		return time.Time{}, fmt.Errorf("profile: failed to stat %s: %w", id, err)
	}

	return info.ModTime(), nil
}

func load(path string) (Profile, error) {
	var p Profile

	data, err := os.ReadFile(path)
	if err != nil {
		return p, fmt.Errorf("profile: failed to read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("profile: failed to parse %s: %w", path, err)
	}

	if p.ID == "" {
		p.ID = strings.TrimSuffix(filepath.Base(path), ".json")
	}

	return p, nil
}
