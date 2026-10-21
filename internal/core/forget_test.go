package core

import (
	"fmt"
	"strings"
	"testing"
)

func localForgetSession(t *testing.T) (*Session, *Task, *RunProvider) {
	t.Helper()
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	task := s.task
	rp := NewTestRunProvider()
	return s, task, rp
}

func interpretLocal(t *testing.T, task *Task, rp *RunProvider, src string) {
	t.Helper()
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	if err := interpretSession(task, rp); err != nil {
		t.Fatalf("interpret: %v\n%s", err, src)
	}
}

func TestForgetInclusiveCutAndReload(t *testing.T) {
	_, task, rp := localForgetSession(t)
	dc := task.dc

	interpretLocal(t, task, rp, `
: stable 1 ;
forget: reload
: scratch 2 ;
`)
	if dc.dict.GetByName("stable") == nil {
		t.Fatal("stable missing before forget")
	}
	if dc.dict.GetByName("reload") == nil {
		t.Fatal("reload mark missing")
	}
	if dc.dict.GetByName("scratch") == nil {
		t.Fatal("scratch missing")
	}
	dictKeep := dc.dict.Len()
	ctKeep := dc.ct.Len()
	dtKeep := dc.dt.Len()
	typeKeep := dc.tr.Len()

	interpretLocal(t, task, rp, `reload`)

	if dc.dict.GetByName("reload") != nil {
		t.Fatal("reload mark should be gone after inclusive forget")
	}
	if dc.dict.GetByName("scratch") != nil {
		t.Fatal("scratch should be gone after forget")
	}
	if dc.dict.GetByName("stable") == nil {
		t.Fatal("stable should survive forget")
	}
	if dc.dict.Len() >= dictKeep {
		t.Fatalf("dict len %d should be < pre-forget %d", dc.dict.Len(), dictKeep)
	}
	if dc.ct.Len() >= ctKeep {
		t.Fatalf("ct len %d should be < pre-forget %d", dc.ct.Len(), ctKeep)
	}
	_ = dtKeep
	_ = typeKeep

	// Second load recreates mark + scratch
	interpretLocal(t, task, rp, `
forget: reload
: scratch 3 ;
`)
	if dc.dict.GetByName("reload") == nil || dc.dict.GetByName("scratch") == nil {
		t.Fatal("reload/scratch missing after second load")
	}
	interpretLocal(t, task, rp, `scratch`)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("scratch = %v, want 3", got)
	}
}

func TestForgetInterpretOnly(t *testing.T) {
	_, task, rp := localForgetSession(t)
	interpretLocal(t, task, rp, `forget: reload`)
	tr, err := NewTextReaderFromString(`: bad reload ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	err = interpretSession(task, rp)
	if err == nil {
		t.Fatal("expected compile of interpret-only mark to fail")
	}
	if !strings.Contains(err.Error(), "interpret-only") {
		t.Fatalf("err = %v, want interpret-only", err)
	}
}

func TestForgetSealedRefuse(t *testing.T) {
	_, task, rp := localForgetSession(t)
	task.dc.Seal()
	tr, err := NewTextReaderFromString(`forget: reload`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	err = interpretSession(task, rp)
	if err == nil {
		t.Fatal("expected forget: on sealed DC to fail")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "sealed") {
		t.Fatalf("err = %v, want sealed", err)
	}
}

func TestForgetFinalMethodGuard(t *testing.T) {
	_, task, rp := localForgetSession(t)
	interpretLocal(t, task, rp, `
TypedObject.newTypeDef^ >ys
"SurvType" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^

forget: reload

: SurvType#m
  1
;
`)
	dictBefore := task.dc.dict.Len()
	ctBefore := task.dc.ct.Len()
	typeBefore := task.dc.tr.Len()

	tr, err := NewTextReaderFromString(`reload`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	err = interpretSession(task, rp)
	if err == nil {
		t.Fatal("expected forget refuse for surviving Final method")
	}
	msg := err.Error()
	if !strings.Contains(msg, "SurvType") || !strings.Contains(msg, "refused") {
		t.Fatalf("err = %v, want SurvType refused", err)
	}
	if task.dc.dict.Len() != dictBefore || task.dc.ct.Len() != ctBefore || task.dc.tr.Len() != typeBefore {
		t.Fatalf("lens mutated on refuse: dict %d/%d ct %d/%d type %d/%d",
			task.dc.dict.Len(), dictBefore, task.dc.ct.Len(), ctBefore, task.dc.tr.Len(), typeBefore)
	}
	if task.dc.dict.GetByName("reload") == nil {
		t.Fatal("mark should remain after refuse")
	}
}

func TestForgetHappyPathTypeUnderMark(t *testing.T) {
	_, task, rp := localForgetSession(t)
	interpretLocal(t, task, rp, `
forget: reload
TypedObject.newTypeDef^ >ys
"GoneType" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^
: GoneType#m 7 ;
`)
	if task.dc.tr.GetLocal("GoneType") == nil {
		t.Fatal("GoneType missing before forget")
	}
	interpretLocal(t, task, rp, `reload`)
	if task.dc.tr.GetLocal("GoneType") != nil {
		t.Fatal("GoneType should be gone after forget")
	}
	if task.dc.dict.GetByName("GoneType") != nil {
		t.Fatal("GoneType word should be gone after forget")
	}
	if task.dc.dict.GetByName("reload") != nil {
		t.Fatal("mark should be gone")
	}
}

func TestForgetPoisonsOutstandingMethodRef(t *testing.T) {
	_, task, rp := localForgetSession(t)
	interpretLocal(t, task, rp, `
forget: reload
TypedObject.newTypeDef^ >ys
"GoneType" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^
: GoneType#m 7 ;
`)
	typ := task.dc.tr.GetLocal("GoneType")
	if typ == nil {
		t.Fatal("GoneType missing before forget")
	}
	ref := NewMethodRef(MakeNewTypedObject(typ), "m")

	rp.DataStack.Clear()
	rp.DataStack.Push(ref)
	RunOps(task, rp, OpMethodInvoke, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("before forget: %v, want 7", got)
	}

	interpretLocal(t, task, rp, `reload`)
	if task.dc.tr.GetLocal("GoneType") != nil {
		t.Fatal("GoneType should be gone after forget")
	}

	expectForgottenInvoke(t, task, rp, ref, "GoneType#m")

	interpretLocal(t, task, rp, `
forget: reload
TypedObject.newTypeDef^ >ys
"GoneType" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^
: GoneType#m 9 ;
`)
	newTyp := task.dc.tr.GetLocal("GoneType")
	if newTyp == nil {
		t.Fatal("GoneType missing after reload")
	}
	newRef := NewMethodRef(MakeNewTypedObject(newTyp), "m")
	rp.DataStack.Clear()
	rp.DataStack.Push(newRef)
	RunOps(task, rp, OpMethodInvoke, OpStop)
	got, ok = rp.DataStack.Pop().(Int)
	if !ok || got != 9 {
		t.Fatalf("reloaded method: %v, want 9", got)
	}

	expectForgottenInvoke(t, task, rp, ref, "GoneType#m")
}

func expectForgottenInvoke(t *testing.T, task *Task, rp *RunProvider, ref *MethodRef, wantName string) {
	t.Helper()
	rp.DataStack.Clear()
	rp.DataStack.Push(ref)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("want forgotten MethodRef panic")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "forgotten") || !strings.Contains(msg, wantName) {
			t.Fatalf("panic = %q, want forgotten %s", msg, wantName)
		}
	}()
	RunOps(task, rp, OpMethodInvoke, OpStop)
}

func TestParseLeadingForgetMarkLine(t *testing.T) {
	cases := []struct {
		line string
		name string
		ok   bool
	}{
		{"forget: reload", "reload", true},
		{"forget: reload\n", "reload", true},
		{"forget: reload\r\n", "reload", true},
		{"forget: reset-invaders", "reset-invaders", true},
		{" forget: reload", "", false},
		{"forget: reload ", "", false},
		{"forget:  reload", "", false},
		{"forget:reload", "", false},
		{"forget: reload // x", "", false},
		{"requiresFyne", "", false},
		{"", "", false},
		{"forget: ", "", false},
		{"forget: reload extra", "", false},
	}
	for _, tc := range cases {
		got, ok := parseLeadingForgetMarkLine(tc.line)
		if ok != tc.ok || got != tc.name {
			t.Errorf("parseLeadingForgetMarkLine(%q) = %q %v, want %q %v",
				tc.line, got, ok, tc.name, tc.ok)
		}
	}
}

func TestForgetMarkWordFlags(t *testing.T) {
	_, task, rp := localForgetSession(t)
	interpretLocal(t, task, rp, `forget: reload`)
	w := task.dc.dict.GetByName("reload")
	if w == nil {
		t.Fatal("missing reload")
	}
	if w.flags&WordFlagInterpretOnly == 0 {
		t.Fatal("mark should be InterpretOnly")
	}
	if w.flags&WordFlagColon == 0 {
		t.Fatal("mark should be Colon")
	}
	if fmt.Sprint(w.colonCode) == "" {
		t.Fatal("mark needs colonCode")
	}
}
