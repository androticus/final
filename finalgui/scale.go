package finalgui

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/fyneui"
	"github.com/go-gl/glfw/v3.4/glfw"
)

const (
	prefKeyUIScale       = "ui.scale"
	prefKeyScaleNagDone  = "ui.scale.nagDismissed"
	scaleEnvKey          = "FYNE_SCALE"
	defaultChromeWidth   = 720
	defaultChromeHeight  = 560
	anomalyMonitorMinPx  = 3000
	anomalyWindowFracMax = 0.25
	anomalyScaleMax      = float32(1.05)
)

// scalePresets are the only values View ▸ Scale… offers (plan: 0.75…2.5).
var scalePresets = []float64{0.75, 1.0, 1.5, 2.0, 2.5}

func clampScale(v float64) float64 {
	if v < scalePresets[0] {
		return scalePresets[0]
	}
	if v > scalePresets[len(scalePresets)-1] {
		return scalePresets[len(scalePresets)-1]
	}
	return v
}

func nearestScalePreset(v float64) float64 {
	best := scalePresets[0]
	bestD := math.Abs(v - best)
	for _, p := range scalePresets[1:] {
		d := math.Abs(v - p)
		if d < bestD {
			best, bestD = p, d
		}
	}
	return best
}

func formatScale(v float64) string {
	v = clampScale(v)
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func parseScaleLabel(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return clampScale(v), true
}

func prefs() fyne.Preferences {
	return fyneui.TheApp.Preferences()
}

func storedScale() (float64, bool) {
	v := prefs().FloatWithFallback(prefKeyUIScale, 0)
	if v == 0 {
		return 0, false
	}
	return clampScale(v), true
}

func setStoredScale(v float64) {
	prefs().SetFloat(prefKeyUIScale, clampScale(v))
	prefs().SetBool(prefKeyScaleNagDone, true)
}

func scaleNagDone() bool {
	return prefs().BoolWithFallback(prefKeyScaleNagDone, false)
}

func setScaleNagDone() {
	prefs().SetBool(prefKeyScaleNagDone, true)
}

// applyStoredScaleEnv sets FYNE_SCALE from Prefs when the env var is unset.
// Call after NewApp (prefs ready) and before the first NewWindow.
func applyStoredScaleEnv() {
	if os.Getenv(scaleEnvKey) != "" {
		return
	}
	v, ok := storedScale()
	if !ok {
		return
	}
	_ = os.Setenv(scaleEnvKey, formatScale(v))
}

func envScaleAuthoritative() bool {
	return os.Getenv(scaleEnvKey) != ""
}

func scalePresetLabels() []string {
	out := make([]string, len(scalePresets))
	for i, p := range scalePresets {
		out[i] = formatScale(p)
	}
	return out
}

func currentCanvasScale() float64 {
	g := FinalGui_
	if g.win == nil {
		return 1
	}
	return float64(g.win.Canvas().Scale())
}

func windowPixelWidth(w fyne.Window) int {
	if w == nil {
		return 0
	}
	c := w.Canvas()
	sz := c.Size()
	if sz.Width <= 0 {
		sz = fyne.NewSize(defaultChromeWidth, defaultChromeHeight)
	}
	x, _ := c.PixelCoordinateForPosition(fyne.NewPos(sz.Width, 0))
	if x <= 0 {
		return int(float32(sz.Width) * c.Scale())
	}
	return x
}

func primaryMonitorPixelWidth() (int, bool) {
	mon := glfw.GetPrimaryMonitor()
	if mon == nil {
		return 0, false
	}
	mode := mon.GetVideoMode()
	if mode == nil {
		return 0, false
	}
	return mode.Width, true
}

// scaleLooksAnomalous reports under-scaled UI on a large display (Pi+4K case).
func scaleLooksAnomalous(canvasScale float32, monitorW, windowPixelW int) bool {
	if canvasScale > anomalyScaleMax {
		return false
	}
	if monitorW >= anomalyMonitorMinPx {
		return true
	}
	if monitorW > 0 && windowPixelW > 0 &&
		float64(windowPixelW) < anomalyWindowFracMax*float64(monitorW) {
		return true
	}
	return false
}

func shouldInviteScaleAnomaly() bool {
	if envScaleAuthoritative() {
		return false
	}
	if scaleNagDone() {
		return false
	}
	if _, ok := storedScale(); ok {
		return false
	}
	g := FinalGui_
	if g.win == nil {
		return false
	}
	monW, ok := primaryMonitorPixelWidth()
	if !ok {
		return false
	}
	return scaleLooksAnomalous(g.win.Canvas().Scale(), monW, windowPixelWidth(g.win))
}

// maybeInviteScaleAnomaly runs on the UI thread after the app has started.
func maybeInviteScaleAnomaly() {
	if !shouldInviteScaleAnomaly() {
		return
	}
	g := FinalGui_
	if g.win == nil {
		return
	}
	dialog.ShowConfirm(
		"Display scale",
		"The UI looks small for this display. Open View ▸ Scale to enlarge it? (Takes effect after relaunch.)",
		func(change bool) {
			if change {
				openScaleDialog()
				return
			}
			setScaleNagDone()
		},
		g.win,
	)
}

// openScaleDialog is View ▸ Scale… (UI thread).
func openScaleDialog() {
	g := FinalGui_
	if g.win == nil {
		return
	}
	labels := scalePresetLabels()
	selected := formatScale(nearestScalePreset(currentCanvasScale()))
	if v, ok := storedScale(); ok {
		selected = formatScale(v)
	}
	radio := widget.NewRadioGroup(labels, nil)
	radio.SetSelected(selected)
	cur := formatScale(nearestScalePreset(currentCanvasScale()))
	hint := widget.NewLabel(fmt.Sprintf(
		"Current render scale ≈ %s. Choice applies after relaunch.\nPresets: %s … %s.",
		cur, labels[0], labels[len(labels)-1],
	))
	hint.Wrapping = fyne.TextWrapWord
	body := container.NewVBox(hint, radio)
	d := dialog.NewCustomConfirm("UI Scale", "Apply", "Cancel", body, func(ok bool) {
		if !ok {
			return
		}
		v, ok := parseScaleLabel(radio.Selected)
		if !ok {
			return
		}
		setStoredScale(v)
		dialog.ShowConfirm(
			"Relaunch required",
			fmt.Sprintf("Scale %s is saved. Quit Final now so it takes effect on the next launch?", formatScale(v)),
			func(quit bool) {
				if quit {
					requestQuit()
				}
			},
			g.win,
		)
	}, g.win)
	d.Resize(fyne.NewSize(420, 320))
	d.Show()
}

func onAppStarted() {
	installNativeQuitHook()
	maybeInviteScaleAnomaly()
}
