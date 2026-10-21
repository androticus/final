package core

import (
	"strings"
	"testing"
)

func mustCompileColonDef(t *testing.T, src string) {
	t.Helper()
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile %q: %v", src, err)
	}
}

func expectCompileColonDefError(t *testing.T, src string, wantSubstr string) {
	t.Helper()
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, src)
	if err == nil {
		t.Fatalf("compile %q: expected error containing %q", src, wantSubstr)
	}
	if wantSubstr != "" && !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("compile %q: err = %v, want substring %q", src, err, wantSubstr)
	}
}

func TestCompileStructureValidIfElseThen(t *testing.T) {
	mustCompileColonDef(t, `: ifblock_smoke if 99 else -1 then ;`)
}

func TestCompileStructureValidIfThen(t *testing.T) {
	mustCompileColonDef(t, `: ifblock_ifonly if 99 then ;`)
}

func TestCompileStructureValidForeachI(t *testing.T) {
	mustCompileColonDef(t, `: ijk_ismoke 0 1 range foreach i /foreach ;`)
}

// Inside :s, i/j/k and a/b/c skip compile-time foreach/switch depth (runtime still panics).
func TestCompileStructureValidIJKABCInSubwordWithoutBlocks(t *testing.T) {
	mustCompileColonDef(t, `: ijk_sub :s use i j k ;s ;`)
	mustCompileColonDef(t, `: abc_sub :s use a b c ;s ;`)
}

func TestCompileStructureValidForeachIJ(t *testing.T) {
	mustCompileColonDef(t, `: ijk_ijsmoke 0 1 range foreach i 2 3 range foreach i j /foreach /foreach ;`)
}

func TestCompileStructureValidDeepForeachNesting(t *testing.T) {
	// No compile-time nesting cap; deeper loops simply lack i/j/k accessors.
	mustCompileColonDef(t, `: deep4 0 1 range foreach 2 3 range foreach 4 5 range foreach 6 7 range foreach nop /foreach /foreach /foreach /foreach ;`)
}

func TestCompileStructureValidBeginAgain(t *testing.T) {
	mustCompileColonDef(t, `: bagain begin again ;`)
}

func TestCompileStructureValidBeginUntil(t *testing.T) {
	mustCompileColonDef(t, `: buntil begin 1 until ;`)
}

func TestCompileStructureValidBeginWhileRepeat(t *testing.T) {
	mustCompileColonDef(t, `: bwhilerep begin 1 while repeat ;`)
}

func TestCompileStructureValidForever(t *testing.T) {
	mustCompileColonDef(t, `: fforever forever /forever ;`)
}

func TestCompileStructureValidSwitch(t *testing.T) {
	mustCompileColonDef(t, `: swsmoke null switch 1 case 9 /case /switch ;`)
}

func TestCompileStructureValidSwitchDefaultOnly(t *testing.T) {
	mustCompileColonDef(t, `: swdef null switch default 9 /default /switch ;`)
}

func TestCompileStructureValidSwitchEmpty(t *testing.T) {
	mustCompileColonDef(t, `: swempty null switch /switch ;`)
}

func TestCompileStructureValidSwitchWithA(t *testing.T) {
	mustCompileColonDef(t, `: swa switch a 1 == case 9 /case /switch ;`)
}

func TestCompileStructureSwitchDirectNestPanics(t *testing.T) {
	expectCompileColonDefError(t, `: bad null switch null switch 1 case 1 /case /switch /switch ;`, "cannot nest directly")
}

func TestCompileStructureValidNestedIfInBegin(t *testing.T) {
	mustCompileColonDef(t, `: nest begin 1 if break then again ;`)
}

func TestCompileStructureValidNestedIfInForever(t *testing.T) {
	mustCompileColonDef(t, `: nestf forever 1 if break then /forever ;`)
}

func TestCompileStructureValidForeachBreak(t *testing.T) {
	mustCompileColonDef(t, `: fbreak 0 1 range foreach break /foreach ;`)
}

func TestCompileStructureValidForeachContinue(t *testing.T) {
	mustCompileColonDef(t, `: fcont 0 1 range foreach continue /foreach ;`)
}

func TestCompileStructureValidBeginWhileRepeatBreak(t *testing.T) {
	mustCompileColonDef(t, `: bwrb begin 1 while break repeat ;`)
}

func TestCompileStructurePanics(t *testing.T) {
	cases := []struct {
		name       string
		src        string
		wantSubstr string
	}{
		// if / else / then
		{name: "double else", src: `: ifblock_2else if 99 else -1 else 0 then ;`, wantSubstr: "'else' requires a preceding 'if'"},
		{name: "else without if", src: `: ifblock_elsenoif else -1 then ;`, wantSubstr: "'else' requires a preceding 'if'"},
		{name: "then without if", src: `: ifblock_thennoif then ;`, wantSubstr: "'then' requires a preceding 'if' or 'else'"},
		{name: "unclosed if at semicolon", src: `: unclosedif if 1 ;`, wantSubstr: semicolonBlockCheckMsg},

		// begin / again / until
		{name: "again without begin", src: `: begin_againnobegin if again ;`, wantSubstr: "'again' requires a preceding 'begin'"},
		{name: "until without begin", src: `: begin_untilnobegin if until ;`, wantSubstr: "'until' requires a preceding 'begin'"},
		{name: "unclosed begin at semicolon", src: `: unclosedbegin begin ;`, wantSubstr: semicolonBlockCheckMsg},

		// while / repeat
		{name: "while without begin", src: `: while_nobegin 1 while ;`, wantSubstr: "'while' requires a preceding 'begin'"},
		{name: "repeat without while", src: `: repeat_nowhile repeat ;`, wantSubstr: "'repeat' requires a preceding 'while'"},
		{name: "repeat without begin", src: `: repeat_nobegin begin repeat ;`, wantSubstr: "'repeat' requires a preceding 'while'"},

		// foreach
		{name: "foreach without iterable", src: `: foreach_noiter foreach ;`, wantSubstr: semicolonBlockCheckMsg},
		{name: "unclosed foreach at semicolon", src: `: unclosedforeach 0 1 range foreach ;`, wantSubstr: semicolonBlockCheckMsg},
		{name: "slash foreach without foreach", src: `: slashforeach /foreach ;`, wantSubstr: "'/foreach' requires a preceding 'foreach'"},
		{name: "slash foreach with open if", src: `: ifforeach if 1 /foreach ;`, wantSubstr: "'/foreach' requires a preceding 'foreach'"},

		// i / j / k depth (main body only; :s skips)
		{name: "i outside foreach", src: `: ijk_inofor i ;`, wantSubstr: "'i/j/k' require 1/2/3 outer 'foreach' loops"},
		{name: "j requires nested foreach", src: `: jonly 0 1 range foreach j ;`, wantSubstr: "'i/j/k' require 1/2/3 outer 'foreach' loops"},
		{name: "k requires three foreach", src: `: ijk_knofor 0 1 range foreach 2 3 range foreach j k /foreach /foreach ;`, wantSubstr: "'i/j/k' require 1/2/3 outer 'foreach' loops"},

		// a / b / c depth (main body only; :s skips)
		{name: "a outside switch", src: `: abc_inosw a ;`, wantSubstr: "'a/b/c' require 1/2/3 outer 'switch' blocks"},
		{name: "b requires nested switch", src: `: bonly null switch a 1 == case b /case /switch ;`, wantSubstr: "'a/b/c' require 1/2/3 outer 'switch' blocks"},

		// switch
		{name: "case without switch", src: `: casenosw case 1 /case /switch ;`, wantSubstr: "'case' requires a preceding 'switch'"},
		{name: "slash case without case", src: `: slashcase /case ;`, wantSubstr: "'/case' requires a preceding 'case'"},
		{name: "second case without slash case", src: `: twocase switch 1 case 9 2 case 8 /case /switch ;`, wantSubstr: "'case' requires a preceding 'switch'"},
		{name: "default without slash case on open case", src: `: defopen switch 1 case 9 default 0 /default /switch ;`, wantSubstr: "'default' requires a preceding 'switch'"},
		{name: "default without switch", src: `: defnosw default 0 /default /switch ;`, wantSubstr: "'default' requires a preceding 'switch'"},
		{name: "slash default without default", src: `: slashdefault /default ;`, wantSubstr: "'/default' requires a preceding 'default'"},
		{name: "slash switch without switch", src: `: slashswitch /switch ;`, wantSubstr: "'/switch' requires a preceding 'switch'"},
		{name: "unclosed switch at semicolon", src: `: unclosedswitch switch ;`, wantSubstr: semicolonBlockCheckMsg},

		// break / continue placement
		{name: "break outside loop", src: `: breaknoloop break ;`, wantSubstr: "'break' requires an enclosing loop"},
		{name: "break only in if", src: `: breakonlyif if break then ;`, wantSubstr: "'break' requires an enclosing loop"},
		{name: "continue outside loop", src: `: contnoloop continue ;`, wantSubstr: "'continue' requires 'foreach' or 'forever'"},
		{name: "continue in until", src: `: contuntil begin continue 1 until ;`, wantSubstr: "'continue' requires 'foreach' or 'forever'"},
		{name: "continue in while/repeat", src: `: contwhile begin 1 while continue repeat ;`, wantSubstr: "'continue' requires 'foreach' or 'forever'"},

		// forever closer mismatches
		{name: "slash forever without forever", src: `: slashforever /forever ;`, wantSubstr: "'/forever' requires a preceding 'forever'"},
		{name: "forever closed with again", src: `: foreveragain forever again ;`, wantSubstr: "'again' requires a preceding 'begin'"},
		{name: "begin closed with slash forever", src: `: beginforever begin /forever ;`, wantSubstr: "'/forever' requires a preceding 'forever'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectCompileColonDefError(t, tc.src, tc.wantSubstr)
		})
	}
}
