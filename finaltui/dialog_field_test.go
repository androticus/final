package finaltui

import (
	"testing"

	"github.com/androticus/final/finalembed"
)

func mintedFields(t *testing.T, params *finalembed.DynamicObject) *finalembed.DynamicObject {
	t.Helper()
	raw := params.Get(DOFieldFields.String())
	fields, ok := raw.(*finalembed.DynamicObject)
	if !ok || fields == nil {
		t.Fatalf("minted params.fields = %T", raw)
	}
	return fields
}

func TestMintFieldDialogParamsDefaults(t *testing.T) {
	params := mintFieldDialogParams()
	if got := string(params.Get(DOFieldTitle.String()).(finalembed.String)); got != "Information Request" {
		t.Fatalf("title=%q", got)
	}
	if Btn(params.Get(DOFieldDefaultButton.String()).(finalembed.Int)) != BtnOk {
		t.Fatalf("defaultButton=%v", params.Get(DOFieldDefaultButton.String()))
	}
	if got := string(params.Get(DOFieldBodyText.String()).(finalembed.String)); got != "" {
		t.Fatalf("bodyText=%q", got)
	}
	fields := mintedFields(t, params)
	if fields.Len() != 0 {
		t.Fatalf("fields.Len()=%d, want 0", fields.Len())
	}
	defer func() {
		if recover() == nil {
			t.Fatal("empty minted fields must fail ValidateFieldDialogParams")
		}
	}()
	ValidateFieldDialogParams(params)
}

func TestValidateFieldDialogParamsOK(t *testing.T) {
	params := mintFieldDialogParams()
	fields := mintedFields(t, params)
	fields.Set("name", finalembed.String(""))
	fields.Set("n", finalembed.Int(0))
	fields.Set("f", finalembed.Float(0))
	fields.Set("b", finalembed.Bool(false))
	title, body, def, specs := ValidateFieldDialogParams(params)
	if title != "Information Request" || body != "" || def != BtnOk {
		t.Fatalf("title=%q body=%q def=%v", title, body, def)
	}
	if !params.HasField(DOFieldBodyText.String()) {
		t.Fatal("factory must set bodyText (empty = hide instructions)")
	}
	if len(specs) != 4 {
		t.Fatalf("specs=%d", len(specs))
	}
	if specs[0].Name != "name" || specs[0].Kind != FieldKindString {
		t.Fatalf("spec0=%+v", specs[0])
	}
}

func TestValidateFieldDialogParamsBadType(t *testing.T) {
	params := mintFieldDialogParams()
	mintedFields(t, params).Set("x", []byte("nope"))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	ValidateFieldDialogParams(params)
}

func TestShowFieldDialogRoundTrip(t *testing.T) {
	Tui_.ArmUI()
	go func() {
		req := <-Tui_.UIRequests()
		if req.Cmd != UiCmdField {
			t.Errorf("cmd=%v", req.Cmd)
		}
		out := finalembed.NewDynamicObject()
		out.Set("name", finalembed.String("Ada"))
		out.Set("n", finalembed.Int(42))
		Tui_.Complete(CompleteFields(BtnOk, out))
	}()

	params := mintFieldDialogParams()
	fields := mintedFields(t, params)
	fields.Set("name", finalembed.String(""))
	fields.Set("n", finalembed.Int(0))
	res := Tui_.raiseField(params)
	if choiceFromResponse(res) != BtnOk {
		t.Fatal("want Ok")
	}
	got := res.Result.Get(DOFieldFields.String()).(*finalembed.DynamicObject)
	if string(got.Get("name").(finalembed.String)) != "Ada" {
		t.Fatalf("name=%v", got.Get("name"))
	}
	if got.Get("n").(finalembed.Int) != 42 {
		t.Fatalf("n=%v", got.Get("n"))
	}
}
