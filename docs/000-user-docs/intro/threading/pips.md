# Pips

A `Pip` is a Final virtual machine with its own stacks and instruction pointer, running on its own lightweight thread — independent of the pip that started it. Think of it as a sprightly little worker.

All pips in a Session share one dictionary context (dictionary, code table, and data table). Only the **main** pip may compile or define new words. Once any child pip is running, the dictionary is locked: compile and define attempts raise an error. A dictionary may also be sealed after a successful compile (for example, core after bootstrap); that is separate from the live-child lock.

Pips are lightweight — each needs only a small set of empty stacks to exist — so you can create many of them when that fits the work, especially for blocking I/O such as files or the network. Still, a machine has a limited number of cores that actually run at once. For CPU-bound “split the work” jobs, keep subdivision modest (for example, on the order of 16 pips), not hundreds.

## Spawning

Spawning starts a word with a parameter object in a new Pip (thread). When that word returns, the pip exits. There is no join: you do not wait on a pip handle. Coordinate with [pipes](pipes.md) (and optionally shared atomics).

To spawn a new Pip (thread):

```forth
Pip.newPipParams^ spawn: doWork
```

`spawn:` reads the next identifier, checks that the word is suitable to spawn, and launches it. For lower-level use, `(spawnPip)` takes the parameter object and an execution token obtained with `&of`. Use `help spawn:` and `help (spawnPip)` for stack effects and details.

The word you spawn should be an ordinary colon word — not immediate, hidden, compile-only, or still being defined.

Note that the root Task in a Session runs under the Final Interpreter. Spawned words do not, and only execute the code in the spawned word.

## PipParams

“PipParams” is a convention, not a separate type. `Pip.newPipParams^` builds an `AtomicDynamicObject` parameter object (threadsafe if the object is shared across pips) with:

| Field | Default | Role |
| ----- | ------- | ---- |
| `_stdout` | current stdout (**TextWrite**) | Required at spawn — child’s stdout |
| `_stderr` | current stderr (**TextWrite**) | Required at spawn — child’s stderr |
| `_name` | *(none)* | Optional label for diagnostics; otherwise the spawn word’s name |

Put your own payload fields directly on the parameter object (for example `.n` or `.pipe`). Keys reserved for the system start with `_`. The whole object is sitting on the child’s data stack when the pip starts.

## Pip API

`Pip` is a registered Final type for class members only, such as `newPipParams^`.

## Faults

Final panics mean programmer error (stack underflow, bad types, and the like) — not normal control flow. Expected failures should use optional results instead.

A pip must not fail silently. By default, the **Session** is the failure domain: if one pip faults, the whole Session run is treated as faulted. Sibling pips are stopped; there is no “catch in the parent and keep going” path, and still no join. Other Sessions are unaffected.

In practice: close pipes (or otherwise unblock waiters) when shutting down work, so a blocked pull does not hang forever after a fault elsewhere.

## Stdio

- By default, `Pip.newPipParams^` copies the current stdout and stderr into the parameter object; the child uses those same writers. Process stdout/stderr are safe for concurrent writes from several pips (including fault reports).
- Custom sinks (builders, files, network, and so on) are **not** automatically locked — your program owns those streams.
- Prefer `fmt` and `write*` on a **TextWrite** / **ByteWrite** for real I/O. Do not treat stdout as a general file or socket mux.
- Stdin: line-oriented REPL input stays on the main pip. Worker pips get null stdin unless you bind an explicit **TextRead**.
