package core

// tranche3Enabled is armed by EnableBootstrapTranche3 (fyneui package init)
// before InitGlobalExecutionEnvironment, so RunBootstrapInit compiles core3.
var tranche3Enabled = false

// EnableBootstrapTranche3 arms core3 (fyne Final sources) for InitGlobalExecutionEnvironment.
// Call from fyneui init() before that Init runs. No-op if already enabled.
func EnableBootstrapTranche3() {
	tranche3Enabled = true
}

const bootstrapTranche3Manifest = `
// Fyne type docs and Final methods (compiled only when tranche3 is enabled)
internal/final/fyneui/init_fyne.fnl
// GameSim types
internal/final/fyneui/0200-sim/sim_simgame.fnl
// Docs -- note: not order-dependent
internal/final/fyneui/1000-doc/fyne/accordion.fnl
internal/final/fyneui/1000-doc/fyne/activity.fnl
internal/final/fyneui/1000-doc/fyne/app.fnl
internal/final/fyneui/1000-doc/fyne/app_tabs.fnl
internal/final/fyneui/1000-doc/fyne/border.fnl
internal/final/fyneui/1000-doc/fyne/box.fnl
internal/final/fyneui/1000-doc/fyne/button.fnl
internal/final/fyneui/1000-doc/fyne/canvas.fnl
internal/final/fyneui/1000-doc/fyne/canvas_image.fnl
internal/final/fyneui/1000-doc/fyne/check.fnl
internal/final/fyneui/1000-doc/fyne/container.fnl
internal/final/fyneui/1000-doc/fyne/dialog.fnl
internal/final/fyneui/1000-doc/fyne/dialogs.fnl
internal/final/fyneui/1000-doc/fyne/disableable.fnl
internal/final/fyneui/1000-doc/fyne/entry.fnl
internal/final/fyneui/1000-doc/fyne/form.fnl
internal/final/fyneui/1000-doc/fyne/form_layout.fnl
internal/final/fyneui/1000-doc/fyne/grid.fnl
internal/final/fyneui/1000-doc/fyne/icon.fnl
internal/final/fyneui/1000-doc/fyne/keys.fnl
internal/final/fyneui/1000-doc/fyne/label.fnl
internal/final/fyneui/1000-doc/fyne/list.fnl
internal/final/fyneui/1000-doc/fyne/menu.fnl
internal/final/fyneui/1000-doc/fyne/menu_bar.fnl
internal/final/fyneui/1000-doc/fyne/mouse.fnl
internal/final/fyneui/1000-doc/fyne/prefs.fnl
internal/final/fyneui/1000-doc/fyne/radio_group.fnl
internal/final/fyneui/1000-doc/fyne/rich_text.fnl
internal/final/fyneui/1000-doc/fyne/scroll.fnl
internal/final/fyneui/1000-doc/fyne/select.fnl
internal/final/fyneui/1000-doc/fyne/separator.fnl
internal/final/fyneui/1000-doc/fyne/slider.fnl
internal/final/fyneui/1000-doc/fyne/split.fnl
internal/final/fyneui/1000-doc/fyne/table.fnl
internal/final/fyneui/1000-doc/fyne/tree.fnl
internal/final/fyneui/1000-doc/fyne/tree_data.fnl
internal/final/fyneui/1000-doc/fyne/window.fnl
`
