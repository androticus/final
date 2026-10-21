/*
This file demonstrates the use of the interative or compiled "paren ops.
( value:a ... ) creates a List ( values:s ... s) creates a string list
{ value:a name:s ... } creates a DynamicObject
*/


`
Demonstrates the use of the interpreted or compiled "paren" ops.
( value:a ... ) creates a List ( values:s ... s) creates a StringList
{ value:a name:s ... } creates a DynamicObject
{ value:a name:s ... sm} creates a StringMap
( do:DynamicObject ) { value:a name:s ... so} creates a StaticObject
from the DynamicObject provided BEFORE the first { .
[ RxC values:Float ] fills the N values in a Mat2f or VecNf
NOTE: every value must be provided; Mat in row major order.
Run paren-ops -- it prints the json for every object.

` $ nl

// --- one word per unique opener/closer pair ---

/*
( … ) → List
( -- list:List )
*/
: parenList
  ( 1 2 "three" true )
;

/*
( … s) → StringList
( -- sl:StringList )
*/
: parenStringList
  ( "alpha" "beta" "gamma" s)
;

/*
{ … } → DynamicObject
( -- do:DynamicObject )
*/
: parenDynamicObject
  { 1 "one" 2 "two" "hi" "msg" }
;

/*
{ … sm} → StringMap
( -- sm:StringMap )
*/
: parenStringMap
  { "Ada" "name" 36 "age" sm}
;

/*
proto { … so} → StaticObject
Prototype on DS before {; fields must exist on the prototype.
( -- so:StaticObject )
*/
: parenStaticObject
  { 0 "x" 0 "y" }
  { 10 "x" 20 "y" so}
;

/*
mat/vec [ … ] → same Mat2f or VecNf, filled (row-major for Mat)
( -- mat:Mat2f )
*/
: parenMatFill
  2 3 Mat2f.new^
  [ 0 1 2  3 4 5 ]
;

/*
Run every paren-op exemplar; print pretty JSON for each.
( -- )
*/
: paren-ops
  "--- ( ) List ---" $ nl
  parenList >jsonp $ nl nl

  "--- ( s) StringList ---" $ nl
  parenStringList >jsonp $ nl nl

  "--- { } DynamicObject ---" $ nl
  parenDynamicObject >jsonp $ nl nl

  "--- { sm} StringMap ---" $ nl
  parenStringMap >jsonp $ nl nl

  "--- proto { so} StaticObject ---" $ nl
  parenStaticObject >jsonp $ nl nl

  "--- mat [ ] Mat2f fill ---" $ nl
  parenMatFill >jsonp $ nl nl

  "--- done ---" $ nl
;

`
Enter:  paren-ops
` $
