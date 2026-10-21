package core

import "fmt"

// These are go routines that implement the final-native bootstrap words.
// Don't confuse with bootstrap* words, that were for the go native repl

func bootstrapFinalFind(f *RunFrame) { // ( name:String -- w true ) | ( -- false )
	ds := f.RP.DataStack
	name := ds.Pop().(string)
	w, _, ok := FindNameInTask(f.Task, f.RP, String(name))
	if !ok {
		ds.Push(false)
		return
	}
	ds.Push(w)
	ds.Push(true)
}

func bootstrapFinalCurrentDictHasWord(f *RunFrame) { // ( name:String -- bool )
	ds := f.RP.DataStack
	name := ds.Pop().(string)
	found := false
	if f.RP.FindScope != nil {
		found = f.RP.FindScope.FindWordByName(name) != nil
	} else if f.Task != nil && f.Task.dc != nil {
		found = f.Task.dc.dict.HasName(name)
	}
	ds.Push(found)
}

func bootstrapFinalReplError(f *RunFrame) { // ( msg:String -- )
	ds := f.RP.DataStack
	msg, ok := ds.Pop().(String)
	if !ok {
		bootstrapHandleReplError(f, fmt.Errorf("replError: expected String message"))
		return
	}
	bootstrapHandleReplError(f, fmt.Errorf("%s", msg))
}

func bootstrapFinalGetNextToken(f *RunFrame) { // ( -- value type ) | ( -- TokenEOF )
	runGetNextToken(f.Task, f.RP)
}

func bootstrapFinalReadIdentifier(f *RunFrame) { // ( -- name:String )
	f.RP.Parser.PushNextAsIdentifier(f.RP.DataStack)
}

func bootstrapFinalValidatePlainIdent(f *RunFrame) { // ( name:String -- name:String )
	ds := f.RP.DataStack
	nameRaw := ds.Pop()
	name, ok := nameRaw.(String)
	if !ok {
		panic(fmt.Errorf("validatePlainIdent: expected String name, got %T", nameRaw))
	}
	if err := ValidatePlainIdentName(name); err != nil {
		panic(err)
	}
	ds.Push(name)
}

func bootstrapFinalCompileFromFile(f *RunFrame) { // ( filename:String -- )
	ds := f.RP.DataStack
	pathRaw := ds.Pop()
	path, ok := pathRaw.(String)
	if !ok {
		panic(fmt.Errorf("compileFromFile: expected String path, got %T", pathRaw))
	}
	if err := compileBootstrapFile(f.Task, f.RP, string(path)); err != nil {
		panic(err)
	}
	if f.RP.Parser != nil {
		f.RP.Parser.clearParserLocation()
	}
}

func bootstrapFinalCompileFromManifest(f *RunFrame) { // ( filename:String -- )
	ds := f.RP.DataStack
	pathRaw := ds.Pop()
	path, ok := pathRaw.(String)
	if !ok {
		panic(fmt.Errorf("compileFromManifest: expected String path, got %T", pathRaw))
	}
	if err := compileBootstrapManifestFile(f.Task, f.RP, string(path)); err != nil {
		panic(err)
	}
	if f.RP.Parser != nil {
		f.RP.Parser.clearParserLocation()
	}
}

func bootstrapFinalSetCompileState(f *RunFrame) { // ( flag:Bool -- )
	if f.Task == nil {
		panic("bootstrapFinalSetCompileState: nil Task")
	}
	v := f.RP.DataStack.Pop().(Bool)
	SetCompileState(f.Task.compilerForTarget(), bool(v))
}

// bootstrapFinalHandleCompound is the OpTransition target for FinalHandleCompound (legacy; interpret uses OpHandleCompound).
func bootstrapFinalHandleCompound(f *RunFrame) {
	handleCompound(f)
}
