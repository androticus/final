# Output Formatting

[`$`](streams/stdio.md) prints one value’s natural string to stdout. [`$$`](streams/stdio.md) prints one value’s natural string to stderr. In `final-gui` and `final`, use **View ▸ Stdout/Stderr** to show the standing stream panes (Prompt stays); **View ▸ Prompt/History** returns to the Response and History panes.

When you need a composed message — several fields, fixed widths, decimals, hex, a quoted string — use **`fmt`** with a **List of values** and a **format string**.

You have already seen this pattern in [Writing Clean Final Code](clean-final-coding.md) and in object examples, e.g. `"%s (%d)" fmt`. [DateTime](datetime.md) also uses `( … )` to build component Lists for factories. This page is the API detail for those sandwiches and for `fmt` itself.

---

## API at a glance

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `(` | `( -- )` r( -- cm ) | open a value sandwich; mark current DS depth |
| `)` | `( …new… -- list:List )` | close sandwich → **List** (Any cells) |
| `s)` | `( …new… -- list:StringList )` | close sandwich → **StringList** (String cells only) |
| `fmt` | `( values:List fmt:s -- result:s )` | substitute values into the format string |

Typical print pipeline:

```forth
( values… ) "format…" fmt $
```

`fmt` returns a **String** on DS. Print it with `$`, concatenate with `str::+`, or hand it to a [TextWriter](streams/intro.md) (`writeString`, `writeLine`).

---

## The `( )` and `( s)` sandwiches

Both sandwiches share the same open word **`(`**. The closer chooses the list type:

| Sandwich | Closer | Result | Typical use |
| -------- | ------ | ------ | ----------- |
| `( … )` | `)` | **List** | `fmt`, DateTime/Interval factories, mixed values |
| `( … s)` | `s)` | **StringList** | `str::join`, string-only lists |

### How they work

1. **`(`** records the current Data Stack depth (a context mark on the return stack). It does **not** push a value.
2. Between `(` and the closer, push the items you want in the list — literals, property reads, method results, anything that leaves a value on DS.
3. **`)`** (or **`s)`**) gathers **only the values pushed after `(`**, packs them into a list in **writing order**, and leaves that list on DS.

List order matches left-to-right source order: the first value after `(` is `list[0]`; the last (TOS before the closer) is `list[n-1]`. That is also the order `fmt` walks for successive `%` markers.

```forth
( 10 20 30 )          // List: 10, 20, 30
( "a" "b" "c" s)      // StringList: "a", "b", "c"
```

`s)` requires every gathered cell to be a **String** (non-strings panic). Use `)` when cells may be Int, Float, Bool, or mixed.

Empty sandwiches are legal: `( )` → empty List; `( s)` → empty StringList.

You can also build a List with `List.new^` / `.append^` and pass that to `fmt`. The sandwich is the concise form for a fixed handful of values.

---

## Only *new* values — a common trip-up

**The sandwich does not include values that were already on DS when `(` ran.** It collects solely the *new* cells placed on DS after the open.

```forth
"Ada" 36
( ) "%s (%d)" fmt $     // WRONG: empty list — Ada and 36 sit *below* the mark
```

```forth
"Ada" 36
( "Ada" 36 ) "%s (%d)" fmt $   // works, but duplicates the literals
```

If the values you need are already on DS (or live on an object you are using), **do not expect `)` to “reach under” the mark**. Produce them *inside* the sandwich, or park them first and push fresh copies/reads after `(`.

### Prefer parking on an auxiliary stack

Final’s aux stacks exist for this kind of problem ([Stack Operators](stacks/operators.md), [Clean Final Code](clean-final-coding.md)). Park handles or operands, then emit values *inside* `( … )`:

```forth
: printFullName ( inst:o -- )
  >ts
  ( t:.prefix@ t:.name@ t:.affix@ )
  "%s%s%s" fmt $
  t:drop
;
```

```forth
"Ada" >ts
36 >xs
( t:peek x:peek ) "%s is %d" fmt $
x:drop t:drop
```

**Cautions**

* Stack juggling after `(` to `pick` / `over` values from *below* the mark can work (a copied cell is “new”), but it is hard to read and easy to get wrong — park instead.
* Ops inside the sandwich that drop or rearrange cells *below* the mark can corrupt the caller’s stack. Keep the sandwich body a straight sequence of pushes (and rooted reads that leave results on DS).
* `s)` will reject non-String cells; if you need mixed types for `fmt`, use `)`.

---

## `fmt`

```forth
// ( values:List fmtString:s -- result:s )
( 42 "widgets" ) "%d %s" fmt $ nl    // 42 widgets
```

Rules of thumb:

* One value in the List per `%` substitution (not counting `%%`).
* Prefer an explicit type marker (`%d`, `%s`, …) over a catch-all when you know the type.
* A count mismatch does not always abort the word; missing or extra substitutions show up as error text inside the result. Keep the List and the format string in sync.

`fmt` is for **building a String**. It does not write to stdout by itself — follow with `$` or a writer method when you want output.

---

## Format codes (cheat sheet)

Format strings are ordinary Final Strings. Substitutions use **`%`** markers (C-style). Literal text copies through unchanged. Use **`%%`** for a literal percent sign.

### Type markers (Final-relevant)

| Code | Expects | Result |
| ---- | ------- | ------ |
| `%v` | any | default rendering of the value |
| `%s` | String | string contents |
| `%d` | Int (also fine for many integral values) | decimal integer |
| `%f` | Float | decimal floating point |
| `%t` | Bool | `true` / `false` |
| `%c` | Int (Unicode code point) | one character |
| `%x` / `%X` | Int | hex (lower / upper) |
| `%o` | Int | octal |
| `%b` | Int | binary |
| `%e` / `%E` | Float | scientific notation |
| `%g` / `%G` | Float | compact float (decimal or scientific) |
| `%q` | String | double-quoted string with escapes |
| `%%` | — | literal `%` |

`%v` is a reasonable default when exploring. For production messages, prefer the typed codes so Int/Float/Bool/String stay obvious at the call site.

`%c` wants an **Int** code point (same idea as [`r>s $`](streams/stdio.md) — interpret the Int as a character), not a one-character String. Prefer `%c` inside `fmt`; use `r>s` when you need a String for `$` / `$$` or other String APIs.

### Width and precision

Insert width and precision between `%` and the type letter:

| Form | Meaning | Example | Sample result |
| ---- | ------- | ------- | ------------- |
| `%wd` | minimum width `w`, pad with spaces on the left | `( 42 ) "%5d" fmt` | `"   42"` |
| `%0wd` | pad with leading zeros | `( 42 ) "%05d" fmt` | `"00042"` |
| `%-ws` | width `w`, pad on the right (left-align) | `( "ab" ) "%-6s" fmt` | `"ab    "` |
| `%.pf` | precision `p` (digits after decimal for `%f`) | `( 3.14159 ) "%.2f" fmt` | `"3.14"` |
| `%w.pf` | width and precision together | `( 3.14159 ) "%8.2f" fmt` | `"    3.14"` |
| `%+.1f` | always show a sign | `( 1.5 ) "%+.1f" fmt` | `"+1.5"` |

Width applies similarly to `%s`, `%d`, `%x`, and friends. Precision for `%s` caps the number of characters taken from the string; for integers it is less often useful than for floats.

```forth
( "n" 7 3.5 ) "%s=%02d rate=%.1f%%" fmt $ nl
// n=07 rate=3.5%
```

---

## DateTime and Interval

`fmt` is for **general** messages (`%d`, `%s`, width/precision). It is **not** the DateTime/Interval layout API.

To format a **DateTime** or **Interval** with a letter-run pattern (or a curated constant such as `DateTime.iso` / `Interval.clock`), use **`dt::fmt`**:

```forth
DateTime.now^ DateTime.iso dt::fmt $ nl
5 Interval.minute dt::* Interval.clock dt::fmt $ nl
```

Full pattern tables, cascade rules, sign handling, and format constants: [DateTime and Interval — Formatting](datetime.md#formatting-with-dtfmt).

---

## Examples

### Several fields, then print

```forth
( "sensors" 3 true ) "%s: count=%d enabled=%t\n" fmt $
```

### Single value still needs a List

```forth
( Int.maxInt ) "%d" fmt $ nl
( 255 ) "0x%02X" fmt $ nl
```

### Join strings with `( s)`

```forth
( "alpha" "beta" "gamma" s) ", " str::join $ nl
// alpha, beta, gamma
```

(`str::join` is covered under [String Operators](strings/ops.md); `( s)` is the handy way to hand it a StringList.)

### Write into a builder / file

```forth
StringBuilder.new^ >ts
( "maxInt" Int.maxInt ) "%s=%d" fmt t:.writeString^
t:.toString^ $ nl
t:drop
```

---

## Notes

* **`$` vs `fmt`** — `$` is one-shot natural print. `fmt` builds a String you can print, store, or send elsewhere.
* **Objects** — prefer pulling primitive fields (with rooted reads) into the sandwich rather than formatting a whole compound with `%v`.
* **Related** — [stdin/out Streams](streams/stdio.md) for `$` / `$$` / `r>s`; [StringBuilder](strings/stringbuilder.md) and [TextWriter](streams/intro.md) for sinks; [DateTime](datetime.md) for `( parts )` factory Lists and for **`dt::fmt`** pattern formatting.
