//go:build linux

package uinput

import (
	"fmt"

	"github.com/holoplot/go-evdev"

	"github.com/elbekmiddle/KeyForge/internal/keyboard"
)

type Keyboard struct {
	device *evdev.InputDevice
}

func NewKeyboard() (*Keyboard, error) {
	capabilities := map[evdev.EvType][]evdev.EvCode{
		evdev.EV_KEY: {
			evdev.KEY_ESC,

			evdev.KEY_1,
			evdev.KEY_2,
			evdev.KEY_3,
			evdev.KEY_4,
			evdev.KEY_5,
			evdev.KEY_6,
			evdev.KEY_7,
			evdev.KEY_8,
			evdev.KEY_9,
			evdev.KEY_0,

			evdev.KEY_Q,
			evdev.KEY_W,
			evdev.KEY_E,
			evdev.KEY_R,
			evdev.KEY_T,
			evdev.KEY_Y,
			evdev.KEY_U,
			evdev.KEY_I,
			evdev.KEY_O,
			evdev.KEY_P,

			evdev.KEY_A,
			evdev.KEY_S,
			evdev.KEY_D,
			evdev.KEY_F,
			evdev.KEY_G,
			evdev.KEY_H,
			evdev.KEY_J,
			evdev.KEY_K,
			evdev.KEY_L,

			evdev.KEY_Z,
			evdev.KEY_X,
			evdev.KEY_C,
			evdev.KEY_V,
			evdev.KEY_B,
			evdev.KEY_N,
			evdev.KEY_M,

			evdev.KEY_SPACE,
			evdev.KEY_ENTER,
			evdev.KEY_TAB,
			evdev.KEY_BACKSPACE,

			evdev.KEY_LEFTCTRL,
			evdev.KEY_LEFTSHIFT,
			evdev.KEY_LEFTALT,
			evdev.KEY_LEFTMETA,

			evdev.KEY_RIGHTCTRL,
			evdev.KEY_RIGHTSHIFT,
			evdev.KEY_RIGHTALT,
			evdev.KEY_RIGHTMETA,

			evdev.KEY_UP,
			evdev.KEY_DOWN,
			evdev.KEY_LEFT,
			evdev.KEY_RIGHT,

			evdev.KEY_F1,
			evdev.KEY_F2,
			evdev.KEY_F3,
			evdev.KEY_F4,
			evdev.KEY_F5,
			evdev.KEY_F6,
			evdev.KEY_F7,
			evdev.KEY_F8,
			evdev.KEY_F9,
			evdev.KEY_F10,
			evdev.KEY_F11,
			evdev.KEY_F12,
		},
	}

	id := evdev.InputID{
		BusType: evdev.BUS_USB,
		Vendor:  0x1234,
		Product: 0x5678,
		Version: 1,
	}

	device, err := evdev.CreateDevice(
		"KeyForge Virtual Keyboard",
		id,
		capabilities,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual keyboard: %w", err)
	}

	return &Keyboard{
		device: device,
	}, nil
}

func (k *Keyboard) Send(event keyboard.KeyEvent) error {
	code, ok := evdev.KEYFromString[event.Code]
	if !ok {
		return fmt.Errorf("unsupported key code: %s", event.Code)
	}

	keyEvent := &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  code,
		Value: int32(event.Value),
	}

	if err := k.device.WriteOne(keyEvent); err != nil {
		return fmt.Errorf("failed to write key event: %w", err)
	}

	syncEvent := &evdev.InputEvent{
		Type:  evdev.EV_SYN,
		Code:  evdev.SYN_REPORT,
		Value: 0,
	}

	if err := k.device.WriteOne(syncEvent); err != nil {
		return fmt.Errorf("failed to write sync event: %w", err)
	}

	return nil
}

func (k *Keyboard) Close() error {
	if k.device == nil {
		return nil
	}

	if err := evdev.DestroyDevice(k.device); err != nil {
		return fmt.Errorf("failed to destroy virtual keyboard: %w", err)
	}

	k.device = nil

	return nil
}
