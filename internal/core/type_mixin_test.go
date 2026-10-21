package core

import "testing"

func TestApplyMixinRecordsCatalogType(t *testing.T) {
	mixin := &Type{name: "MixinProbe"}
	client := &Type{name: "ClientProbe"}
	ApplyMixin(client, mixin, func(t *Type) {
		t.registerNativeMethod("probe", func(*RunFrame, *TypeObjPair) {})
	})
	got := client.GetMixinTypes()
	if len(got) != 1 || got[0] != mixin {
		t.Fatalf("GetMixinTypes() = %v, want [%v]", got, mixin)
	}
	mm := client.GetMethodMetadata("probe")
	if mm == nil {
		t.Fatal("expected local probe method from mixin register")
	}
	if mm.mixinType != mixin {
		t.Fatalf("probe.mixinType = %v, want MixinProbe", mm.mixinType)
	}
	if client.mixinRegistering != nil {
		t.Fatal("mixinRegistering must be cleared after ApplyMixin")
	}
	if mixin.GetMethodMetadata("probe") != nil {
		t.Fatal("mixin catalog type must not receive client's registered methods")
	}
}

func TestTryGetMixinTypeOnMethodMetadata(t *testing.T) {
	ResetTestRuntime()
	pmT := GetTypeRegistrar().Get("Pixmap")
	mm := pmT.GetMethodMetadata("savePngToFile")
	if mm == nil || mm.mixinType == nil || mm.mixinType.Name() != "Image" {
		t.Fatalf("Pixmap.savePngToFile.mixinType = %v, want Image", mm)
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, mm)
	meta := ths.Typ.GetMethodMetadata("tryGetMixinType")
	if meta == nil {
		t.Fatal("missing tryGetMixinType on MethodMetadata")
	}
	meta.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("tryGetMixinType want true")
	}
	typ, ok := rp.DataStack.Pop().(*Type)
	if !ok || typ.Name() != "Image" {
		t.Fatalf("got %v, want Image", typ)
	}

	native := pmT.GetMethodMetadata("clear")
	if native == nil {
		t.Fatal("Pixmap.clear missing")
	}
	if native.mixinType != nil {
		t.Fatal("Pixmap.clear must not have mixinType")
	}
	f, ths = memberRunFrame(rp, native)
	ths.Typ.GetMethodMetadata("tryGetMixinType").invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatal("clear tryGetMixinType want false")
	}
}

func TestTryGetMixinTypeOnPropertyMetadata(t *testing.T) {
	ResetTestRuntime()
	pmT := GetTypeRegistrar().Get("Pixmap")
	pm := pmT.GetPropertyMetadata("rows")
	if pm == nil {
		t.Fatal("Pixmap.rows missing")
	}
	if pm.mixinType != nil {
		t.Fatal("Pixmap.rows must not have mixinType")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pm)
	meta := ths.Typ.GetMethodMetadata("tryGetMixinType")
	if meta == nil {
		t.Fatal("missing tryGetMixinType on PropertyMetadata")
	}
	meta.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatal("rows tryGetMixinType want false")
	}
}

func TestTryGetMixinTypesEmpty(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	if intT == nil {
		t.Fatal("Int not registered")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, intT)
	mm := ths.Typ.GetMethodMetadata("tryGetMixinTypes")
	if mm == nil {
		t.Fatal("missing tryGetMixinTypes on Type facet")
	}
	mm.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatalf("Int tryGetMixinTypes want false")
	}
}

func TestTryGetMixinTypesPixmapImage(t *testing.T) {
	ResetTestRuntime()
	pmT := GetTypeRegistrar().Get("Pixmap")
	if pmT == nil {
		t.Fatal("Pixmap not registered")
	}
	mixins := pmT.GetMixinTypes()
	if len(mixins) != 1 || mixins[0].Name() != "Image" {
		t.Fatalf("Pixmap mixins = %v, want [Image]", mixins)
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pmT)
	mm := ths.Typ.GetMethodMetadata("tryGetMixinTypes")
	if mm == nil {
		t.Fatal("missing tryGetMixinTypes on Type facet")
	}
	mm.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatalf("tryGetMixinTypes want true")
	}
	list, isList := rp.DataStack.Pop().(*AnyList)
	if !isList || list.Len() != 1 {
		t.Fatalf("got %T len=%d, want AnyList len 1", list, 0)
	}
	typ, isType := list.At(0).(*Type)
	if !isType || typ.Name() != "Image" {
		t.Fatalf("mixin[0] = %v, want Image *Type", list.At(0))
	}
}

func TestPixmapInheritsImageCodecMethods(t *testing.T) {
	pmT := GetTypeRegistrar().Get("Pixmap")
	mm := pmT.GetMethodMetadata("savePngToFile")
	if mm == nil {
		t.Fatal("Pixmap missing savePngToFile from Image mixin")
	}
	if mm.mixinType != imageType {
		t.Fatalf("savePngToFile.mixinType = %v, want Image", mm.mixinType)
	}
	if imageType.GetMethodMetadata("savePngToFile") == nil {
		t.Fatal("Image catalog missing savePngToFile")
	}
	if imageType.GetMethodMetadata("savePngToFile").mixinType != nil {
		t.Fatal("Image catalog savePngToFile must not be stamped as mixin")
	}
}

func TestMat2cInheritsImageCodecMethods(t *testing.T) {
	mcT := GetTypeRegistrar().Get("Mat2c")
	if mcT.GetMethodMetadata("savePngToFile") == nil {
		t.Fatal("Mat2c missing savePngToFile from Image mixin")
	}
	mixins := mcT.GetMixinTypes()
	if len(mixins) != 1 || mixins[0].Name() != "Image" {
		t.Fatalf("Mat2c mixins = %v, want [Image]", mixins)
	}
	if mcT.GetParent() != matProtoType {
		t.Fatalf("Mat2c parent = %v, want ProtoMat", mcT.GetParent())
	}
}

func TestHelpPixmapCompletes(t *testing.T) {
	ResetTestRuntime()
	task, err := newSessionTask(SessionKindLocal)
	if err != nil {
		t.Fatal(err)
	}
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp, "help Pixmap\n")
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after help Pixmap", rp.DataStack.Len())
	}
}

func TestHelpMat2cCompletes(t *testing.T) {
	ResetTestRuntime()
	task, err := newSessionTask(SessionKindLocal)
	if err != nil {
		t.Fatal(err)
	}
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp, "help Mat2c\n")
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after help Mat2c", rp.DataStack.Len())
	}
}
