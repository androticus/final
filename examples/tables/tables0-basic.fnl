forget: reset-tables0
`
This example uses a DynamicObject prototype for a Table,
and StaticObject entries made from that DO.
It demonstrates stable sorting, meaning that
the order of other columns is preserved for ties.
Try:

0 tables0-demo
1 tables0-demo

` const: tables0-blurb

/*
Make the Prototype DynamicObject.
This provides the column names, and will be 
used to create rows.
( -- proto:DO )
*/
: makePrototype
  {
    "" "name"
    0  "age"
  }
;

makePrototype const: proto

/*
Make the table itself, then we will add data rows.
( -- tbl:Table )
*/
: makeTable
  proto Table.newFromDynamicObject^
;

makeTable const: table

/*
( name:s age:i -- )
*/
: addRow
  proto.newStatic^
  >ts
    t:.age!
    t:.name!
  ts> table.append^
;

: addRows
  "Gopher" 13 addRow
  "Alice"  20 addRow
  "Bob"    24 addRow
  "Alice"  55 addRow
  "Foobie" 16 addRow
  "Barbie" 16 addRow 
;

addRows

/*
Print the table before sorting,
then sort and print again.
( colIdx:i -- )
*/
: tables0-demo
  table >ts
  "Before sorting: \n\n" $
  t:.rows >jsonp $ nl
  
  Table.SortAsc t:.sortByOrd^
  "After sorting: \n\n" $
  t:.rows >jsonp $ t:drop
;

tables0-blurb $ nl
