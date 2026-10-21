// Package finalgo is "turnkey" public API for embedding the Final language in Go programs.
//
// Quick start:
//
//	sess, err := finalgo.NewSession()
//	if err != nil {
//		panic(err)
//	}
//	defer sess.Close()
//
//	out, err := sess.GetSender().Send("1 2 + $", 0)
//
// Implementation details live in internal/core. For the terminal REPL, see RunInteractiveREPL.
//
// See https://github.com/androticus/final/finalembed for the advanced embedding API. for more details.
package finalgo
