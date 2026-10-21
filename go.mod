module github.com/androticus/final

go 1.25.0

require (
	fyne.io/fyne/v2 v2.8.1 // 2.8.x only; pin via replace — do not take 2.9 until we opt in
	github.com/aquilax/go-perlin v1.1.0
	github.com/fogleman/gg v1.3.0
	github.com/gdamore/tcell/v2 v2.13.10
	github.com/go-sql-driver/mysql v1.10.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/gopxl/beep/v2 v2.1.1
	github.com/jackc/pgx/v5 v5.10.0
	github.com/rivo/tview v0.42.0
	github.com/warthog618/go-gpiocdev v0.9.1
	golang.design/x/clipboard v0.8.0
	golang.org/x/crypto v0.54.0
	golang.org/x/image v0.28.0
	golang.org/x/term v0.45.0
	gonum.org/v1/gonum v0.17.0
	modernc.org/sqlite v1.54.0
)

require (
	github.com/go-gl/glfw/v3.4/glfw v0.1.0-pre.1.0.20260707082822-2a407d02d01a
	go.bug.st/serial v1.6.2
	golang.org/x/sys v0.47.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	fyne.io/systray v1.12.3-0.20260810170012-af4e8e793ec4 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/FyshOS/fancyfs v0.0.1 // indirect
	github.com/anthonynsimon/bild v0.14.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/oto/v3 v3.3.2 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fyne-io/gl-js v0.2.1-0.20260315212741-029c47fd27e8 // indirect
	github.com/fyne-io/glfw-js v0.4.0 // indirect
	github.com/fyne-io/image v0.1.1 // indirect
	github.com/fyne-io/oksvg v0.2.0 // indirect
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276 // indirect
	github.com/go-text/render v0.2.1 // indirect
	github.com/go-text/typesetting v0.3.4 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hack-pad/go-indexeddb v0.3.2 // indirect
	github.com/hack-pad/safejs v0.1.0 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/jsummers/gobmp v0.0.0-20230614200233-a9de23ed2e25 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.5.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	github.com/rymdport/portal v0.4.2 // indirect
	github.com/srwiley/oksvg v0.0.0-20221011165216-be6e8873101c // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/yuin/goldmark v1.8.2 // indirect
	golang.design/x/x11 v0.2.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/mobile v0.0.0-20250606033058-a2a15c67f36f // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.74.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

// Fyne 2.8.x only. `go get -u` would otherwise take 2.9 (same module path fyne.io/fyne/v2).
// Bump both sides of this replace for a 2.8 patch; take 2.9 only by an explicit versioned get.
replace fyne.io/fyne/v2 => fyne.io/fyne/v2 v2.8.1
