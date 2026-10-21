# Regex

**Regex** is Final’s sealed type for regular-expression matching. It wraps Go’s RE2 engine (`regexp`): linear-time matching, no backreferences, and no lookaround. Prefer [String Operators](strings/ops.md) (`str::startsWith`, `str::indexOf`, …) when an exact substring or fixed separator is enough; reach for Regex when the pattern itself varies.

The [Quick Tour](quick/features.md) already flagged “compiled regex, matches, selectors.” [Strings](strings/index.md) pointed here for pattern matching. This page is the API and the day-to-day patterns.

Type `help Regex` (or `help Regex#find`) in the TUI for the same stack effects inline.

---

## When to use what

| Approach | Use when | How |
| -------- | -------- | --- |
| Exact string probes | Fixed prefix, suffix, or substring | `str::…` ([ops](strings/ops.md)) |
| **One-shot** Regex | A single check or extract; pattern is not reused | Class methods `Regex.match^` / `Regex.find^` |
| **Compiled** Regex | Same pattern over many subjects, or you need find-all / split / replace / captures | `Regex.new^`, park on **ts**, call instance methods |

Stack order is **subject then tool**: the String is deeper; the pattern (or replacement) is TOS — same posture as other Final APIs.

Patterns are easiest as **raw strings** (backticks), so `\d` and friends need no double-escaping ([Interpreter](quick/interpreter.md)).

Invalid pattern text is a programming error: `Regex.new^` and the one-shot helpers **panic**. While drafting in the REPL, probe with `Regex.isPatternValid?` first.

---

## One-shot (class)

No `Regex` instance is left on the stack. Fine for a one-off validation or first-match extract.

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `new` | `( pattern:s -- re:Regex )` | compile; panics if invalid |
| `isPatternValid?` | `( pattern:s -- b )` | `true` if the pattern would compile |
| `match` | `( s:String pattern:s -- b )` | `true` if the pattern matches somewhere in `s` |
| `find` | `( s:String pattern:s -- found:s true \| false )` | first match substring, or modal `false` |
| `quoteMeta` | `( s -- q:s )` | escape metacharacters so `s` matches as a literal |

```forth
"42" `^\d+$` Regex.match^ $ nl              // true

"hello 42 there" `\d+` Regex.find^ not if
  "no digits" $ nl
else
  $ nl                                      // 42
then

`(unclosed` Regex.isPatternValid? $ nl      // false
`price $` Regex.quoteMeta $ nl              // price \$
```

`find` is modal ([Writing Clean Final Code](clean-final-coding.md)): success leaves the substring then `true`; failure leaves only `false`. A successful **empty** match is `"" true`, not `false`.

---

## Compiled (instance)

Build once with `Regex.new^`, reuse on many subjects. Park the receiver on **ts** for a short session of method calls ([Methods](objects/methods.md), [Clean Final](clean-final-coding.md)).

### Properties

| Name | Type | Description |
| ---- | ---- | ----------- |
| `pattern` | String | pattern text |
| `numGroups` | Int | number of capturing groups (not counting the whole match) |
| `groupNames` | StringList | keys for groups in Go order: `"0"` (whole match), then `"1"`… or names; length is `.numGroups@ + 1` |

### Instance methods

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `match` | `( s -- b )` | `true` if this Regex matches somewhere in `s` |
| `find` | `( s -- found:s true \| false )` | first match substring (modal) |
| `findAll` | `( s -- parts:StringList true \| false )` | all non-overlapping matches (modal; `false` if none) |
| `findAllN` | `( s n:i -- parts:StringList true \| false )` | at most `n` matches (`n < 0` → all) |
| `findIndex` | `( s -- start:i end:i true \| false )` | **byte** indexes of the first match (modal) |
| `split` | `( s -- parts:StringList )` | split on matches (always a list; one element if no split) |
| `splitN` | `( s n:i -- parts:StringList )` | split into at most `n` parts |
| `matchObject` | `( s -- o:DynamicObject true \| false )` | whole match + captures as a DO (modal) |
| `replaceAll` | `( s repl:s -- out:s )` | replace every match; `repl` may use `$1`, `$name`, `$0`, … |
| `replaceAllLiteral` | `( s repl:s -- out:s )` | replace every match; `repl` is literal (no `$` expansion) |

```forth
`\w+` Regex.new^ >ts

"one two three" t:.findAll^ not if
  "no words" $ nl
else
  // StringList of matches on DS
  foreach
    i $ sp
  /foreach
  nl
then

"a1b2c3" t:.findIndex^ not if
  "no match" $ nl
else
  // start end on DS (byte indexes)
  $ sp $ nl
then

t:drop
```

### Captures with `matchObject`

On success, fields follow standard numbering: **`"0"`** is the whole match; **`"1"`**, **`"2"`**, … are capturing groups. Named groups use `(?P<name>…)` and appear under that name (still numbered as well). Creation order matches Go group indexes, so `k @()` is group `k`.

Access ([Collection Accessors](objects/collection-accessors.md), [DynamicObject](compounds/dynamicobject.md)):

* named group — `.name@` (Final ident) or `"name" @{}`
* any key including `"0"` — `@{}` (string key)
* by position — `@()` with an Int
* `.0@` is **invalid** (`"0"` is not a Final identifier)

```forth
`(?P<code>[A-Z]+)-(?P<num>\d+)` Regex.new^ >ts
"AB-42" t:.matchObject^ not if
  "no match" $ nl
else
  >xs
  x:.code@ $ sp          // AB
  x:.num@ $ nl           // 42
  "0" x:@{} $ nl         // AB-42  (whole match)
  1 x:@() $ nl           // AB
  x:drop
then
t:drop
```

### Replace

```forth
`(\d+)` Regex.new^ >ts
"a12b34" "#$1" t:.replaceAll^ $ nl           // a#12b#34
t:drop

`\$` Regex.new^ >ts
"a$b" "USD" t:.replaceAllLiteral^ $ nl       // aUSDb
t:drop
```

---

## Examples

### Validate once (one-shot)

```forth
: isZip? ( s -- b )
  `^\d{5}(-\d{4})?$` Regex.match^
;

"80302" isZip? $ nl          // true
"80302-1234" isZip? $ nl     // true
"8030" isZip? $ nl           // false
```

### Reuse a compiled pattern

```forth
`\d+` Regex.new^ >ts
"x=10 y=20" t:.findAll^ not if
  "none" $ nl
else
  foreach i $ sp /foreach nl
then
"only 7 here" t:.find^ not if "none" $ nl else $ nl then
t:drop
```

### Split on whitespace

Unlike `str::split` (exact separator), Regex split treats the pattern as the delimiter:

```forth
`\s+` Regex.new^ >ts
"alpha   beta\tgamma" t:.split^ >xs
x:.len@ $ nl                 // 3
0 x:@() $ nl                 // alpha
x:drop
t:drop
```

---

## Pattern cheat sheet (RE2)

Final patterns are Go/RE2. There are **no** `\1` backreferences and **no** `(?=…)` / `(?!…)` lookaround. For the full grammar, see Go’s [`regexp/syntax`](https://pkg.go.dev/regexp/syntax) docs.

### Characters and classes

| Pattern | Meaning |
| ------- | ------- |
| `.` | any character except newline (unless `(?s)`) |
| `[abc]` / `[^abc]` | character class / negated |
| `[a-z0-9]` | ranges inside a class |
| `\d` `\D` | digit / non-digit (`[0-9]` / `[^0-9]`) |
| `\w` `\W` | word char `[0-9A-Za-z_]` / non-word |
| `\s` `\S` | whitespace / non-whitespace |
| `\p{Greek}` / `\P{Greek}` | Unicode property / negated |
| `[[:alpha:]]` | ASCII POSIX class (inside `[]`) |

### Anchors

| Pattern | Meaning |
| ------- | ------- |
| `^` `$` | start / end of text (also line starts/ends with `(?m)`) |
| `\A` `\z` | absolute start / end of text |
| `\b` `\B` | ASCII word boundary / non-boundary |

### Repetition

| Pattern | Meaning |
| ------- | ------- |
| `x*` `x+` `x?` | 0+, 1+, 0 or 1 (greedy) |
| `x*?` `x+?` `x??` | same, prefer fewer (non-greedy) |
| `x{n}` `x{n,}` `x{n,m}` | exactly n; n or more; n…m |

### Groups and alternation

| Pattern | Meaning |
| ------- | ------- |
| `x&#124;y` | `x` or `y` (prefer left) |
| `(re)` | capturing group |
| `(?:re)` | non-capturing group |
| `(?P<name>re)` | named capturing group (use with `matchObject`) |

### Flags (inline)

| Pattern | Meaning |
| ------- | ------- |
| `(?i)` | case-insensitive |
| `(?m)` | multi-line: `^` / `$` also match line edges |
| `(?s)` | “dotall”: `.` matches newline |
| `(?U)` | ungreedy default for quantifiers |
| `(?flags:re)` | apply flags only to `re` |

### Literals and replace templates

| Need | How |
| ---- | --- |
| Metacharacter as literal | escape with `\`, or wrap text via `Regex.quoteMeta` |
| Whole-string match | anchor with `^…$` |
| `$` in `replaceAll` repl | `$1`, `$name`, `$0` expand; use `replaceAllLiteral` when `$` must stay literal |

Indexes from `findIndex` are **Utf-8 byte** offsets (same caution as `str::len` / `str::indexOf` — do not invent a byte-range substring API; use the match String from `find` / `matchObject` instead).
