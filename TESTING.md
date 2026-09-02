# KeyForge — real-environment testing checklist

Everything below was written/verified inside a headless Linux sandbox
(no display, no Docker daemon, restricted network egress). This
checklist is what to actually run on real machines to confirm it all
works end to end. Check items off in order — each section depends on
the one before it.

## 0. What's already verified vs. what isn't

Verified in the sandbox:
- `go build ./...`, `go vet ./...`, `gofmt -l .` all clean (Linux, native)
- CLI (`keyforge`) and GUI (`keyforge-gui`) both cross-compile cleanly
  for `GOOS=linux` and `GOOS=windows` (see `dist/` — all four binaries
  are real, verified with `file`, not just "should work")
- Backend (`keyforge-backend`): full auth/device/profile-sync/realtime
  flow tested end to end with `curl` + a Node ws client against a real
  Postgres instance; TypeORM migration tested against a fresh DB twice
  (source and compiled)

NOT verified (needs a real environment — no display, no GNOME Shell, no
physical keyboard/mouse, no Docker daemon here):
- The GUI actually opening a window and being clickable
- `keyforge run` doing real keyboard/mouse remapping
- The GNOME extension actually installing/enabling/responding inside a
  real GNOME Shell session
- `docker compose up` end to end (Dockerfile/compose commands were
  each verified individually, but never run through Compose itself)
- The Windows GUI/CLI actually launching on real Windows (cross-compiled
  and confirmed to be a valid PE32+ binary, nothing more)

## 1. Backend — Docker Compose (Linux or Windows host with Docker)

- [ ] `cd keyforge-backend && cp .env.example .env`, fill in real
      `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` / `DATABASE_PASSWORD`
- [ ] `docker compose up --build`
- [ ] `migrate` service exits 0, `db` reports healthy, `backend` stays up
- [ ] `curl -X POST localhost:3000/auth/register -d '{"email":"a@b.com","password":"testpass123"}' -H 'Content-Type: application/json'`
      returns `accessToken`/`refreshToken`
- [ ] Restart just the `backend` container — data survives (Postgres
      volume persists)

## 2. Linux client — CLI (`dist/keyforge-linux-amd64`)

- [ ] `sudo usermod -aG input $USER`, log out/in (or use the udev rule
      from the earlier setup so you don't need `sudo` every run)
- [ ] `./keyforge devices` lists your real keyboard/mouse under
      `/dev/input/eventN`
- [ ] `./keyforge extension` — installs and enables the GNOME extension;
      if GNOME says it needs a reload, do X11 `Alt+F2 → r` or, on
      Wayland, log out/in, then re-run
- [ ] `./keyforge active-app` reports the actually-focused window
      (switch focus between two apps and re-run to confirm it changes)
- [ ] `KEYFORGE_BACKEND_URL=http://localhost:3000 ./keyforge register`
      — creates an account, confirm `session.json` appears under
      `~/.local/share/keyforge/<GUID>/` with `-rw-------` permissions
- [ ] `./keyforge sync` — pushes local profiles, confirm via
      `GET /profiles?deviceGuid=...` on the backend that they landed
- [ ] `sudo ./keyforge run /dev/input/eventN` — **on a spare/test
      keyboard, not your main one, in case a mapping goes wrong** —
      confirm `KEY_A` actually types `KEY_B` (the default test mapping)
- [ ] While `run` is active, switch focus to a different app and
      confirm the "active application changed" log line appears
- [ ] While `run` is active, edit a profile via the backend (or push a
      newer version from another device) and confirm the real-time
      websocket triggers a re-pull without restarting the daemon
- [ ] Unplug the network — confirm remapping keeps working (doc section
      28: local-first)

## 3. Linux client — GUI (`dist/keyforge-gui-linux-amd64`)

- [ ] Launches, shows a window with 4 tabs (Status/Account/Profiles/Settings)
- [ ] Status tab shows the real device GUID
- [ ] Account tab: register/login work, password field is masked,
      "Signed in (user ...)" appears after success
- [ ] Settings tab: changing the backend URL and clicking "Check GNOME
      extension" actually installs/enables it, same as the CLI command
- [ ] Profiles tab lists local profiles, "Sync now" round-trips with
      the backend same as `keyforge sync`
- [ ] Status tab "Start" actually starts remapping (same caveat as
      above — test keyboard), log lines stream into the text view live
- [ ] "Stop" cleanly stops it, "Start" again works a second time
      without restarting the app
- [ ] Closing the window while running stops the daemon cleanly (no
      orphaned process left holding `/dev/uinput`)

## 4. Windows client (`dist/keyforge-gui-windows-amd64.exe`, `dist/keyforge-windows-amd64.exe`)

- [ ] `.exe` launches without missing-DLL errors (Fyne on Windows
      should be statically linked enough not to need a separate GLFW/GL
      DLL, but confirm on a clean machine, not one with dev tools
      installed)
- [ ] Window renders, all 4 tabs work
- [ ] Account tab: register/login against the same backend works
      identically to Linux (this is pure HTTP/WS, no OS-specific code)
- [ ] Status tab "Start" — **expected to fail with a clear error**
      ("remapping is not implemented on windows yet"). This is correct,
      documented behavior, not a bug — see the package doc comment at
      the top of `cmd/keyforge-gui/main.go`
- [ ] Confirm the CLI's `keyforge-windows-amd64.exe identity` and
      `register`/`login`/`sync` commands work the same way from
      PowerShell/cmd

## 5. Known gaps to keep in mind while testing

- `session.json` is plain JSON with restrictive file permissions, not
  OS keyring-backed. Don't test this against a real production account
  on a shared machine.
- Sync conflict resolution is last-write-wins by timestamp only — no
  merge. Editing the same profile on two devices between syncs will
  silently drop one side's change.
- Windows has no remapping backend at all yet — GUI/CLI account and
  sync features work, keyboard/mouse remapping does not.
- The GNOME extension's `Ping()`/`GetActiveApplication()` D-Bus calls
  go through `gdbus` as a subprocess (not a native D-Bus library) —
  confirm `gdbus` is actually on `$PATH` on whatever distro you test
  (it ships with GLib/GNOME by default, but minimal installs may lack it).

## Reproducing the cross-compiles yourself

```bash
# Linux (native, no extra setup)
go build -o keyforge ./cmd/keyforge
go build -o keyforge-gui ./cmd/keyforge-gui

# Windows cross-compile from Linux (needs gcc-mingw-w64-x86-64)
sudo apt install gcc-mingw-w64-x86-64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -o keyforge.exe ./cmd/keyforge          # CLI: no CGO needed

CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H windowsgui" -o keyforge-gui.exe ./cmd/keyforge-gui
# ^ this one is slow (5-10+ min) the first time — it's compiling Fyne's
#   OpenGL/GLFW C bindings via cgo. Subsequent builds reuse Go's build
#   cache and are much faster.
```

The `replace` directives in `go.mod` route `fyne.io/*`, `golang.org/x/*`,
and `gopkg.in/*` through their official GitHub mirrors — only needed if
your network can't reach the vanity import domains directly (e.g. this
sandbox). On a normal machine with unrestricted network, these
`replace` lines are harmless but unnecessary.
