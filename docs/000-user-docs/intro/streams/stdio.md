# stdin/out Streams

Every Final **Session** / run has three text streams bound for you:

| Stream | Type | Role |
| ------ | ---- | ---- |
| **stdin** | **TextRead** (usually a **TextReader**) | input (REPL lines, scripted input) |
| **stdout** | **TextWrite** (usually a **TextWriter**) | ordinary program output |
| **stderr** | **TextWrite** (usually a **TextWriter**) | diagnostics and fault reports |

You do **not** open these with the file or in-memory factories. The runtime wires them when the Session starts — process stdio for an interactive TUI or headless REPL, capture buffers when a host embeds Final, or null readers/writers when a worker has no input/output.

## Everyday printing: `$` and `$$`

For most programs you never hold the stdout handle yourself. Words you already used write there for you:

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `$` | `( val:a -- )` | print the value’s natural string to **stdout** |
| `$$` | `( val:a -- )` | same natural string, to **stderr** |

```forth
> "Hello from Final!\n" $

Hello from Final!
```

```forth
> "warning" $$
```

Acceptable `$` / `$$` inputs include **String**, **Int**, **Uint**, **Float**, **Bool**, and **DateTime**. Other objects typically print only their type name — not a full dump. For richer formatting of several values, see [Output Formatting](../out-formatting.md) (`fmt` and friends).

### Printing a Rune

A **Rune** is a Unicode code point carried as an **`Int`**. Ordinary `$` on an Int prints the **decimal** code point (`65 $` → `65`), which is rarely what you want for text. Convert first with **`r>s`**, then print:

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `r>s` | `( rune:i -- s )` | Int code point → one-character String |

```forth
> 65 r>s $ 10 r>s $

A
```

When you are already building a `fmt` message, use **`%c`** with the Int code point instead — no `r>s` needed:

```forth
> ( 65 ) "char=%c" fmt $ nl

char=A
```

On a **TextWrite** handle, call **`writeRune`** directly; conversion is only for the `$` / `$$` path (and any other String sink).

## When you care about the handle

* **Pips** copy the parent’s stdout/stderr into the spawn parameter object by default; process stdout/stderr are safe for concurrent writes. Custom writers you create (files, builders) are **not** automatically locked — your program owns them. See [Pips](../threading/pips.md).
* Worker pips get a **null** stdin unless you bind an explicit reader; line-oriented REPL input stays on the main pip.
* Prefer **TextWrite** / **ByteWrite** `write*` methods (and `fmt`) for anything beyond quick `$` prints. Do not treat stdout as a general mux for files or network traffic.

## Null streams

When no real input or output is bound, reads EOF immediately and writes are discarded. That is normal for some embedded or worker setups — not an error by itself.
