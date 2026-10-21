forget: reset-pls
`
Pip Local Storage (pls) is a Pip-local DynamicObject.
You can use it for any code that might be executed by multiple
Pips, each of which would like to keep track of some things.
This demo simply shows how to read and write to it.
Doing so uses the same syntax as local variables, i.e. l.
` const: pls-blurb

: pls-demo
  // write a field to pls
  "Foo" pls.name!
  
  // now fetch again and read the field
  // as with all DynamicObjects, you must
  // write to a field first before reading it
  "pls.name is: " pls.name@ str::cat $ nl
;


pls-blurb $ nl pls-demo
