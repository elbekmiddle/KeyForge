package activeapp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type LinuxDetector struct{}

func NewLinuxDetector() *LinuxDetector {
	return &LinuxDetector{}
}

type focusedWindow struct {
	Name  string `json:"name"`
	Class string `json:"class"`
	AppID string `json:"app_id"`
	PID   int    `json:"pid"`
}

func (d *LinuxDetector) Current() (Application, error) {
	script := `
		(() => {
			const window = global.display.get_focus_window();

			if (!window) {
				return JSON.stringify({});
			}

			let appID = "";

			try {
				appID = window.get_gtk_application_id() || "";
			} catch (e) {}

			return JSON.stringify({
				name: window.get_title() || "",
				class: window.get_wm_class() || "",
				app_id: appID,
				pid: window.get_pid() || 0
			});
		})()
	`

	cmd := exec.Command(
		"gdbus",
		"call",
		"--session",
		"--dest",
		"org.gnome.Shell",
		"--object-path",
		"/org/gnome/Shell",
		"--method",
		"org.gnome.Shell.Eval",
		script,
	)

	output, err := cmd.Output()
	if err != nil {
		return Application{}, fmt.Errorf("failed to query GNOME Shell: %w", err)
	}

	result, err := parseGDBusResult(string(output))
	if err != nil {
		return Application{}, err
	}

	var window focusedWindow

	if err := json.Unmarshal([]byte(result), &window); err != nil {
		return Application{}, fmt.Errorf(
			"failed to parse focused window: %w",
			err,
		)
	}

	if window.PID == 0 {
		return Application{}, fmt.Errorf("no active application found")
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

func parseGDBusResult(output string) (string, error) {
	output = strings.TrimSpace(output)

	start := strings.Index(output, "'")
	end := strings.LastIndex(output, "'")

	if start == -1 || end <= start {
		return "", fmt.Errorf(
			"invalid gdbus response: %s",
			output,
		)
	}

	value := output[start+1 : end]

	value = strings.ReplaceAll(value, "\\'", "'")
	value = strings.ReplaceAll(value, "\\\"", "\"")
	value = strings.ReplaceAll(value, "\\\\", "\\")

	return value, nil
}

func processExecutable(pid int) (string, error) {
	exePath := filepath.Join(
		"/proc",
		strconv.Itoa(pid),
		"exe",
	)

	executable, err := os.Readlink(exePath)
	if err != nil {
		return "", fmt.Errorf(
			"failed to resolve executable for pid %d: %w",
			pid,
			err,
		)
	}

	return executable, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
