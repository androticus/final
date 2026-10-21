package finaltui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/androticus/final/finalembed"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	pageFields       = "fields"
	narrowFieldWidth = 14
)

type fieldCtrl struct {
	spec  FieldSpec
	input *tview.InputField // String / Int / Float
	check *tview.Checkbox   // Bool
}

func (ui *appUI) buildFieldsPage() tview.Primitive {
	ui.fieldsInstr = tview.NewTextView().SetWrap(true).SetWordWrap(true)
	ui.fieldsError = tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetWrap(true)
	ui.fieldsForm = tview.NewForm()

	ui.fieldsOk = tview.NewButton("Ok").
		SetSelectedFunc(ui.acceptFields)
	ui.fieldsCancel = tview.NewButton("Cancel").
		SetSelectedFunc(func() { ui.finishFields(BtnCancel, nil) })

	btnRow := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(ui.fieldsOk, 10, 0, true).
		AddItem(nil, 2, 0, false).
		AddItem(ui.fieldsCancel, 10, 0, false).
		AddItem(nil, 0, 1, false)

	// instr + form grow; error + buttons pinned to the bottom of the dialog.
	ui.fieldsRoot = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.fieldsInstr, 0, 0, false).
		AddItem(ui.fieldsForm, 0, 1, true).
		AddItem(ui.fieldsError, 1, 0, false).
		AddItem(btnRow, 3, 0, false)
	ui.fieldsRoot.SetBorder(true).SetTitle(" Fields ")
	ui.fieldsRoot.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			ui.finishFields(BtnCancel, nil)
			return nil
		}
		return ev
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(ui.fieldsRoot, 0, 2, true).
			AddItem(nil, 0, 1, false), 64, 0, true).
		AddItem(nil, 0, 1, false)
}

func (ui *appUI) showFieldReq(req *UiRequest) {
	title, body, _, specs := ValidateFieldDialogParams(req.Params)
	ui.fieldsSpecs = specs
	ui.fieldsError.SetText("")
	ui.fieldsRoot.SetTitle(" " + title + " ")

	if strings.TrimSpace(body) == "" {
		ui.fieldsInstr.SetText("")
		ui.fieldsRoot.ResizeItem(ui.fieldsInstr, 0, 0)
	} else {
		ui.fieldsInstr.SetText(strings.TrimSpace(body))
		ui.fieldsRoot.ResizeItem(ui.fieldsInstr, 2, 0)
	}

	ui.fieldsForm.Clear(true)
	ui.fieldsCtrls = make([]fieldCtrl, 0, len(specs))
	for _, spec := range specs {
		c := fieldCtrl{spec: spec}
		switch spec.Kind {
		case FieldKindBool:
			checked := bool(spec.Default.(finalembed.Bool))
			cb := tview.NewCheckbox().SetLabel(spec.Name + " ").SetChecked(checked)
			c.check = cb
			ui.fieldsForm.AddFormItem(cb)
		case FieldKindInt, FieldKindFloat:
			in := tview.NewInputField().
				SetLabel(spec.Name + ": ").
				SetText(formatDefault(spec)).
				SetFieldWidth(narrowFieldWidth)
			ui.trackEditTarget(in)
			c.input = in
			ui.fieldsForm.AddFormItem(in)
		default:
			in := tview.NewInputField().
				SetLabel(spec.Name + ": ").
				SetText(formatDefault(spec)).
				SetFieldWidth(0)
			ui.trackEditTarget(in)
			c.input = in
			ui.fieldsForm.AddFormItem(in)
		}
		ui.fieldsCtrls = append(ui.fieldsCtrls, c)
	}
	// Buttons live in the bottom row (not Form.AddButton) so they stay bottom-centered.

	ui.pages.SwitchToPage(pageFields)
	ui.app.SetFocus(ui.fieldsForm)
}

func formatDefault(spec FieldSpec) string {
	switch spec.Kind {
	case FieldKindString:
		return string(spec.Default.(finalembed.String))
	case FieldKindInt:
		return strconv.FormatInt(int64(spec.Default.(finalembed.Int)), 10)
	case FieldKindFloat:
		f := float64(spec.Default.(finalembed.Float))
		s := strconv.FormatFloat(f, 'g', -1, 64)
		// Keep a decimal so the control reads as float (0 → "0.0").
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s
	default:
		return ""
	}
}

func (ui *appUI) acceptFields() {
	out := finalembed.NewDynamicObject()
	for _, c := range ui.fieldsCtrls {
		switch c.spec.Kind {
		case FieldKindBool:
			out.Set(c.spec.Name, finalembed.Bool(c.check.IsChecked()))
		case FieldKindString:
			out.Set(c.spec.Name, finalembed.String(c.input.GetText()))
		case FieldKindInt:
			text := strings.TrimSpace(c.input.GetText())
			n, err := strconv.ParseInt(text, 10, 64)
			if err != nil {
				ui.fieldsError.SetText(c.spec.Name + ": not an integer")
				ui.app.SetFocus(c.input)
				return
			}
			out.Set(c.spec.Name, finalembed.Int(n))
		case FieldKindFloat:
			text := strings.TrimSpace(c.input.GetText())
			n, err := strconv.ParseFloat(text, 64)
			if err != nil {
				ui.fieldsError.SetText(c.spec.Name + ": not a float")
				ui.app.SetFocus(c.input)
				return
			}
			out.Set(c.spec.Name, finalembed.Float(n))
		}
	}
	ui.finishFields(BtnOk, out)
}

func (ui *appUI) finishFields(btn Btn, fields *finalembed.DynamicObject) {
	name, _ := ui.pages.GetFrontPage()
	if name != pageFields {
		return
	}
	ui.pages.SwitchToPage(pageMain)
	ui.app.SetFocus(ui.prompt)
	if btn == BtnOk {
		Tui_.Complete(CompleteFields(btn, fields))
		return
	}
	Tui_.Complete(CompleteChoice(BtnCancel))
}

func (ui *appUI) dispatchUIReq(req *UiRequest) {
	if req == nil {
		Tui_.Complete(&UiResponse{
			Status: UiStatusError,
			Err:    "nil UiRequest",
		})
		return
	}
	switch req.Cmd {
	case UiCmdPrompt:
		ui.showPromptReq(req)
	case UiCmdField:
		ui.showFieldReq(req)
	default:
		Tui_.Complete(&UiResponse{
			Status: UiStatusError,
			Err:    fmt.Sprintf("unsupported UiCmd %d", req.Cmd),
		})
	}
}
