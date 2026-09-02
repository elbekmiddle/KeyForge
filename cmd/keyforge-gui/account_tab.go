package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/elbekmiddle/KeyForge/internal/identity"
	"github.com/elbekmiddle/KeyForge/internal/sync"
)

func buildAccountTab(state *appState) fyne.CanvasObject {
	state.sessionLabel = widget.NewLabel("Not signed in")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("email@example.com")

	// Masked input — an improvement over the CLI's plain-text prompt
	// (documented there as a known gap); the password is still never
	// written to disk either way (doc section 6).
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("password")

	feedback := widget.NewLabel("")

	loginBtn := widget.NewButton("Login", func() {
		authenticate(state, "login", emailEntry.Text, passwordEntry.Text, feedback)
	})

	registerBtn := widget.NewButton("Create account", func() {
		authenticate(state, "register", emailEntry.Text, passwordEntry.Text, feedback)
	})

	logoutBtn := widget.NewButton("Logout", func() {
		if state.identity.GUID == "" {
			return
		}

		if err := sync.ClearSession(state.identity.GUID); err != nil {
			feedback.SetText("Logout failed: " + err.Error())
			return
		}

		feedback.SetText("Logged out.")
		refreshSessionLabel(state)
	})

	return container.NewVBox(
		state.sessionLabel,
		widget.NewForm(
			widget.NewFormItem("Email", emailEntry),
			widget.NewFormItem("Password", passwordEntry),
		),
		container.NewHBox(loginBtn, registerBtn, logoutBtn),
		feedback,
	)
}

func authenticate(state *appState, mode, email, password string, feedback *widget.Label) {
	if state.identity.GUID == "" {
		feedback.SetText("Device identity not bootstrapped, cannot save session.")
		return
	}

	feedback.SetText("Working...")

	go func() {
		client := sync.NewClient()

		var (
			session sync.Session
			err     error
		)

		if mode == "register" {
			session, err = client.Register(email, password)
		} else {
			session, err = client.Login(email, password)
		}

		if err != nil {
			feedback.SetText(fmt.Sprintf("%s failed: %v", mode, err))
			return
		}

		if err := sync.SaveSession(state.identity.GUID, session); err != nil {
			feedback.SetText("Signed in, but failed to save session locally: " + err.Error())
			return
		}

		if fp, err := identity.Fingerprint(); err == nil {
			_ = client.RegisterDevice(session.AccessToken, state.identity.GUID, fp, state.identity.Device.Name)
		}

		feedback.SetText("Success.")
		refreshSessionLabel(state)
	}()
}
