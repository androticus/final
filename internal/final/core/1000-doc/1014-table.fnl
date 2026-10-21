// Table — named-column row collection

/*
A collection of rows organized by named columns.
Column names and order are fixed when the table is
created; rows may be appended afterward (including none).
Each column may hold any Final value, but cells in a
column should share one type — especially if you sort.

Create with a factory that matches your row shape:
  newFromDynamicObject — DO prototype (rows: DO clones
    or StaticObjects from that proto)
  newFromTypedObject — TypedObject-derived Type (needs @())
  newFromType — Type property schema (e.g. FileInfo)
  newFromColumnNames — StringList of names; rows are
    Lists (@()) or Maps (@{}) with columnar sameness

Do not mix incompatible row kinds in one table.
Iterable: foreach yields rows in current order.
Sort is stable (successive sorts compose). Use
Table.SortAsc / Table.SortDesc with sortByOrd or
sortByName. Sortable cell types: Bool, Int, Uint,
Float, String, DateTime — no mixed numeric kinds
in a sorted column. Append does not keep sort order.
eg: `` proto Table.newFromDynamicObject^ ``
*/
:doc Table

/*
Create a Table whose columns are the names in the list.
Rows may be Lists (column i via @()) or Maps (keys via
@{}). Sort picks access from the first row when both
are possible. Trust the caller for row shape.
eg: `` ("name" "age") Table.newFromColumnNames^ ``
( names:StringList -- table:Table )
*/
:doc Table#newFromColumnNames

/*
Create a Table from a DynamicObject prototype.
Column names come from the prototype's fields (creation
order). Append DO clones or StaticObjects from that
proto — same schema, ordinal access for sorting.
eg: `` proto Table.newFromDynamicObject^ ``
( proto:DynamicObject -- table:Table )
*/
:doc Table#newFromDynamicObject

/*
Create a Table from a TypedObject-derived Type.
Column names come from the type's fields. The type's
parent must be TypedObject.
Distinct from newFromType: clear TO shape assumptions
(ordinal fields; room for stricter append later).
eg: `` MyRow Table.newFromTypedObject^ ``
( t:Type -- table:Table )
*/
:doc Table#newFromTypedObject

/*
Create a Table from a Type's property schema.
Column names are the type's properties at call time.
Use for types with a fixed field set (e.g. FileInfo).
Does not require TypedObject parentage.
( t:Type -- table:Table )
*/
:doc Table#newFromType

/*
Ascending sort direction for sortByOrd / sortByName.
Prefer this over bare false for readable call sites.
eg: `` 0 Table.SortAsc t:.sortByOrd^ ``
( -- b )
*/
:doc Table#SortAsc

/*
Descending sort direction for sortByOrd / sortByName.
Prefer this over bare true for readable call sites.
eg: `` "age" Table.SortDesc t:.sortByName^ ``
( -- b )
*/
:doc Table#SortDesc

/*
The fixed column names, in schema order.
( -- names:StringList )
*/
:doc Table#columnNames

/*
The underlying List of row values, in current order.
( -- rows:List )
*/
:doc Table#rows

/*
Append a row compatible with this table's schema.
No shape check — wrong row kinds panic later on sort.
( item:a -- )
*/
:doc Table#append

/*
The row at the given ordinal index (0-based).
( row:i -- item:a )
*/
:doc Table#at

/*
Stable-sort rows by column ordinal.
dir is Table.SortAsc or Table.SortDesc.
Apply last key first when multi-column sorting.
Panics if the column is out of range, cell types
mismatch, or a cell type is not sortable.
eg: `` 0 Table.SortAsc t:.sortByOrd^ ``
( col:i dir:b -- )
*/
:doc Table#sortByOrd

/*
Stable-sort rows by column name.
dir is Table.SortAsc or Table.SortDesc.
Same rules as sortByOrd; panics if the name is unknown.
eg: `` "age" Table.SortDesc t:.sortByName^ ``
( name:s dir:b -- )
*/
:doc Table#sortByName
