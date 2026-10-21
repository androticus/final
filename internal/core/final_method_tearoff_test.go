package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseFinalMethodHeader(t *testing.T) {
	t.Parallel()
	typeName, methodName, isClass, ok, err := parseFinalMethodHeader("Person#print")
	if err != nil || !ok || isClass || typeName != "Person" || methodName != "print" {
		t.Fatalf("Person#print → %q %q class=%v ok=%v err=%v", typeName, methodName, isClass, ok, err)
	}
	typeName, methodName, isClass, ok, err = parseFinalMethodHeader("Person##newWithAge")
	if err != nil || !ok || !isClass || typeName != "Person" || methodName != "newWithAge" {
		t.Fatalf("Person##newWithAge → %q %q class=%v ok=%v err=%v", typeName, methodName, isClass, ok, err)
	}
	_, _, _, ok, err = parseFinalMethodHeader("plain")
	if err != nil || ok {
		t.Fatalf("plain → ok=%v err=%v", ok, err)
	}
	_, _, _, _, err = parseFinalMethodHeader("A#B#C")
	if err == nil {
		t.Fatal("A#B#C: want error")
	}
	_, _, _, _, err = parseFinalMethodHeader("A###B")
	if err == nil {
		t.Fatal("A###B: want error")
	}
}

func TestFinalMethodTearOff_instanceNoDictWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
TypedObject.newTypeDef^ >ys
"TearHost" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^

: TearHost#answer
  42
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	if globalDictionaryContext.dict.GetByName("TearHost#answer") != nil {
		t.Fatal("tear-off left TearHost#answer in dictionary")
	}
	typ := globalDictionaryContext.tr.GetLocal("TearHost")
	if typ == nil {
		t.Fatal("TearHost not in local catalog")
	}
	mm := typ.MethodByName("answer")
	if mm == nil || !mm.isFinalMethod {
		t.Fatal("answer Final method missing")
	}
	ct := ModuleTable.DCForModule(nil, mm.finalModuleID).ct
	off := mm.finalColonCode
	foundODrop := false
	for i := off; i < ct.Len()-1; i++ {
		if ct.Read(i) == int(OpODrop) && ct.Read(i+1) == int(OpUnnest) {
			foundODrop = true
			break
		}
	}
	if !foundODrop {
		t.Fatalf("TearHost#answer CT from %d: want OpODrop before OpUnnest", off)
	}
	if mm.doc != "" {
		// no prelude comment in this snippet
	}
	rp.DataStack.Push(MakeNewTypedObject(typ))
	runHandleCompoundWithSegs(task, rp, []string{".", "answer", "^"})
	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d after invoke", rp.OStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want 42", rp.DataStack.Peek())
	}
}

func TestFinalMethodTearOff_preludeDoc(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
TypedObject.newTypeDef^ >ys
"DocHost" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^

/*
answers forty-two
*/
: DocHost#answer
  42
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	mm := globalDictionaryContext.tr.GetLocal("DocHost").MethodByName("answer")
	if mm == nil {
		t.Fatal("missing answer")
	}
	if !strings.Contains(mm.doc, "forty-two") {
		t.Fatalf("doc = %q, want prelude comment", mm.doc)
	}
}

func TestFinalMethodTearOff_classMethod(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
TypedObject.newTypeDef^ >ys
"ClassHost" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^

: ClassHost##forty
  40
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	if globalDictionaryContext.dict.GetByName("ClassHost##forty") != nil {
		t.Fatal("class tear-off left dict word")
	}
	typ := globalDictionaryContext.tr.GetLocal("ClassHost")
	mm := typ.GetClassMethodMetadata("forty")
	if mm == nil || !mm.isFinalMethod {
		t.Fatal("class method forty missing")
	}
	// Invalidate path: touch FinalType before late add already happened at `;`.
	_ = typ.FinalType()
	rp.DataStack.Push(typ)
	runHandleCompoundWithSegs(task, rp, []string{".", "forty", "^"})
	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d", rp.OStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 40 {
		t.Fatalf("TOS = %v, want 40", got)
	}
}

func TestFinalMethodTearOff_rejectsForeignType(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.Parser = NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("NoSuchType#x\n")),
		label: "tearoff-reject",
	}})
	err := bootstrapColonStart(rp.Parser, task, rp)
	if err == nil {
		t.Fatal("want error for unknown type")
	}
	if !strings.Contains(err.Error(), "not owned") && !strings.Contains(err.Error(), "NoSuchType") {
		t.Fatalf("err = %v", err)
	}
}

func TestFinalMethodTearOff_nestedInstanceMethods(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
TypedObject.newTypeDef^ >ys
"NestHost" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^

: NestHost#inner
  7
;

: NestHost#outer
  o:.inner^
  10 +
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	typ := globalDictionaryContext.tr.GetLocal("NestHost")
	rp.DataStack.Push(MakeNewTypedObject(typ))
	runHandleCompoundWithSegs(task, rp, []string{".", "outer", "^"})
	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d, want 0", rp.OStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 17 {
		t.Fatalf("TOS = %v, want 17", got)
	}
}

func TestInvalidateClassTypeCache(t *testing.T) {
	ResetTestRuntime()
	typ := &Type{name: "CacheHost"}
	facet := typ.FinalType()
	if facet == nil {
		t.Fatal("FinalType nil")
	}
	if typ.classType != facet {
		t.Fatal("classType not cached")
	}
	typ.AddClassMethod("m", &MethodMetadata{name: "m", call: func(*RunFrame, *TypeObjPair) {}})
	if typ.classType != nil {
		t.Fatal("AddClassMethod should InvalidateClassTypeCache")
	}
	facet2 := typ.FinalType()
	if facet2.methodMetadata == nil || !facet2.methodMetadata.HasKey("m") {
		t.Fatal("rebuilt facet missing class method m")
	}
}
