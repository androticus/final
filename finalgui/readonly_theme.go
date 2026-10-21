package finalgui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// readOnlyTheme remaps ColorNameDisabled to a darker/readable gray so disabled
// Entries (response / history) stay selectable for copy but look non-editable.
// It also pins idle scrollbar thickness to half of SizeNameScrollBar so the
// overlay gutter stays at the expanded width (aimable) instead of collapsing
// to SizeNameScrollBarSmall (3px) until hover.
type readOnlyTheme struct {
	base fyne.Theme
}

func newReadOnlyTheme() fyne.Theme {
	base := fyne.CurrentApp().Settings().Theme()
	if base == nil {
		base = theme.DefaultTheme()
	}
	return &readOnlyTheme{base: base}
}

func (t *readOnlyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameDisabled {
		if v == theme.VariantDark {
			// muted light gray on dark backgrounds
			return color.NRGBA{R: 0xb8, G: 0xb8, B: 0xb8, A: 0xff}
		}
		// dark gray on light backgrounds (not black — signals read-only)
		return color.NRGBA{R: 0x3d, G: 0x3d, B: 0x3d, A: 0xff}
	}
	return t.base.Color(n, v)
}

func (t *readOnlyTheme) Font(s fyne.TextStyle) fyne.Resource { return t.base.Font(s) }
func (t *readOnlyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(n)
}
func (t *readOnlyTheme) Size(n fyne.ThemeSizeName) float32 {
	// Fyne MinSize of the overlay is ScrollBarSmall*2 when idle, ScrollBar when
	// hovered. Matching those keeps the strip at full width so the target does
	// not jump. The thumb still thickens slightly on hover.
	if n == theme.SizeNameScrollBarSmall {
		return t.base.Size(theme.SizeNameScrollBar) / 2
	}
	return t.base.Size(n)
}
