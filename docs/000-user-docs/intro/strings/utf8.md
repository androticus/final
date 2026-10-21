# Unicode and Utf-8

Final stores every **`String`** as a sequence of **Utf-8** bytes. A **Rune** (documented under [Basic Types](../quick/basic-types.md) as an `Int` synonym) is a Unicode **code point** — the abstract character value — not a byte. One Rune encodes as **one to four** Utf-8 bytes.

You already know that ASCII-7 is a proper subset of Utf-8: English source, Json keys, and HTML tags that stay in ASCII look identical as bytes whether you call them “ASCII” or “Utf-8.” Final source itself is Utf-8; old DOS/Windows “code pages” are not accepted ([Interpreter](../quick/interpreter.md)).

## How the encoding works

Utf-8 uses the **high bits** of each byte to mark where characters start and how long they are:

| Byte pattern (binary) | Role | Payload bits |
| --------------------- | ---- | ------------ |
| `0xxxxxxx` | single-byte character (U+0000…U+007F, i.e. ASCII) | 7 |
| `110xxxxx` | start of a **2**-byte sequence | 5 + … |
| `1110xxxx` | start of a **3**-byte sequence | 4 + … |
| `11110xxx` | start of a **4**-byte sequence | 3 + … |
| `10xxxxxx` | **continuation** byte (never a start) | 6 |

Examples:

* `'A'` → one byte `0x41`
* `'é'` (U+00E9) → two bytes `0xC3 0xA9`
* `'日'` (U+65E5) → three bytes
* `'🙂'` (U+1F642) → four bytes

Only the **shortest** encoding of a code point is legal. Overlong forms (for example encoding ASCII as a two-byte sequence) are invalid. That rule is part of what keeps the encoding unambiguous.

## Qualities that matter in practice

These properties are why Final (like Go) can treat Utf-8 as the everyday string encoding without a large runtime tax.

### Deterministic find; no ambiguous overlaps

Because ASCII bytes **never** appear inside a multi-byte sequence, a search for an ASCII needle — `"/"`, `".fnl"`, `"\n"`, a Json delimiter — cannot false-match in the middle of an emoji or accented letter. **`str::startsWith`**, **`str::endsWith`**, **`str::indexOf`**, and **`str::split`** are byte scans that stay **boundary-safe** when the probe itself is well-formed Utf-8 (including plain ASCII).

The same bit pattern that marks continuation bytes (`10xxxxxx`) also makes Utf-8 **self-synchronizing**: from a random offset you can walk forward to the next start byte without restarting from the beginning of the string. That is the engineering reason “find” and “split on ASCII” stay simple and reliable.

### ASCII-majority text stays fast

Most Final programs spend their time on source text, paths, protocol tokens, and similar ASCII-heavy material. For those, length checks, concatenation, prefix/suffix tests, and splits are ordinary **byte** operations — the same cost model as a classic C string, without a hidden per-character decode on every call.

You pay for decoding when you **ask** for characters: `foreach` over a String, **`TextReader.getNextRune`**, or converting with **`r>s`** before `$` / `$$`. That cost is linear in the text you walk, and for ASCII the decoder’s fast path is essentially “one byte → one rune.”

Kernighan and Donovan’s *The Go Programming Language* walks through this model in detail (strings as bytes, range/decode for runes, and why indexing by “character” is the wrong default). The short version for Final users: **Utf-8 is not a tax on every string op**; it is a compact encoding with a clear, cheap ASCII fast path and an explicit decode step when you need code points.

### What “length” means

| Question | Answer in Final |
| -------- | ---------------- |
| How many **bytes** of storage? | `str::len` |
| How many **code points** (Runes)? | Count them with `foreach` (see [Iterating Runes](iter-runes.md)) |
| How many **grapheme clusters** (user-perceived characters)? | Not a core API — combining marks and emoji sequences need a higher-level library |

Because multi-byte Runes exist, **`str::len` and the rune count are different** whenever the text leaves ASCII. That is expected, not a bug.

## Safety model (short)

| Operation | Safe for general Unicode? | Notes |
| --------- | ------------------------- | ----- |
| `str::startsWith` / `endsWith` / `indexOf` / `split` | Yes for well-formed Utf-8 probes | Byte match; ASCII probes cannot land inside a multi-byte rune |
| `str::indexOf` result | Byte offset | Useful for ASCII wire tokens; do **not** invent a byte slice from it |
| Byte-indexed substring | **Not provided** | Would risk splitting a character; build with [StringBuilder](stringbuilder.md) or walk runes |
| Ordering (`<` `>` on Strings) | Byte order | Fine for ASCII; not locale collation |

Next: the concrete [String Operators](ops.md).
