package uinput

import (
	"fmt"

	"github.com/holoplot/go-evdev"

	"github.com/elbekmiddle/KeyForge/internal/mouse"
)

// MouseOutput is the mouse counterpart of Output: it accepts remapped
// button events and passed-through motion/scroll events and re-emits them
// as a virtual mouse.
type MouseOutput interface {
	SendButton(event mouse.ButtonEvent) error
	SendMotion(event mouse.MotionEvent) error
	Close() error
}

type Mouse struct {
	device *evdev.InputDevice
}

func NewMouse() (*Mouse, error) {
	capabilities := map[evdev.EvType][]evdev.EvCode{
		evdev.EV_KEY: {
			evdev.BTN_LEFT,
			evdev.BTN_RIGHT,
			evdev.BTN_MIDDLE,
			evdev.BTN_SIDE,
			evdev.BTN_EXTRA,
		},
		evdev.EV_REL: {
			evdev.REL_X,
			evdev.REL_Y,
			evdev.REL_WHEEL,
			evdev.REL_HWHEEL,
		},
	}

	id := evdev.InputID{
		BusType: evdev.BUS_USB,
		Vendor:  0x1234,
		Product: 0x5679,
		Version: 1,
	}

	device, err := evdev.CreateDevice(
		"KeyForge Virtual Mouse",
		id,
		capabilities,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual mouse: %w", err)
	}

	return &Mouse{device: device}, nil
}

func (m *Mouse) SendButton(event mouse.ButtonEvent) error {
	code, ok := evdev.KEYFromString[event.Code]
	if !ok {
		return fmt.Errorf("unsupported button code: %s", event.Code)
	}

	if err := m.device.WriteOne(&evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  code,
		Value: int32(event.Value),
	}); err != nil {
		return fmt.Errorf("failed to write button event: %w", err)
	}

	return m.sync()
}

func (m *Mouse) SendMotion(event mouse.MotionEvent) error {
	code, ok := evdev.RELFromString[event.Axis]
	if !ok {
		return fmt.Errorf("unsupported motion axis: %s", event.Axis)
	}

	if err := m.device.WriteOne(&evdev.InputEvent{
		Type:  evdev.EV_REL,
		Code:  code,
		Value: event.Value,
	}); err != nil {
		return fmt.Errorf("failed to write motion event: %w", err)
	}

	return m.sync()
}

func (m *Mouse) sync() error {
	if err := m.device.WriteOne(&evdev.InputEvent{
		Type:  evdev.EV_SYN,
		Code:  evdev.SYN_REPORT,
		Value: 0,
	}); err != nil {
		return fmt.Errorf("failed to write sync event: %w", err)
	}

	return nil
}

func (m *Mouse) Close() error {
	if m.device == nil {
		return nil
	}

	if err := evdev.DestroyDevice(m.device); err != nil {
		return fmt.Errorf("failed to destroy virtual mouse: %w", err)
	}

	m.device = nil

	return nil
}
