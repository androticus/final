# DynamicObject

**DynamicObject** (DO) is Final's lightweight named-field object. You create fields as you go; there is no separate schema. It is a frequent choice for Json objects, ad-hoc records, and small configuration-style payloads.

Compared with **StringMap**, a DO is typically leaner and faster when the number of fields stays modest (on the order of tens). It also **preserves field creation order**, and supports both name and ordinal access.

## Creation

Empty instance:

```forth
DynamicObject.new^   // ( -- o:DynamicObject )
```

`newdo` is a short alias for the same factory.

### `{` … `}` literal

Prefer the `{` … `}` builders when you already know the fields. `{` marks stack depth (like `(`). `}` gathers **value / name** pairs down to that mark into a DynamicObject. Use `sm}` instead of `}` to build a **StringMap** the same way.

```forth
{ 1 "x" 2 "y" 3 "z" } >ts
t:.x@ $ sp t:.y@ $ sp t:.z@ $ nl    // 1 2 3
t:drop

{ "Ada" "name" 36 "age" } >jsonp $ nl
{ "London" "city" "UK" "country" } sm} >ts
"city" t:@{} $ nl
t:drop
```

Stack order is **value then name** (same as `!{}`).

## Access

| Form | Effect | Description |
| ---- | ------ | ----------- |
| `.name!` | `( value -- )` | create or set field `name` |
| `.name@` | `( -- value )` | read field; error if missing |
| `!{}` | `( value key:s -- )` | create or set by string key |
| `@{}` | `( key:s -- value )` | read by key; error if missing |
| `?{}` | `( key:s -- b )` | true if the field exists |
| `@()` | `( index:i -- value )` | read by **creation ordinal** |
| `!()` | `( value index:i -- )` | write by creation ordinal |

A field must be created (with `.name!` or `!{}`) before it can be read with `.name@` or `@{}`.

DynamicObject does not expose Map's `len` / `clear` / `hasKey` methods; use `?{}` for existence, and build a fresh DO when you want a clean slate.

## Example

```forth
{ 1 "x" 2 "y" 3 "z" } >ts
t:.x@ $ sp t:.y@ $ sp t:.z@ $ nl    // 1 2 3
0 t:@() $ nl                         // 1  (first field created)
"x" t:?{} $ nl                       // true
t:drop
```

When you need to grow fields after creation, start empty and set as you go:

```forth
newdo >ts
1 t:.x!
2 t:.y!
t:drop
```

## Cloning (prototypes)

`.clone^` returns a **shallow** copy with the same fields and values. Use a prototype DO with defaults, then clone for each working instance so mutations stay isolated.

```forth
{ "" "string" 0 "int" } >xs

x:.clone^ >ys
"y mutate" y:.string!
1 y:.int!

x:.clone^ >zs
"z mutate" z:.string!
2 z:.int!

xs> >jsonp $ nl    // prototype unchanged
ys> >jsonp $ nl
zs> >jsonp $ nl
```

For a fixed field set with lower per-instance overhead, mint a [StaticObject](staticobject.md) from the same prototype via `.newStatic^`.

## Iteration

`foreach` walks fields in **creation order**. The loop value `i` is a **KeyValue**: `.key@` is the field name (String), `.value@` is the stored value.

```forth
{ "Ada" "name" 36 "age" } >ts
t:peek foreach
  i >xs
  x:.key@ $ ": " $ x:.value@ $ nl
  x:drop
/foreach
t:drop
```

A DynamicObject is a one-off instance: its field set is not a reusable type. When you need a named type you can mint repeatedly (optionally with Final methods), use [TypedObject](typedobject.md).

## Optimizations for Large Field Sets

Since DynamicObject is the object returned when parsing JSON, it is optimized "under the hood" to support almost unlimited field sizes. This is not a typical use case in normal use of DO as an Object-type, but helps it become an efficient hybrid of an Object and a Map for large machine-generated field sets, such as cache key results that can run into the thousands. If DO has more than 64 fields added, it switches under the hood to use a map-like ordinal lookup, rather than an interated lookup. This operation does not affect how DO behaves or is used. For an object that follows purely Map-based semantics, the Final Map or StringMap types are still recommended.
