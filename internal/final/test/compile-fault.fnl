forget: reset-compile-fault

`
This file has compiler faults, and is used to test error reporting.
The individual faults can be enable/disabled via comment.
The enableable faults have flush left comments.
Note: compiler will stop on first fault.

Try:

* enable one of the faults, and reload the file
` const: compile-fault-blurb
/*
Block comment to insure line counts.
myWord has a fault
( -- )
*/
: myWord
  nop
//foo // nonexistent word
;

: badLit
//12bad
  nop
;

{
  1 "one"
//2 two // bad field in interpreted paren sandwich
} const: fields

compile-fault-blurb $ nl

/*
unclosed comment should fault at eof
*/
// delete the */ to trigger the unclosed comment fault.
// don't forget to add it back when finished with the demo


