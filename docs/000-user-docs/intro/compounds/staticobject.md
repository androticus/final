# StaticObject

**StaticObject** (SO) is a fixed-schema record whose **field names live on a prototype DynamicObject**; each instance holds only a dense value array. Access looks like DynamicObject (named fields, keyed and ordinal items, `foreach`), but the field set cannot grow after construction.

Typical uses: many rows that share one schema — Table rows, SQL-style result shapes, directory-listing style records — when you want DO-like accessors with lower per-instance overhead than cloning full DynamicObjects.

## Creation

Build (or obtain) a prototype DO, then call **`.newStatic^`** on that instance:

```forth
{ "" "name" 0 "age" } >xs          // prototype: field names + defaults

x:.newStatic^ >ts
"Ada" t:.name!
36 t:.age!
t:.name@ $ sp t:.age@ $ nl         // Ada 36
t:drop
x:drop
```

`.newStatic^` is an **instance method on DynamicObject**:

```forth
proto.newStatic^   // ( -- so:StaticObject )   cs( proto -- proto )
```

It shallow-copies the prototype’s current field values into the new StaticObject. Later stores on the instance do not change the prototype’s values; later stores on the prototype do not change existing instances.

There is no `StaticObject.new^` factory — every SO is minted from a DO prototype.

## Access

Same API as DynamicObject:

| Form | Effect | Description |
| ---- | ------ | ----------- |
| `.name!` | `( value -- )` | set an **existing** schema field |
| `.name@` | `( -- value )` | read field; error if not on the prototype |
| `!{}` | `( value key:s -- )` | set by string key (schema field only) |
| `@{}` | `( key:s -- value )` | read by key |
| `?{}` | `( key:s -- b )` | true if the name is on the prototype |
| `@()` | `( index:i -- value )` | read by **prototype field ordinal** |
| `!()` | `( value index:i -- )` | write by ordinal |

**Difference from DynamicObject:** setting an unknown name panics — the schema is frozen at `newStatic`. Use `?{}` to test membership against the prototype’s names.

## Example (Table-style rows)

```forth
: makePrototype
  { "" "name" 0 "age" }
;

makePrototype const: proto

: addRow   // ( name:s age:i tbl:Table -- )
  >ys
  proto.newStatic^ >ts
  t:.age! t:.name!
  ts> y:.append^
  y:drop
;

proto Table.newFromDynamicObject^ >zs
"Ada" 36 z:peek addRow
"Bob" 24 z:peek addRow
z:drop
```

## Iteration

`foreach` walks fields in **prototype registration order**. The loop value `i` is a **KeyValue** (`.key@` / `.value@`), same as DynamicObject.

## Related

- Prototype / ad-hoc DynamicObjects: [DynamicObject](dynamicobject.md) (including `.clone^`)
- Named types with typed fields and Final methods: [TypedObject](typedobject.md)
- Columnar grids: [Table](table.md)
