# Final for command line scripts and OS Pipes

Final supports command line operation to run scripts written in Final, or to be the target of a pipe. It uses semantics similar to Python and node.

All of the below assume you have put the Final folder on your path. Otherwise, use ./final.

To launch `final` and run a script, without launching the interactive tui, use this format:

```bash
final commandScript.fnl [ param1 param2 ...]
```

To pipe text into Final's stdin, do the following:

```bash
someShellCmd | final commandScript.fnl [ param1 param2 ... ]
```

When you actually want to pipe in Final interpretable text, do the following. The - is only needed if you add params to the command line -- it is a placeholder for the stdin input.
```bash
echo '1 2 + $ nl' | final - [ param1 param2 ... ]
>3
```
Final also support creating executable final scripts on Mac/Linux, using shebang invocation. (Note: as with all executable scripts, you must mark it execuable using chmod.) `final` must be on the PATH.

eg
```
#!/usr/bin/env final
sys.argv >json $
```

In all cases, the command and any parameters are available in Final via the Sys.params StringList object. `Sys` instance is available via the axiom `sys`.
params[0] contains the command, or `-`
params[1..] contain any params you passed.

```forth
// this assumes there was a command. otherwise params is empty
0 sys.params.@() $
```

If you need to access the raw, unprocessed command line arguments, C-style, they are available in sess.argv -- [0] is the text that launched Final, eg `./final`, then the remaining values are the individual arguments. Note: you should generally not need to access this.
