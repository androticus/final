forget: reset-csv-export-tests

/*
Tests for modules/csv/10-csv-export.fnl

From repo root (final-go / final-tui REPL or Send):
  "modules/csv/10-csv-export.fnl" compileFromFile
  "modules/csv/10-csv-export-test.fnl" compileFromFile
  csvExportTests
*/


: csvAssertEq
  // ( expected:s actual:s label:s -- )
  >ts
  dup2 == if
    drop2
    "PASS " t:peek str::+ $ nl
  else
    "FAIL " t:peek str::+ $ nl
    "  expected: " $ swap $ nl
    "  actual:   " $ $ nl
    "csv export test failed" error
  then
  t:drop
;

: csvMakeProto
  newdo >ts
    "" t:.name!
    0 t:.age!
    "" t:.note!
  ts>
;

/*
Build sample table (does not touch caller's DS below).
( -- tbl:Table )
*/
: csvMakeSampleTable
  csvMakeProto >ts
  t:peek Table.newFromDynamicObject^ >ys
  t:peek .newStatic^ >xs
    "Alice" x:.name!
    20 x:.age!
    "hi" x:.note!
  xs> y:.append^
  t:peek .newStatic^ >xs
    "Bob" x:.name!
    null x:.age!
    "" x:.note!
  xs> y:.append^
  t:peek .newStatic^ >xs
    "Carol" x:.name!
    3 x:.age!
    "a,b" x:.note!
  xs> y:.append^
  t:drop
  ys>
;

: testCsvWhenNeeded
  getDefaultCsvOptions >ys
  csvMakeSampleTable >xs
  y:peek x:peek table>csv-string
  "name,age,note\nAlice,20,hi\nBob,,\nCarol,3,\"a,b\"\n"
  swap
  "csv WhenNeeded"
  csvAssertEq
  x:drop y:drop
;

: testCsvQuoteAll
  getDefaultCsvOptions >ys
  true y:.quoteAllStrings!
  csvMakeSampleTable >xs
  y:peek x:peek table>csv-string
  "\"name\",\"age\",\"note\"\n\"Alice\",20,\"hi\"\n\"Bob\",,\n\"Carol\",3,\"a,b\"\n"
  swap
  "csv QuoteAll"
  csvAssertEq
  x:drop y:drop
;

: testCsvNoHeaders
  getDefaultCsvOptions >ys
  false y:.exportHeaders!
  csvMakeSampleTable >xs
  y:peek x:peek table>csv-string
  "Alice,20,hi\nBob,,\nCarol,3,\"a,b\"\n"
  swap
  "csv no headers"
  csvAssertEq
  x:drop y:drop
;

: testTsvBasic
  getDefaultTsvOptions >ys
  csvMakeSampleTable >xs
  y:peek x:peek table>tsv-string
  "name\tage\tnote\nAlice\t20\thi\nBob\t\t\nCarol\t3\ta,b\n"
  swap
  "tsv basic"
  csvAssertEq
  x:drop y:drop
;

: testCsvQuotedQuote
  getDefaultCsvOptions >ys
  false y:.exportHeaders!
  csvMakeProto >ts
  t:peek Table.newFromDynamicObject^ >xs
  t:peek .newStatic^ >zs
    "x" z:.name!
    1 z:.age!
    "say \"hi\"" z:.note!
  zs> x:.append^
  t:drop
  y:peek x:peek table>csv-string
  "x,1,\"say \"\"hi\"\"\"\n"
  swap
  "csv embedded quote"
  csvAssertEq
  x:drop y:drop
;

: testCsvThrowNewline
  getDefaultCsvOptions >ys
  false y:.exportHeaders!
  csvMakeProto >ts
  t:peek Table.newFromDynamicObject^ >xs
  t:peek .newStatic^ >zs
    "x" z:.name!
    1 z:.age!
    "line1\nline2" z:.note!
  zs> x:.append^
  t:drop
  try
    y:peek x:peek table>csv-string
    drop
    "expected throw for newline" error
  catch
    drop
    "PASS csv throw newline" $ nl
  /try
  x:drop y:drop
;

: testTsvThrowTab
  getDefaultTsvOptions >ys
  false y:.exportHeaders!
  csvMakeProto >ts
  t:peek Table.newFromDynamicObject^ >xs
  t:peek .newStatic^ >zs
    "x" z:.name!
    1 z:.age!
    "has\ttab" z:.note!
  zs> x:.append^
  t:drop
  try
    y:peek x:peek table>tsv-string
    drop
    "expected throw for tab" error
  catch
    drop
    "PASS tsv throw tab" $ nl
  /try
  x:drop y:drop
;

: testCsvFileRoundtrip
  getDefaultCsvOptions >ys
  "tmp/csv-export-test.csv" y:.filename!
  csvMakeSampleTable >xs
  y:peek x:peek table>csv-string >zs
  y:peek x:peek table>csv-file
  "tmp/csv-export-test.csv" File.readText^
  z:peek
  swap
  "csv file roundtrip"
  csvAssertEq
  z:drop x:drop y:drop
;

: csvExportTests
  testCsvWhenNeeded
  testCsvQuoteAll
  testCsvNoHeaders
  testTsvBasic
  testCsvQuotedQuote
  testCsvThrowNewline
  testTsvThrowTab
  testCsvFileRoundtrip
  "csvExportTests done" $ nl
;
