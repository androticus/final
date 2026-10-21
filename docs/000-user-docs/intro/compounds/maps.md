# Maps

A Map is a table of key → value entries. The first write of a key creates the entry; later writes replace the value. Reads of a missing key are an error unless you test first with `?{}` or `hasKey`.

Final provides two map types that share the same programming model.

## Common API

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( -- m )` | empty map |
| prop | `len` | `( -- len:i )` | number of entries |
| method | `clear` | `( -- )` | remove all entries |
| method | `hasKey` | `( key:a -- b )` | true if the key exists |
| method | `deleteKey` | `( key:a -- )` | remove the entry if present |

### Keyed accessors

| Op | Effect | Description |
| -- | ------ | ----------- |
| `@{}` | `( key:a -- value )` | get; missing key is an error |
| `!{}` | `( value key:a -- )` | create or replace |
| `?{}` | `( key:a -- b )` | existence test (same idea as `hasKey`) |

### Iteration

Put the map on DS, then `foreach` … `/foreach`. Order is **not** defined. The loop value `i` is a **KeyValue** with `.key@` and `.value@`.

## Map

**Map** accepts keys that are comparable primitives (typically **String** or numbers). Values may be any Final type. Compound objects cannot be used as keys.

Prefer **Map** when you need **numeric** (or other non-string) keys.

```forth
Map.new^ >ts
"red" 1 t:!{}
"green" 2 t:!{}
1 t:@{} $ nl           // red
1 t:.hasKey^ $ nl      // true
1 t:.deleteKey^
t:.len@ $ nl           // 1
t:drop
```

## StringMap

**StringMap** is the usual choice for string-keyed dictionaries. Only a **String** is accepted as a key; values may still be any type. It inherits the common Map members (`len`, `clear`, and so on).

```forth
StringMap.new^ >ts
100 "count" t:!{}
"Ada" "name" t:!{}
"name" t:@{} $ nl
"missing" t:?{} $ nl   // false
t:peek foreach
  i >xs
  x:.key@ $ "=" $ x:.value@ $ nl
  x:drop
/foreach
t:drop
```

For small named records where field order matters, or where dotted `.field` access is nicer, see [DynamicObject](dynamicobject.md). For a reusable schema, see [TypedObject](typedobject.md).
