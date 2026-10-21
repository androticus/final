forget: reset-csv-demo
`
Build a tiny Table of critters, then render it as CSV with the
csv module (table>csv-string).

Load this project (final-tui Load, or):

  "projects/csv-demo/csv-demo.fnlp" loadProject

Try: csv-demo

` const: csv-demo-blurb

/*
Prototype DO → column headers Creature, Legs, Eyes.
( -- proto:DynamicObject )
*/
: makeCsvDemoProto
  newdo >ts
    "" t:.Creature!
    0  t:.Legs!
    0  t:.Eyes!
  ts>
;

/*
( -- tbl:Table )
*/
: makeCsvDemoTable
  makeCsvDemoProto >ts
  t:peek Table.newFromDynamicObject^ >ys
  t:peek .newStatic^ >xs
    "ant" x:.Creature!
    6 x:.Legs!
    2 x:.Eyes!
  xs> y:.append^
  t:peek .newStatic^ >xs
    "spider" x:.Creature!
    8 x:.Legs!
    8 x:.Eyes!
  xs> y:.append^
  t:peek .newStatic^ >xs
    "octopus" x:.Creature!
    8 x:.Legs!
    2 x:.Eyes!
  xs> y:.append^
  t:drop
  ys>
;

/*
Render the critter table as CSV and print it.
( -- )
*/
: csv-demo
  makeCsvDemoTable >ts
  getDefaultCsvOptions >ys
  y:peek t:peek table>csv-string
  "CSV:\n" $
  $ nl
  y:drop
  t:drop
;

csv-demo-blurb $
