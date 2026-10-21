/*
auto-stats ingest — fetch unhuman.autos listings into SQLite.

Run from repo root:
  "internal/final/core/0060-http/0068-http_util.fnl" compileFromFile
  "internal/final/apps/auto-stats/ingest.fnl" compileFromFile
  ingest

N=20 fit-ready rows per model×year; Atlanta 30301 / radius 100; years 2019–2025.
Used inventory only (inventory_type=used); reject new/CPO and miles < csMinMiles.
Upsert → unhuman_listing; progress → ingest_goal.
HTTP/JSON/status failures throw; outer catch ends the session and prints
remaining page-fetch estimate.
*/

// --- config ---

"internal/final/apps/auto-stats/db/auto-stats.db" const: csDbPath
"internal/final/apps/auto-stats/db/unhuman_listing.sql" const: csListingSqlPath
"internal/final/apps/auto-stats/db/ingest_goal.sql" const: csGoalSqlPath

"30301" const: csZip
100 const: csRadius
20 const: csTargetN
25 const: csPageRows
2019 const: csYearLo
2025 const: csYearHi
500 const: csStartCap
500 const: csMinMiles // drop lot-new / demo miles that skew linear age

"unhuman.autos" const: csHost
"api/cars" const: csPath

0 var: csDb
"" var: seedMake

// --- row transform ---

/*
Raw unhuman listing DO → flat DB-shaped DO.
Missing fields become Null (fit? filters them out).
( unhuman:DO -- row:DO )
*/
: unhuman-to-db-rec
  >xs newdo >ys
  null y:.id!
  null y:.price!
  null y:.year!
  null y:.make!
  null y:.model!
  null y:.miles!
  null y:.vehicle_type!
  null y:.body_type!
  null y:.doors!
  null y:.cylinders!
  null y:.drivetrain!
  null y:.inventory_type!
  "id" x:?{} if "id" x:@{} y:.id! then
  "price" x:?{} if "price" x:@{} y:.price! then
  "miles" x:?{} if "miles" x:@{} y:.miles! then
  "inventory_type" x:?{} if "inventory_type" x:@{} y:.inventory_type! then
  "build" x:?{} if
    "build" x:@{} >zs
    "year" z:?{} if "year" z:@{} y:.year! then
    "make" z:?{} if "make" z:@{} y:.make! then
    "model" z:?{} if "model" z:@{} y:.model! then
    "vehicle_type" z:?{} if "vehicle_type" z:@{} y:.vehicle_type! then
    "body_type" z:?{} if "body_type" z:@{} y:.body_type! then
    "doors" z:?{} if "doors" z:@{} y:.doors! then
    "cylinders" z:?{} if "cylinders" z:@{} y:.cylinders! then
    "drivetrain" z:?{} if "drivetrain" z:@{} y:.drivetrain! then
    z:drop
  then
  x:drop ys>
;

/*
True if row is fit-ready used inventory: non-null id/price/miles/year,
inventory_type is used (or absent), miles >= csMinMiles.
( row:DO -- b )
*/
: db-rec-fit?
  >ts
  t:.id@ Null? if
    false
  else
    t:.price@ Null? if
      false
    else
      t:.miles@ Null? if
        false
      else
        t:.year@ Null? if
          false
        else
          t:.miles@ csMinMiles < if
            false
          else
            t:.inventory_type@ Null? if
              true
            else
              t:.inventory_type@ "used" ==
            then
          then
        then
      then
    then
  then
  t:drop
;

// --- DB open / schema ---

: connectCsDb
  DbProvider.sqlite csDbPath DbProvider.new^
  nif
    "Can't open auto-stats DB: " swap str::+ error
  then
  csDb !
;

: execSqlFile
  File.readText^
  csDb @ >ts
    t:.exec^
  t:drop
;

: ensureCsSchema
  csListingSqlPath execSqlFile
  csGoalSqlPath execSqlFile
;

// --- goal seed ---

`
INSERT OR IGNORE INTO ingest_goal
  (make, model, year, db_make, db_model, zip, radius, target_n, have_n, next_start)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
`
const: sqlInsertGoal

/*
( make:s model:s year:i dbMake:s dbModel:s -- )
*/
: seedGoalRow
  >zs >ys >xs >ts
  seedMake !
  (
    seedMake @
    t:peek
    x:peek
    y:peek
    z:peek
    csZip
    csRadius
    csTargetN
  )
  sqlInsertGoal swap
  csDb @ >ys
    y:.execWithArgs^
  y:drop
  t:drop x:drop y:drop z:drop
;

/*
( apiMake:s apiModel:s dbMake:s dbModel:s -- )
*/
: seedModelYears
  >zs >ys >xs >ts
  csYearLo
  begin
    dup csYearHi <= while
      dup
      t:peek x:peek rot y:peek z:peek
      seedGoalRow
      1 +
  repeat
  drop
  t:drop x:drop y:drop z:drop
;

: seedGoalsIfEmpty
  "SELECT COUNT(*) AS n FROM ingest_goal"
  csDb @ >ts t:.query^ t:drop
  iter>list 0 @() >ts
  t:.n@ 0 >
  t:drop
  nif
    "Seeding ingest_goal catalog…" $ nl
    "toyota" "camry" "Toyota" "Camry" seedModelYears
    "honda" "civic" "Honda" "Civic" seedModelYears
    "honda" "accord" "Honda" "Accord" seedModelYears
    "nissan" "altima" "Nissan" "Altima" seedModelYears
    "chevrolet" "malibu" "Chevrolet" "Malibu" seedModelYears
    "dodge" "charger" "Dodge" "Charger" seedModelYears
    "lexus" "es" "Lexus" "ES" seedModelYears
    "bmw" "3 series" "BMW" "3 Series" seedModelYears
    "audi" "a4" "Audi" "A4" seedModelYears
    "toyota" "rav4" "Toyota" "RAV4" seedModelYears
    "honda" "cr-v" "Honda" "CR-V" seedModelYears
    "nissan" "rogue" "Nissan" "Rogue" seedModelYears
    "chevrolet" "equinox" "Chevrolet" "Equinox" seedModelYears
    "ford" "explorer" "Ford" "Explorer" seedModelYears
    "subaru" "outback" "Subaru" "Outback" seedModelYears
    "lexus" "rx" "Lexus" "RX" seedModelYears
    "bmw" "x5" "BMW" "X5" seedModelYears
    "audi" "q5" "Audi" "Q5" seedModelYears
    "lincoln" "nautilus" "Lincoln" "Nautilus" seedModelYears
    "lincoln" "navigator" "Lincoln" "Navigator" seedModelYears
    "ford" "f-150" "Ford" "F-150" seedModelYears
    "chevrolet" "silverado 1500" "Chevrolet" "Silverado 1500" seedModelYears
    "ram" "ram 1500" "RAM" "Ram 1500" seedModelYears
    "ingest_goal seeded." $ nl
  then
;

// --- SQL ---

`
SELECT COUNT(*) AS n FROM unhuman_listing
WHERE make = ? AND model = ? AND year = ?
  AND price IS NOT NULL AND miles IS NOT NULL AND year IS NOT NULL
  AND miles >= 500
  AND (inventory_type IS NULL OR inventory_type = 'used')
`
const: sqlCountFit

`
UPDATE ingest_goal SET have_n = ?, next_start = ? WHERE make = ? AND model = ? AND year = ?
`
const: sqlUpdateGoal

`
SELECT COALESCE(SUM((target_n - have_n + 24) / 25), 0) AS pages
FROM ingest_goal
WHERE have_n < target_n AND next_start < 500
`
const: sqlPagesRemaining

`
SELECT make, model FROM ingest_goal
WHERE have_n < target_n AND next_start < 500
GROUP BY make, model
ORDER BY make, model
`
const: sqlPendingModels

`
SELECT make, model, year, db_make, db_model, zip, radius, target_n, have_n, next_start
FROM ingest_goal
WHERE make = ? AND model = ? AND have_n < target_n AND next_start < 500
ORDER BY year
`
const: sqlPendingYears

: pagesRemaining
  sqlPagesRemaining csDb @ >ts t:.query^ t:drop
  iter>list 0 @() >ts
  t:.pages@
  t:drop
;

: printPagesRemaining
  pagesRemaining
  ( peek ) "Page fetches remaining (est.): %d\n" fmt $
  drop
;

/*
Recount fit rows and persist have_n + next_start for this goal DO.
( goal:DO -- )
*/
: writeGoalProgress
  >ts
  (
    t:.db_make@
    t:.db_model@
    t:.year@
  )
  sqlCountFit swap
  csDb @ >xs x:.queryWithArgs^ x:drop
  iter>list 0 @() >xs
  x:.n@ t:.have_n!
  (
    t:.have_n@
    t:.next_start@
    t:.make@
    t:.model@
    t:.year@
  )
  sqlUpdateGoal swap
  csDb @ >ys y:.execWithArgs^ y:drop
  x:drop t:drop
;

`
INSERT INTO unhuman_listing
  (id, price, miles, year, make, model, body_type, vehicle_type, drivetrain, doors, cylinders, inventory_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  price=excluded.price,
  miles=excluded.miles,
  year=excluded.year,
  make=excluded.make,
  model=excluded.model,
  body_type=excluded.body_type,
  vehicle_type=excluded.vehicle_type,
  drivetrain=excluded.drivetrain,
  doors=excluded.doors,
  cylinders=excluded.cylinders,
  inventory_type=excluded.inventory_type
`
const: sqlUpsertListing

: upsertDbRec
  >ts
  (
    t:.id@
    t:.price@
    t:.miles@
    t:.year@
    t:.make@
    t:.model@
    t:.body_type@
    t:.vehicle_type@
    t:.drivetrain@
    t:.doors@
    t:.cylinders@
    t:.inventory_type@
  )
  sqlUpsertListing swap
  csDb @ >xs x:.execWithArgs^ x:drop
  t:drop
;

/*
( goal:DO rows:i start:i -- req:HttpReq )
*/
: makeCarsReq
  >ys >xs >ts
  HttpReq.new^ >zs
  csHost z:.hostAddress.host!
  csPath z:.path!
  30 z:.timeoutS!
  t:.make@ "make" z:.queryParams!{}
  t:.model@ "model" z:.queryParams!{}
  ( t:.year@ ) "%d" fmt "year" z:.queryParams!{}
  t:.zip@ "zip" z:.queryParams!{}
  ( t:.radius@ ) "%d" fmt "radius" z:.queryParams!{}
  "used" "inventory_type" z:.queryParams!{}
  ( x:peek ) "%d" fmt "rows" z:.queryParams!{}
  ( y:peek ) "%d" fmt "start" z:.queryParams!{}
  t:drop x:drop y:drop
  zs>
;

/*
( req:HttpReq -- res:HttpRes )
Non-200: throw, except exhausted pagination (start > num_found) → leave
a sentinel false? — handled in fetchGoalPage via softDial.
*/

/*
Modal dial: res true | errMsg false. Non-200 includes status in errMsg.
( req:HttpReq -- res:HttpRes true | err:s false )
*/
: dialCarsModal
  httpDial
  nif
    "HTTP dial failed: " swap str::+ false
  else
    >ts
    t:.status@ 200 == if
      ts> true
    else
      ( t:.status@ ) "HTTP status %d" fmt >xs
      t:.getBodyAsString^
      x:peek " " str::+ swap str::+
      x:drop
      t:drop
      false
    then
  then
;

/*
( need:i -- rows:i )
*/
: clampPageRows
  dup 0 <= if
    drop 0
  else
    dup csPageRows > if
      drop csPageRows
    then
  then
;

/*
True if err string looks like pagination past end (not throttle).
( err:s -- b )
*/
: isPastEndErr?
  dup "num_found" str::indexOf 0 < if
    drop false
  else
    drop true
  then
;

/*
Fetch one page; upsert fit listings; update goal progress.
Past-end pagination exhausts the cell; other HTTP errors throw.
( goal:DO -- )
*/
: fetchGoalPage
  >ts
  t:.target_n@ t:.have_n@ - clampPageRows
  dup 0 == if
    drop
    t:drop
  else
    >xs
    t:peek x:peek t:.next_start@ makeCarsReq
    dialCarsModal
    nif
      // err on DS
      dup isPastEndErr? if
        drop
        csStartCap t:.next_start!
        t:peek writeGoalProgress
        ( t:.make@ t:.model@ t:.year@ t:.have_n@ )
        "    year %s %s %d exhausted at have %d\n" fmt $
        x:drop t:drop
      else
        throw
      then
    else
      >ys
      y:.getBodyAsString^ json> >zs
      y:drop
      z:.listings@ >ys
      y:.len@ >zs
      z:peek 0 == if
        csStartCap t:.next_start!
      else
        y:peek
        foreach
          i unhuman-to-db-rec
          dup db-rec-fit? if
            upsertDbRec
          else
            drop
          then
        /foreach
        t:.next_start@ z:peek + t:.next_start!
        // fewer than requested → no more pages
        z:peek x:peek < if
          csStartCap t:.next_start!
        then
      then
      t:peek writeGoalProgress
      ( t:.make@ t:.model@ t:.year@ t:.have_n@ t:.target_n@ )
      "    year %s %s %d → have %d / %d\n" fmt $
      z:drop y:drop x:drop t:drop
    then
  then
;

/*
Fill one goal year cell (may take several pages).
( goal:DO -- )
*/
: ingestGoalCell
  >ts
  begin
    t:.have_n@ t:.target_n@ <
    t:.next_start@ csStartCap <
    && while
      t:peek fetchGoalPage
  repeat
  t:drop
;

/*
Ingest all pending years for one make/model.
( make:s model:s -- )
*/
: ingestOneModel
  >xs >ts
  ( t:peek x:peek ) "START model %s %s\n" fmt $
  (
    t:peek
    x:peek
  )
  sqlPendingYears swap
  csDb @ >ys y:.queryWithArgs^ y:drop
  iter>list >ys
  y:peek
  foreach
    i ingestGoalCell
  /foreach
  ( t:peek x:peek ) "DONE  model %s %s\n" fmt $
  y:drop t:drop x:drop
;

: ingestAllPendingModels
  sqlPendingModels csDb @ >ts t:.query^ t:drop
  iter>list >ts
  t:peek
  foreach
    i >xs
    x:.make@ x:.model@ ingestOneModel
    x:drop
  /foreach
  t:drop
;

/*
Main entry: open DB, seed goals, ingest until done or throw.
( -- )
*/
: ingest
  connectCsDb
  ensureCsSchema
  seedGoalsIfEmpty
  "=== auto-stats ingest begin ===" $ nl
  printPagesRemaining
  try
    ingestAllPendingModels
    "=== auto-stats ingest complete ===" $ nl
  catch
    ( peek ) "=== auto-stats ingest stopped: %s ===\n" fmt $
    drop
  /try
  printPagesRemaining
;
