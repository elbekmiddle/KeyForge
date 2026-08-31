package activeapp

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// ExtensionDetector reads the focused window from the KeyForge GNOME
// extension over D-Bus (com.elbekmiddle.KeyForge), instead of calling
// org.gnome.Shell.Eval directly (see doc sections 16-18: Eval-based
// introspection is unreliable/denied on recent GNOME/Wayland sessions).
type ExtensionDetector struct{}

func NewExtensionDetector() *ExtensionDetector {
	return &ExtensionDetector{}
}

func (d *ExtensionDetector) Current() (Application, error) {
	output, err := exec.Command(
		"gdbus", "call", "--session",
		"--dest", "com.elbekmiddle.KeyForge",
		"--object-path", "/com/elbekmiddle/KeyForge",
		"--method", "com.elbekmiddle.KeyForge.GetActiveApplication",
	).Output()
	if err != nil {
		return Application{}, fmt.Errorf(
			"activeapp: failed to query keyforge gnome extension: %w", err,
		)
	}

	result, err := parseGDBusResult(string(output))
	if err != nil {
		return Application{}, err
	}

	var window focusedWindow

	if err := json.Unmarshal([]byte(result), &window); err != nil {
		return Application{}, fmt.Errorf(
			"activeapp: failed to parse focused window: %w", err,
		)
	}

	if window.PID == 0 {
		return Application{}, fmt.Errorf("activeapp: no active application found")
	}

	executable, err := processExecutable(window.PID)
	if err != nil {
		return Application{}, err
	}

	name := window.Name
	if name == "" {
		name = window.AppID
	}
	if name == "" {
		name = window.Class
	}

	return Application{
		Name:       name,
		Class:      firstNonEmpty(window.AppID, window.Class),
		Executable: executable,
		PID:        window.PID,
	}, nil
}
