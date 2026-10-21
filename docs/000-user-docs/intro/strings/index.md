# Strings

The [Quick Tour](../quick/basic-types.md) already introduced **`String`**: immutable Utf-8 text, ASCII-7 as a proper subset, and the caution that **`str::len` is a byte length**, not a character count. Literals (quoted, raw, and rune ticks) are covered under [Interpreter](../quick/interpreter.md). Comparison of two Strings uses Utf-8 **byte** ordering ([Basic Operators](../quick/basic-ops.md)).

This section is the day-to-day **API** for working with text:

| Page | Contents |
| ---- | -------- |
| [Unicode and Utf-8](utf8.md) | How Utf-8 is coded, why finds stay unambiguous, and why that does not slow ordinary string work |
| [String Operators](ops.md) | The full `str::` family: length, trim, case, concat, probes, split/join |
| [Iterating Runes](iter-runes.md) | `foreach` over a String (code points as `Int`) plus a worked census example |
| [StringBuilder](stringbuilder.md) | Mutable accumulation when many concatenations would be costly |

A few design points that shape every API below:

* **Immutability** — every `str::` “mutation” returns a **new** String; the inputs are unchanged.
* **Stack operators, not instance methods** — text helpers are the `str::…` words (same family style as other axiom groups). Park a String on **ts** only when you need several rooted object ops around it; for pure string transforms, keep values on **DS**.
* **No byte-indexed substring** — cutting at an arbitrary byte offset can split a multi-byte character. Build pieces with **`StringBuilder`**, use `str::indexOf` /`str::sub` / `str::split`, or iterate runes (this section and [TextReader](../streams/intro.md)).
* **Truthy** — `""` is false; any non-empty String is true ([Basic Types](../quick/basic-types.md)).

For line-oriented or file-backed text, prefer [Text and Byte Streams](../streams/index.md). For pattern matching, see [Regex](../regex.md) (later in the survey).
