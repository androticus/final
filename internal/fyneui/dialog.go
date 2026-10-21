package fyneui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

// Button constants match TUI Dialog (BtnCancel=0, BtnOk=1).
const (
	btnCancel = 0
	btnOk     = 1
)

const (
	doFieldTitle         = "title"
	doFieldBodyText      = "bodyText"
	doFieldDefaultButton = "defaultButton"
	doFieldFields        = "fields"
	maxFieldDialogFields = 8
	radioYes             = "Yes"
	radioNo              = "No"
	// Entry MinSize is tiny; floor the custom dialog so string fields are usable.
	fieldDialogMinWidth float32 = 420
)

type fieldKind int

const (
	fieldKindString fieldKind = iota
	fieldKindBool
	fieldKindInt
	fieldKindFloat
)

type fieldSpec struct {
	Name    string
	Kind    fieldKind
	Default any
}

type fieldRow struct {
	spec  fieldSpec
	entry *widget.Entry
	radio *widget.RadioGroup
}

func firstAppWindow() fyne.Window {
	var w fyne.Window
	fyne.DoAndWait(func() {
		if TheApp == nil {
			return
		}
		wins := TheApp.Driver().AllWindows()
		if len(wins) > 0 {
			w = wins[0]
		}
	})
	if w == nil {
		panic(fmt.Errorf("Dialog: no open Window"))
	}
	return w
}

func classifyFieldValue(v any) (fieldKind, bool) {
	switch v.(type) {
	case core.String:
		return fieldKindString, true
	case core.Bool:
		return fieldKindBool, true
	case core.Int:
		return fieldKindInt, true
	case core.Float:
		return fieldKindFloat, true
	default:
		return 0, false
	}
}

func mintFieldDialogParams() *core.DynamicObject {
	params := core.NewDynamicObject()
	params.Set(doFieldTitle, core.String("Information Request"))
	params.Set(doFieldDefaultButton, core.Int(btnOk))
	params.Set(doFieldBodyText, core.String(""))
	params.Set(doFieldFields, core.NewDynamicObject())
	return params
}

func validateFieldDialogParams(params *core.DynamicObject) (title, body string, defBtn int, specs []fieldSpec) {
	if params == nil {
		panic(fmt.Errorf("Dialog.showFieldDialog: nil params"))
	}

	title = "Information Request"
	defBtn = btnOk
	if params.HasField(doFieldTitle) {
		s, ok := params.Get(doFieldTitle).(core.String)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: title must be String"))
		}
		title = strings.TrimSpace(string(s))
		if title == "" {
			panic(fmt.Errorf("Dialog.showFieldDialog: title must be non-empty"))
		}
	}
	if params.HasField(doFieldDefaultButton) {
		i, ok := params.Get(doFieldDefaultButton).(core.Int)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: defaultButton must be Int"))
		}
		defBtn = int(i)
		if defBtn != btnOk && defBtn != btnCancel {
			panic(fmt.Errorf("Dialog.showFieldDialog: defaultButton must be btnOk or btnCancel"))
		}
	}
	if params.HasField(doFieldBodyText) {
		s, ok := params.Get(doFieldBodyText).(core.String)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: bodyText must be String"))
		}
		body = string(s)
	}

	if !params.HasField(doFieldFields) {
		panic(fmt.Errorf("Dialog.showFieldDialog: missing fields"))
	}
	raw := params.Get(doFieldFields)
	fields, ok := raw.(*core.DynamicObject)
	if !ok || fields == nil {
		panic(fmt.Errorf("Dialog.showFieldDialog: fields must be a DynamicObject"))
	}
	n := fields.Len()
	if n < 1 || n > maxFieldDialogFields {
		panic(fmt.Errorf("Dialog.showFieldDialog: fields count %d not in 1..%d", n, maxFieldDialogFields))
	}
	specs = make([]fieldSpec, 0, n)
	for i := 0; i < n; i++ {
		item := fields.ItemAt(i)
		kind, ok := classifyFieldValue(item.Value)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: field %q has unsupported type %T", item.Name, item.Value))
		}
		specs = append(specs, fieldSpec{Name: item.Name, Kind: kind, Default: item.Value})
	}
	return title, body, defBtn, specs
}

func formatFieldDefault(spec fieldSpec) string {
	switch spec.Kind {
	case fieldKindString:
		return string(spec.Default.(core.String))
	case fieldKindInt:
		return strconv.FormatInt(int64(spec.Default.(core.Int)), 10)
	case fieldKindFloat:
		f := float64(spec.Default.(core.Float))
		s := strconv.FormatFloat(f, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s
	default:
		return ""
	}
}

func parseIntField(name, text string) (core.Int, string) {
	n, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if err != nil {
		return 0, name + ": not an integer"
	}
	return core.Int(n), ""
}

func parseFloatField(name, text string) (core.Float, string) {
	n, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		return 0, name + ": not a float"
	}
	return core.Float(n), ""
}

func harvestFields(rows []fieldRow) (*core.DynamicObject, string) {
	out := core.NewDynamicObject()
	for _, row := range rows {
		switch row.spec.Kind {
		case fieldKindBool:
			out.Set(row.spec.Name, core.Bool(row.radio.Selected == radioYes))
		case fieldKindString:
			out.Set(row.spec.Name, core.String(row.entry.Text))
		case fieldKindInt:
			n, errMsg := parseIntField(row.spec.Name, row.entry.Text)
			if errMsg != "" {
				return nil, errMsg
			}
			out.Set(row.spec.Name, n)
		case fieldKindFloat:
			n, errMsg := parseFloatField(row.spec.Name, row.entry.Text)
			if errMsg != "" {
				return nil, errMsg
			}
			out.Set(row.spec.Name, n)
		}
	}
	return out, ""
}

func newFieldRow(spec fieldSpec) fieldRow {
	row := fieldRow{spec: spec}
	switch spec.Kind {
	case fieldKindBool:
		rg := widget.NewRadioGroup([]string{radioYes, radioNo}, nil)
		rg.Horizontal = true
		if bool(spec.Default.(core.Bool)) {
			rg.SetSelected(radioYes)
		} else {
			rg.SetSelected(radioNo)
		}
		row.radio = rg
	default:
		e := widget.NewEntry()
		e.SetText(formatFieldDefault(spec))
		row.entry = e
	}
	return row
}

// ( prompt:s -- btn:i ) — TUI-compatible modal wait.
func showDefaultPrompt(prompt string) int {
	parent := firstAppWindow()
	w := beginModalWait()
	var d *dialog.ConfirmDialog
	fyne.DoAndWait(func() {
		d = dialog.NewConfirm("Prompt", prompt, func(yes bool) {
			w.complete(confirmCloserIndex(yes), nil)
		}, parent)
		d.Show()
	})
	res := w.wait()
	fyne.DoAndWait(func() {
		d.Hide()
	})
	return res.CloserIndex
}

// ( fdp:DO -- fields:DO true | false ) — TUI-compatible modal wait.
func showFieldDialog(params *core.DynamicObject) (fields *core.DynamicObject, ok bool) {
	title, body, _, specs := validateFieldDialogParams(params)
	parent := firstAppWindow()
	w := beginModalWait()

	var d dialog.Dialog
	var rows []fieldRow
	fyne.DoAndWait(func() {
		form := widget.NewForm()
		form.SubmitText = "Ok"
		form.CancelText = "Cancel"
		rows = make([]fieldRow, 0, len(specs))
		for _, spec := range specs {
			row := newFieldRow(spec)
			rows = append(rows, row)
			if row.radio != nil {
				form.Append(spec.Name, row.radio)
			} else {
				form.Append(spec.Name, row.entry)
			}
		}

		errLabel := widget.NewLabel("")
		var bodyLabel fyne.CanvasObject
		if strings.TrimSpace(body) != "" {
			lbl := widget.NewLabel(strings.TrimSpace(body))
			lbl.Wrapping = fyne.TextWrapWord
			bodyLabel = lbl
		}

		var content fyne.CanvasObject
		if bodyLabel != nil {
			content = container.NewVBox(bodyLabel, form, errLabel)
		} else {
			content = container.NewVBox(form, errLabel)
		}

		form.OnSubmit = func() {
			harvested, errMsg := harvestFields(rows)
			if errMsg != "" {
				errLabel.SetText(errMsg)
				return
			}
			w.complete(btnOk, harvested)
			d.Hide()
		}
		form.OnCancel = func() {
			w.complete(btnCancel, nil)
			d.Hide()
		}
		form.Refresh()

		d = dialog.NewCustomWithoutButtons(title, content, parent)
		d.SetOnClosed(func() {
			w.complete(btnCancel, nil)
		})
		d.Show()
		min := d.MinSize()
		width := min.Width
		if width < fieldDialogMinWidth {
			width = fieldDialogMinWidth
		}
		d.Resize(fyne.NewSize(width, min.Height))
	})
	res := w.wait()
	fyne.DoAndWait(func() {
		d.Hide()
	})
	if res.CloserIndex != btnOk {
		return nil, false
	}
	return res.Data.(*core.DynamicObject), true
}

func registerDialogMembers(t *core.Type) {
	t.RegisterClassProperty("btnOk", "Int",
		func(_ any) any { return core.Int(btnOk) }, nil)
	t.RegisterClassProperty("btnCancel", "Int",
		func(_ any) any { return core.Int(btnCancel) }, nil)

	// ( prompt:s -- btn:i )
	t.RegisterNativeClassMethod("showDefaultPrompt", func(f *core.RunFrame, _ *core.TypeObjPair) {
		prompt := string(f.RP.DataStack.Pop().(core.String))
		f.RP.DataStack.Push(core.Int(showDefaultPrompt(prompt)))
	})

	// ( -- params:DynamicObject )
	t.RegisterNativeClassMethod("newFieldDialogParams", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(mintFieldDialogParams())
	})

	// ( fdp:DynamicObject -- fields:DO true | false )
	t.RegisterNativeClassMethod("showFieldDialog", func(f *core.RunFrame, _ *core.TypeObjPair) {
		params := f.RP.DataStack.Pop().(*core.DynamicObject)
		fields, ok := showFieldDialog(params)
		if !ok {
			f.RP.DataStack.Push(core.Bool(false))
			return
		}
		f.RP.DataStack.Push(fields)
		f.RP.DataStack.Push(core.Bool(true))
	})
}

func CreateRegisterDialogType() {
	t := core.NewRegisteringType("Dialog",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Dialog"),
	)
	registerDialogMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterDialogType()
}
