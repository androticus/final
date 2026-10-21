package finaltui

import (
	"fmt"
	"strings"

	"github.com/androticus/final/finalembed"
)

// FieldKind classifies a fields-bag entry by its default value's Final type.
type FieldKind int

const (
	FieldKindString FieldKind = iota
	FieldKindBool
	FieldKindInt
	FieldKindFloat
)

// FieldSpec is one row in a field dialog (ordinal order).
type FieldSpec struct {
	Name    string
	Kind    FieldKind
	Default any
}

func classifyFieldValue(v any) (FieldKind, bool) {
	switch v.(type) {
	case finalembed.String:
		return FieldKindString, true
	case finalembed.Bool:
		return FieldKindBool, true
	case finalembed.Int:
		return FieldKindInt, true
	case finalembed.Float:
		return FieldKindFloat, true
	default:
		return 0, false
	}
}

// ValidateFieldDialogParams checks outer params + nested fields. Panics on invalid.
// Returns title, body (may be empty), default button, and field specs in ordinal order.
func ValidateFieldDialogParams(params *finalembed.DynamicObject) (title, body string, defBtn Btn, specs []FieldSpec) {
	if params == nil {
		panic(fmt.Errorf("Dialog.showFieldDialog: nil params"))
	}

	title = "Information Request"
	defBtn = BtnOk
	if params.HasField(DOFieldTitle.String()) {
		s, ok := params.Get(DOFieldTitle.String()).(finalembed.String)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: title must be String"))
		}
		title = strings.TrimSpace(string(s))
		if title == "" {
			panic(fmt.Errorf("Dialog.showFieldDialog: title must be non-empty"))
		}
	}
	if params.HasField(DOFieldDefaultButton.String()) {
		i, ok := params.Get(DOFieldDefaultButton.String()).(finalembed.Int)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: defaultButton must be Int"))
		}
		defBtn = Btn(i)
		if defBtn != BtnOk && defBtn != BtnCancel {
			panic(fmt.Errorf("Dialog.showFieldDialog: defaultButton must be btnOk or btnCancel"))
		}
	}
	if params.HasField(DOFieldBodyText.String()) {
		s, ok := params.Get(DOFieldBodyText.String()).(finalembed.String)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: bodyText must be String"))
		}
		body = string(s)
	}

	if !params.HasField(DOFieldFields.String()) {
		panic(fmt.Errorf("Dialog.showFieldDialog: missing fields"))
	}
	raw := params.Get(DOFieldFields.String())
	fields, ok := raw.(*finalembed.DynamicObject)
	if !ok || fields == nil {
		panic(fmt.Errorf("Dialog.showFieldDialog: fields must be a DynamicObject"))
	}
	n := fields.Len()
	if n < 1 || n > MaxFieldDialogFields {
		panic(fmt.Errorf("Dialog.showFieldDialog: fields count %d not in 1..%d", n, MaxFieldDialogFields))
	}
	specs = make([]FieldSpec, 0, n)
	for i := 0; i < n; i++ {
		item := fields.ItemAt(i)
		kind, ok := classifyFieldValue(item.Value)
		if !ok {
			panic(fmt.Errorf("Dialog.showFieldDialog: field %q has unsupported type %T", item.Name, item.Value))
		}
		specs = append(specs, FieldSpec{Name: item.Name, Kind: kind, Default: item.Value})
	}
	return title, body, defBtn, specs
}

func mintFieldDialogParams() *finalembed.DynamicObject {
	params := finalembed.NewDynamicObject()
	params.Set(DOFieldTitle.String(), finalembed.String("Information Request"))
	params.Set(DOFieldDefaultButton.String(), finalembed.Int(BtnOk))
	params.Set(DOFieldBodyText.String(), finalembed.String(""))
	params.Set(DOFieldFields.String(), finalembed.NewDynamicObject())
	return params
}
