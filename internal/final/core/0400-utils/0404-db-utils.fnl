/*
Database-motivated utility words.
Some of these words are general purpose.
*/

/*
Return the value and true if not null, false if null.
( val:a -- val true | false )
*/
: val?
  dup Null? if
    drop false
  then
;

/*
Return the value if not null, otherwise return the default.
( val:a default:a -- val | default )
*/
: val|def 
  over Null? if
    nip
  else
    drop
  then
;

: _iter>target
  >ts
    foreach
      i t:.append^
    /foreach
  ts>
;

/*
Put all the values returned by the iterator into a List.
Works with any `foreach`-compatible Iterable or Iterator.
When used with a DbProvider.query^ it returns a List of
the StaticObject value returned for each row.
( iterable:Iterable|Iterator -- l:List )
*/
: iter>list
  List.new^
  _iter>target
;

/*
Create a new Table from a Database query result.
Expects a DbIter; returns a Table of StaticObject.
( iter:DbIter -- tbl:Table )
*/
: dbIter>table
  .protoObj Table.newFromDynamicObject^
  _iter>target
;
