package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/elbekmiddle/KeyForge/internal/identity"
	"github.com/elbekmiddle/KeyForge/internal/profile"
	"github.com/elbekmiddle/KeyForge/internal/sync"
)

func buildProfilesTab(state *appState) fyne.CanvasObject {
	list := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {},
	)

	var current []profile.Profile

	reload := func() {
		if state.identity.GUID == "" {
			return
		}

		profiles, err := profile.LoadAll(state.identity.GUID)
		if err != nil {
			return
		}

		current = profiles
		list.Length = func() int { return len(current) }
		list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			p := current[i]
			label := p.Name
			if p.App != "" {
				label = fmt.Sprintf("%s  (app: %s)", p.Name, p.App)
			}
			o.(*widget.Label).SetText(label)
		}
		list.Refresh()
	}

	reload()

	feedback := widget.NewLabel("")

	refreshBtn := widget.NewButton("Refresh", func() {
		reload()
		feedback.SetText("")
	})

	syncBtn := widget.NewButton("Sync now", func() {
		syncProfilesNow(state, feedback, reload)
	})

	top := container.NewHBox(refreshBtn, syncBtn)

	return container.NewBorder(
		container.NewVBox(top, feedback),
		nil, nil, nil,
		list,
	)
}

func syncProfilesNow(state *appState, feedback *widget.Label, onDone func()) {
	if state.identity.GUID == "" {
		feedback.SetText("Device identity not bootstrapped.")
		return
	}

	session, ok := sync.LoadSession(state.identity.GUID)
	if !ok {
		feedback.SetText("Not signed in — see the Account tab.")
		return
	}

	feedback.SetText("Syncing...")

	go func() {
		client := sync.NewClient()

		if fp, err := identity.Fingerprint(); err == nil {
			_ = client.RegisterDevice(session.AccessToken, state.identity.GUID, fp, state.identity.Device.Name)
		}

		local, err := profile.LoadAll(state.identity.GUID)
		if err != nil {
			feedback.SetText("Failed to read local profiles: " + err.Error())
			return
		}

		accepted, skipped, remote, err := client.PushProfiles(session.AccessToken, state.identity.GUID, local)
		if err != nil {
			feedback.SetText("Sync failed: " + err.Error())
			return
		}

		for _, p := range remote {
			_ = profile.Save(state.identity.GUID, p)
		}

		feedback.SetText(fmt.Sprintf(
			"Synced: %d accepted, %d skipped, %d total on server.",
			len(accepted), len(skipped), len(remote),
		))

		onDone()
	}()
}
