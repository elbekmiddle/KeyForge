// Package extension manages the KeyForge GNOME Shell extension: it is
// embedded into the KeyForge binary, installed under the user's
// gnome-shell extensions directory on first run, and kept enabled and
// up to date on every subsequent run — the user never runs
// `cp`/`mkdir`/`gnome-extensions install` by hand.
package extension

import (
	"embed"
	"encoding/json"
	"fmt"
)

// UUID is the GNOME Shell extension identifier. It is a fixed identity —
// see doc section 1 — distinct from any per-device or per-user identity.
const UUID = "keyforge@elbekmiddle"

//go:embed keyforge@elbekmiddle/extension.js keyforge@elbekmiddle/metadata.json
var assets embed.FS

// assetDir is the directory name inside the embedded FS.
const assetDir = "keyforge@elbekmiddle"

type metadata struct {
	UUID    string `json:"uuid"`
	Version int    `json:"version"`
}

// EmbeddedVersion returns the "version" field from the embedded
// metadata.json — the version KeyForge ships with, as opposed to whatever
// is currently installed on disk.
func EmbeddedVersion() (int, error) {
	data, err := assets.ReadFile(assetDir + "/metadata.json")
	if err != nil {
		return 0, fmt.Errorf("extension: failed to read embedded metadata.json: %w", err)
	}

	var meta metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return 0, fmt.Errorf("extension: failed to parse embedded metadata.json: %w", err)
	}

	return meta.Version, nil
}

// files returns the embedded extension's file names, e.g.
// ["extension.js", "metadata.json"].
func files() ([]string, error) {
	entries, err := assets.ReadDir(assetDir)
	if err != nil {
		return nil, fmt.Errorf("extension: failed to list embedded assets: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		names = append(names, entry.Name())
	}

	return names, nil
}

func readAsset(name string) ([]byte, error) {
	data, err := assets.ReadFile(assetDir + "/" + name)
	if err != nil {
		return nil, fmt.Errorf("extension: failed to read embedded %s: %w", name, err)
	}

	return data, nil
}
