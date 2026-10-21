package core

// ModuleFile wraps a DynamicObject parsed from a `_MODULE_` JSON document.
// Fields match the module file schema: name, description, final, version, requires, files.
type ModuleFile struct {
	data *DynamicObject
}

func NewModuleFile(data *DynamicObject) *ModuleFile {
	return &ModuleFile{data: data}
}

func (mf *ModuleFile) FinalType() *Type {
	return moduleFileType
}

func (mf *ModuleFile) GetName() string {
	return mf.stringField("name")
}

func (mf *ModuleFile) GetDescription() string {
	return mf.stringField("description")
}

func (mf *ModuleFile) GetFinal() string {
	return mf.stringField("final")
}

func (mf *ModuleFile) GetVersion() string {
	return mf.stringField("version")
}

func (mf *ModuleFile) GetRequires() *AnyList {
	return mf.listField("requires")
}

func (mf *ModuleFile) GetFiles() *AnyList {
	return mf.listField("files")
}

func (mf *ModuleFile) stringField(name string) string {
	if mf == nil || mf.data == nil || !mf.data.HasField(name) {
		return ""
	}
	return moduleFileAsString(mf.data.Get(name))
}

func (mf *ModuleFile) listField(name string) *AnyList {
	if mf == nil || mf.data == nil || !mf.data.HasField(name) {
		return nil
	}
	return moduleFileAsAnyList(mf.data.Get(name))
}

// moduleFileAsString returns the string value, or "" if missing/wrong type.
// Final JSON uses String (alias of string); reject non-string types without panic.
func moduleFileAsString(v any) string {
	if s, ok := v.(String); ok {
		return string(s)
	}
	return ""
}

// moduleFileAsAnyList returns *AnyList, or nil if missing/wrong type.
func moduleFileAsAnyList(v any) *AnyList {
	if l, ok := v.(*AnyList); ok {
		return l
	}
	return nil
}

func registerModuleFileMembers(t *Type) {
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*ModuleFile).GetName()) },
		nil,
	)
	t.registerProperty("description", "String",
		func(inst Any) Any { return String(inst.(*ModuleFile).GetDescription()) },
		nil,
	)
	t.registerProperty("final", "String",
		func(inst Any) Any { return String(inst.(*ModuleFile).GetFinal()) },
		nil,
	)
	t.registerProperty("version", "String",
		func(inst Any) Any { return String(inst.(*ModuleFile).GetVersion()) },
		nil,
	)
	t.registerProperty("requires", "List",
		func(inst Any) Any { return inst.(*ModuleFile).GetRequires() },
		nil,
	)
	t.registerProperty("files", "List",
		func(inst Any) Any { return inst.(*ModuleFile).GetFiles() },
		nil,
	)
}

func CreateRegisterModuleFileType() {
	t := &Type{
		name:            "ModuleFile",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[ModuleFile]("ModuleFile"),
	}
	registerModuleFileMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	moduleFileType = t
}

var moduleFileType *Type

func init() {
	CreateRegisterModuleFileType()
}

var _ Typed = (*ModuleFile)(nil)
