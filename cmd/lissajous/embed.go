package main

import (
	"fmt"
	"os"

	"github.com/androticus/final/finalembed"
)

// registerLissajous exposes *Lissajous to Final: a `new` class method (book defaults),
// read/write properties for the parameters, and a `generate` instance method that writes
// the animated GIF to a file path supplied by the script.
func registerLissajous(t *finalembed.Type) {
	t.RegisterNativeClassMethod("new", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		f.RP.DataStack.Push(NewLissajous())
	})

	t.RegisterProperty("cycles", "Int",
		func(inst any) any { return finalembed.Int(inst.(*Lissajous).cycles) },
		func(inst any, v any) { inst.(*Lissajous).cycles = int(v.(finalembed.Int)) },
	)
	t.RegisterProperty("res", "Float",
		func(inst any) any { return finalembed.Float(inst.(*Lissajous).res) },
		func(inst any, v any) { inst.(*Lissajous).res = float64(v.(finalembed.Float)) },
	)
	t.RegisterProperty("size", "Int",
		func(inst any) any { return finalembed.Int(inst.(*Lissajous).size) },
		func(inst any, v any) { inst.(*Lissajous).size = int(v.(finalembed.Int)) },
	)
	t.RegisterProperty("nframes", "Int",
		func(inst any) any { return finalembed.Int(inst.(*Lissajous).nframes) },
		func(inst any, v any) { inst.(*Lissajous).nframes = int(v.(finalembed.Int)) },
	)
	t.RegisterProperty("delay", "Int",
		func(inst any) any { return finalembed.Int(inst.(*Lissajous).delay) },
		func(inst any, v any) { inst.(*Lissajous).delay = int(v.(finalembed.Int)) },
	)

	t.RegisterNativeMethod("generate", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(finalembed.String))
		file, err := os.Create(path)
		if err != nil {
			panic(fmt.Errorf("Lissajous.generate: create %q: %w", path, err))
		}
		defer file.Close()
		if err := ths.Obj.(*Lissajous).Generate(file); err != nil {
			panic(fmt.Errorf("Lissajous.generate: %w", err))
		}
	})
}

// runEmbedded registers Lissajous with an embedded Final session and drives it from a
// Final script that configures the generator and writes the GIF to outPath.
func runEmbedded(outPath string) error {
	sess, err := finalembed.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	if err := sess.RegisterTypes([]finalembed.TypeReg{
		finalembed.NewTypeReg[*Lissajous]("Lissajous", registerLissajous),
	}); err != nil {
		return err
	}

	script := fmt.Sprintf(`Lissajous.new^ >ts
8 t:.cycles!
150 t:.size!
"%s" t:.generate^
`, outPath)

	if _, err := sess.GetSender().Send(script, 0); err != nil {
		return err
	}
	return nil
}
