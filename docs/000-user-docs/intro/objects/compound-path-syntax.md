# Compound Path Syntax

A **compound path** is a single lexical token that chains object operations with **no whitespace**. It is Final's compact notation for property access, method invoke, and collection get/set. Paths compile and interpret as a left-to-right walk of fixed **segments**.

The Quick Tour covered the everyday forms (`t:.name@`, `List.new^`, `t:@()`). This page enumerates the **path elements**, stacking rules, and longer examples.

## Why paths

* **Clarity** — one token states the whole access intent
* **Brevity** — nested objects need not be parked mid-walk
* **Efficiency** — the walker materializes and elides intermediates for you

Rooted paths **preserve** the root receiver on the context stack, so you do not `dup` before each member op. That is the pattern recommended in [Writing Clean Final Code](../clean-final-coding.md).

## Path shape

```text
[ cs: ] [ RootIdent ] { Segment }…
```

Examples:

| Path | Meaning |
| ---- | ------- |
| `Math.pi` | class const on Type `Math` |
| `List.new^` | class method `new`, invoke |
| `t:.len@` | fetch `.len` on TS TOS |
| `t:.append^` | invoke `.append` on TS TOS |
| `1 t:!()` | store DS value at index into TS list |
| `dc.dict.len` | nested property walk from `dc` |
| `t:.name~` | materialize PropertyRef for `.name` |

## Elements

### 1. Context selector (`cs:`)

Optional prefix naming which stack holds the **receiver**:

| Prefix | Stack |
| ------ | ----- |
| *(none)* | DS by default for reads and ident-rooted stores (`sess.config.paramA!`); stores onto a receiver parked on DS need `t:` / `x:` / … |
| `t:` `x:` `y:` `z:` | auxilliary stacks |
| `o:` | Object Stack (Final methods only) |

```forth
List.new^ >ts
t:.len@ $ nl
```

See [Context Selectors](../stacks/cs.md).

### 2. Root

How the walk starts:

| Form | Role |
| ---- | ---- |
| Leading `.name…` | receiver already on the selected stack |
| CapCase / ident then `.…` | look up word (usually a Type), use its value as focus; ephemeral Type roots are elided when the path finishes |
| Leading `@()` / `!()` / `@{}` / `!{}` / `?{}` | collection op on the receiver |

```forth
TokenType.TokenEOF $     // Type root + class const → value on DS
"dup" find .name@ $      // instance already on DS + property fetch
```

### 3. Member segment: `.name`

Dot plus a member name. What happens depends on the **following affix** (or whether more path follows).

| Follow-on | Property | Method |
| --------- | -------- | ------ |
| end of path or mid-path continue | fetch value (implied `@`) | **error** — need `^` |
| `@` | fetch (fused) | n/a |
| `!` | store (fused) | n/a |
| `^` | n/a | invoke (fused) |
| `~` | leave PropertyRef | leave MethodRef |

Mid-path `.inner` always means “fetch this property and continue with its value”:

```forth
dc.dict.len $
```

### 4. Affixes

| Segment | Role |
| ------- | ---- |
| `@` | fetch via PropertyRef (or fused after `.name`) |
| `!` | store via PropertyRef (or fused after `.name`) |
| `^` | invoke MethodRef (must be **terminal**) |
| `~` | materialize ref after `.name` (must be **terminal** in that token) |

Fused forms you will see constantly:

```forth
t:.count@          // .  count  @
t:.count!          // .  count  !
t:.reset^          // .  reset  ^
t:.count~          // .  count  ~
```

Illegal: `t:.count~@` (`~` cannot be followed inside the same token). Use `t:.count~` then `@`.

### 5. Collection affixes

Each of these is **one** segment cell:

| Segment | DS operands | Result |
| ------- | ----------- | ------ |
| `@()` | `index:i` | value |
| `!()` | `value` then `index:i` | *(store)* |
| `@[]` | indexes (Mat2N: `r` then `c`) | value |
| `![]` | `value` then indexes | *(store)* |
| `@{}` | `key` | value |
| `!{}` | `value` then `key` | *(store)* |
| `?{}` | `key` | `exists:b` |

```forth
0 t:@()
99 0 t:!()
"name" t:@{}
"Ada" "name" t:!{}
"name" t:?{}
```

Details and type coverage: [Collection Accessors](collection-accessors.md).

## Receiver and DS roles

| Location | Holds |
| -------- | ----- |
| Context stack (`t:` …) | receiver for the path |
| DS | indexes, keys, values to store, and fetched / method results |

Common factory then park:

```forth
DynamicObject.new^ >ts     // inst on DS → TS
"Ada" t:.name!
t:.name@ $ nl
t:drop
```

**Stores** (`.name!`, `!()`, `!{}`) onto a receiver parked on a stack need `t:` / `x:` / `y:` / `z:` / `o:`. Ident-rooted paths do not (`1 sess.config.paramA!`). `l.` / `pls.` stores are the same. A bare `.name!` on DS is rejected (`d:` does not change that).

## Class paths vs instance paths

```forth
// Class — Type name in the path
List.new^
Math.pi
Int.maxInt

// Instance — receiver on a stack
List.new^ >ts
10 t:.append^
t:.len@ $
t:drop
```

## Nested walks and elision

Interior objects fetched mid-path are **not** left on stacks. Only the root receiver (when rooted) and the terminal product (on DS) remain.

```forth
// TypedObject instance on TS with a List field .tags
Person.new^ >ts
List.new^ >xs
"lang" x:.append^
"math" x:.append^
xs> t:.tags!

t:.tags@ >xs
0 x:@() $ nl       // lang
x:drop
t:drop
```

When a property yields another object you will keep using, park it (`>xs`) rather than relying on a one-shot DS value.

## Refs and read–modify–write

```forth
"dup" find >ts
t:.name~
dup @
drop
"DUP" swap !
t:.name@ $ nl      // DUP
t:drop
```

Standalone `@` / `!` / `^` consume the ref. Fused path terminals elide the ref automatically.

## Worked examples

### Build and inspect a list

```forth
List.new^ >ts
10 t:.append^
20 t:.append^
30 t:.append^
t:.len@ $ nl       // 3
1 t:@() $ nl       // 20
40 1 t:!()
1 t:@() $ nl       // 40
t:drop
```

### Map upsert and probe

```forth
StringMap.new^ >ts
1 "a" t:!{}
2 "b" t:!{}
"a" t:@{} $ nl
"c" t:?{} if
  "c" t:@{} $
else
  "missing" $
then
nl
t:drop
```

### DynamicObject fields and ordinals

```forth
DynamicObject.new^ >ts
1 t:.x!
2 t:.y!
3 t:.z!
t:.y@ $ nl         // 2
1 t:@() $ nl       // 2  (creation order)
"z" t:?{} $ nl     // true
t:drop
```

### Class const plus rooted instance work

Type constants and factories use a CapCase root; instance work parks the handle:

```forth
Math.pi 2 * $ nl

StringBuilder.new^ >ts
"maxInt=" t:.writeString^
Int.maxInt "%d" fmt t:.writeString^
t:.toString^ $ nl
t:drop
```
### Final method with `o:`

```forth
: Box#label
  o:.title@
;

Box.new^ >ts
"inbox" t:.title!
t:.label^ $ nl
t:drop
```

Inside `Box#label`, `o:.title@` is an ordinary rooted path on OS.

## Segment checklist

Allowed punctuation / affix cells inside a compound token:

`.`  `@`  `!`  `^`  `~`  `@()`  `!()`  `@{}`  `!{}`  `?{}`

Rules of thumb:

* no whitespace inside the token
* `^` and `~` only as the last segment of that token
* member names: letter or `_` start; then letters, digits, `_`, `-`, `$`, `?`
* prefer `t:` / `x:` / … for any sequence of ops on one object

Related: [Properties](properties.md), [Methods](methods.md), [Collection Accessors](collection-accessors.md), [Stacks](../stacks/index.md).
