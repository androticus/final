package finaltui

import (
	"fmt"

	"github.com/androticus/final/finalembed"
)

// Dialog is a class-only host type (no instances).
// Phase 2: showDefaultPrompt. Phase 3: newFieldDialogParams + showFieldDialog.
type Dialog struct{}

func registerDialog(ty *finalembed.Type) {
	ty.RegisterClassProperty("btnOk", "Int",
		func(_ any) any { return finalembed.Int(BtnOk) }, nil)
	ty.RegisterClassProperty("btnCancel", "Int",
		func(_ any) any { return finalembed.Int(BtnCancel) }, nil)

	ty.RegisterNativeClassMethod("showDefaultPrompt", func(f *finalembed.RunFrame, _ *finalembed.TypeObjPair) {
		body := string(f.RP.DataStack.Pop().(finalembed.String))
		params := finalembed.NewDynamicObject()
		params.Set(DOFieldTitle.String(), finalembed.String("Prompt"))
		params.Set(DOFieldBodyText.String(), finalembed.String(body))
		params.Set(DOFieldDefaultButton.String(), finalembed.Int(BtnOk))

		res := Tui_.raisePrompt(params)
		f.RP.DataStack.Push(finalembed.Int(choiceFromResponse(res)))
	})

	ty.RegisterNativeClassMethod("newFieldDialogParams", func(f *finalembed.RunFrame, _ *finalembed.TypeObjPair) {
		f.RP.DataStack.Push(mintFieldDialogParams())
	})

	ty.RegisterNativeClassMethod("showFieldDialog", func(f *finalembed.RunFrame, _ *finalembed.TypeObjPair) {
		params := f.RP.DataStack.Pop().(*finalembed.DynamicObject)
		ValidateFieldDialogParams(params) // panic if invalid before UI
		res := Tui_.raiseField(params)
		pushFieldDialogResult(f, res)
	})
}

func choiceFromResponse(res *UiResponse) Btn {
	if res == nil {
		panic(fmt.Errorf("Dialog: nil UiResponse"))
	}
	if res.Status != UiStatusOK {
		panic(fmt.Errorf("Dialog: %s", res.Err))
	}
	if res.Result == nil {
		panic(fmt.Errorf("Dialog: nil Result"))
	}
	v := res.Result.Get(DOFieldChoice.String())
	if v == nil {
		panic(fmt.Errorf("Dialog: missing %s", DOFieldChoice))
	}
	switch c := v.(type) {
	case finalembed.Int:
		return Btn(c)
	default:
		panic(fmt.Errorf("Dialog: %s has type %T", DOFieldChoice, v))
	}
}

func pushFieldDialogResult(f *finalembed.RunFrame, res *UiResponse) {
	btn := choiceFromResponse(res)
	if btn == BtnCancel {
		f.RP.DataStack.Push(finalembed.Bool(false))
		return
	}
	// Final stack gets the edited fields bag only — never the outer params wrapper
	// (title/defaultButton/bodyText/fields).
	if !res.Result.HasField(DOFieldFields.String()) {
		panic(fmt.Errorf("Dialog.showFieldDialog: Ok response missing fields"))
	}
	raw := res.Result.Get(DOFieldFields.String())
	fields, ok := raw.(*finalembed.DynamicObject)
	if !ok || fields == nil {
		panic(fmt.Errorf("Dialog.showFieldDialog: Ok response fields not a DynamicObject (got %T)", raw))
	}
	f.RP.DataStack.Push(fields)
	f.RP.DataStack.Push(finalembed.Bool(true))
}
