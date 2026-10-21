/*
auto-stats Stage-1 ladder — age alone, miles alone, then both.

Analyst order (Judd C-vs-A):
  C₀ mean-only
  age vs mean · miles vs mean
  both vs each single (incremental after the first predictor)

Run from repo root:
  bin/final-go -core
  "internal/final/apps/auto-stats/ladder_age_miles.fnl" compileFromFile
  ladder

Uses frozen SQLite from ingest (fit-ready used rows, miles >= 500).
*/

"internal/final/apps/auto-stats/db/auto-stats.db" const: csDbPath
2026 const: csAsOfYear
500 const: csMinMiles

`
SELECT year, miles, price
FROM unhuman_listing
WHERE price IS NOT NULL
  AND miles IS NOT NULL
  AND year IS NOT NULL
  AND miles >= 500
  AND (inventory_type IS NULL OR inventory_type = 'used')
`
const: sqlFitRows

0 var: csDb
0 var: fitRows
0 var: nFit
0 var: fitMiles
0 var: yPrice
0 var: xMean
0 var: xAge
0 var: xMiles
0 var: xBoth
0 var: mMean
0 var: mAge
0 var: mMiles
0 var: mBoth

: connectCsDb
  DbProvider.sqlite csDbPath DbProvider.new^
  nif
    "Can't open auto-stats DB: " swap str::+ error
  then
  csDb !
;

/*
Load fit-ready listing rows into fitRows / nFit.
( -- )
*/
: loadFitRows
  sqlFitRows csDb @ >ts t:.query^ t:drop
  iter>list
  dup fitRows !
  .len@ nFit !
;

/*
Build y and design mats from fitRows (as-of − year → ageYears).
( -- )
*/
: buildDesign
  nFit @ VecNf.new^ yPrice !
  nFit @ 1 Mat2f.new^ xMean !
  nFit @ 2 Mat2f.new^ xAge !
  nFit @ 2 Mat2f.new^ xMiles !
  nFit @ 3 Mat2f.new^ xBoth !
  0 nFit @ 1 - range
  foreach
    // i = row index
    i >xs
      fitRows @ >ts
        x:peek t:@()
      t:drop
      >ys
        y:.year@ >Int
        csAsOfYear swap - >Float >zs
        y:.miles@ >Float fitMiles !
        yPrice @ >ts
          y:.price@ >Float x:peek t:!()
        t:drop
        xMean @ >ts
          1.0 x:peek t:!()
        t:drop
        xAge @ >ts
          1.0 x:peek 0 t:![]
          z:peek x:peek 1 t:![]
        t:drop
        xMiles @ >ts
          1.0 x:peek 0 t:![]
          fitMiles @ x:peek 1 t:![]
        t:drop
        xBoth @ >ts
          1.0 x:peek 0 t:![]
          z:peek x:peek 1 t:![]
          fitMiles @ x:peek 2 t:![]
        t:drop
        z:drop
      y:drop
    x:drop
  /foreach
;



: fitAll
  yPrice @ xMean @ OLSModel.fit^ mMean !
  yPrice @ xAge @ OLSModel.fit^ mAge !
  yPrice @ xMiles @ OLSModel.fit^ mMiles !
  yPrice @ xBoth @ OLSModel.fit^ mBoth !
;

/*
( model:OLSModel label:s -- )
*/
: printModel
  >xs >ts
  t:.beta@ >ys
  (
    x:peek
    t:.p@
    t:.sse@
    t:.rmse@
    t:.r2@
  )
  "  %-12s  p=%d  SSE=%.4e  RMSE=$%.0f  R²=%.4f\n" fmt $
  "    β =" $
  0
  begin
    dup y:.len@ < while
      dup y:@() >zs
        ( z:peek ) "  %.6g" fmt $
      z:drop
      1 +
  repeat
  drop
  nl
  y:drop t:drop x:drop
;

/*
( c:OLSModel a:OLSModel label:s -- )
*/
: printCompare
  >xs
  OLSModel.compare^ >ts
  (
    x:peek
    t:.pre@
    t:.f@
    t:.df1@
    t:.df2@
    t:.sseC@
    t:.sseA@
  )
  "  %-28s  PRE=%.4f  F=%.2f  df1=%d df2=%d  SSE_C=%.4e  SSE_A=%.4e\n" fmt $
  t:drop x:drop
;

: printHeader
  "============================================================" $ nl
  " auto-stats Stage-1 ladder — age / miles (OLS + PRE/F)" $ nl
  "============================================================" $ nl
  (
    nFit @
    csAsOfYear
    csMinMiles
  )
  " n=%d   as-of year=%d   miles >= %d (used inventory)\n" fmt $
  " y = asking price (USD)\n" $
  nl
;

: printModels
  "Fitted models" $ nl
  mMean @ "mean-only" printModel
  mAge @ "age" printModel
  mMiles @ "miles" printModel
  mBoth @ "age+miles" printModel
  nl
;

: printCompares
  "C-vs-A comparisons" $ nl
  mMean @ mAge @ "mean → age" printCompare
  mMean @ mMiles @ "mean → miles" printCompare
  mMean @ mBoth @ "mean → age+miles" printCompare
  mAge @ mBoth @ "age → age+miles" printCompare
  mMiles @ mBoth @ "miles → age+miles" printCompare
  nl
;

/*
Report which single predictor wins on SSE, then both vs that single.
( -- )
*/
: printBetterSingle
  mAge @ >ts t:.sse@ t:drop
  mMiles @ >ts t:.sse@ t:drop
  < if
    "Better single (lower SSE): age" $ nl
    mAge @ mBoth @ "age → age+miles (incr.)" printCompare
  else
    "Better single (lower SSE): miles" $ nl
    mMiles @ mBoth @ "miles → age+miles (incr.)" printCompare
  then
  nl
  "Read: alone first for interpretation; both for incremental PRE after the better single." $ nl
;

/*
Main entry.
( -- )
*/
: ladder
  connectCsDb
  loadFitRows
  nFit @ 0 <= if
    "No fit-ready rows in " csDbPath str::+ error
  then
  buildDesign
  fitAll
  printHeader
  printModels
  printCompares
  printBetterSingle
  "Done." $ nl
;
