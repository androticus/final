package core

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestBeginTransactionTwicePanics(t *testing.T) {
	ResetTestRuntime()
	cs := NewCompilerState(globalDictionaryContext)
	cs.beginTransaction()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nested beginTransaction")
		}
	}()
	cs.beginTransaction()
}

func TestEndTransactionNoOpWhenInactive(t *testing.T) {
	ResetTestRuntime()
	cs := NewCompilerState(globalDictionaryContext)
	cs.endTransaction(false)
	cs.endTransaction(true)
}

func TestUnitRollbackRemovesPartialColonWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	cs := task.compilerForTarget()
	dictLenBefore := len(globalDictionaryContext.dict.byOrdinal)
	cs.beginTransaction()
	bootstrapCreate(cs, "partial")
	if !globalDictionaryContext.dict.IsBuilding() {
		t.Fatal("word should be on building stack after :")
	}
	if globalDictionaryContext.dict.GetByName("partial") != nil {
		t.Fatal("building word should not be in dictionary until ;")
	}
	cs.rollbackTransaction()
	if globalDictionaryContext.dict.IsBuilding() {
		t.Fatal("building stack should be clear after rollback")
	}
	if globalDictionaryContext.dict.GetByName("partial") != nil {
		t.Fatal("word should be gone after rollback")
	}
	if len(globalDictionaryContext.dict.byOrdinal) != dictLenBefore {
		t.Fatalf("dict len = %d, want %d", len(globalDictionaryContext.dict.byOrdinal), dictLenBefore)
	}
	if GetCompileState(cs) {
		t.Fatal("STATE should be false after rollback")
	}
}

func TestRollbackSetsCompileStateFalse(t *testing.T) {
	ResetTestRuntime()
	cs := NewCompilerState(globalDictionaryContext)
	cs.beginTransaction()
	bootstrapCreate(cs, "x")
	cs.rollbackTransaction()
	if GetCompileState(cs) {
		t.Fatal("STATE should be false")
	}
}

func TestCompileWithTransactionCommitKeepsWords(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: ok ;`)
	if err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("ok") == nil {
		t.Fatal("ok should be defined after successful compile")
	}
	cs := task.compilerForTarget()
	if cs.txnActiveForTest() {
		t.Fatal("txn should be inactive after commit")
	}
}

func TestCompileWithTransactionErrorRollsBack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dictBefore := len(globalDictionaryContext.dict.byOrdinal)
	err := compileBootstrapSource(task, rp, `: bad nosuchword ;`)
	if err == nil {
		t.Fatal("expected compile error")
	}
	if globalDictionaryContext.dict.GetByName("bad") != nil {
		t.Fatal("partial word should be rolled back")
	}
	if len(globalDictionaryContext.dict.byOrdinal) != dictBefore {
		t.Fatalf("dict len = %d, want %d", len(globalDictionaryContext.dict.byOrdinal), dictBefore)
	}
}

func TestCompileWithTransactionUnfinishedColonRollsBack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dictBefore := len(globalDictionaryContext.dict.byOrdinal)
	err := compileBootstrapSource(task, rp, `: unfinished dup`)
	if err == nil {
		t.Fatal("expected unfinished colon error")
	}
	if globalDictionaryContext.dict.GetByName("unfinished") != nil {
		t.Fatal("unfinished word should be rolled back")
	}
	if len(globalDictionaryContext.dict.byOrdinal) != dictBefore {
		t.Fatalf("dict len = %d, want %d", len(globalDictionaryContext.dict.byOrdinal), dictBefore)
	}
}

func TestDrainParserToEOF(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("line1\nline2\n")),
		label: "test",
	}})
	_, _, _ = p.Next()
	drainParserToEOF(p)
	if !p.InputEOF() {
		t.Fatal("parser should be at EOF after drain")
	}
}

func TestWarmBootResetClearsStateAndTxn(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	cs.beginTransaction()
	SetCompileState(cs, true)
	warmBootReset(task, rp)
	if GetCompileState(cs) {
		t.Fatal("STATE should be false after warmBootReset")
	}
	if cs.txnActiveForTest() {
		t.Fatal("txn should be inactive after warmBootReset")
	}
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("byeReq should be false after warmBootReset")
	}
}

func TestDictionaryRollbackToRemovesNames(t *testing.T) {
	ResetTestRuntime()
	d := globalDictionaryContext.dict
	start := len(d.byOrdinal)
	w := NewWord("rollbackme", WordFlagColon)
	d.Add(w)
	if d.GetByName("rollbackme") == nil {
		t.Fatal("word should exist")
	}
	d.RollbackTo(start)
	if d.GetByName("rollbackme") != nil {
		t.Fatal("name should be removed from byName")
	}
	if len(d.byOrdinal) != start {
		t.Fatalf("len = %d, want %d", len(d.byOrdinal), start)
	}
}

func compileLocalUnit(t *testing.T, task *Task, rp *RunProvider, src string) error {
	t.Helper()
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	return compileWithTransaction(task, rp, func() error {
		return interpretSession(task, rp)
	})
}

func typedObjectMakeTypeSrc(typeName string) string {
	return fmt.Sprintf(`
TypedObject.newTypeDef^ >ys
"%s" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
ys> TypedObject.makeType^
`, typeName)
}

func TestCompileWithTransactionRollbackDropsMakeType(t *testing.T) {
	_, task, rp := localForgetSession(t)
	dc := task.dc
	const typeName = "TxnRbType"
	src := typedObjectMakeTypeSrc(typeName) + `: boom nosuchword ;`
	err := compileLocalUnit(t, task, rp, src)
	if err == nil {
		t.Fatal("expected compile error after makeType")
	}
	if dc.tr.GetLocal(typeName) != nil {
		t.Fatal("type should be gone after unit rollback")
	}
	if dc.dict.GetByName(typeName) != nil {
		t.Fatal("type proxy word should be gone after unit rollback")
	}

	err = compileLocalUnit(t, task, rp, typedObjectMakeTypeSrc(typeName))
	if err != nil {
		t.Fatalf("reload makeType: %v", err)
	}
	if dc.tr.GetLocal(typeName) == nil {
		t.Fatal("type should exist after successful reload")
	}
}

func TestCompileWithTransactionCommitKeepsMakeType(t *testing.T) {
	_, task, rp := localForgetSession(t)
	const typeName = "TxnKeepType"
	if err := compileLocalUnit(t, task, rp, typedObjectMakeTypeSrc(typeName)); err != nil {
		t.Fatal(err)
	}
	if task.dc.tr.GetLocal(typeName) == nil {
		t.Fatal("type should remain after commit")
	}
}

func TestCompileWithTransactionRollbackRefusesSurvivingFinalMethod(t *testing.T) {
	_, task, rp := localForgetSession(t)
	const typeName = "SurvTxnType"
	if err := compileLocalUnit(t, task, rp, typedObjectMakeTypeSrc(typeName)); err != nil {
		t.Fatal(err)
	}
	typ := task.dc.tr.GetLocal(typeName)
	if typ == nil {
		t.Fatal("type missing after commit")
	}
	dictBefore := task.dc.dict.Len()
	ctBefore := task.dc.ct.Len()
	typeBefore := task.dc.tr.Len()

	err := compileLocalUnit(t, task, rp, fmt.Sprintf(`
: %s#m
  1
;
: boom nosuchword ;
`, typeName))
	if err == nil {
		t.Fatal("expected compile-unit rollback refuse for surviving Final method")
	}
	msg := err.Error()
	if !strings.Contains(msg, typeName) || !strings.Contains(msg, "refused") {
		t.Fatalf("err = %v, want %s refused", err, typeName)
	}
	if task.dc.dict.Len() != dictBefore {
		t.Fatalf("dict rolled back on refuse: %d want %d", task.dc.dict.Len(), dictBefore)
	}
	if task.dc.ct.Len() < ctBefore {
		t.Fatalf("CT truncated on refuse: %d < keep %d", task.dc.ct.Len(), ctBefore)
	}
	if task.dc.tr.Len() != typeBefore {
		t.Fatalf("type catalog mutated on refuse: %d want %d", task.dc.tr.Len(), typeBefore)
	}
	if !typ.HasMethod("m") {
		t.Fatal("surviving type method should remain after refused rollback")
	}
	cs := task.compilerForTarget()
	if cs.txnActiveForTest() {
		t.Fatal("txn should be inactive after refuse")
	}
	if GetCompileState(cs) {
		t.Fatal("STATE should be false after refuse")
	}
	if err := compileLocalUnit(t, task, rp, `: okAfterRefuse 1 ;`); err != nil {
		t.Fatalf("session should compile after refuse: %v", err)
	}
}
