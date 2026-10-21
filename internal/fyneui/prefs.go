package fyneui

import (
	"github.com/androticus/final/internal/core"
)

func copyToStringList(src []string) *core.StringList {
	l := core.NewStringListWithCapacity(len(src))
	for _, s := range src {
		l.Append(s)
	}
	return l
}

func copyToIntList(src []int) *core.IntList {
	l := core.NewIntListWithCapacity(len(src))
	for _, n := range src {
		l.Append(n)
	}
	return l
}

func copyToFloatList(src []float64) *core.FloatList {
	l := core.NewFloatListWithCapacity(len(src))
	for _, n := range src {
		l.Append(n)
	}
	return l
}

func sliceFromStringList(l *core.StringList) []string {
	n := l.Len()
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = l.At(i)
	}
	return out
}

func sliceFromIntList(l *core.IntList) []int {
	n := l.Len()
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = l.At(i)
	}
	return out
}

func sliceFromFloatList(l *core.FloatList) []float64 {
	n := l.Len()
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = l.At(i)
	}
	return out
}

func registerPrefsMembers(t *core.Type) {
	// ( key:s default:b -- b )
	t.RegisterNativeClassMethod("getBool", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		def := bool(ds.Pop().(core.Bool))
		key := string(ds.Pop().(core.String))
		ds.Push(core.Bool(TheApp.Preferences().BoolWithFallback(key, def)))
	})
	// ( val:b key:s -- )
	t.RegisterNativeClassMethod("setBool", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		val := bool(ds.Pop().(core.Bool))
		TheApp.Preferences().SetBool(key, val)
	})

	// ( key:s default:i -- i )
	t.RegisterNativeClassMethod("getInt", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		def := int(ds.Pop().(core.Int))
		key := string(ds.Pop().(core.String))
		ds.Push(core.Int(TheApp.Preferences().IntWithFallback(key, def)))
	})
	// ( val:i key:s -- )
	t.RegisterNativeClassMethod("setInt", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		val := int(ds.Pop().(core.Int))
		TheApp.Preferences().SetInt(key, val)
	})

	// ( key:s default:f -- f )
	t.RegisterNativeClassMethod("getFloat", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		def := float64(ds.Pop().(core.Float))
		key := string(ds.Pop().(core.String))
		ds.Push(core.Float(TheApp.Preferences().FloatWithFallback(key, def)))
	})
	// ( val:f key:s -- )
	t.RegisterNativeClassMethod("setFloat", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		val := float64(ds.Pop().(core.Float))
		TheApp.Preferences().SetFloat(key, val)
	})

	// ( key:s default:s -- s )
	t.RegisterNativeClassMethod("getString", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		def := string(ds.Pop().(core.String))
		key := string(ds.Pop().(core.String))
		ds.Push(core.String(TheApp.Preferences().StringWithFallback(key, def)))
	})
	// ( val:s key:s -- )
	t.RegisterNativeClassMethod("setString", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		val := string(ds.Pop().(core.String))
		TheApp.Preferences().SetString(key, val)
	})

	// ( key:s -- l:IntList true | false )
	t.RegisterNativeClassMethod("getIntList", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		vals := TheApp.Preferences().IntListWithFallback(key, nil)
		if vals == nil {
			ds.Push(core.Bool(false))
			return
		}
		ds.Push(copyToIntList(vals))
		ds.Push(core.Bool(true))
	})
	// ( val:IntList key:s -- )
	t.RegisterNativeClassMethod("setIntList", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		l := ds.Pop().(*core.IntList)
		TheApp.Preferences().SetIntList(key, sliceFromIntList(l))
	})

	// ( key:s -- l:FloatList true | false )
	t.RegisterNativeClassMethod("getFloatList", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		vals := TheApp.Preferences().FloatListWithFallback(key, nil)
		if vals == nil {
			ds.Push(core.Bool(false))
			return
		}
		ds.Push(copyToFloatList(vals))
		ds.Push(core.Bool(true))
	})
	// ( val:FloatList key:s -- )
	t.RegisterNativeClassMethod("setFloatList", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		l := ds.Pop().(*core.FloatList)
		TheApp.Preferences().SetFloatList(key, sliceFromFloatList(l))
	})

	// ( key:s -- l:StringList true | false )
	t.RegisterNativeClassMethod("getStringList", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		vals := TheApp.Preferences().StringListWithFallback(key, nil)
		if vals == nil {
			ds.Push(core.Bool(false))
			return
		}
		ds.Push(copyToStringList(vals))
		ds.Push(core.Bool(true))
	})
	// ( val:StringList key:s -- )
	t.RegisterNativeClassMethod("setStringList", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ds := f.RP.DataStack
		key := string(ds.Pop().(core.String))
		l := ds.Pop().(*core.StringList)
		TheApp.Preferences().SetStringList(key, sliceFromStringList(l))
	})

	// ( key:s -- )
	t.RegisterNativeClassMethod("remove", func(f *core.RunFrame, _ *core.TypeObjPair) {
		key := string(f.RP.DataStack.Pop().(core.String))
		TheApp.Preferences().RemoveValue(key)
	})
}

func CreateRegisterPrefsType() {
	t := core.NewRegisteringType("Prefs",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Prefs"),
	)
	registerPrefsMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterPrefsType()
}
