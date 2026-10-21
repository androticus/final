package core

import (
	"fmt"
	"strings"
	"unicode"
)

// String helper methods — registered on the sealed String Type.
// See docs/003-design-notes/_archived/126-string-helpers.md.

func stringTrimSpace(s String) String {
	return String(strings.TrimSpace(string(s)))
}

func stringLower(s String) String {
	return String(strings.ToLower(string(s)))
}

func stringUpper(s String) String {
	return String(strings.ToUpper(string(s)))
}

func stringTrimStart(s String) String {
	return String(strings.TrimLeftFunc(string(s), unicode.IsSpace))
}

func stringTrimEnd(s String) String {
	return String(strings.TrimRightFunc(string(s), unicode.IsSpace))
}

func stringIsEmpty(s String) Bool {
	return Bool(len(s) == 0)
}

func stringStartsWith(s, prefix String) Bool {
	return Bool(strings.HasPrefix(string(s), string(prefix)))
}

func stringEndsWith(s, suffix String) Bool {
	return Bool(strings.HasSuffix(string(s), string(suffix)))
}

func stringAppend(s, other String) String {
	return String(string(s) + string(other))
}

func stringIndexOf(s, sub String) Int {
	return Int(strings.Index(string(s), string(sub)))
}

func stringSub(s, find, replace String) String {
	return String(strings.ReplaceAll(string(s), string(find), string(replace)))
}

func stringJoin(parts []string, sep String) String {
	if len(parts) == 0 {
		return ""
	}
	return String(strings.Join(parts, string(sep)))
}

func joinListItems[T any](list *List[T], sep String, itemString func(T) string) String {
	n := list.Len()
	if n == 0 {
		return ""
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = itemString(list.At(i))
	}
	return stringJoin(parts, sep)
}

func joinElementString(v any) string {
	s, ok := v.(String)
	if !ok {
		panic(fmt.Errorf("str::join: list item %T is not String", v))
	}
	return string(s)
}

func stringSplit(s, sep String) *StringList {
	parts := strings.Split(string(s), string(sep))
	out := NewStringList()
	for _, p := range parts {
		out.Append(p)
	}
	return out
}

func stringSplitLines(s String) *StringList {
	// strings.Split after TrimSuffix trailing newline behavior — keep empty trailing line if file ends with \n
	raw := string(s)
	if raw == "" {
		return NewStringList()
	}
	parts := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	out := NewStringList()
	for _, p := range parts {
		out.Append(p)
	}
	return out
}

func registerStringMembers(t *Type) {
	// ( s:String -- bytes:ByteList ) UTF-8 bytes copy
	t.registerNativeMethod("asBytes", func(f *RunFrame, ths *TypeObjPair) {
		s := string(ths.Obj.(String))
		f.RP.DataStack.Push(NewByteListFromString(s))
	})
}

// String stack ops share OpStrOp; subcode selects the operation (str::len, str::cat, …).

type SCStrOp = int

const (
	SC_str_len SCStrOp = iota
	SC_str_trimStart
	SC_str_trimEnd
	SC_str_trim
	SC_str_cat
	SC_str_isEmpty
	SC_str_startsWith
	SC_str_endsWith
	SC_str_indexOf
	SC_str_split
	SC_str_splitLines
	SC_str_join
	SC_str_toLower
	SC_str_toUpper
	SC_str_sub

	SCStrOpCount
)

type StrOpFn func(ds *Stack[any])

var strOpFns = [SCStrOpCount]StrOpFn{
	SC_str_len:        runStrLen,
	SC_str_trimStart:  runStrTrimStart,
	SC_str_trimEnd:    runStrTrimEnd,
	SC_str_trim:       runStrTrim,
	SC_str_cat:        runStrCat,
	SC_str_isEmpty:    runStrIsEmpty,
	SC_str_startsWith: runStrStartsWith,
	SC_str_endsWith:   runStrEndsWith,
	SC_str_indexOf:    runStrIndexOf,
	SC_str_split:      runStrSplit,
	SC_str_splitLines: runStrSplitLines,
	SC_str_join:       runStrJoin,
	SC_str_toLower:      runStrLower,
	SC_str_toUpper:      runStrUpper,
	SC_str_sub:        runStrSub,
}

func registerStrOps(d *[]*Word) {
	regSubcodeAxiom(d, "str::len", OpStrOp, SC_str_len, WordFlagAxiom)
	regSubcodeAxiom(d, "str::trimStart", OpStrOp, SC_str_trimStart, WordFlagAxiom)
	regSubcodeAxiom(d, "str::trimEnd", OpStrOp, SC_str_trimEnd, WordFlagAxiom)
	regSubcodeAxiom(d, "str::trim", OpStrOp, SC_str_trim, WordFlagAxiom)
	regSubcodeAxiom(d, "str::cat", OpStrOp, SC_str_cat, WordFlagAxiom)
	regSubcodeAxiom(d, "str::+", OpStrOp, SC_str_cat, WordFlagAxiom)
	regSubcodeAxiom(d, "str::isEmpty", OpStrOp, SC_str_isEmpty, WordFlagAxiom)
	regSubcodeAxiom(d, "str::startsWith", OpStrOp, SC_str_startsWith, WordFlagAxiom)
	regSubcodeAxiom(d, "str::endsWith", OpStrOp, SC_str_endsWith, WordFlagAxiom)
	regSubcodeAxiom(d, "str::indexOf", OpStrOp, SC_str_indexOf, WordFlagAxiom)
	regSubcodeAxiom(d, "str::split", OpStrOp, SC_str_split, WordFlagAxiom)
	regSubcodeAxiom(d, "str::splitLines", OpStrOp, SC_str_splitLines, WordFlagAxiom)
	regSubcodeAxiom(d, "str::join", OpStrOp, SC_str_join, WordFlagAxiom)
	regSubcodeAxiom(d, "str::toLower", OpStrOp, SC_str_toLower, WordFlagAxiom)
	regSubcodeAxiom(d, "str::toUpper", OpStrOp, SC_str_toUpper, WordFlagAxiom)
	regSubcodeAxiom(d, "str::sub", OpStrOp, SC_str_sub, WordFlagAxiom)
}

func popString(ds *Stack[any], op string) String {
	v := ds.Pop()
	s, ok := v.(String)
	if !ok {
		panic(fmt.Errorf("str::%s: want String, got %T", op, v))
	}
	return s
}

func runStrLen(ds *Stack[any]) {
	s := popString(ds, "len")
	ds.Push(Int(len(s)))
}

func runStrTrimStart(ds *Stack[any]) {
	ds.Push(stringTrimStart(popString(ds, "trimStart")))
}

func runStrTrimEnd(ds *Stack[any]) {
	ds.Push(stringTrimEnd(popString(ds, "trimEnd")))
}

func runStrTrim(ds *Stack[any]) {
	ds.Push(stringTrimSpace(popString(ds, "trim")))
}

func runStrCat(ds *Stack[any]) {
	op2 := popString(ds, "cat")
	op1 := popString(ds, "cat")
	ds.Push(stringAppend(op1, op2))
}

func runStrIsEmpty(ds *Stack[any]) {
	ds.Push(stringIsEmpty(popString(ds, "isEmpty")))
}

func runStrStartsWith(ds *Stack[any]) {
	prefix := popString(ds, "startsWith")
	s := popString(ds, "startsWith")
	ds.Push(stringStartsWith(s, prefix))
}

func runStrEndsWith(ds *Stack[any]) {
	suffix := popString(ds, "endsWith")
	s := popString(ds, "endsWith")
	ds.Push(stringEndsWith(s, suffix))
}

func runStrIndexOf(ds *Stack[any]) {
	sub := popString(ds, "indexOf")
	s := popString(ds, "indexOf")
	ds.Push(stringIndexOf(s, sub))
}

func runStrSplit(ds *Stack[any]) {
	sep := popString(ds, "split")
	s := popString(ds, "split")
	ds.Push(stringSplit(s, sep))
}

func runStrSplitLines(ds *Stack[any]) {
	ds.Push(stringSplitLines(popString(ds, "splitLines")))
}

func runStrJoin(ds *Stack[any]) {
	sep := popString(ds, "join")
	items := ds.Pop()
	switch list := items.(type) {
	case *StringList:
		ds.Push(joinListItems(list, sep, func(s string) string { return s }))
	case *List[any]:
		ds.Push(joinListItems(list, sep, joinElementString))
	default:
		panic(fmt.Errorf("str::join: want List or StringList, got %T", items))
	}
}

func runStrLower(ds *Stack[any]) {
	ds.Push(stringLower(popString(ds, "lower")))
}

func runStrUpper(ds *Stack[any]) {
	ds.Push(stringUpper(popString(ds, "upper")))
}

func runStrSub(ds *Stack[any]) {
	replace := popString(ds, "sub")
	find := popString(ds, "sub")
	s := popString(ds, "sub")
	ds.Push(stringSub(s, find, replace))
}
