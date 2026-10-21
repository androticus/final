package core

import "testing"

func TestParserTypeRegisteredEmpty(t *testing.T) {
	ResetTestRuntime()
	pT := GetTypeRegistrar().Get("Parser")
	if pT == nil {
		t.Fatal("missing Parser type")
	}
	if !pT.HasProperty("lastComment") {
		t.Fatal("Parser missing lastComment property")
	}
	if pT.MethodByName("getClearLastComment") == nil {
		t.Fatal("Parser missing getClearLastComment method")
	}
}

func TestParserLastCommentRegistration(t *testing.T) {
	ResetTestRuntime()
	pT := GetTypeRegistrar().Get("Parser")
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(mustTextReader(t, "// hello\n"))
	_, _, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	ths := &TypeObjPair{Obj: p, Typ: pT}
	f := &RunFrame{RP: rp}
	pT.PropertyByName("lastComment").invokeGet(f, ths)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("Parser.lastComment = %v, want hello", got)
	}
	pT.MethodByName("getClearLastComment").invokeCall(f, ths)
	got, ok = rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("getClearLastComment = %v, want hello", got)
	}
	if p.LastComment() != "" {
		t.Fatalf("LastComment after getClearLastComment = %q", p.LastComment())
	}
}

func mustTextReader(t *testing.T, s string) *TextReader {
	t.Helper()
	tr, err := NewTextReaderFromString(s)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

func TestTokenTypeClassConstants(t *testing.T) {
	ResetTestRuntime()
	tt := GetTypeRegistrar().Get("TokenType")
	if tt == nil {
		t.Fatal("missing TokenType type")
	}
	facet := tt.FinalType()
	want := map[string]Int{
		"TokenEOF":           Int(TokenEOF),
		"TokenIdent":         Int(TokenIdent),
		"TokenCompoundIdent": Int(TokenCompoundIdent),
		"TokenInt":           Int(TokenInt),
		"TokenUint":          Int(TokenUint),
		"TokenFloat":         Int(TokenFloat),
		"TokenBool":          Int(TokenBool),
		"TokenString":        Int(TokenString),
		"TokenRawString":     Int(TokenRawString),
		"TokenRune":          Int(TokenRune),
		"TokenLiteral": Int(TokenLiteral),
		"TokenNull":          Int(TokenNull),
		"TokenComment":       Int(TokenComment),
		"TokenNewline":       Int(TokenNewline),
	}
	if facet.propertyMetadata.Len() != len(want) {
		t.Fatalf("TokenType facet property count = %d, want %d", facet.propertyMetadata.Len(), len(want))
	}
	rp := NewTestRunProvider()
	for name, wantVal := range want {
		if !facet.HasProperty(name) {
			t.Fatalf("TokenType facet missing class constant property %q", name)
		}
		pm := facet.PropertyByName(name)
		if pm.set != nil {
			t.Fatalf("TokenType.%s should be readonly", name)
		}
		f, ths := memberRunFrame(rp, tt)
		pm.invokeGet(f, ths)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != wantVal {
			t.Fatalf("TokenType.%s = %v, want %v", name, got, wantVal)
		}
	}
}

func TestTokenTypeDictionaryWord(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"Parser", "TokenType"} {
		if globalDictionaryContext.dict.GetByName(name) == nil {
			t.Fatalf("missing dictionary type word %q", name)
		}
	}
}

func TestSessionTokenTypeTokenEOF(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	res := s.RunScript("TokenType .TokenEOF")
	if res.Status != RunCompleted {
		t.Fatalf("status = %v fault=%v err=%v", res.Status, res.Fault, res.Err)
	}
}

func TestSessionTokenTypeTokenEOFCompoundElided(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	cp := CompoundPathForTest([]string{"TokenType", ".", "TokenEOF"})
	if err := dispatchCompoundToken(p, task, rp, cp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (terminal only)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != Int(TokenEOF) {
		t.Fatalf("TOS = %v, want Int(TokenEOF)", got)
	}
}

func TestSessionTokenTypeTokenEOFFusedCompound(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	res := s.RunScript("TokenType.TokenEOF")
	if res.Status != RunCompleted {
		t.Fatalf("status = %v fault=%v err=%v", res.Status, res.Fault, res.Err)
	}
}
