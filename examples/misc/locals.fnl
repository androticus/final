forget: reset-locals

`
Simple Demo of word-local variables, on l.

Try: 
demo-locals
` const: locals-blurb

/*
Create a small DynamicObject and set it on l.person.
Then retrieve and print it.
( )
*/
: demo-locals
  { "Brad" "name"
    21 "age"
  }
  l.person!
  
  //l >ts t:.person@ t:drop
  l.person@
  >jsonp $ nl
;

locals-blurb $ nl

