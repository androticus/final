/*
auto-stats Stage-2b — age + miles + make (no vehicle class).

Drop class (it correlates with model mix). Keep wear core; add make
one-hots. Honda is the reference make (mass-market mid); other makes
get a dummy column.

Compact C = age+miles; Augmented A = age+miles+make.

β columns (class model):
  0 intercept  1 ageYears  2 miles
  3 Audi  4 BMW  5 Chevrolet  6 Dodge  7 Ford
  8 Lexus  9 Lincoln  10 Nissan  11 RAM  12 Subaru  13 Toyota

Run from repo root:
  bin/final-go -core
  "internal/final/apps/auto-stats/ladder_make.fnl" compileFromFile
  ladderMake
*/

"internal/final/apps/auto-stats/db/auto-stats.db" const: csDbPath
2026 const: csAsOfYear
500 const: csMinMiles
"Honda" const: csRefMake
14 const: csMakeP // intercept + age + miles + 11 make dummies

`
SELECT year, miles, price, make, model, body_type
FROM unhuman_listing
WHERE price IS NOT NULL
  AND miles IS NOT NULL
  AND year IS NOT NULL
  AND make IS NOT NULL
  AND miles >= 500
  AND (inventory_type IS NULL OR inventory_type = 'used')
`
const: sqlFitRows

0 var: csDb
0 var: fitRows
0 var: nFit
0 var: fitMiles
0 var: fitAge
0 var: fitMakeCol
0 var: yPrice
0 var: xWear
0 var: xMake
0 var: mWear
0 var: mMake

: connectCsDb
  DbProvider.sqlite csDbPath DbProvider.new^
  nif
    "Can't open auto-stats DB: " swap str::+ error
  then
  csDb !
;

/*
Absolute Mat column for a make dummy, or -1 if reference (Honda),
or -2 if unknown / empty.
( make:s -- col:i )
*/
: makeDummyCol
  >xs
  x:peek "" == if
    -2
  else
    x:peek csRefMake == if
      -1
    else
      x:peek "Audi" == if
        3
      else
        x:peek "BMW" == if
          4
        else
          x:peek "Chevrolet" == if
            5
          else
            x:peek "Dodge" == if
              6
            else
              x:peek "Ford" == if
                7
              else
                x:peek "Lexus" == if
                  8
                else
                  x:peek "Lincoln" == if
                    9
                  else
                    x:peek "Nissan" == if
                      10
                    else
                      x:peek "RAM" == if
                        11
                      else
                        x:peek "Subaru" == if
                          12
                        else
                          x:peek "Toyota" == if
                            13
                          else
                            -2
                          then
                        then
                      then
                    then
                  then
                then
              then
            then
          then
        then
      then
    then
  then
  x:drop
;

: loadFitRows
  sqlFitRows csDb @ >ts t:.query^ t:drop
  iter>list >ts
  List.new^ fitRows !
  t:peek
  foreach
    i >ys
      y:.make@ makeDummyCol >xs
      x:peek -2 == if
        x:drop
      else
        fitRows @ >zs
          y:peek z:.append^
        z:drop
        x:drop
      then
      y:drop
  /foreach
  t:drop
  fitRows @ .len@ nFit !
;

: buildDesign
  nFit @ VecNf.new^ yPrice !
  nFit @ 3 Mat2f.new^ xWear !
  nFit @ csMakeP Mat2f.new^ xMake !
  0 nFit @ 1 - range
  foreach
    i >xs
      fitRows @ >ts
        x:peek t:@()
      t:drop
      >ys
        y:.year@ >Int
        csAsOfYear swap - >Float fitAge !
        y:.miles@ >Float fitMiles !
        y:.make@ makeDummyCol fitMakeCol !
        yPrice @ >ts
          y:.price@ >Float x:peek t:!()
        t:drop
        xWear @ >ts
          1.0 x:peek 0 t:![]
          fitAge @ x:peek 1 t:![]
          fitMiles @ x:peek 2 t:![]
        t:drop
        xMake @ >ts
          1.0 x:peek 0 t:![]
          fitAge @ x:peek 1 t:![]
          fitMiles @ x:peek 2 t:![]
          fitMakeCol @ 0 < if
            // Honda reference — all make dummies stay 0
          else
            1.0 x:peek fitMakeCol @ t:![]
          then
        t:drop
      y:drop
    x:drop
  /foreach
;

: fitAll
  yPrice @ xWear @ OLSModel.fit^ mWear !
  yPrice @ xMake @ OLSModel.fit^ mMake !
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
Print make block of β with names (reference Honda omitted = $0 shift).
( model:OLSModel -- )
*/
: printMakeBetas
  >ts
  t:.beta@ >ys
  "  Make level vs Honda (same age/miles):" $ nl
  3 y:@() >zs ( z:peek ) "    Audi       %+.0f\n" fmt $ z:drop
  4 y:@() >zs ( z:peek ) "    BMW        %+.0f\n" fmt $ z:drop
  5 y:@() >zs ( z:peek ) "    Chevrolet  %+.0f\n" fmt $ z:drop
  6 y:@() >zs ( z:peek ) "    Dodge      %+.0f\n" fmt $ z:drop
  7 y:@() >zs ( z:peek ) "    Ford       %+.0f\n" fmt $ z:drop
  "    Honda      (reference)\n" $
  8 y:@() >zs ( z:peek ) "    Lexus      %+.0f\n" fmt $ z:drop
  9 y:@() >zs ( z:peek ) "    Lincoln    %+.0f\n" fmt $ z:drop
  10 y:@() >zs ( z:peek ) "    Nissan     %+.0f\n" fmt $ z:drop
  11 y:@() >zs ( z:peek ) "    RAM        %+.0f\n" fmt $ z:drop
  12 y:@() >zs ( z:peek ) "    Subaru     %+.0f\n" fmt $ z:drop
  13 y:@() >zs ( z:peek ) "    Toyota     %+.0f\n" fmt $ z:drop
  y:drop t:drop
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
  " auto-stats Stage-2b — age/miles + make (no vehicle class)" $ nl
  "============================================================" $ nl
  (
    nFit @
    csAsOfYear
    csMinMiles
  )
  " n=%d   as-of year=%d   miles >= %d (used inventory)\n" fmt $
  " y = asking price (USD)\n" $
  " Reference make: Honda. Dummies for other catalog makes.\n" $
  " No vehicle-class columns (class correlates with model mix).\n" $
  nl
;

: printModels
  "Fitted models" $ nl
  mWear @ "age+miles" printModel
  mMake @ "age+miles+make" printModel
  mMake @ printMakeBetas
  nl
;

: printCompares
  "C-vs-A comparisons" $ nl
  mWear @ mMake @ "age+miles → +make" printCompare
  nl
  "Read: make asks whether brand level shifts remain after age and miles.\n" $
;

: ladderMake
  connectCsDb
  loadFitRows
  nFit @ 0 <= if
    "No fit-ready make-mapped rows in " csDbPath str::+ error
  then
  buildDesign
  fitAll
  printHeader
  printModels
  printCompares
  "Done." $ nl
;
