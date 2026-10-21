package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// formatNaturalFloat is shortest round-trip float text with a decimal (or exponent)
// so finite Float is visually distinct from Int. Inf/NaN are unchanged.
func formatNaturalFloat(x Float) string {
	s := strconv.FormatFloat(float64(x), 'g', -1, 64)
	if !math.IsInf(float64(x), 0) && !math.IsNaN(float64(x)) &&
		!strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

// formatNaturalString returns the natural string for v ($ semantics, doc 09).
func formatNaturalString(v any) string {
	if v == nil {
		return "null" // note: Final null is "null", not "nil"
	}
	if ns, ok := v.(NaturalStringer); ok {
		return ns.NaturalString()
	}
	switch x := v.(type) {
	case String:
		return string(x)
	case time.Time:
		return string(dateTimeToString(x))
	case Bool:
		if x {
			return "true"
		}
		return "false"
	case Int:
		return strconv.Itoa(int(x))
	case Uint:
		return strconv.FormatUint(uint64(x), 10) + "u"
	case Float:
		return formatNaturalFloat(x)
	case *AnyList:
		return "List"
	case *TextReader:
		return "TextReader"
	case *TextWriter:
		return "TextWriter"
	default:
		return fmt.Sprintf("<%T>", v)
	}
}

func flushTextWrite(w TextWrite) {
	if w == nil {
		return
	}
	if f, ok := w.(Flushable); ok {
		f.Flush()
	}
}

// runS pops ds.tos and prints it to out as a natural string ($ / $$).
// panics if ds is empty
func runS(ds *Stack[any], out TextWrite) {
	if out == nil {
		panic("runS: nil TextWrite")
	}
	if ds.Len() == 0 {
		panic("runS: ds is empty")
	}
	out.WriteString(formatNaturalString(ds.Pop()))
	flushTextWrite(out)
}

// runRtoS pops an Int code point and pushes a one-character String (r>s).
func runRtoS(ds *Stack[any]) {
	if ds.Len() == 0 {
		panic("runRtoS: ds is empty")
	}
	r := rune(int(ds.Pop().(Int)))
	ds.Push(String(string(r)))
}

// runDsS prints all data-stack values deepest-first to stdout (-> . behavior: natural string + space).
// Bootstrap debug only; stack is unchanged.
func runDsS(ds *Stack[any], stdout TextWrite) {
	if stdout == nil {
		panic("runDsS: nil stdout")
	}
	n := ds.Len()
	for i := n - 1; i >= 0; i-- {
		if i < n-1 {
			stdout.WriteString(" ")
		}
		stdout.WriteString(formatNaturalString(*ds.AddrAt(i)))
	}
	if n > 0 {
		stdout.WriteString(" ")
	}
	flushTextWrite(stdout)
}

// runStarS prints RS/DS/TS snapshot to stdout (*$ layout; RS * = current). Stacks unchanged.
func runStarS(rp *RunProvider, stdout TextWrite, execDC *DictionaryContext, ip IP, lastOp OpCode) {
	if stdout == nil {
		panic("runStarS: nil stdout")
	}
	stdout.WriteString(FormatStackSnapshotReport(rp, execDC, ip, lastOp))
	flushTextWrite(stdout)
}
