package main

import (
	"context"
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/elbekmiddle/KeyForge/internal/daemon"
)

func buildStatusTab(state *appState) fyne.CanvasObject {
	guidLabel := widget.NewLabel(identityLine(state))

	state.statusLabel = widget.NewLabel("Stopped")

	state.startStop = widget.NewButton("Start", func() {
		toggleDaemon(state)
	})

	top := container.NewVBox(
		guidLabel,
		container.NewHBox(widget.NewLabel("Status:"), state.statusLabel),
		state.startStop,
	)

	return container.NewBorder(top, nil, nil, nil, container.NewScroll(state.logView))
}

func identityLine(state *appState) string {
	if state.identity.GUID == "" {
		return "Device: (not bootstrapped)"
	}

	return fmt.Sprintf("Device: %s (%s)", state.identity.GUID, state.identity.Device.Name)
}

func toggleDaemon(state *appState) {
	state.mu.Lock()
	running := state.daemonRunning
	state.mu.Unlock()

	if running {
		stopDaemon(state)
		return
	}

	startDaemon(state)
}

func startDaemon(state *appState) {
	if runtime.GOOS != "linux" {
		state.appendLog(fmt.Sprintf(
			"cannot start: keyboard/mouse remapping is only implemented on Linux (evdev/uinput/GNOME extension). Running on %s.\n",
			runtime.GOOS,
		))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	state.mu.Lock()
	state.daemonCancel = cancel
	state.daemonRunning = true
	state.mu.Unlock()

	state.statusLabel.SetText("Running")
	state.startStop.SetText("Stop")

	go func() {
		err := daemon.RunContext(ctx, state.log, daemon.Options{})

		state.mu.Lock()
		state.daemonRunning = false
		state.daemonCancel = nil
		state.mu.Unlock()

		// NOTE: Fyne 2.5 doesn't yet have fyne.Do() for guaranteed
		// main-thread-safe UI updates from a goroutine; these direct
		// widget calls are the pragmatic choice for this version and
		// work in practice for simple text/state updates like these.
		state.statusLabel.SetText("Stopped")
		state.startStop.SetText("Start")

		if err != nil {
			state.appendLog(fmt.Sprintf("daemon stopped with error: %v\n", err))
		}
	}()
}

func stopDaemon(state *appState) {
	state.mu.Lock()
	cancel := state.daemonCancel
	state.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}
