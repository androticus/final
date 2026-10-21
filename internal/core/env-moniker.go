package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Env is a class-only catalog type for reading host configuration.
// Final: Env.get^ ( key:s -- val:s true | false )
// Lookup: cwd .env (priority), then process environment. See docs/003-design-notes/155-env.md.
type Env struct{}

var envType *Type

func CreateRegisterEnvType() {
	t := &Type{
		name:    "Env",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Env.checkCoerceType: Env is a class-only type (got %T)", value))
		},
	}
	registerEnvMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	envType = t
}

func registerEnvMembers(t *Type) {
	// ( key:String -- val:String true | false )
	t.registerNativeClassMethod("get", func(f *RunFrame, _ *TypeObjPair) {
		key := string(f.RP.DataStack.Pop().(String))
		if val, ok := envLookup(key); ok {
			f.RP.DataStack.Push(String(val))
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})

	// ( mon:Moniker keyname:String -- b:Bool )
	// true if key found (all $keyname occurrences replaced); false if missing (mon unchanged).
	// Never requires SR_OS. Does not push the secret or resolved string.
	t.registerNativeClassMethod("subMoniker", func(f *RunFrame, _ *TypeObjPair) {
		keyname := string(f.RP.DataStack.Pop().(String))
		mon := f.RP.DataStack.Pop().(*Moniker)
		val, ok := envLookup(keyname)
		if !ok {
			f.RP.DataStack.Push(Bool(false))
			return
		}
		mon.subKey(keyname, val)
		f.RP.DataStack.Push(Bool(true))
	})
}

func init() {
	CreateRegisterEnvType()
}

// envLookup reads cwd ./.env (no cache), then falls back to os.LookupEnv.
func envLookup(key string) (string, bool) {
	if val, ok := parseDotEnvFile(".env")[key]; ok {
		return val, true
	}
	return os.LookupEnv(key)
}

// parseDotEnvFile parses path into a key→value map.
// Missing file → empty map. Stops at the first duplicate key (keeps earlier binding).
// Malformed lines are skipped.
func parseDotEnvFile(path string) map[string]string {
	out := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue
			}
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue // malformed: no '='
		}
		key := strings.TrimSpace(line[:eq])
		if key == "" {
			continue // malformed: empty key
		}
		val := strings.TrimSpace(line[eq+1:])
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		if _, exists := out[key]; exists {
			return out // first duplicate: keep earlier binding, stop
		}
		out[key] = val
	}
	if err := sc.Err(); err != nil {
		// I/O error mid-read: keep bindings parsed so far (same spirit as skip-malformed).
		return out
	}
	return out
}

// Moniker is a runtime-opaque type that can contain secrets, such as the value of a .env var.
// It can be used as a DSN, etc.
// Use Env.subMoniker^ to substitute the value of a .env var into a string.
// DO NOT surface Moniker.Value() to Final, especially not in NaturalString.
type Moniker struct {
	value string
}

var monikerType *Type

func NewMoniker(value string) *Moniker {
	return &Moniker{value: value}
}

// Value returns the current string for trusted Go consumers (e.g. DbProvider). Not a Final API.
func (m *Moniker) Value() string {
	if m == nil {
		return ""
	}
	return m.value
}

func (m *Moniker) FinalType() *Type { return monikerType }

// NaturalString redacts the payload for .s / diagnostics.
func (m *Moniker) NaturalString() string {
	return "Moniker(<opaque>)"
}

func (m *Moniker) subKey(keyname, val string) {
	m.value = strings.ReplaceAll(m.value, "$"+keyname, val)
}

func registerMonikerMembers(t *Type) {
	// ( init:String -- mon:Moniker )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		init := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(NewMoniker(init))
	})
}

func CreateRegisterMonikerType() {
	t := &Type{
		name:            "Moniker",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Moniker]("Moniker"),
	}
	registerMonikerMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	monikerType = t
}

func init() {
	CreateRegisterMonikerType()
}

var _ Typed = (*Moniker)(nil)
var _ NaturalStringer = (*Moniker)(nil)

// dsnFromStackCell returns a DSN string from a Final String or *Moniker cell.
func dsnFromStackCell(v any) string {
	switch x := v.(type) {
	case String:
		return string(x)
	case *Moniker:
		return x.Value()
	default:
		panic(fmt.Errorf("DbProvider.new: dsn must be String or Moniker, got %T", v))
	}
}
