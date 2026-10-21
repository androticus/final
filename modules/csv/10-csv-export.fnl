/*
Final export of a Table to csv/tsv file.

Options DynamicObject (see getDefaultCsvOptions / getDefaultTsvOptions):
  .exportHeaders    Bool   — write columnNames as first line
  .quoteAllStrings  Bool   — CSV only: true=All, false=WhenNeeded
  .filename         String — required for *file words (panic if empty)

String cells with embedded newline or return throw (external data).
TSV also throws on embedded tab. Empty filename on file export panics.
Null and "" render as an empty cell. Numbers/Bool use ._toString, never quoted.
Line endings are \n only.
*/

/*
Get a DynamicObject with default options for csv export.
( -- defaultOptions:DynamicObject )
*/
: getDefaultCsvOptions
  newdo >ts
    true t:.exportHeaders!
    false t:.quoteAllStrings!
    "" t:.filename!
  ts>
;

/*
Get a DynamicObject with default options for tsv export.
( -- defaultOptions:DynamicObject )
*/
: getDefaultTsvOptions
  newdo >ts
    true t:.exportHeaders!
    "" t:.filename!
  ts>
;

// --- string field helpers (x:ctx  z:sb) ---

/*
Throw if s contains newline or return.
( s:String -- s )
*/
: xsv_rejectNlCr
  dup "\n" str::indexOf 0 < not if
    drop "xsv: string contains newline" throw
  then
  dup "\r" str::indexOf 0 < not if
    drop "xsv: string contains return" throw
  then
;

/*
True if CSV WhenNeeded should quote s (has comma or quote).
( s:String -- s b )
*/
: xsv_csvNeedsQuote
  dup "," str::contains if
    true
  else
    dup "\"" str::contains if
      true
    else
      false
    then
  then
;

/*
Write a CSV-quoted string (doubles embedded quotes).
( s:String -- ) x:ctx z:sb
*/
: xsv_writeCsvQuoted
  "\"" z:.writeString^
  "\"" "\"\"" str::sub
  z:.writeString^
  "\"" z:.writeString^
;

/*
Write a non-empty CSV string field (quoting per ctx.quoteAll).
( s:String -- ) x:ctx z:sb
*/
: xsv_writeCsvString
  x:.quoteAll@ if
    xsv_writeCsvQuoted
  else
    xsv_csvNeedsQuote if
      xsv_writeCsvQuoted
    else
      z:.writeString^
    then
  then
;

/*
Write a non-empty TSV string field; throw if tab present.
( s:String -- ) x:ctx z:sb
*/
: xsv_writeTsvString
  dup "\t" str::indexOf 0 < not if
    drop "tsv: string contains tab" throw
  then
  z:.writeString^
;

/*
Write a String field (empty → empty cell). CSV vs TSV from ctx.isCsv.
( s:String -- ) x:ctx z:sb
*/
: xsv_writeStringField
  dup str::isEmpty if
    drop
  else
    xsv_rejectNlCr
    x:.isCsv@ if
      xsv_writeCsvString
    else
      xsv_writeTsvString
    then
  then
;

/*
Write one cell value.
( cell:a -- ) x:ctx z:sb
*/
: xsv_writeCell
  dup Null? if
    drop
  else
    dup String? if
      xsv_writeStringField
    else
      ._toString^ z:.writeString^
    then
  then
;

/*
Write header line from ctx.cols.
( -- ) x:ctx z:sb
*/
: xsv_writeHeader
  0 >ts
  x:.cols@ foreach
    t:peek 0 > if
      x:.sep@ z:.writeString^
    then
    i xsv_writeStringField
    drop
    ts> 1 + >ts
  /foreach
  t:drop
  "\n" z:.writeString^
;

/*
Write one data row (cells by column ordinal via @()).
( row:a -- ) x:ctx z:sb
*/
: xsv_writeRow
  >ys
  0 >ts
  x:.cols@ foreach
    t:peek 0 > if
      x:.sep@ z:.writeString^
    then
    t:peek y:@()
    xsv_writeCell
    ts> 1 + >ts
  /foreach
  t:drop
  y:drop
  "\n" z:.writeString^
;

/*
Render table using ctx (isCsv, quoteAll, sep, cols, tbl, sb).
( -- ) x:ctx z:sb
*/
: xsv_render
  x:.opt@ .exportHeaders@ if
    xsv_writeHeader
  then
  x:.tbl@ foreach
    i xsv_writeRow
    drop
  /foreach
;

/*
Build ctx, render, leave string.
( options:DO tbl:Table isCsv:b sep:s quoteAll:b -- s )
*/
: table>xsv-string_
  newdo >xs
    x:.quoteAll!
    x:.sep!
    x:.isCsv!
    x:.tbl!
    x:.opt!
    x:.tbl.columnNames@ x:.cols!
    StringBuilder.new^ dup x:.sb! >zs
    xsv_render
    z:.toString^
  z:drop
  x:drop
;

/*
Export a Table to a csv-formatted string, with the options provided.
( options:DynamicObject tbl:Table -- csv:s ) throws on bad strings
*/
: table>csv-string
  swap >ys
  y:peek swap true "," y:.quoteAllStrings@
  table>xsv-string_
  y:drop
;

/*
Export a Table to a tsv-formatted string, with the options provided.
( options:DynamicObject tbl:Table -- tsv:s ) throws on bad strings
*/
: table>tsv-string
  swap >ys
  y:peek swap false "\t" false
  table>xsv-string_
  y:drop
;

/*
Export a Table to a csv file, with the options provided.
( options:DynamicObject tbl:Table -- ) panics if filename empty; throws on bad strings
*/
: table>csv-file
  swap >ys
  y:.filename@ str::isEmpty if
    "xsv: filename required for file export" error
  then
  y:.filename@ >ts
  y:peek swap
  table>csv-string
  t:peek swap File.writeText^
  t:drop
  y:drop
;

/*
Export a Table to a tsv file, with the options provided.
( options:DynamicObject tbl:Table -- ) panics if filename empty; throws on bad strings
*/
: table>tsv-file
  swap >ys
  y:.filename@ str::isEmpty if
    "xsv: filename required for file export" error
  then
  y:.filename@ >ts
  y:peek swap
  table>tsv-string
  t:peek swap File.writeText^
  t:drop
  y:drop
;
