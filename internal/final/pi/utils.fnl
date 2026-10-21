/*
Final Pi support utilities.
*/

/*
Prints a map of the gpio pins, and how they are assigned.
Two column map in logical, not header order.
( -- )
*/
: pinMap
  // ( entry:s -- entry':s )
  :s entryOrDummy
    dup nif drop "-" then 
  ;s
  
  // ( -- )
  :s writeEntry
    ( i t:@() entryOrDummy i  i l.offs + dup  t:@() entryOrDummy )
    "%10s  gpio %2d  %2d gpio  %s\n" fmt $  
  ;s
  
  GpioCtrl.osPinMap // ( -- map:StringList[0-27] )
  nl
  >ts
  t:.len 2 / dup l.offs!
  0 swap 1 - range
  foreach
    writeEntry
  /foreach
  nl
  tdrop
;
