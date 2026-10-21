package core

import "fmt"

// bootstrapFinalFmt implements fmt: ( args:AnyList fmt:String -- result:String )
func bootstrapFinalFmt(f *RunFrame) {
	ds := f.RP.DataStack
	format := ds.Pop().(string)
	args := ds.Pop().(*AnyList)
	ds.Push(String(fmt.Sprintf(format, args.items...)))
}
