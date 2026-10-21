/*
Native Throw exemplars (ThrowProbe class methods).
Same catch story as Final "msg" throw, but raised from Go via Throw().
*/

/*
Single native Throw → catch.
( -- )
*/
: demoNativeBoom
  try
    ThrowProbe.boom^
    "unreachable" $ nl
  catch
    "CAUGHT: " swap str::+ $ nl
  /try
;

/*
Nested Throw in one native (outer-first caused by: join).
( -- )
*/
: demoNativeBoomJoin
  try
    ThrowProbe.boomJoin^
    "unreachable" $ nl
  catch
    "CAUGHT: " swap str::+ $ nl
  /try
;

`
Native ThrowProbe demos loaded (not run).
Enter:  demoNativeBoom
   or:  demoNativeBoomJoin
` $
