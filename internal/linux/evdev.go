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
		info, err := inspectInputDevice(path)
		if err != nil {
			continue
		}

		devices = append(devices, *info)
	}

	return devices, nil
}

func inspectInputDevice(path string) (*device.Device, error) {
	data, err := os.ReadFile("/proc/bus/input/devices")
	if err != nil {
		return nil, fmt.Errorf("failed to read input devices: %w", err)
	}

	base := filepath.Base(path)

	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	var name string
	var bus string
	var vendorID string
	var productID string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "I: Bus="):
			bus, vendorID, productID = parseDeviceID(line)

		case strings.HasPrefix(line, "N: Name="):
			name = strings.Trim(strings.TrimPrefix(line, "N: Name="), `"`)

		case strings.HasPrefix(line, "H: Handlers="):
			handlers := strings.TrimPrefix(line, "H: Handlers=")

			if handlersContains(handlers, base) == base {
				return &device.Device{
					Name:         name,
					Type:         detectDeviceType(handlers, name),
					Path:         path,
					VendorID:     vendorID,
					ProductID:    productID,
					Bus:          bus,
					Manufacturer: manufacturerFromName(name),
				}, nil
			}

			name = ""
			bus = ""
			vendorID = ""
			productID = ""
		}
	}

	return nil, scanner.Err()
}

func parseDeviceID(line string) (string, string, string) {
	fields := strings.Fields(line)

	if len(fields) < 4 {
		return "", "", ""
	}

	bus := strings.TrimPrefix(fields[1], "Bus=")
	vendor := strings.TrimPrefix(fields[2], "Vendor=")
	product := strings.TrimPrefix(fields[3], "Product=")

	return bus, vendor, product
}

func handlersContains(handlers, target string) string {
	for _, handler := range strings.Fields(handlers) {
		if handler == target {
			return target
		}
	}

	return ""
}

func detectDeviceType(handlers, name string) device.Type {
	nameLower := strings.ToLower(name)
	handlersLower := strings.ToLower(handlers)

	if strings.Contains(nameLower, "keyboard") ||
		strings.Contains(handlersLower, "kbd") {
		return device.TypeKeyboard
	}

	if strings.Contains(nameLower, "mouse") ||
		strings.Contains(nameLower, "touchpad") {
		return device.TypeMouse
	}

	return device.TypeUnknown
}

func manufacturerFromName(name string) string {
	parts := strings.Fields(name)

	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}
