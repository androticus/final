# StringBuilder

**StringBuilder** is the mutable partner to immutable **String**. Append runes, strings, or lines as you go, then take a snapshot with **`toString`**. Use it when a loop would otherwise call `str::+` over and over (each concat allocates a new String).

You already saw a short builder example under [Methods](../objects/methods.md) and the stream-shaped write members under [String Streams](../streams/string.md). This page is the full API.

Park the builder on **ts** (or another aux stack) for a session of writes, then `t:drop` — same nesting style as other objects ([Writing Clean Final Code](../clean-final-coding.md)).

---

## Factories

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( -- sb )` | empty builder |
| class | `newWithPretty` | `( indent:String -- sb )` | pretty mode: indentation string for nested structured output |

```forth
StringBuilder.new^
"  " StringBuilder.newWithPretty^
```

Pretty builders are aimed at structured renderers (for example Json). For ordinary path joins, messages, and file assembly, plain **`new`** is enough.

---

## Properties

| Name | Type | Description |
| ---- | ---- | ----------- |
| `pretty` | Bool | `true` if created with `newWithPretty` (read-only) |
| `indentString` | String | indent unit used when pretty (read-only; empty for a plain builder) |

---

## Write and snapshot

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| method | `writeRune` | `( r:i -- )` | append one Unicode code point |
| method | `writeString` | `( s:String -- )` | append as-is (no added newline) |
| method | `writeLine` | `( s:String -- )` | append string, then `\n` |
| method | `toString` | `( -- s:String )` | current contents as an immutable String |

```forth
StringBuilder.new^ >ts
"hello" t:.writeString^
32 t:.writeRune^
"world" t:.writeLine^
t:.toString^ $              // hello world\n
t:drop
```

`toString` does not clear the builder; further writes keep appending. Build a fresh **StringBuilder** when you want a clean buffer.

---

## Pretty / nesting helpers

Only meaningful when **`pretty`** is true (`newWithPretty`). Plain builders treat the pretty writes as ordinary writes (or no-ops where noted).

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| method | `writeStringPretty` | `( s:String -- )` | write current indent (if nested), then `s` |
| method | `writeSpacePretty` | `( -- )` | write a space if pretty; otherwise nothing |
| method | `writeWhitespacePretty` | `( -- )` | write `\n` if pretty, else a single space |
| method | `nest` | `( -- )` | increase indent level by one |
| method | `unnest` | `( -- )` | decrease indent level by one (stops at 0) |

```forth
"  " StringBuilder.newWithPretty^ >ts
"{" t:.writeString^
t:.writeWhitespacePretty^
t:.nest^
"\"name\"" t:.writeStringPretty^
t:.writeSpacePretty^
":" t:.writeString^
t:.writeSpacePretty^
"\"Ada\"" t:.writeString^
t:.unnest^
t:.writeWhitespacePretty^
"}" t:.writeString^
t:.toString^ $ nl
t:drop
```

Typical pretty output:

```
{
  "name" : "Ada"
}
```

---

## When to choose what

| Need | Prefer |
| ---- | ------ |
| One or two concatenations | `str::+` / `str::cat` ([String Operators](ops.md)) |
| Many appends in a loop | **StringBuilder** |
| Character-at-a-time parse/build | Runes via `foreach` / **TextReader**, append with `writeRune` |
| Line-oriented file I/O | **TextReader** / **TextWriter** ([Streams](../streams/index.md)) |
| Indented structured text | `newWithPretty` + nest/unnest |

Strings remain Utf-8 end to end: every `writeRune` encodes a code point; `toString` yields a normal immutable **String** ready for `str::` operators or `$`.

---

## Using a StringBuilder as a TextWrite value

StringBuilder implements the TextWrite interface. This means you can supply a StringBuilder to contexts that accept a TextWrite, such as when you want to capture the stdout/stderr of a [Pip (thread)](../threading/index.md); or in Http bodies expecting a TextWrite Stream.

