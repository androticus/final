package core

const replProgramName = "Final"

// bootstrapWriteREPLBanner prints the program name and version on stdout (interactive REPL).
func bootstrapWriteREPLBanner(rp *RunProvider) {
	if rp == nil || rp.Stdout == nil {
		return
	}
	rp.Stdout.WriteString(replProgramName + " v" + Version() + "\n")
	rp.Stdout.WriteString("Type 'bye' to exit; 1-Ctrl+C to break program, 2-quick to terminate;\n")
	flushTextWrite(rp.Stdout)
}
