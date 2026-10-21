# String Operators

All everyday String helpers are **`str::` stack operators**. They take String operands on **DS** and leave results on **DS**. Non-String inputs panic at runtime.

Operands are never modified in place: “mutating” words return a **new** String.

Type `list-string` and then eg `help str::len` (and friends) in the UI for the operator description an stack effects.

---

## Inquiry

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::len` | `( s:String -- len:i )` | **Byte** length of the Utf-8 encoding (not rune count) |
| `str::isEmpty` | `( s:String -- b )` | `true` if `s` has zero bytes |

```forth
"hello" str::len $ nl          // 5
"é" str::len $ nl              // 2  (one rune, two bytes)
"" str::isEmpty $ nl           // true
```

Empty String is also false under Truthiness, so `str::isEmpty` and a bare `if` on the string agree for control flow — prefer `str::isEmpty` when you want the intent obvious.

---

## Trim

Whitespace follows Unicode space classes (not only ASCII `' '`).

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::trimStart` | `( s -- s' )` | drop leading whitespace |
| `str::trimEnd` | `( s -- s' )` | drop trailing whitespace |
| `str::trim` | `( s -- s' )` | drop both ends |

```forth
"  x.fnl  " str::trim ".fnl" str::endsWith $ nl    // true
```

---

## Case

Applies Unicode casing rules (not ASCII-only).

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::toLower` | `( s -- s' )` | lowercase |
| `str::toUpper` | `( s -- s' )` | uppercase |

```forth
"Hello" str::toLower $ nl    // hello
"αβγ" str::toUpper $ nl      // ΑΒΓ
```

---

## Substitution

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::sub` | `( str:s find:s replace:s -- str' )` | replace **every** occurrence of `find` with `replace` |

Returns a **new** String. If `find` is absent, the result equals `str`. Substitution is exact string match (not Regex).

```forth
"banana" "a" "o" str::sub $ nl    // bonono
"foo" "x" "y" str::sub $ nl       // foo
"aa" "a" "aa" str::sub $ nl       // aaaa
```

For pattern-based replace, see [Regex](../regex.md).

---

## Concatenation

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::cat` | `( a:String b:String -- ab:String )` | concatenate `a` then `b` |
| `str::+` | `( a:String b:String -- ab:String )` | synonym for `str::cat` |

```forth
"hello" " " str::+ "world" str::+ $ nl    // hello world
```

For a few pieces, `str::+` is fine. For building a large result in a loop, use [StringBuilder](stringbuilder.md) instead of repeated `str::+` (each concat allocates a fresh String).

---

## Prefix / suffix / find

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::startsWith` | `( s:String prefix:String -- b )` | `true` if `s` begins with `prefix` |
| `str::endsWith` | `( s:String suffix:String -- b )` | `true` if `s` ends with `suffix` |
| `str::contains` | `( s:String sub:String -- b )` | `true` if `s` contains `sub` |
| `str::indexOf` | `( s:String sub:String -- idx:i )` | **byte** index of first `sub`, or `-1` if absent |

```forth
"internal/final/core/0046-parse.fnl" >ts
t:peek "internal/" str::startsWith $ nl     // true
t:peek ".fnl" str::endsWith $ nl            // true
t:peek "/" str::indexOf $ nl                // 8
t:drop

"banana" "an" str::contains $ nl             // true
"banana" "xy" str::contains $ nl             // false
```
---

## Split and join

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `str::split` | `( s:String sep:String -- parts:StringList )` | split on exact separator sequence |
| `str::splitLines` | `( s:String -- lines:StringList )` | split on `\n`; one trailing `\n` is stripped before splitting |
| `str::join` | `( items:List\|StringList sep:String -- s:String )` | join string items with `sep` |

`str::split` and `str::join` round-trip cleanly with **StringList**. `str::join` also accepts a **List** whose every element is a **String** (non-String items panic).

```forth
"a,b,c" "," str::split >ts
t:peek foreach
  i $ sp
/foreach
nl
"," t:peek str::join $ nl    // a,b,c
t:drop
```

```forth
"alpha\nbeta\n" str::splitLines >ts
t:.len@ $ nl                  // 2
0 t:@() $ nl                  // alpha
1 t:@() $ nl                  // beta
t:drop
```

Notes on `str::splitLines`:

* The empty String yields an **empty** StringList (not a single empty line).
* A final `\n` does not create an extra empty trailing line.
* Only `\n` is the separator; a lone `\r` is left in the line text (CRLF files are usually normalized by [TextReader.readLine](../streams/intro.md), which strips a trailing `\r`).

---

## Patterns you will see in core

Error messages and path checks often chain concat and probes without parking:

```forth
"WARNING: word not found: " "compileFromFile" str::+ $ nl

"notes.fnl" ".fnl" str::endsWith not if
  "expected a .fnl path" $ nl
then
```

When the next step is list processing (sort, foreach, Json), keep the **StringList** from `str::split` / `str::splitLines` on **ts** and use the [List API](../compounds/lists/index.md).

Next: [Iterating Runes](iter-runes.md) when you need code points rather than bytes.
