/*
Final Db words, for a Harry Potter-based example in Postegres.

The code is factored into the words that connect to and create the 
table in the Db, then the words common to all examples.

This example also shows how to use a Moniker, to keep an Env value
secret.
*/

// Utilities

/*
Convert the query to a list, then to json.
( query:Iterator -- )
*/
: query>jsonp
  List.new^ >ts
  foreach
    i t:.append^
  /foreach
  ts> >jsonp
;

// substitued your unamd/pwd literally if you want, and fix up connectDb
`postgres://$POSTGRES_UNAME@localhost:5432/final_demo?sslmode=disable&password=$POSTGRES_PWD`
const: hpDsn

/*
Create a Moniker to opaquely hold the DSN after secret substitution.
DbProvider.new^ accepts a string or Moniker.
( -- mon:Moniker ) error if can't sub
*/
: createDsnMoniker
  hpDsn Moniker.new^ >ts
  t:peek "POSTGRES_UNAME" Env.subMoniker^ >xs
  t:peek "POSTGRES_PWD" Env.subMoniker^
  xs> && if
    ts>
  else
    t:drop
    "Could not substitute required Postgres parameters" error
  then
;  
 
`
DROP TABLE IF EXISTS "hpCharacters";
CREATE TABLE "hpCharacters" (
  "id"                 SERIAL PRIMARY KEY,  -- auto-assign / increment
  "firstName"          TEXT    NOT NULL,
  "lastName"           TEXT    NOT NULL,
  "isGood"             BOOLEAN NOT NULL,
  "firstMovieSeenYear" INT     NOT NULL
);
`
const: ddlPostgres 



// The hpDb will be saved in here, across words

0 var: hpDb

/*
Connect to the database using a Moniker
( ) save to hpDb, else raise error if can't create
*/
: connectDb
  DbProvider.postgres createDsnMoniker
//DbProvider.postgres hpDsn // if you put your creds in the hpDsn
  DbProvider.new^
  nif
    dup 
    "Can't open DbProvider: " $ $ nl
    error
  then
  hpDb !
;


// Drop/Create the table
// ( ddl:s -- )
: dropCreateTable
  ddlPostgres
  hpDb @ >ts
    t:.exec^
  t:drop
;

/*
Insert a movie character into the Db.
Shows how to use arguments.
( args:List -- )
*/
: insertCharList
`
INSERT INTO "hpCharacters" ("firstName", "lastName", "isGood", "firstMovieSeenYear")
VALUES ($1, $2, $3, $4)
`
  hpDb @ >ts
    swap t:.execWithArgs^
  t:drop
;

// ( -- )
: insertData
  ( "Harry" "Potter" true 2001 ) insertCharList
  ( "Hermione" "Granger" true 2001 ) insertCharList
  ( "Ron" "Weasley" true 2001 ) insertCharList
  ( "Albus" "Dumbledore" true 2001 ) insertCharList
  ( "Minerva" "McGonagall" true 2001 ) insertCharList
  ( "Rubeus" "Hagrid" true 2001 ) insertCharList
  ( "Draco" "Malfoy" false 2001 ) insertCharList
  ( "Severus" "Snape" false 2001 ) insertCharList
  ( "Lord" "Voldemort" false 2001 ) insertCharList
  ( "Ginny" "Weasley" true 2002 ) insertCharList
  ( "Sirius" "Black" true 2004 ) insertCharList
  ( "Remus" "Lupin" true 2004 ) insertCharList
  ( "Cedric" "Diggory" true 2005 ) insertCharList
  ( "Cho" "Chang" true 2005 ) insertCharList
  ( "Luna" "Lovegood" true 2007 ) insertCharList
  ( "Dolores" "Umbridge" false 2007 ) insertCharList
  ( "Nymphadora" "Tonks" true 2007 ) insertCharList
  ( "Bellatrix" "Lestrange" false 2007 ) insertCharList
  ( "Horace" "Slughorn" true 2009 ) insertCharList
  ( "Neville" "Longbottom" true 2001 ) insertCharList
;

// Connect, and add the data. We stay connected via hpDb.
: connectAddData
  connectDb
  dropCreateTable
  insertData
;

// Queries

`
SELECT "id", "firstName", "lastName", "isGood", "firstMovieSeenYear"
FROM "hpCharacters"
ORDER BY "firstMovieSeenYear", "lastName", "firstName"
`
const: sqlAllByYearSeenLastFirst

`
SELECT *
FROM "hpCharacters"
WHERE "firstMovieSeenYear" = $1
ORDER BY id
`
const: sqlYearSeenById

/*
Run a query, iterate the result and print.
The `i` is a DynamicObject with field names from the query.
The type of each field is the expected type for the Db column type.
*/
: queryAllByYearSeenLastFirst

  hpDb @ >ts
    sqlAllByYearSeenLastFirst
    t:.query^
  t:drop

  foreach
    // each i: DynamicObject clone of schema prototype
    i >ts
      ( t:.firstName@ t:.lastName@ t:.firstMovieSeenYear@ )
      "name: %s %s  year first seen: %d\n" fmt $
    t:drop
  /foreach
;

/*
Use this query, that returns the iterator, to call query>json 
or the framework iter>list to create a list of the values.
( year:i -- iter:Iterator)
*/
: queryYearSeenById
    ( peek ) swap drop
    sqlYearSeenById
    swap
  hpDb @ >ts  
    t:.queryWithArgs^
  t:drop
;

`
select count(*), sum("firstMovieSeenYear")
from "hpCharacters"
`
const: sqlAnonCols

/*
This demonstrates the names automatically given to the fields
by the DbProvider for unnamed columns.
( sql:s -- iter:Iterator )
*/
: runQuery
  hpDb @ >ts  
    t:.query^
  t:drop
;
