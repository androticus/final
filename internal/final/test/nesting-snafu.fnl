forget: nesting

/*
Tests of Final nesting

*/

: nest-snafus
  0 10 range
  foreach
    // first, the ContextMark is where i is supposed to be
    // j is RS[1] but is actually i
    "( op: supposed to be to be j but will be i:\n" $
    -10 0 range
    foreach
     ( j )
     "%d " fmt $
    /foreach
    nl
    
    // next, the CatchMark is where i is supposed to be
    // j is RS[1] but is actually i
    "try/catch supposed to be to be j but will be i:\n" $
    -10 0 range
    foreach
      try
        j $ sp
      catch
      nop
      /try 
    /foreach
    nl
    
  /foreach
;

