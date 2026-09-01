package profile

import "github.com/elbekmiddle/KeyForge/internal/mapping"

// Profile is a per-application (or per-device) remapping configuration,
// loaded from ~/.local/share/keyforge/<guid>/profiles/<id>.json (doc
// section 15).
//
// App is a case-insensitive substring match against the active window's
// class/app-id (doc section 23) — e.g. "firefox" matches "Firefox",
// "firefox.desktop", etc. Leave it empty for a fallback/default profile.
type Profile struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	App    string `json:"app"`
	Device string `json:"device,omitempty"`
	Modes  []Mode `json:"modes"`
}

type Mode struct {
	ID       uint8             `json:"id"`
	Name     string            `json:"name"`
	Mappings []mapping.Mapping `json:"mappings"`
}
