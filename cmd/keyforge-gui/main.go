// Command keyforge-gui is the cross-platform desktop shell for KeyForge:
// account (register/login/logout), daemon start/stop with a live log
// view, local profile browsing, and backend URL settings.
//
// Built with Fyne (fyne.io/fyne/v2), which targets Linux, Windows, and
// macOS from one codebase. IMPORTANT PLATFORM NOTE: this window and
// everything in it (auth, sync, settings) runs the same on every
// platform. The actual key/mouse REMAPPING ENGINE underneath
// (internal/keyboard, internal/mouse, internal/uinput, the GNOME
// extension) is Linux-only right now — it's built on evdev/uinput and a
// GNOME Shell D-Bus extension, neither of which exist on Windows. A
// Windows remapping backend (a low-level keyboard hook plus a virtual
// input driver) is a separate, substantial piece of work that has not
// been started. On Windows this GUI will run, and account/sync
// screens work, but the Status tab's Start button will report that
// remapping isn't available on this OS yet.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/elbekmiddle/KeyForge/internal/device"
	kflog "github.com/elbekmiddle/KeyForge/internal/logger"
	kfsync "github.com/elbekmiddle/KeyForge/internal/sync"
)

// appState is the GUI's shared, mutable state. Every field that's
// touched from more than one goroutine (daemon goroutine vs. UI click
// handlers) is guarded by mu.
type appState struct {
	mu sync.Mutex

	log     *slog.Logger
	logView *widget.Entry
	logLock sync.Mutex // serializes appends to logView's text

	identity device.Identity

	daemonCancel  context.CancelFunc
	daemonRunning bool

	statusLabel *widget.Label
	startStop   *widget.Button

	sessionLabel *widget.Label
}

func (s *appState) appendLog(line string) {
	s.logLock.Lock()
	defer s.logLock.Unlock()

	if s.logView == nil {
		return
	}

	text := s.logView.Text + line
	// Cap the buffer so a long-running daemon doesn't grow this
	// unbounded — keep roughly the last ~500 lines.
	const maxLines = 500
	if lines := countLines(text); lines > maxLines {
		text = trimToLastLines(text, maxLines)
	}

	s.logView.SetText(text)
	s.logView.CursorRow = countLines(text)
}

func main() {
	a := app.NewWithID("com.elbekmiddle.keyforge")
	w := a.NewWindow("KeyForge")
	w.Resize(fyne.NewSize(720, 480))

	state := &appState{}

	logView := widget.NewMultiLineEntry()
	logView.Wrapping = fyne.TextWrapWord
	logView.Disable() // read-only log view
	state.logView = logView

	state.log = kflog.NewWithWriter(newLineWriter(state.appendLog))

	id, err := device.Bootstrap()
	if err != nil {
		state.appendLog(fmt.Sprintf("failed to bootstrap device identity: %v\n", err))
	} else {
		state.identity = id
	}

	statusTab := buildStatusTab(state)
	accountTab := buildAccountTab(state)
	profilesTab := buildProfilesTab(state)
	settingsTab := buildSettingsTab(state)

	tabs := container.NewAppTabs(
		container.NewTabItem("Status", statusTab),
		container.NewTabItem("Account", accountTab),
		container.NewTabItem("Profiles", profilesTab),
		container.NewTabItem("Settings", settingsTab),
	)

	if runtime.GOOS != "linux" {
		state.appendLog(fmt.Sprintf(
			"note: remapping engine is Linux-only today; running on %s — account/sync screens work, Start will refuse.\n",
			runtime.GOOS,
		))
	}

	refreshSessionLabel(state)

	w.SetContent(tabs)
	w.SetOnClosed(func() {
		state.mu.Lock()
		cancel := state.daemonCancel
		state.mu.Unlock()

		if cancel != nil {
			cancel()
		}
	})

	w.ShowAndRun()
}

func refreshSessionLabel(state *appState) {
	if state.sessionLabel == nil || state.identity.GUID == "" {
		return
	}

	if session, ok := kfsync.LoadSession(state.identity.GUID); ok {
		state.sessionLabel.SetText("Signed in (user " + shorten(session.UserID) + ")")
	} else {
		state.sessionLabel.SetText("Not signed in")
	}
}

func shorten(id string) string {
	if len(id) <= 8 {
		return id
	}

	return id[:8]
}
