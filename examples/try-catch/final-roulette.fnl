/*
Simple example of how to use try/catch, both catching and throwing.
*/


/*
5/6 times, return the expected result.
1/6 times, an abnormal condition arose--throw.
( -- res | throws )
*/
: doFr_T
  6 Rand.int^
  <>0
  if
    "The dinglehopper was successfully snarfblatted!"
  else
    "The thingamajig could not be dragooned!" throw
  then
;

/*
Outer caller, with a try/catch block
( -- )
*/
: fr
  try
    doFr_T // ( res:s -- )
    "YEH!: " swap str::+
  catch // ( err:s -- )
    "DANG: " swap str::+
  /try
  $ nl
;

`
Welcome to Final Roulette! You have 1/6 chance of busting up!
Enter fr repeatedly...
 
` $
