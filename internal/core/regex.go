package core

import (
	"regexp"
	"strconv"
)

// Regex is a compiled regular expression (Final sealed type).
// The class also offers one-shot match and find methods.
type Regex struct {
	re *regexp.Regexp
}

func (r *Regex) FinalType() *Type { return regexType }

// NaturalString formats for stack traces / .s / $ (not a Final _toString API).
func (r *Regex) NaturalString() string {
	if r == nil || r.re == nil {
		return "Regex pattern: <nil>"
	}
	return "Regex pattern: " + r.re.String()
}

// NewRegex creates a Regex from pattern text. Panics on a bad pattern.
func NewRegex(pattern string) *Regex {
	return &Regex{re: regexp.MustCompile(pattern)}
}

var regexType *Type

var _ Typed = (*Regex)(nil)
var _ NaturalStringer = (*Regex)(nil)

// ( sb:StringBuilder -- sb )
func nativeRenderJsonRegex(f *RunFrame, ths *TypeObjPair) {
	sb := f.RP.ZStack.Peek().(*StringBuilder)
	sb.WriteString(stringJsonWire(String(ths.Obj.(*Regex).re.String())))
}

func regexPopString(f *RunFrame) String {
	return f.RP.DataStack.Pop().(String)
}

func pushModalFalse(f *RunFrame) {
	f.RP.DataStack.Push(Bool(false))
}

func pushModalString(f *RunFrame, s String) {
	f.RP.DataStack.Push(s)
	f.RP.DataStack.Push(Bool(true))
}

func pushModalStringList(f *RunFrame, parts []string) {
	if len(parts) == 0 {
		pushModalFalse(f)
		return
	}
	out := NewStringListWithCapacity(len(parts))
	for _, p := range parts {
		out.Append(p)
	}
	f.RP.DataStack.Push(out)
	f.RP.DataStack.Push(Bool(true))
}

func regexGroupKey(names []string, i int) string {
	if i < len(names) && names[i] != "" {
		return names[i]
	}
	return strconv.Itoa(i)
}

func regexMatchObject(re *regexp.Regexp, s string) (*DynamicObject, bool) {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, false
	}
	names := re.SubexpNames()
	o := NewDynamicObject()
	// Include Go group 0 (whole match) then capturing groups 1…n — standard regex numbering.
	for i := 0; i < len(m); i++ {
		o.Set(regexGroupKey(names, i), String(m[i]))
	}
	return o, true
}

func registerRegexMembers(t *Type) {
	t.registerProperty("pattern", "String",
		func(inst Any) Any { return String(inst.(*Regex).re.String()) },
		nil,
	)
	t.registerProperty("numGroups", "Int",
		func(inst Any) Any { return Int(inst.(*Regex).re.NumSubexp()) },
		nil,
	)
	t.registerProperty("groupNames", "StringList",
		func(inst Any) Any {
			names := inst.(*Regex).re.SubexpNames()
			out := NewStringListWithCapacity(len(names))
			for i := range names {
				out.Append(regexGroupKey(names, i))
			}
			return out
		},
		nil,
	)

	// --- class methods ---

	// ( pattern:String -- re:Regex )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewRegex(string(regexPopString(f))))
	})

	// ( pattern:String -- b:Bool )
	t.registerNativeClassMethod("isPatternValid?", func(f *RunFrame, _ *TypeObjPair) {
		_, err := regexp.Compile(string(regexPopString(f)))
		f.RP.DataStack.Push(Bool(err == nil))
	})

	// ( s:String pattern:String -- b:Bool )
	t.registerNativeClassMethod("match", func(f *RunFrame, _ *TypeObjPair) {
		pattern := string(regexPopString(f))
		s := string(regexPopString(f))
		f.RP.DataStack.Push(Bool(regexp.MustCompile(pattern).MatchString(s)))
	})

	// ( s:String pattern:String -- found:String true | false )
	t.registerNativeClassMethod("find", func(f *RunFrame, _ *TypeObjPair) {
		pattern := string(regexPopString(f))
		s := string(regexPopString(f))
		loc := regexp.MustCompile(pattern).FindStringIndex(s)
		if loc == nil {
			pushModalFalse(f)
			return
		}
		pushModalString(f, String(s[loc[0]:loc[1]]))
	})

	// ( s:String -- q:String )
	t.registerNativeClassMethod("quoteMeta", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(String(regexp.QuoteMeta(string(regexPopString(f)))))
	})

	// --- instance methods ---

	// ( s:String -- b:Bool )
	t.registerNativeMethod("match", func(f *RunFrame, ths *TypeObjPair) {
		s := string(regexPopString(f))
		f.RP.DataStack.Push(Bool(ths.Obj.(*Regex).re.MatchString(s)))
	})

	// ( s:String -- found:String true | false )
	t.registerNativeMethod("find", func(f *RunFrame, ths *TypeObjPair) {
		s := string(regexPopString(f))
		loc := ths.Obj.(*Regex).re.FindStringIndex(s)
		if loc == nil {
			pushModalFalse(f)
			return
		}
		pushModalString(f, String(s[loc[0]:loc[1]]))
	})

	// ( s:String -- parts:StringList true | false )
	t.registerNativeMethod("findAll", func(f *RunFrame, ths *TypeObjPair) {
		s := string(regexPopString(f))
		pushModalStringList(f, ths.Obj.(*Regex).re.FindAllString(s, -1))
	})

	// ( s:String n:Int -- parts:StringList true | false )
	t.registerNativeMethod("findAllN", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		s := string(regexPopString(f))
		pushModalStringList(f, ths.Obj.(*Regex).re.FindAllString(s, n))
	})

	// ( s:String -- start:Int end:Int true | false )
	t.registerNativeMethod("findIndex", func(f *RunFrame, ths *TypeObjPair) {
		s := string(regexPopString(f))
		loc := ths.Obj.(*Regex).re.FindStringIndex(s)
		if loc == nil {
			pushModalFalse(f)
			return
		}
		f.RP.DataStack.Push(Int(loc[0]))
		f.RP.DataStack.Push(Int(loc[1]))
		f.RP.DataStack.Push(Bool(true))
	})

	// ( s:String -- parts:StringList )
	t.registerNativeMethod("split", func(f *RunFrame, ths *TypeObjPair) {
		s := string(regexPopString(f))
		parts := ths.Obj.(*Regex).re.Split(s, -1)
		out := NewStringListWithCapacity(len(parts))
		for _, p := range parts {
			out.Append(p)
		}
		f.RP.DataStack.Push(out)
	})

	// ( s:String n:Int -- parts:StringList )
	t.registerNativeMethod("splitN", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		s := string(regexPopString(f))
		parts := ths.Obj.(*Regex).re.Split(s, n)
		out := NewStringListWithCapacity(len(parts))
		for _, p := range parts {
			out.Append(p)
		}
		f.RP.DataStack.Push(out)
	})

	// ( s:String -- o:DynamicObject true | false )
	t.registerNativeMethod("matchObject", func(f *RunFrame, ths *TypeObjPair) {
		s := string(regexPopString(f))
		o, ok := regexMatchObject(ths.Obj.(*Regex).re, s)
		if !ok {
			pushModalFalse(f)
			return
		}
		f.RP.DataStack.Push(o)
		f.RP.DataStack.Push(Bool(true))
	})

	// ( s:String repl:String -- out:String )
	t.registerNativeMethod("replaceAll", func(f *RunFrame, ths *TypeObjPair) {
		repl := string(regexPopString(f))
		s := string(regexPopString(f))
		f.RP.DataStack.Push(String(ths.Obj.(*Regex).re.ReplaceAllString(s, repl)))
	})

	// ( s:String repl:String -- out:String )
	t.registerNativeMethod("replaceAllLiteral", func(f *RunFrame, ths *TypeObjPair) {
		repl := string(regexPopString(f))
		s := string(regexPopString(f))
		f.RP.DataStack.Push(String(ths.Obj.(*Regex).re.ReplaceAllLiteralString(s, repl)))
	})

	t.registerNativeMethod("_renderJson", nativeRenderJsonRegex)
}

func CreateRegisterRegexType() {
	t := &Type{
		name:            "Regex",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Regex]("Regex"),
	}
	registerRegexMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	regexType = t
}

func init() {
	CreateRegisterRegexType()
}
