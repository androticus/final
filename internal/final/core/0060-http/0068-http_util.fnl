/*
Http subsytem utilities and helpers
*/

/*
Send an HttpReq and return the HttpRes | error
( req:HttpReq -- res:HttpRes true | err:s false )
*/
: httpDial
  HttpDialer.new^
  >ts
  t:.send^
  t:drop
;

/*
Print the result of an httpDial.
If successful, the result body is printed as a string.
If an error occurred, print the error.
Note that Http errors are reported in the status field
of the result; actual error results are timeouts, etc.
( res:HttpRes true | err:s false -- )
*/
: printHttpRes
  if
    >ts
    ( t:.status t:.getBodyAsString^ )
    "(status: %d): %s" fmt
    t:drop
  else // error
    ( peek )
    "ERROR: %s" fmt
    nip
  then
  $ nl
;
