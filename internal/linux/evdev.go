package linux

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elbekmiddle/KeyForge/internal/device"
)

func ListInputDevices() ([]device.Device, error) {
	paths, err := filepath.Glob("/dev/input/event*")
	if err != nil {
		return nil, fmt.Errorf("failed to scan input devices: %w", err)
	}

	var devices []device.Device

	for _, path := range paths {
		name, err := getDeviceName(path)
		if err != nil {
			continue
		}

		if name == "" {
			name = "Unknown"
		}

		devices = append(devices, device.Device{
			Name: name,
			Path: path,
		})
	}

	return devices, nil
}

func getDeviceName(path string) (string, error) {
	base := filepath.Base(path)

	data, err := os.ReadFile("/proc/bus/input/devices")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	var name string
	var handlers string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "N: Name=") {
			name = strings.Trim(strings.TrimPrefix(line, "N: Name="), `"`)
		}

		if strings.HasPrefix(line, "H: Handlers=") {
			handlers = strings.TrimPrefix(line, "H: Handlers=")

			if strings.Contains(handlers, base) {
				return name, nil
			}

			name = ""
			handlers = ""
		}
	}

	return "", scanner.Err()
}
