# Properties

A **property** is a named member of a Type accessed with `.name` plus a fetch or store affix. Backing may be a stored slot or a computed getter/setter; the API is the same.

Property names and instance method names on one Type are disjoint. Use `help TypeName` to see which is which.

## Fetch and store

| Form | Effect | Notes |
| ---- | ------ | ----- |
| `cs:.name` or `cs:.name@` | `( -- value )` cs( obj -- obj ) | read; bare `.name` implies `@` |
| `cs:.name!` | `( value -- )` cs( obj -- obj ) | write |
| `TypeName.constProp` | `( -- value )` | class const; Type receiver is elided |

Value in / value out always on **DS**. The rooted receiver stays on the context stack.

```forth
DynamicObject.new^ >ts
"Ada" t:.name!
36 t:.age!
t:.name@ $ sp t:.age@ $ nl    // Ada 36
t:drop
```

**Store paths onto a parked receiver** need an explicit context selector (`t:` / `x:` / …). A bare store on DS without a prefix is rejected (`d:` does not help). Ident-rooted paths are allowed without a selector (`1 sess.config.paramA!`), as are `l.field!` and `pls.field!`.

Readonly / const properties (including many class constants) may omit `@`:

```forth
Math.pi $
Int.maxInt $
Float.NaN $
TokenType.TokenEOF $
```

## Readonly vs writable

* **Readonly** — no setter (`set` nil). Help shows `readonly`. Class constants (e.g. `Math.pi`, `TokenType.TokenEOF`) are readonly class properties.
* **Writable** — get and set. Help shows `get/set`. Fused `.name@` / `.name!` are one-shot (no ref left on the stack). Mid-path `.name` fetches the value and continues the walk with that value as the next focus.

## Property refs (`~`)

When you need the ref itself (for example read–modify–write with separate `@` / `!`), materialize with the **`~`** affix. `~` must be the **last** segment of that path token.

```forth
// Word on DS; materialize name ref, then Forth-style R-M-W
"dup" find
.name~
dup @          // fetch current name
drop
"fresh" swap ! // store; consumes ref
```

Standalone `@` and `!` consume a PropertyRef on the context stack. Fused `.name@` / `.name!` never leave a ref behind.

You cannot fuse `~` with `@` in one token (`name~@` is illegal). Use `.name~` then a separate `@` or `!`.

## Nested property walks

Intermediate `.prop` segments fetch the contained object and continue; intermediates are not parked on a stack:

```forth
dc.dict.len $     // DictionaryContext → dict → len
```

See [Compound Path Syntax](compound-path-syntax.md) for the full segment vocabulary.
