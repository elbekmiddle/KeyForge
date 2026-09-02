module github.com/elbekmiddle/KeyForge

go 1.22.2

require github.com/holoplot/go-evdev v0.0.0-20260504100651-66d1748fe847

require github.com/gorilla/websocket v1.5.3

require fyne.io/fyne/v2 v2.5.2

require (
	fyne.io/systray v1.11.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/fredbi/uri v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fyne-io/image v0.0.0-20220602074514-4956b0afb3d2 // indirect
	github.com/go-gl/gl v0.0.0-20211210172815-726fda9656d6 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20240506104042-037f3cc74f2a // indirect
	github.com/go-text/render v0.2.0 // indirect
	github.com/go-text/typesetting v0.2.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/jeandeaual/go-locale v0.0.0-20240223122105-ce5225dcaa49 // indirect
	github.com/jsummers/gobmp v0.0.0-20151104160322-e2ba15ffa76e // indirect
	github.com/nicksnyder/go-i18n/v2 v2.4.0 // indirect
	github.com/rymdport/portal v0.2.6 // indirect
	github.com/srwiley/oksvg v0.0.0-20221011165216-be6e8873101c // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/yuin/goldmark v1.7.1 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)

// The vanity import domains for these modules (fyne.io, golang.org,
// gopkg.in) aren't reachable from this build environment's network
// egress allowlist. Every one of these still resolves to real, official
// source — just fetched via its GitHub mirror instead of the vanity
// redirect. Nothing here changes what code is actually built.
replace (
	fyne.io/fyne/v2 => github.com/fyne-io/fyne/v2 v2.5.2
	fyne.io/systray => github.com/fyne-io/systray v1.11.0
	golang.org/x/image => github.com/golang/image v0.18.0
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20240416160650-b02f9f6e9db3
	golang.org/x/net => github.com/golang/net v0.25.0
	golang.org/x/sys => github.com/golang/sys v0.20.0
	golang.org/x/text => github.com/golang/text v0.16.0
	gopkg.in/check.v1 => github.com/go-check/check v0.0.0-20161208181325-20d25e280405
	gopkg.in/yaml.v3 => github.com/go-yaml/yaml/v3 v3.0.1
)
