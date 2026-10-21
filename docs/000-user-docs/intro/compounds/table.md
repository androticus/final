# Table

**Table** is a named-column row collection. Column names are fixed at construction; you append rows afterward. Sort is **stable**, so successive sorts compose (later keys break ties without scrambling earlier order).

Rows can be any objects that expose the schema’s columns — commonly **StaticObject** instances minted from a DynamicObject prototype, DynamicObject clones, TypedObject instances, or other typed values with matching properties / `@()` / `@{}`.

---

## Construction

| Class method | Effect | Schema source |
| ------------ | ------ | ------------- |
| `newFromDynamicObject` | `( proto:DynamicObject -- table )` | property names of the DO (ordinal access for DO/StaticObject rows) |
| `newFromTypedObject` | `( t:Type -- table )` | TypedObject-derived type with `@()` |
| `newFromType` | `( t:Type -- table )` | any Type’s property names (`@()` and/or `@{}`) |
| `newFromColumnNames` | `( names:StringList -- table )` | explicit column name list |

## Members

| Kind | Name | Effect / Type | Description |
| ---- | ---- | ------------- | ----------- |
| prop | `columnNames` | StringList | fixed schema names |
| prop | `rows` | List | row list |
| method | `append` | `( row -- )` | append a compatible row |
| method | `at` | `( idx:i -- row )` | row by ordinal |
| method | `sortByOrd` | `( idx:i sortDirection:b -- )` | stable sort by column ordinal |
| method | `sortByName` | `( name:s sortDirection:b -- )` | stable sort by column name |
| class prop | `SortAsc` | Bool (`false`) | ascending direction for sort methods |
| class prop | `SortDesc` | Bool (`true`) | descending |

Sortable cell types: **Bool**, **Int**, **Uint**, **Float**, **String**, **DateTime**. Compared cells must share a type.

Table is **Iterable** — `foreach` walks rows.

---

## Example (DO proto + StaticObject rows)

```forth
: makePrototype  ( -- proto:DynamicObject )
  { "" "name" 0 "age" }
;

makePrototype const: proto
proto Table.newFromDynamicObject^ const: table

: addRow  ( name:s age:i -- )
  proto.newStatic^ >ts
  t:.age! t:.name!
  ts> table.append^
;

"Gopher" 13 addRow
"Alice"  20 addRow
"Bob"    24 addRow
"Alice"  55 addRow

table >ts
0 Table.SortAsc t:.sortByOrd^     // sort by name (column 0)
t:.rows >jsonp $
t:drop
```

`proto.newStatic^` builds a **StaticObject**: field names live on the prototype DO; the instance holds a values array. Same `.name!` / `.age@` API as a DynamicObject for those fields.

Successive sorts: sort by age, then by name (or the reverse) — stability preserves relative order for ties.

```forth
table >ts
1 Table.SortAsc t:.sortByOrd^           // by age
"name" Table.SortAsc t:.sortByName^     // then by name
t:drop
```

---

## Notes

* Append does not deeply validate row shape — supply rows that match the schema accessors the Table was built for.
* Prefer a DO prototype + `newStatic` when you want light row instances with a shared column layout.
* For reusable typed rows with Final methods, mint a TypedObject type and use `Table.newFromTypedObject^`.
* `dbIter>table` (from a `DbProvider.query^` iterator) builds the Table from the query’s DynamicObject prototype and appends StaticObject rows. See [SQL Databases](../database.md).

Use `help Table` in the TUI for the full member list.
