package core

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func parseLineTokens(t *testing.T, line string) []Token {
	t.Helper()
	p := NewParser(nil)
	tokens, stillOpen, err := p.ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine(%q): %v", line, err)
	}
	if stillOpen {
		t.Fatalf("ParseLine(%q): stillOpen", line)
	}
	if len(p.errors) != 0 {
		t.Fatalf("ParseLine(%q): errors %v", line, p.errors)
	}
	return tokens
}

func parseLineError(t *testing.T, line string) {
	t.Helper()
	p := NewParser(nil)
	_, _, err := p.ParseLine(line)
	if err == nil {
		t.Fatalf("ParseLine(%q): want error", line)
	}
}

func expectCompoundPath(t *testing.T, val any, want ...string) {
	t.Helper()
	cp, ok := val.(*compoundPathValue)
	if !ok {
		t.Fatalf("compound value type %T, want *compoundPathValue", val)
	}
	wantPath := strings.Join(want, " ")
	if cp.path != String(wantPath) {
		t.Fatalf("compound path = %q, want %q", cp.path, wantPath)
	}
}

func TestPushNextStreamEOF(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("")),
		label: "test",
	}})
	ds := NewStack[any]("test", 8)
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if !StreamEOFOnStack(ds) {
		t.Fatalf("stack = %v, want lone TokenEOF", ds)
	}
}

func TestPushNextEmitsNewlineOnEmptyLines(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("\n\n99\n")),
		label: "test",
	}})
	ds := NewStack[any]("test", 8)
	for range 2 {
		if err := pushNextToken(p, ds); err != nil {
			t.Fatal(err)
		}
		if TokenType(ds.Pop().(Int)) != TokenNewline {
			t.Fatal("want TokenNewline")
		}
		ds.Pop() // nil value
	}
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if !HasTokenFlag(TokenType(ds.Pop().(Int)), TokenInt) {
		t.Fatal("want TokenInt")
	}
	if ds.Pop().(Int) != 99 {
		t.Fatal("want 99")
	}
}

func TestParseNewlineToken(t *testing.T) {
	tokens := parseLineTokens(t, "a\nb")
	if len(tokens) != 3 {
		t.Fatalf("tokens = %d, want 3 (a, newline, b)", len(tokens))
	}
	if tokens[0].Type != TokenIdent || tokens[0].Value != String("a") {
		t.Fatalf("token[0] = %+v", tokens[0])
	}
	if tokens[1].Type != TokenNewline {
		t.Fatalf("token[1] = %+v, want TokenNewline", tokens[1])
	}
	if tokens[2].Type != TokenIdent || tokens[2].Value != String("b") {
		t.Fatalf("token[2] = %+v", tokens[2])
	}
}

func TestParserLineIncrementsOnNewline(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("a\nb\n")),
		label: "test",
	}})
	if p.line != 1 {
		t.Fatalf("initial line = %d, want 1", p.line)
	}
	ds := NewStack[any]("test", 8)
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if p.line != 1 {
		t.Fatalf("after first token line = %d, want 1", p.line)
	}
	ds.Pop()
	ds.Pop()
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if p.line != 2 {
		t.Fatalf("after first newline line = %d, want 2", p.line)
	}
	ds.Pop()
	ds.Pop()
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if p.line != 2 {
		t.Fatalf("after second-line token line = %d, want 2", p.line)
	}
}

func TestSkipParserLineDrainToEndOfLine(t *testing.T) {
	tr, err := NewTextReaderFromString("a b\nc d")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	if _, _, err := p.Next(); err != nil {
		t.Fatal(err)
	}
	skipParserLine(p)
	val, typ, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if typ != TokenIdent || val != String("c") {
		t.Fatalf("after skip got %v %v, want ident c", val, typ)
	}
}

func TestSkipParserLineDoesNotDrainMultiLineStream(t *testing.T) {
	tr, err := NewTextReaderFromString("bad rest\nok\nbye")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	if _, _, err := p.Next(); err != nil {
		t.Fatal(err)
	}
	skipParserLine(p)
	val, typ, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if typ != TokenIdent || val != String("ok") {
		t.Fatalf("after skip got %v %v, want ident ok", val, typ)
	}
}

func TestParseLineDsDollarIsSimpleIdent(t *testing.T) {
	tokens := parseLineTokens(t, "1 2 ds$")
	if len(tokens) != 3 {
		t.Fatalf("len = %d", len(tokens))
	}
	if !IsPlainIdentToken(tokens[2].Type) || tokens[2].Value != String("ds$") {
		t.Fatalf("token[2] = %+v, want plain ident ds$", tokens[2])
	}
}

func TestParseLineHyphenatedIdent(t *testing.T) {
	tokens := parseLineTokens(t, ": test-if 1 - 2 ;")
	if !IsPlainIdentToken(tokens[1].Type) || tokens[1].Value != String("test-if") {
		t.Fatalf("token[1] = %+v, want ident test-if", tokens[1])
	}
	if !IsPlainIdentToken(tokens[3].Type) || tokens[3].Value != String("-") {
		t.Fatalf("token[3] = %+v, want operator -", tokens[3])
	}
}

func TestParseLineColonSemicolon(t *testing.T) {
	tokens := parseLineTokens(t, ": inc 1 + ;")
	want := []struct {
		typ TokenType
		val any
	}{
		{TokenIdent, String(":")},
		{TokenIdent, String("inc")},
		{LiteralTokenType(TokenInt), Int(1)},
		{TokenIdent, String("+")},
		{TokenIdent, String(";")},
	}
	if len(tokens) != len(want) {
		t.Fatalf("len(tokens) = %d, want %d", len(tokens), len(want))
	}
	for i, w := range want {
		if tokens[i].Type != w.typ {
			t.Fatalf("token[%d] type = %v, want %v", i, tokens[i].Type, w.typ)
		}
		if tokens[i].Value != w.val {
			t.Fatalf("token[%d] value = %v, want %v", i, tokens[i].Value, w.val)
		}
	}
}

func TestParseLineDotAloneIsError(t *testing.T) {
	parseLineError(t, ".")
}

func TestParseLineDotFiveIsFloat(t *testing.T) {
	tokens := parseLineTokens(t, ".5 3.14")
	if len(tokens) != 2 || !HasTokenFlag(tokens[0].Type, TokenFloat) || !HasTokenFlag(tokens[1].Type, TokenFloat) {
		t.Fatalf("tokens = %+v", tokens)
	}
}

func TestParseLineLeadingDotFieldIsCompound(t *testing.T) {
	tokens := parseLineTokens(t, ".field")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, ".", "field")
}

func TestParseLineLocalsRootCompound(t *testing.T) {
	tokens := parseLineTokens(t, "l.v!")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "l", ".", "v", "!")
	cp := tokens[0].Value.(*compoundPathValue)
	if cp.stackSelector != compoundPathNoSelector {
		t.Fatalf("stackSelector = %v, want no selector (l is not t:)", cp.stackSelector)
	}
}

func TestParseLinePlsRootCompound(t *testing.T) {
	tokens := parseLineTokens(t, "pls.v!")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "pls", ".", "v", "!")
	cp := tokens[0].Value.(*compoundPathValue)
	if cp.stackSelector != compoundPathNoSelector {
		t.Fatalf("stackSelector = %v, want no selector (pls is not t:)", cp.stackSelector)
	}
}

func TestParseLineDotMethodCaretIsCompound(t *testing.T) {
	tokens := parseLineTokens(t, ".findWordByName^")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, ".", "findWordByName", "^")
}

func TestParseLineDotMethodTildeIsCompound(t *testing.T) {
	tokens := parseLineTokens(t, ".findWordByName~")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, ".", "findWordByName", "~")
}

func TestParseLineDSelectorDepthCompound(t *testing.T) {
	tokens := parseLineTokens(t, `d:depth`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "depth")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csDataStack {
		t.Fatalf("stackSelector = %v, want csDataStack", cp.stackSelector)
	}
}

func TestParseLineDSelectorAloneEmptyPathCompound(t *testing.T) {
	tokens := parseLineTokens(t, "d:")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csDataStack || cp.path != "" {
		t.Fatalf("compound = selector %v path %q, want csDataStack and empty path", cp.stackSelector, cp.path)
	}
}

func TestParseLineTSelectorDropCompound(t *testing.T) {
	tokens := parseLineTokens(t, `t:drop`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "drop")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csTStack {
		t.Fatalf("stackSelector = %v, want csTStack", cp.stackSelector)
	}
}

func TestParseLineTSelectorDotNameFetchCompound(t *testing.T) {
	tokens := parseLineTokens(t, `t:hello.name@`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "hello", ".", "name", "@")
}

func TestParseLineTSelectorAloneEmptyPathCompound(t *testing.T) {
	tokens := parseLineTokens(t, "t:")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csTStack || cp.path != "" {
		t.Fatalf("compound = selector %v path %q, want csTStack and empty path", cp.stackSelector, cp.path)
	}
}

func TestParseLineXSelectorDropCompound(t *testing.T) {
	tokens := parseLineTokens(t, `x:drop`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "drop")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csXStack {
		t.Fatalf("stackSelector = %v, want csXStack", cp.stackSelector)
	}
}

func TestParseLineYSelectorDropCompound(t *testing.T) {
	tokens := parseLineTokens(t, `y:drop`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "drop")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csYStack {
		t.Fatalf("stackSelector = %v, want csYStack", cp.stackSelector)
	}
}

func TestParseLineTSelectorDollarCompound(t *testing.T) {
	tokens := parseLineTokens(t, "t:$")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "$")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csTStack {
		t.Fatalf("stackSelector = %v, want csTStack", cp.stackSelector)
	}
}

func TestParseLineXSelectorAloneEmptyPathCompound(t *testing.T) {
	tokens := parseLineTokens(t, "x:")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csXStack || cp.path != "" {
		t.Fatalf("compound = selector %v path %q, want csXStack and empty path", cp.stackSelector, cp.path)
	}
}

func TestParseLineYSelectorAloneEmptyPathCompound(t *testing.T) {
	tokens := parseLineTokens(t, "y:")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csYStack || cp.path != "" {
		t.Fatalf("compound = selector %v path %q, want csYStack and empty path", cp.stackSelector, cp.path)
	}
}

func TestParseLineZSelectorDropCompound(t *testing.T) {
	tokens := parseLineTokens(t, `z:drop`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "drop")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csZStack {
		t.Fatalf("stackSelector = %v, want csZStack", cp.stackSelector)
	}
}

func TestParseLineZSelectorAloneEmptyPathCompound(t *testing.T) {
	tokens := parseLineTokens(t, "z:")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csZStack || cp.path != "" {
		t.Fatalf("compound = selector %v path %q, want csZStack and empty path", cp.stackSelector, cp.path)
	}
}

func TestParseLineOSelectorDropCompound(t *testing.T) {
	tokens := parseLineTokens(t, `o:drop`)
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "drop")
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csOStack {
		t.Fatalf("stackSelector = %v, want csOStack", cp.stackSelector)
	}
}

func TestParseLineOSelectorAloneEmptyPathCompound(t *testing.T) {
	tokens := parseLineTokens(t, "o:")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want compound", tokens[0].Type)
	}
	cp, ok := tokens[0].Value.(*compoundPathValue)
	if !ok || cp.stackSelector != csOStack || cp.path != "" {
		t.Fatalf("compound = selector %v path %q, want csOStack and empty path", cp.stackSelector, cp.path)
	}
}

func TestParseLineEmptyIndexFieldCompound(t *testing.T) {
	tokens := parseLineTokens(t, "().field")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "()", ".", "field")
}

func TestParseLineFQNToken(t *testing.T) {
	tokens := parseLineTokens(t, "math::sin")
	if !HasTokenFlag(tokens[0].Type, TokenIdent) || HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want plain Ident", tokens[0].Type)
	}
	if tokens[0].Value != String("math::sin") {
		t.Fatalf("value = %v", tokens[0].Value)
	}
}

func TestParseLineMultiSegmentFQN(t *testing.T) {
	tokens := parseLineTokens(t, "math::util::sin")
	if !HasTokenFlag(tokens[0].Type, TokenIdent) || HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v, want plain Ident", tokens[0].Type)
	}
}

func TestParseLineCompoundNamespaceSegment(t *testing.T) {
	tokens := parseLineTokens(t, "foo::bar.baz")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("type = %v", tokens[0].Type)
	}
	expectCompoundPath(t, tokens[0].Value, "foo::bar", ".", "baz")
}

func TestParseLineCompoundFieldPath(t *testing.T) {
	tokens := parseLineTokens(t, "foo.bar.baz")
	expectCompoundPath(t, tokens[0].Value, "foo", ".", "bar", ".", "baz")
}

func TestParseLineCompoundIndexSuffix(t *testing.T) {
	tokens := parseLineTokens(t, "myList()")
	expectCompoundPath(t, tokens[0].Value, "myList", "()")
}

func TestParseLineCompoundAffixes(t *testing.T) {
	tokens := parseLineTokens(t, "myObj.field@")
	expectCompoundPath(t, tokens[0].Value, "myObj", ".", "field", "@")

	tokens = parseLineTokens(t, "counter!")
	expectCompoundPath(t, tokens[0].Value, "counter", "!")
}

func TestParseLineBackslashNotCompoundAffix(t *testing.T) {
	parseLineError(t, `myObj.field!\`)

	tokens := parseLineTokens(t, `\`)
	if len(tokens) != 1 || !IsPlainIdentToken(tokens[0].Type) || tokens[0].Value != String(`\`) {
		t.Fatalf("lone \\ = %+v, want plain TokenIdent", tokens)
	}

	tokens = parseLineTokens(t, `core\dup`)
	if len(tokens) != 1 || !IsPlainIdentToken(tokens[0].Type) || tokens[0].Value != String(`core\dup`) {
		t.Fatalf("core\\dup = %+v, want plain TokenIdent lookup form", tokens)
	}
}

func TestParseLineCaretMustBeTerminal(t *testing.T) {
	parseLineError(t, "w.name^.other")
	parseLineError(t, "dc.dict.foo^.bar")
}

func TestParseLineTildeMustBeTerminal(t *testing.T) {
	parseLineError(t, "w.name~@")
	parseLineError(t, "t:hello.name~.other")
}

func TestParseLineTrailingDotErrors(t *testing.T) {
	parseLineError(t, "foo.bar.")
}

func TestParseLineLiteralMemberNamesInCompound(t *testing.T) {
	for _, line := range []string{"myObject.null", "t:.Null!"} {
		tokens := parseLineTokens(t, line)
		if len(tokens) != 1 || !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
			t.Fatalf("%q: type = %v, want single compound", line, tokens[0].Type)
		}
	}
}

func TestParseLineStandaloneDoubleColonErrors(t *testing.T) {
	parseLineError(t, "::")
	parseLineError(t, "::foo")
}

func TestParseLineEmptyParensWordIsSimpleIdent(t *testing.T) {
	tokens := parseLineTokens(t, "()")
	if !IsPlainIdentToken(tokens[0].Type) || tokens[0].Value != String("()") {
		t.Fatalf("token = %+v", tokens[0])
	}
}

func TestParseLineHiddenWordStillSimple(t *testing.T) {
	tokens := parseLineTokens(t, "(do)")
	if !IsPlainIdentToken(tokens[0].Type) || tokens[0].Value != String("(do)") {
		t.Fatalf("token = %+v", tokens[0])
	}
}

func TestParseLineLeadingIndexCompound(t *testing.T) {
	tokens := parseLineTokens(t, "().field")
	if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
		t.Fatalf("().field type = %v, want compound", tokens[0].Type)
	}
}

func TestParseLineRenamedArmAndMarkWorkers(t *testing.T) {
	tokens := parseLineTokens(t, "enqueueFromMark fillSetFromMark armCs armSubCode")
	for i, want := range []string{"enqueueFromMark", "fillSetFromMark", "armCs", "armSubCode"} {
		if !IsPlainIdentToken(tokens[i].Type) || tokens[i].Value != String(want) {
			t.Fatalf("token[%d] = %+v, want ident %s", i, tokens[i], want)
		}
	}
}

func TestParseLineCommaInSimpleIdent(t *testing.T) {
	tokens := parseLineTokens(t, "literal, drop ,")
	if !IsPlainIdentToken(tokens[0].Type) || tokens[0].Value != String("literal,") {
		t.Fatalf("token[0] = %+v, want plain ident literal,", tokens[0])
	}
	if tokens[1].Value != String("drop") {
		t.Fatalf("token[1] = %+v", tokens[1])
	}
	if tokens[2].Value != String(",") {
		t.Fatalf("token[2] = %+v, want standalone ,", tokens[2])
	}
}

func TestParseLineDoNotSplitHiddenWordAsCompound(t *testing.T) {
	tokens := parseLineTokens(t, "(do) dup")
	if tokens[0].Value != String("(do)") {
		t.Fatalf("token[0] = %+v", tokens[0])
	}
}

func TestParseLineNullLiteral(t *testing.T) {
	tokens := parseLineTokens(t, "null true false")
	if tokens[0].Type != LiteralTokenType(TokenNull) || tokens[0].Value != Any(nil) {
		t.Fatalf("null = %+v", tokens[0])
	}
	if tokens[1].Type != LiteralTokenType(TokenBool) || tokens[1].Value != Bool(true) {
		t.Fatalf("true = %+v", tokens[1])
	}
	if tokens[2].Type != LiteralTokenType(TokenBool) || tokens[2].Value != Bool(false) {
		t.Fatalf("false = %+v", tokens[2])
	}
	for i := range tokens {
		if !HasTokenFlag(tokens[i].Type, TokenLiteral) {
			t.Fatalf("token[%d] missing TokenLiteral: %+v", i, tokens[i])
		}
	}
}

func TestClassifyNameLiteralCaseSensitive(t *testing.T) {
	for _, lit := range []string{"null", "true", "false"} {
		typ, _, err := classifyNameLiteral(lit)
		if err != nil {
			t.Fatalf("%q: %v", lit, err)
		}
		if !HasTokenFlag(typ, TokenLiteral) {
			t.Fatalf("%q: want literal token, got %v", lit, typ)
		}
	}
	for _, tc := range []struct {
		lit  string
		want TokenType
	}{
		{"Null", TokenIdent},
		{"TRUE", TokenIdent},
		{"False", TokenIdent},
		{"NULL", TokenIdent},
	} {
		typ, val, err := classifyNameLiteral(tc.lit)
		if err != nil {
			t.Fatalf("%q: %v", tc.lit, err)
		}
		if typ != tc.want {
			t.Fatalf("%q: typ = %v, want %v", tc.lit, typ, tc.want)
		}
		if HasTokenFlag(typ, TokenLiteral) {
			t.Fatalf("%q: must not be a literal token", tc.lit)
		}
		if val.(String) != String(tc.lit) {
			t.Fatalf("%q: val = %v, want preserved spelling", tc.lit, val)
		}
		if err := ValidatePlainIdentName(String(tc.lit)); err != nil {
			t.Fatalf("ValidatePlainIdentName(%q) = %v, want nil", tc.lit, err)
		}
	}
}

func TestPushNextAsIdentifierNullTypeName(t *testing.T) {
	p := NewParser(nil)
	p.bindString("Null")
	ds := NewStack[any]("test", 8)
	p.PushNextAsIdentifier(ds)
	if got, ok := ds.Pop().(String); !ok || got != "Null" {
		t.Fatalf("ident = %v, want String(Null)", got)
	}
}

func TestParseLineCommentsEmitted(t *testing.T) {
	tokens := parseLineTokens(t, "// hello")
	if !HasTokenFlag(tokens[0].Type, TokenComment) {
		t.Fatalf("type = %v", tokens[0].Type)
	}
	if tokens[0].Value != String("hello") {
		t.Fatalf("value = %v", tokens[0].Value)
	}
	if tokens[0].Literal != "// hello" {
		t.Fatalf("literal = %q", tokens[0].Literal)
	}

	tokens = parseLineTokens(t, "/* block */ 1")
	if !HasTokenFlag(tokens[0].Type, TokenComment) {
		t.Fatalf("block type = %v", tokens[0].Type)
	}
	if tokens[0].Value != String("block") {
		t.Fatalf("block value = %v", tokens[0].Value)
	}
	if tokens[0].Literal != "/* block */" {
		t.Fatalf("block literal = %q", tokens[0].Literal)
	}
	if !HasTokenFlag(tokens[1].Type, TokenInt) || !HasTokenFlag(tokens[1].Type, TokenLiteral) {
		t.Fatalf("token[1] = %+v", tokens[1])
	}
}

func TestParseLineCommentElision(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"//  hello", "hello"},
		{"//", ""},
		{"//\thello", "hello"},
		{"/*  block  */", "block"},
		{"/**/", ""},
	}
	for _, tc := range cases {
		tokens := parseLineTokens(t, tc.line)
		if len(tokens) != 1 {
			t.Fatalf("ParseLine(%q): tokens = %+v", tc.line, tokens)
		}
		if tokens[0].Value != String(tc.want) {
			t.Fatalf("ParseLine(%q) value = %q, want %q", tc.line, tokens[0].Value, tc.want)
		}
	}
}

func TestParseBlockCommentElisionMultiline(t *testing.T) {
	src := "/* some text\n// more text\n*/"
	p := NewParser(nil)
	tokens, stillOpen, err := p.ParseLine(src)
	if err != nil {
		t.Fatal(err)
	}
	if stillOpen {
		t.Fatal("stillOpen")
	}
	if len(tokens) != 1 {
		t.Fatalf("tokens = %+v", tokens)
	}
	want := "some text\n// more text"
	if tokens[0].Value != String(want) {
		t.Fatalf("value = %q, want %q", tokens[0].Value, want)
	}
}

func TestParseHashColumn1Comment(t *testing.T) {
	tokens := parseLineTokens(t, "# hello")
	if len(tokens) != 1 {
		t.Fatalf("tokens = %+v", tokens)
	}
	if !HasTokenFlag(tokens[0].Type, TokenComment) {
		t.Fatalf("type = %v", tokens[0].Type)
	}
	if tokens[0].Value != String("hello") {
		t.Fatalf("value = %v", tokens[0].Value)
	}
	if tokens[0].Literal != "# hello" {
		t.Fatalf("literal = %q", tokens[0].Literal)
	}
	if tokens[0].Col != 1 {
		t.Fatalf("col = %d, want 1", tokens[0].Col)
	}

	tokens = parseLineTokens(t, "#!/usr/bin/env final")
	if len(tokens) != 1 || !HasTokenFlag(tokens[0].Type, TokenComment) {
		t.Fatalf("shebang tokens = %+v", tokens)
	}
	if tokens[0].Value != String("!/usr/bin/env final") {
		t.Fatalf("shebang value = %v", tokens[0].Value)
	}
	if tokens[0].Literal != "#!/usr/bin/env final" {
		t.Fatalf("shebang literal = %q", tokens[0].Literal)
	}

	tokens = parseLineTokens(t, "#")
	if len(tokens) != 1 || !HasTokenFlag(tokens[0].Type, TokenComment) {
		t.Fatalf("bare # tokens = %+v", tokens)
	}
	if tokens[0].Value != String("") {
		t.Fatalf("bare # value = %v", tokens[0].Value)
	}
}

func TestParseIndentedHashIsIdent(t *testing.T) {
	tokens := parseLineTokens(t, "  #not-shebang")
	if len(tokens) != 1 {
		t.Fatalf("tokens = %+v", tokens)
	}
	if !HasTokenFlag(tokens[0].Type, TokenIdent) {
		t.Fatalf("type = %v, want ident", tokens[0].Type)
	}
	if tokens[0].Value != String("#not-shebang") {
		t.Fatalf("value = %v", tokens[0].Value)
	}
	if tokens[0].Col != 3 {
		t.Fatalf("col = %d, want 3", tokens[0].Col)
	}

	tokens = parseLineTokens(t, "1 # x")
	if len(tokens) != 3 {
		t.Fatalf("inline tokens = %+v", tokens)
	}
	if !HasTokenFlag(tokens[1].Type, TokenIdent) || tokens[1].Value != String("#") {
		t.Fatalf("inline # = %+v, want ident #", tokens[1])
	}
}

func TestParseHashColumn1DoesNotNoteLastComment(t *testing.T) {
	p := NewParser(nil)
	p.bindString("# shebang only\n")
	tok, err := p.nextToken()
	if err != nil {
		t.Fatal(err)
	}
	if !HasTokenFlag(tok.Type, TokenComment) {
		t.Fatalf("type = %v", tok.Type)
	}
	if p.LastComment() != "" {
		t.Fatalf("LastComment after # = %q, want empty", p.LastComment())
	}

	p.bindString("// keep me\n#!/usr/bin/env final\n42")
	sawHash := false
	for {
		tok, err := p.nextToken()
		if err != nil {
			t.Fatal(err)
		}
		if tok.Type == TokenEOF {
			break
		}
		if HasTokenFlag(tok.Type, TokenComment) && strings.HasPrefix(tok.Literal, "#") {
			sawHash = true
			if p.LastComment() != "keep me" {
				t.Fatalf("LastComment after # = %q, want keep me", p.LastComment())
			}
		}
	}
	if !sawHash {
		t.Fatal("missing # comment token")
	}
	if p.LastComment() != "keep me" {
		t.Fatalf("LastComment after stream = %q, want keep me", p.LastComment())
	}
}

func TestParserLastComment(t *testing.T) {
	p := NewParser(nil)
	parseLineTokensWithParser(t, p, "// first")
	if p.LastComment() != "first" {
		t.Fatalf("LastComment = %q, want first", p.LastComment())
	}
	parseLineTokensWithParser(t, p, "// second")
	if p.LastComment() != "second" {
		t.Fatalf("LastComment = %q, want second", p.LastComment())
	}
	if got := p.GetClearLastComment(); got != "second" {
		t.Fatalf("GetClearLastComment = %q", got)
	}
	if p.LastComment() != "" {
		t.Fatalf("LastComment after clear = %q", p.LastComment())
	}
	if got := p.GetClearLastComment(); got != "" {
		t.Fatalf("GetClearLastComment empty = %q", got)
	}
}

func parseLineTokensWithParser(t *testing.T, p *Parser, line string) []Token {
	t.Helper()
	tokens, stillOpen, err := p.ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine(%q): %v", line, err)
	}
	if stillOpen {
		t.Fatalf("ParseLine(%q): stillOpen", line)
	}
	if len(p.errors) > 0 {
		t.Fatalf("ParseLine(%q): errors %v", line, p.errors)
	}
	return tokens
}

func TestParseLineStrayBlockCloseErrors(t *testing.T) {
	parseLineError(t, "*/")
}

func TestParseLineRuneLiterals(t *testing.T) {
	tokens := parseLineTokens(t, "'A' '\\n' '\\u0041'")
	if tokens[0].Type != LiteralTokenType(TokenRune) || tokens[0].Value != Rune('A') {
		t.Fatalf("token0 = %+v", tokens[0])
	}
	if tokens[1].Type != LiteralTokenType(TokenRune) || tokens[1].Value != Rune('\n') {
		t.Fatalf("token1 = %+v", tokens[1])
	}
	if tokens[2].Type != LiteralTokenType(TokenRune) || tokens[2].Value != Rune('A') {
		t.Fatalf("token2 = %+v", tokens[2])
	}
}

func TestParseLineInvalidRuneLiterals(t *testing.T) {
	parseLineError(t, "''")
	parseLineError(t, "'ab'")
	parseLineError(t, "'\\U12345678'")
}

func TestParseLineNumberFormats(t *testing.T) {
	tokens := parseLineTokens(t, "42 -7 .5 -.5 1e2 0x10 0b11 0o7 100u")
	if tokens[0].Value != Int(42) {
		t.Fatalf("42 = %v", tokens[0].Value)
	}
	if tokens[1].Value != Int(-7) {
		t.Fatalf("-7 = %v", tokens[1].Value)
	}
	if !HasTokenFlag(tokens[2].Type, TokenFloat) {
		t.Fatalf(".5 type = %v", tokens[2].Type)
	}
	if !HasTokenFlag(tokens[3].Type, TokenFloat) {
		t.Fatalf("-.5 type = %v", tokens[3].Type)
	}
	if !HasTokenFlag(tokens[4].Type, TokenFloat) {
		t.Fatalf("1e2 type = %v", tokens[4].Type)
	}
	if tokens[5].Value != Int(16) {
		t.Fatalf("0x10 = %v", tokens[5].Value)
	}
	if tokens[6].Value != Int(3) {
		t.Fatalf("0b11 = %v", tokens[6].Value)
	}
	if tokens[7].Value != Int(7) {
		t.Fatalf("0o7 = %v", tokens[7].Value)
	}
	if !HasTokenFlag(tokens[8].Type, TokenUint) {
		t.Fatalf("100u type = %v", tokens[8].Type)
	}
}

func TestParseLineHexDigitsEAndE(t *testing.T) {
	tokens := parseLineTokens(t, "0xe0e0e0 0xE0E0E0")
	want := Int(0xE0E0E0)
	if !HasTokenFlag(tokens[0].Type, TokenInt) || tokens[0].Value != want {
		t.Fatalf("0xe0e0e0 = type %v value %v, want TokenInt %v", tokens[0].Type, tokens[0].Value, want)
	}
	if !HasTokenFlag(tokens[1].Type, TokenInt) || tokens[1].Value != want {
		t.Fatalf("0xE0E0E0 = type %v value %v, want TokenInt %v", tokens[1].Type, tokens[1].Value, want)
	}
}

func TestParseLineInvalidNumbers(t *testing.T) {
	t.Run("decimal_digit_start", func(t *testing.T) {
		for _, line := range []string{
			"2drop",
			"42foo",
			"100ufoo",
			"3efoo",
		} {
			parseLineError(t, line)
		}
	})

	t.Run("minus_digit_start", func(t *testing.T) {
		for _, line := range []string{
			"-7drop",
			"-42foo",
			"-5u",
		} {
			parseLineError(t, line)
		}
	})

	t.Run("minus_dot_digit_start", func(t *testing.T) {
		for _, line := range []string{
			"-.5foo",
			"-.5drop",
		} {
			parseLineError(t, line)
		}
	})

	t.Run("dot_digit_start", func(t *testing.T) {
		for _, line := range []string{
			".5foo",
			".5drop",
		} {
			parseLineError(t, line)
		}
	})

	t.Run("base_prefix_requires_valid_integral", func(t *testing.T) {
		for _, line := range []string{
			"0x",
			"0X",
			"0x10g",
			"0xdeadbeefdrop",
			"0b",
			"0B",
			"0b102",
			"0b101x",
			"0o",
			"0o78",
			"0o77g",
			"0O7", // oct prefix is 0o only (02-lexical-parsing.md)
		} {
			parseLineError(t, line)
		}
	})

	t.Run("malformed_float_body", func(t *testing.T) {
		for _, line := range []string{
			"1.2.3",
		} {
			parseLineError(t, line)
		}
	})
}

// TestParseLineNumberStartBoundary confirms forms that look related but are *not* number starts.
func TestParseLineNumberStartBoundary(t *testing.T) {
	t.Run("minus_not_followed_by_number", func(t *testing.T) {
		tokens := parseLineTokens(t, "- -foo")
		if tokens[0].Value != String("-") {
			t.Fatalf("standalone - = %v", tokens[0].Value)
		}
		if tokens[1].Value != String("-foo") {
			t.Fatalf("-foo = %v", tokens[1].Value)
		}
	})
	t.Run("minus_dot_not_followed_by_digit", func(t *testing.T) {
		tokens := parseLineTokens(t, "-.field")
		if IsLiteralTokenType(tokens[0].Type) {
			t.Fatalf("-.field type = %v, want non-literal ident/compound", tokens[0].Type)
		}
	})
}

func TestParseLineFalsePositiveSimpleNotCompound(t *testing.T) {
	cases := []string{"dup", "test-if", "ds$", "+", "==", ">=", ">resolve", "math", "foo"}
	for _, c := range cases {
		tokens := parseLineTokens(t, c)
		if !IsPlainIdentToken(tokens[0].Type) {
			t.Fatalf("%q classified as %v, want plain ident", c, tokens[0].Type)
		}
		if tokens[0].Value != String(c) {
			t.Fatalf("%q value = %v", c, tokens[0].Value)
		}
	}
}

func TestParseLineFalseNegativeMustBeCompound(t *testing.T) {
	cases := []struct {
		line string
		segs []string
	}{
		{".x", []string{".", "x"}},
		{"a.b", []string{"a", ".", "b"}},
		{"a()", []string{"a", "()"}},
		{"a@", []string{"a", "@"}},
		{"foo::bar.baz", []string{"foo::bar", ".", "baz"}},
	}
	for _, tc := range cases {
		tokens := parseLineTokens(t, tc.line)
		if !HasTokenFlag(tokens[0].Type, TokenCompoundIdent) {
			t.Fatalf("%q type = %v, want compound", tc.line, tokens[0].Type)
		}
		expectCompoundPath(t, tokens[0].Value, tc.segs...)
	}
}

func TestParseLineRawString(t *testing.T) {
	tokens := parseLineTokens(t, "`hello`")
	if tokens[0].Type != LiteralTokenType(TokenRawString) || tokens[0].Value != String("hello") {
		t.Fatalf("tokens = %+v", tokens)
	}
}

func TestParseRawStringNoNewlineTokens(t *testing.T) {
	tokens := parseLineTokens(t, "`a\nb` after")
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2 (raw string + ident), got %+v", len(tokens), tokens)
	}
	if tokens[0].Type != LiteralTokenType(TokenRawString) || tokens[0].Value != String("a\nb") {
		t.Fatalf("token[0] = %+v, want raw string a\\nb", tokens[0])
	}
	for _, tok := range tokens {
		if tok.Type == TokenNewline {
			t.Fatalf("newline inside raw string must not emit TokenNewline, got %+v", tokens)
		}
	}
}

func TestParseRawStringMultiline(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("`line1\nline2`")),
		label: "test",
	}})
	val, typ, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if typ != LiteralTokenType(TokenRawString) || val != String("line1\nline2") {
		t.Fatalf("got %v %v, want raw string line1\\nline2", val, typ)
	}
}

func TestParseLineRawStringUnclosed(t *testing.T) {
	p := NewParser(nil)
	_, stillOpen, err := p.ParseLine("`no close")
	if err == nil || !stillOpen {
		t.Fatalf("stillOpen=%v err=%v, want unclosed raw string", stillOpen, err)
	}
}

func TestParseLineCommaIsOperatorWord(t *testing.T) {
	tokens := parseLineTokens(t, "here 0 ,")
	if tokens[2].Type != TokenIdent || tokens[2].Value != String(",") {
		t.Fatalf("token[2] = %+v, want ident ,", tokens[2])
	}
}

func TestPushNextAsIdentifierSimple(t *testing.T) {
	p := NewParser(nil)
	p.bindString("foo bar")
	ds := NewStack[any]("test", 8)
	p.PushNextAsIdentifier(ds)
	if got, ok := ds.Pop().(String); !ok || got != "foo" {
		t.Fatalf("ident = %v, want String(foo)", got)
	}
}

func TestPushNextAsIdentifierCompoundPanics(t *testing.T) {
	p := NewParser(nil)
	p.bindString(".field")
	ds := NewStack[any]("test", 8)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for compound identifier")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "compound identifier") {
			t.Fatalf("panic = %q", msg)
		}
	}()
	p.PushNextAsIdentifier(ds)
}

func TestPushNextAsIdentifierFQNSucceeds(t *testing.T) {
	p := NewParser(nil)
	p.bindString("math::sin")
	ds := NewStack[any]("test", 8)
	p.PushNextAsIdentifier(ds)
	if ds.Len() != 1 || ds.Pop().(String) != "math::sin" {
		t.Fatalf("stack = %v, want String(math::sin)", ds)
	}
}

func TestPushNextAsIdentifierNonIdentPanics(t *testing.T) {
	p := NewParser(nil)
	p.bindString("42")
	ds := NewStack[any]("test", 8)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-identifier")
		}
	}()
	p.PushNextAsIdentifier(ds)
}

func TestSplitCompoundSegmentsUnit(t *testing.T) {
	segs, err := splitCompoundSegments("foo::bar.baz")
	if err != nil {
		t.Fatal(err)
	}
	if len(segs) != 3 || segs[0] != "foo::bar" || segs[1] != "." || segs[2] != "baz" {
		t.Fatalf("segs = %v", segs)
	}
	_, err = splitCompoundSegments("a.")
	if err == nil {
		t.Fatal("want trailing dot error")
	}
}

func TestSplitCompoundAffixLookahead(t *testing.T) {
	cases := []struct {
		lit  string
		want []string
	}{
		{"myList@()", []string{"myList", "@()"}},
		{"myList!()", []string{"myList", "!()"}},
		{"map@{}", []string{"map", "@{}"}},
		{"map!{}", []string{"map", "!{}"}},
		{"mat@[]", []string{"mat", "@[]"}},
		{"mat![]", []string{"mat", "![]"}},
		{"x@y", []string{"x", "@", "y"}},
		{".field@()", []string{".", "field", "@()"}},
	}
	for _, tc := range cases {
		segs, err := splitCompoundSegments(tc.lit)
		if err != nil {
			t.Fatalf("%q: %v", tc.lit, err)
		}
		if len(segs) != len(tc.want) {
			t.Fatalf("%q: segs = %v, want %v", tc.lit, segs, tc.want)
		}
		for i := range tc.want {
			if segs[i] != tc.want[i] {
				t.Fatalf("%q: segs = %v, want %v", tc.lit, segs, tc.want)
			}
		}
	}
}

func TestSplitCompoundLoneGtRejected(t *testing.T) {
	_, err := splitCompoundSegments(".name>@")
	if err == nil {
		t.Fatal("want error for lone > in compound path")
	}
}

func TestSplitCompoundGtShuffleRejected(t *testing.T) {
	for _, lit := range []string{"find>os.name@", "find>ts.name@", ".name>os@"} {
		if _, err := splitCompoundSegments(lit); err == nil {
			t.Fatalf("%q: want error for > shuffle in compound path", lit)
		}
	}
}

func TestRemovedStackShuffleWordsAbsent(t *testing.T) {
	ResetTestRuntime()
	if globalDictionaryContext.dict.GetByName(">os") != nil {
		t.Fatal("word >os should be removed from dictionary")
	}
	if w := globalDictionaryContext.dict.GetByName("drop"); w == nil || w.opCode != OpDrop {
		t.Fatalf("word drop = %v, want OpDrop axiom", w)
	}
}

func TestStackArrowWordsStaySimpleIdent(t *testing.T) {
	for _, lit := range []string{">ts", "ts>", ">mark", ">resolve"} {
		typ, _, err := classifyNameLiteral(lit)
		if err != nil {
			t.Fatalf("%q: %v", lit, err)
		}
		if HasTokenFlag(typ, TokenCompoundIdent) {
			t.Fatalf("%q classified as compound, want simple ident", lit)
		}
	}
}

func TestParserLineIncrementsOnCRLF(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("a\r\nb\r")),
		label: "test",
	}})
	ds := NewStack[any]("test", 8)
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	ds.Pop()
	ds.Pop()
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if p.line != 2 {
		t.Fatalf("after CRLF newline line = %d, want 2", p.line)
	}
}

func TestFormatFaultAtWithFile(t *testing.T) {
	got := formatFaultAt("/tmp/foo.fnl", 2, "    foo", 5, fmt.Errorf(`unknown word "foo"`))
	want := "Error compiling file \"/tmp/foo.fnl\"\nline 2; column 5 :\n" +
		">     foo\n" +
		">     ^\n" +
		`Error: unknown word "foo"`
	if got != want {
		t.Fatalf("formatFaultAt = %q, want %q", got, want)
	}
}

func TestFormatFaultAtWithoutFile(t *testing.T) {
	got := formatFaultAt("", 1, "foo", 1, fmt.Errorf(`unknown word "foo"`))
	want := "Error compiling; line 1; column 1 :\n" +
		"> foo\n" +
		"> ^\n" +
		`Error: unknown word "foo"`
	if got != want {
		t.Fatalf("formatFaultAt = %q, want %q", got, want)
	}
}

func TestFormatFaultAtIdleOmitsLocation(t *testing.T) {
	got := formatFaultAt("", 1, "", 0, fmt.Errorf("plain"))
	if got != "plain" {
		t.Fatalf("formatFaultAt idle = %q, want plain", got)
	}
}

func TestFormatFaultAtEmptyLastLine(t *testing.T) {
	got := formatFaultAt("/tmp/x.fnl", 3, "", 9, fmt.Errorf("unclosed raw string"))
	want := "Error compiling file \"/tmp/x.fnl\"\nline 3; column 1 :\n" +
		"> \n" +
		"> ^\n" +
		"Error: unclosed raw string"
	if got != want {
		t.Fatalf("formatFaultAt empty = %q, want %q", got, want)
	}
}

func TestFormatFaultAtElidesLongLine(t *testing.T) {
	prefix := strings.Repeat("x", 100)
	line := prefix + "foo"
	col := 101 // 'f' of foo at end
	got := formatFaultAt("", 1, line, col, fmt.Errorf(`unknown word "foo"`))
	if !strings.Contains(got, "> ... ") {
		t.Fatalf("want elision prefix in %q", got)
	}
	if !strings.Contains(got, "foo\n") {
		t.Fatalf("want foo in display line: %q", got)
	}
	lines := strings.Split(got, "\n")
	if len(lines) < 3 {
		t.Fatalf("lines = %d, want >= 3", len(lines))
	}
	src := lines[1]
	caret := lines[2]
	if !strings.HasPrefix(src, "> ... ") || !strings.HasPrefix(caret, "> ... ") {
		t.Fatalf("src=%q caret=%q", src, caret)
	}
}

func TestClearParserLocation(t *testing.T) {
	p := NewParser(nil)
	p.setParserLocation("/tmp/x.fnl", 4, 2)
	p.lineText = "hi"
	p.col = 2
	p.clearParserLocation()
	if p.file != "" || p.line != 1 || p.col != 0 || p.lineText != "" {
		t.Fatalf("after clear: file=%q line=%d col=%d lineText=%q",
			p.file, p.line, p.col, p.lineText)
	}
}

func TestDrainToStreamEOFPreservesLastLine(t *testing.T) {
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("foo bar\nbaz\n")),
		label: "test",
	}})
	ds := NewStack[any]("test", 8)
	if err := pushNextToken(p, ds); err != nil {
		t.Fatal(err)
	}
	if p.lineText == "" || p.col == 0 {
		t.Fatalf("after token lineText=%q col=%d", p.lineText, p.col)
	}
	savedLine, savedCol, savedText := p.line, p.col, p.lineText
	p.drainToStreamEOF()
	if p.lineText != savedText || p.col != savedCol || p.line != savedLine {
		t.Fatalf("after drain lineText=%q col=%d line=%d, want %q %d %d",
			p.lineText, p.col, p.line, savedText, savedCol, savedLine)
	}
}
