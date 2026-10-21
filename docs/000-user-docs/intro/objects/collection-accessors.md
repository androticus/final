# Collection Accessors

Some Types expose **item** access in addition to named properties and methods. Items are not properties: they use the fused collection affixes below. Index or key operands are on **DS**; with a context selector (`t:@()`, …) the collection **receiver stays** on that stack.

| Affix | Effect | Typical types |
| ----- | ------ | ------------- |
| `cs:@()` | `( index:i -- value )` | List family, DynamicObject / TypedObject by ordinal; `VecNf` / `Mat2f` flat |
| `cs:!()` | `( value index:i -- )` | same; writes an **existing** slot |
| `cs:@[]` | `( i0:i i1:i … -- value )` | `Mat2f`: `( r:i c:i -- f )` |
| `cs:![]` | `( value i0:i i1:i … -- )` | `Mat2f`: `( f r:i c:i -- )` |
| `cs:@{}` | `( key -- value )` | Map family, DynamicObject, TypedObject; missing key errors |
| `cs:!{}` | `( value key -- )` | create or replace |
| `cs:?{}` | `( key -- exists:b )` | Map family, DynamicObject, TypedObject |

`help TypeName` shows which of `@()` / `!()` / `@[]` / `![]` / `@{}` / `!{}` / `?{}` that type supports. Most types support either the `()` set or the `{}` set; DynamicObject and TypedObject support both. `Mat2f` supports flat `()` and coordinate `[]`.

## Ordinal `@()` / `!()`

Zero-based index. `!()` does not grow a list — use `.append^` / `.insertAt^` for that.

```forth
List.new^ >ts
10 t:.append^
20 t:.append^
30 t:.append^
1 t:@() $ nl          // 20
99 1 t:!()            // replace index 1
1 t:@() $ nl          // 99
t:drop
```

On DynamicObject / TypedObject, `@()` / `!()` use **field creation (or declaration) order**, not a hash order:

```forth
DynamicObject.new^ >ts
1 t:.x!
2 t:.y!
0 t:@() $ nl          // 1
t:drop
```

On **`Mat2f`**, `@()` / `!()` are **flat** row-major. Use `@[]` / `![]` for `(row, col)` — [Vectors and Matrices](../la/vec-mat.md).

## Mat `@[]` / `![]`

These accessors take a row/col pair, and return or set the value in that cell.

```forth
2 3 Mat2f.new^ >ts
// set row 0, col 2 to a value
1.0 0 2 t:![]
// retrive the value
0 2 t:@[] $ // 1
tdrop
```

When iterating in nested `foreach` loops, the following pattern is common:

```forth
( mat -- )
>ts
t:.getRowIter^
foreach
  t:.getColIter^
  foreach
    j i t:@[]
    // do something with the value
  /foreach
/foreach
tdrop
```


## Named `@{}` / `!{}` / `?{}`

```forth
StringMap.new^ >ts
100 "count" t:!{}
"Ada" "name" t:!{}
"name" t:@{} $ nl
"missing" t:?{} $ nl  // false
t:drop
```

`@{}` on a missing key is an error. Probe first with `?{}` (or a type's `hasKey` method when provided).

```forth
"city" t:?{} if
  "city" t:@{} $ nl
else
  "n/a" $ nl
then
```

## Prefixed vs bare

Day-to-day code should park the collection and use a selector:

```forth
0 t:@()
"k" t:@{}
```

A bare `@()` / `@{}` path whose receiver is on **DS** (no `t:` / …) is a one-shot form that **elides** the receiver after the access. Prefer the rooted `cs:` form so the handle remains for the next operation.

## Chaining after a get

Because `@()` / `@{}` leave a **value** on DS, further member access on that value usually means moving it to a stack first:

```forth
"tags" t:@{} >xs
0 x:@() $ nl
x:drop
```

Or nest through properties when the path stays inside one walk — see [Compound Path Syntax](compound-path-syntax.md).

Collection type APIs (append, clear, iteration, …) are documented under [Collection Types](../compounds/index.md).
