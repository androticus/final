// Package finalenbed is the a advanced public API for embedding the Final language in Go programs.
// It enables you to declare and register your own Types into Final.
//
// Quick start:
//
//	sess, err := finalembed.NewSession()
//	if err != nil {
//		panic(err)
//	}
//	defer sess.Close()
//
//	out, err := sess.GetSender().Send("1 2 + $", 0)
//
// Implementation details live in internal/core. For the terminal REPL, see RunInteractiveREPL.
//
// See https://github.com/androticus/final/finalgo for the "turnkey" embedding API.
package finalembed
