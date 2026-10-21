package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func mintedFields(t *testing.T, params *core.DynamicObject) *core.DynamicObject {
	t.Helper()
	raw := params.Get(doFieldFields)
	fields, ok := raw.(*core.DynamicObject)
	if !ok || fields == nil {
		t.Fatalf("minted params.fields = %T", raw)
	}
	return fields
}

func TestMintFieldDialogParamsDefaults(t *testing.T) {
	params := mintFieldDialogParams()
	if got := string(params.Get(doFieldTitle).(core.String)); got != "Information Request" {
		t.Fatalf("title=%q", got)
	}
	if int(params.Get(doFieldDefaultButton).(core.Int)) != btnOk {
		t.Fatalf("defaultButton=%v", params.Get(doFieldDefaultButton))
	}
	if got := string(params.Get(doFieldBodyText).(core.String)); got != "" {
		t.Fatalf("bodyText=%q", got)
	}
	fields := mintedFields(t, params)
	if fields.Len() != 0 {
		t.Fatalf("fields.Len()=%d, want 0", fields.Len())
	}
	defer func() {
		if recover() == nil {
			t.Fatal("empty minted fields must fail validateFieldDialogParams")
		}
	}()
	validateFieldDialogParams(params)
}

func TestValidateFieldDialogParamsOK(t *testing.T) {
	params := mintFieldDialogParams()
	fields := mintedFields(t, params)
	fields.Set("name", core.String(""))
	fields.Set("n", core.Int(0))
	fields.Set("f", core.Float(0))
	fields.Set("b", core.Bool(false))
	title, body, def, specs := validateFieldDialogParams(params)
	if title != "Information Request" || body != "" || def != btnOk {
		t.Fatalf("title=%q body=%q def=%v", title, body, def)
	}
	if len(specs) != 4 {
		t.Fatalf("specs=%d", len(specs))
	}
	if specs[0].Name != "name" || specs[0].Kind != fieldKindString {
		t.Fatalf("spec0=%+v", specs[0])
	}
	if specs[3].Kind != fieldKindBool {
		t.Fatalf("spec3=%+v", specs[3])
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
	validateFieldDialogParams(params)
}

func TestParseIntFloatFields(t *testing.T) {
	n, err := parseIntField("age", "36")
	if err != "" || n != 36 {
		t.Fatalf("parseInt=%v %q", n, err)
	}
	if _, err := parseIntField("age", "x"); err == "" {
		t.Fatal("want int error")
	}
	f, err := parseFloatField("speed", "1.5")
	if err != "" || f != 1.5 {
		t.Fatalf("parseFloat=%v %q", f, err)
	}
	if _, err := parseFloatField("speed", "nope"); err == "" {
		t.Fatal("want float error")
	}
}

func TestDialogTypeRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Dialog")
	if typ == nil {
		t.Fatal("Dialog missing")
	}
	for _, name := range []string{"btnOk", "btnCancel"} {
		if typ.GetClassPropertyMetadata(name) == nil {
			t.Fatalf("missing class property %s", name)
		}
	}
	for _, name := range []string{"showDefaultPrompt", "showFieldDialog", "newFieldDialogParams"} {
		if typ.GetClassMethodMetadata(name) == nil {
			t.Fatalf("missing class method %s", name)
		}
	}
	if typ.GetClassMethodMetadata("_dispatch") != nil {
		t.Fatal("_dispatch should be removed (modal wait, no EventLoop mint)")
	}
}

func TestFormatFieldDefault(t *testing.T) {
	if got := formatFieldDefault(fieldSpec{Kind: fieldKindInt, Default: core.Int(0)}); got != "0" {
		t.Fatalf("int default=%q", got)
	}
	if got := formatFieldDefault(fieldSpec{Kind: fieldKindFloat, Default: core.Float(0)}); got != "0.0" {
		t.Fatalf("float default=%q", got)
	}
	if got := formatFieldDefault(fieldSpec{Kind: fieldKindString, Default: core.String("hi")}); got != "hi" {
		t.Fatalf("string default=%q", got)
	}
}
