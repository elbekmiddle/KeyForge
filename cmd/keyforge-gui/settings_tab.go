package main

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/elbekmiddle/KeyForge/internal/extension"
	"github.com/elbekmiddle/KeyForge/internal/sync"
)

func buildSettingsTab(state *appState) fyne.CanvasObject {
	urlEntry := widget.NewEntry()
	urlEntry.SetText(sync.BaseURL())

	urlFeedback := widget.NewLabel("")

	saveURLBtn := widget.NewButton("Save backend URL", func() {
		sync.SetBaseURLOverride(urlEntry.Text)
		urlFeedback.SetText("Saved for this session. (KEYFORGE_BACKEND_URL env var still works too.)")
	})

	extFeedback := widget.NewLabel("")

	checkExtBtn := widget.NewButton("Check GNOME extension", func() {
		if runtime.GOOS != "linux" {
			extFeedback.SetText("Not applicable on " + runtime.GOOS + ".")
			return
		}

		extFeedback.SetText("Checking...")

		go func() {
			if err := extension.EnsureInstalled(state.log); err != nil {
				extFeedback.SetText("Failed: " + err.Error())
				return
			}

			if err := extension.Ping(); err != nil {
				extFeedback.SetText(fmt.Sprintf(
					"Installed & enabled, but not reachable yet (%v) — GNOME may need a logout/login.", err,
				))
				return
			}

			extFeedback.SetText("Installed, enabled, and reachable.")
		}()
	})

	return container.NewVBox(
		widget.NewLabel("Backend URL"),
		urlEntry,
		saveURLBtn,
		urlFeedback,
		widget.NewSeparator(),
		widget.NewLabel("GNOME Shell extension (Linux only)"),
		checkExtBtn,
		extFeedback,
	)
}
