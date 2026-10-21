package core

import "testing"

func TestModuleFileStringFieldWrongTypeEmpty(t *testing.T) {
	o := NewDynamicObject()
	o.Set("name", Int(42))
	o.Set("description", Bool(true))
	o.Set("final", NewAnyList())
	o.Set("version", Int(1))
	mf := NewModuleFile(o)
	if got := mf.GetName(); got != "" {
		t.Fatalf("GetName = %q, want \"\"", got)
	}
	if got := mf.GetDescription(); got != "" {
		t.Fatalf("GetDescription = %q, want \"\"", got)
	}
	if got := mf.GetFinal(); got != "" {
		t.Fatalf("GetFinal = %q, want \"\"", got)
	}
	if got := mf.GetVersion(); got != "" {
		t.Fatalf("GetVersion = %q, want \"\"", got)
	}
}

func TestModuleFileStringFieldFinalString(t *testing.T) {
	o := NewDynamicObject()
	o.Set("name", String("modules/csv"))
	o.Set("description", String("CSV helpers"))
	o.Set("final", String("v0.0.7-rc1"))
	o.Set("version", String("v0.1.0"))
	mf := NewModuleFile(o)
	if got := mf.GetName(); got != "modules/csv" {
		t.Fatalf("GetName = %q, want modules/csv", got)
	}
	if got := mf.GetDescription(); got != "CSV helpers" {
		t.Fatalf("GetDescription = %q", got)
	}
	if got := mf.GetFinal(); got != "v0.0.7-rc1" {
		t.Fatalf("GetFinal = %q", got)
	}
	if got := mf.GetVersion(); got != "v0.1.0" {
		t.Fatalf("GetVersion = %q, want v0.1.0", got)
	}
}

func TestModuleFileListFieldWrongTypeNil(t *testing.T) {
	o := NewDynamicObject()
	o.Set("requires", String("nope"))
	o.Set("files", Int(1))
	mf := NewModuleFile(o)
	if mf.GetRequires() != nil {
		t.Fatalf("GetRequires = %v, want nil", mf.GetRequires())
	}
	if mf.GetFiles() != nil {
		t.Fatalf("GetFiles = %v, want nil", mf.GetFiles())
	}
}

func TestModuleFileListFieldOK(t *testing.T) {
	files := NewAnyList()
	files.Append(String("10-a.fnl"))
	o := NewDynamicObject()
	o.Set("files", files)
	mf := NewModuleFile(o)
	got := mf.GetFiles()
	if got == nil || got.Len() != 1 {
		t.Fatalf("GetFiles = %#v", got)
	}
}

func TestModuleFileTypeRegistered(t *testing.T) {
	if GetTypeRegistrar().Get("ModuleFile") == nil {
		t.Fatal("ModuleFile type not registered")
	}
	mf := NewModuleFile(NewDynamicObject())
	if mf.FinalType() == nil || mf.FinalType().Name() != "ModuleFile" {
		t.Fatalf("FinalType = %#v", mf.FinalType())
	}
	for _, name := range []string{"name", "description", "final", "version", "requires", "files"} {
		pm := mf.FinalType().GetPropertyMetadata(name)
		if pm == nil || pm.set != nil {
			t.Fatalf("%s property = %#v, want const", name, pm)
		}
	}
	if pm := mf.FinalType().GetPropertyMetadata("path"); pm != nil {
		t.Fatal("path must not be a ModuleFile property")
	}
}
