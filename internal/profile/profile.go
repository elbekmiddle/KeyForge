package profile

import "github.com/elbekmiddle/KeyForge/internal/mapping"

type Profile struct {
	ID     string
	Name   string
	App    string
	Device string
	Modes  []Mode
}

type Mode struct {
	ID       uint8
	Name     string
	Mappings []mapping.Mapping
}
