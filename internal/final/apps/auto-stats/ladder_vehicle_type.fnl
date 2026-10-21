/*
auto-stats Stage-2a — age + miles + vehicle class (car / SUV / truck).

The feed's vehicle_type is only Car|Truck (SUVs lumped under Truck).
We derive body class from body_type for the three-way story:
  Pickup → truck; SUV|Wagon → suv; Sedan|Coupe|Hatchback → car.
Car is the reference dummy (isSUV, isTruck with intercept).

Compact C = age+miles; Augmented A = age+miles+class.

Run from repo root:
  bin/final-go -core
  "internal/final/apps/auto-stats/ladder_vehicle_type.fnl" compileFromFile
  ladderVt
*/

"internal/final/apps/auto-stats/db/auto-stats.db" const: csDbPath
2026 const: csAsOfYear
500 const: csMinMiles

`
SELECT year, miles, price, body_type, vehicle_type, make, model
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
0 var: fitAge
0 var: fitIsSuv
0 var: fitIsTruck
0 var: yPrice
0 var: xWear
0 var: xClass
0 var: mWear
0 var: mClass
0 var: nCar
0 var: nSuv
0 var: nTruck
0 var: sumCar
0 var: sumSuv
0 var: sumTruck

: connectCsDb
  DbProvider.sqlite csDbPath DbProvider.new^
  nif
    "Can't open auto-stats DB: " swap str::+ error
  then
  csDb !
;

/*
Map listing body_type → class string, or null if unmapped.
( row:SO -- class:s|null )
*/
: rowBodyClass
  >ts
  t:.body_type@ Null? if
    null
  else
    t:.body_type@ >xs
      x:peek "Pickup" == if
        "truck"
      else
        x:peek "SUV" == if
          "suv"
        else
          x:peek "Wagon" == if
            "suv"
          else
            x:peek "Sedan" == if
              "car"
            else
              x:peek "Coupe" == if
                "car"
              else
                x:peek "Hatchback" == if
                  "car"
                else
                  null
                then
              then
            then
          then
        then
      then
      x:drop
  then
  t:drop
;

/*
Keep only rows with a mapped body class; tally class counts / price sums.
( -- )
*/
: loadFitRows
  sqlFitRows csDb @ >ts t:.query^ t:drop
  iter>list >ts
  List.new^ fitRows !
  0 nCar !
  0 nSuv !
  0 nTruck !
  0.0 sumCar !
  0.0 sumSuv !
  0.0 sumTruck !
  t:peek
  foreach
    i >ys
      y:peek rowBodyClass >xs
      x:peek Null? if
        x:drop
      else
        fitRows @ >zs
          y:peek z:.append^
        z:drop
        x:peek "car" == if
          nCar @ 1 + nCar !
          sumCar @ y:.price@ >Float + sumCar !
        else
          x:peek "suv" == if
            nSuv @ 1 + nSuv !
            sumSuv @ y:.price@ >Float + sumSuv !
          else
            nTruck @ 1 + nTruck !
            sumTruck @ y:.price@ >Float + sumTruck !
          then
        then
        x:drop
      then
      y:drop
  /foreach
  t:drop
  fitRows @ .len@ nFit !
;

/*
Build y, wear X (intercept+age+miles), class X (+ isSUV + isTruck).
( -- )
*/
: buildDesign
  nFit @ VecNf.new^ yPrice !
  nFit @ 3 Mat2f.new^ xWear !
  nFit @ 5 Mat2f.new^ xClass !
  0 nFit @ 1 - range
  foreach
    i >xs
      fitRows @ >ts
        x:peek t:@()
      t:drop
      >ys
        y:peek rowBodyClass >zs
        y:.year@ >Int
        csAsOfYear swap - >Float fitAge !
        y:.miles@ >Float fitMiles !
        z:peek "suv" == if 1.0 else 0.0 then fitIsSuv !
        z:peek "truck" == if 1.0 else 0.0 then fitIsTruck !
        yPrice @ >ts
          y:.price@ >Float x:peek t:!()
        t:drop
        xWear @ >ts
          1.0 x:peek 0 t:![]
          fitAge @ x:peek 1 t:![]
          fitMiles @ x:peek 2 t:![]
        t:drop
        xClass @ >ts
          1.0 x:peek 0 t:![]
          fitAge @ x:peek 1 t:![]
          fitMiles @ x:peek 2 t:![]
          fitIsSuv @ x:peek 3 t:![]
          fitIsTruck @ x:peek 4 t:![]
        t:drop
        z:drop
      y:drop
    x:drop
  /foreach
;

: fitAll
  yPrice @ xWear @ OLSModel.fit^ mWear !
  yPrice @ xClass @ OLSModel.fit^ mClass !
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
  "  %-18s  p=%d  SSE=%.4e  RMSE=$%.0f  R²=%.4f\n" fmt $
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
  "  %-36s  PRE=%.4f  F=%.2f  df1=%d df2=%d  SSE_C=%.4e  SSE_A=%.4e\n" fmt $
  t:drop x:drop
;

: printHeader
  "============================================================" $ nl
  " auto-stats Stage-2a — age/miles + vehicle class" $ nl
  "============================================================" $ nl
  (
    nFit @
    csAsOfYear
    csMinMiles
  )
  " n=%d   as-of year=%d   miles >= %d (used inventory)\n" fmt $
  " y = asking price (USD)\n" $
  " Class from body_type (feed vehicle_type collapses SUV→Truck).\n" $
  " Coding: Car = reference; β cols = [int, age, miles, isSUV, isTruck]\n" $
  nl
;

: printClassMeans
  "Class sample means (raw asking price)" $ nl
  nCar @ 0 > if
    (
      nCar @
      sumCar @ nCar @ >Float / 
    )
    "  car    n=%d  mean=$%.0f\n" fmt $
  then
  nSuv @ 0 > if
    (
      nSuv @
      sumSuv @ nSuv @ >Float /
    )
    "  suv    n=%d  mean=$%.0f\n" fmt $
  then
  nTruck @ 0 > if
    (
      nTruck @
      sumTruck @ nTruck @ >Float /
    )
    "  truck  n=%d  mean=$%.0f\n" fmt $
  then
  nl
;

: printModels
  "Fitted models" $ nl
  mWear @ "age+miles" printModel
  mClass @ "age+miles+class" printModel
  "  β legend (class model): intercept, ageYears, miles, isSUV, isTruck\n" $
  nl
;

: printCompares
  "C-vs-A comparisons" $ nl
  mWear @ mClass @ "age+miles → +vehicle class" printCompare
  nl
  "Read: class asks whether car/SUV/truck level shifts remain after age and miles.\n" $
;

: ladderVt
  connectCsDb
  loadFitRows
  nFit @ 0 <= if
    "No fit-ready class-mapped rows in " csDbPath str::+ error
  then
  buildDesign
  fitAll
  printHeader
  printClassMeans
  printModels
  printCompares
  "Done." $ nl
;
