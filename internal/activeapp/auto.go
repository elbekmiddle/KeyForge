package activeapp

import "github.com/elbekmiddle/KeyForge/internal/extension"

// AutoDetector prefers the KeyForge GNOME extension's D-Bus service and
// falls back to the direct org.gnome.Shell.Eval query when the extension
// isn't installed/enabled/reachable yet (e.g. right after first launch,
// before GNOME has picked it up).
type AutoDetector struct {
	primary  Detector
	fallback Detector
}

func NewAutoDetector() *AutoDetector {
	return &AutoDetector{
		primary:  NewExtensionDetector(),
		fallback: NewLinuxDetector(),
	}
}

func (d *AutoDetector) Current() (Application, error) {
	if err := extension.Ping(); err == nil {
		if app, err := d.primary.Current(); err == nil {
			return app, nil
		}
	}

	return d.fallback.Current()
}
